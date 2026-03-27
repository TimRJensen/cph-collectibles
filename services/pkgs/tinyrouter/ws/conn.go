package ws

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"
)

type OpCode = byte

const (
	Text  OpCode = 0x1
	Close OpCode = 0x8
	Ping  OpCode = 0x9
	Pong  OpCode = 0xA
)

const (
	guid          = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	acceptMessage = "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: %s\r\n" +
		"\r\n"
	wTimeout = 5 * time.Second
)

var (
	errors = [8]error{
		fmt.Errorf("missing 'Upgrade' header"),
		fmt.Errorf("missing 'Connection' header"),
		fmt.Errorf("missing 'Sec-WebSocket-Key' header"),
		fmt.Errorf("unsuported RequestWriter"),
		fmt.Errorf("fragmented frames not supported"),
		fmt.Errorf("frames must be masked"),
		fmt.Errorf("write timeout"),
		fmt.Errorf("stream full"),
	}
)

/*
Internal helper function. Reads a single WebSocket frame.
The function is RFC complaint
*/
func readFrame(r *bufio.Reader) (byte, []byte, error) {
	// First byte: FIN + RSV + opcode
	b1, err := r.ReadByte()
	if err != nil {
		return 0, nil, err
	}

	// Fragmentation isn't supported currently
	if b1&0x80 == 0 {
		return 0, nil, errors[4]
	}

	// Second byte: Mask key + length7
	b2, err := r.ReadByte()
	if err != nil {
		return 0, nil, err
	}

	// Frames must be masked
	if b2&0x80 == 0 {
		return 0, nil, errors[5]
	}

	// Payloadlength7, 16 or 64
	length := int64(0)
	switch l := int64(b2 & 0x7F); l {
	case 126:
		buff := [2]byte{}
		if _, err := io.ReadFull(r, buff[:]); err != nil {
			return 0, nil, err
		}
		length = int64(binary.BigEndian.Uint16(buff[:]))
	case 127:
		buff := [8]byte{}
		if _, err := io.ReadFull(r, buff[:]); err != nil {
			return 0, nil, err

		}
		length = int64(binary.BigEndian.Uint64(buff[:]))
	default:
		length = l
	}

	// Mask key
	mask := make([]byte, 4)
	if _, err := io.ReadFull(r, mask); err != nil {
		return 0, nil, err
	}

	// Payload
	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return 0, nil, err
	}

	// Unmask
	for i := int64(0); i < length; i++ {
		payload[i] ^= mask[i%4]
	}

	// OpCode
	opcode := b1 & 0x0F

	return opcode, payload, nil
}

/*
Internal helper function. Writes a single WebSocket frame.
The function is RFC complaint
*/
func writeFrame(w *bufio.Writer, opcode byte, payload []byte) error {
	// First byte: FIN + OpCode
	b1 := 0x80 | (opcode & 0x0F)
	if err := w.WriteByte(b1); err != nil {
		return err
	}

	// Second + third byte: No mask + payloadlength7, 16, 64
	l := len(payload)
	switch {
	case l < 126:
		if err := w.WriteByte(byte(l)); err != nil {
			return err
		}
	case l <= 65535:
		buff := [3]byte{}
		buff[0] = 126
		binary.BigEndian.PutUint16(buff[1:], uint16(l))
		if _, err := w.Write(buff[:]); err != nil {
			return err
		}
	default:
		buff := [9]byte{}
		buff[0] = 127
		binary.BigEndian.PutUint64(buff[1:], uint64(l))
		if _, err := w.Write(buff[:]); err != nil {
			return err
		}
	}

	// Last bytes: Payload
	if _, err := w.Write(payload); err != nil {
		return err
	}
	return w.Flush()
}

type frame struct {
	code OpCode
	buff []byte
}

func newFrame(code OpCode, payload []byte) frame {
	buff := make([]byte, len(payload))
	copy(buff, payload)
	return frame{code, buff}
}

type Conn struct {
	conn net.Conn
	br   *bufio.Reader
	bw   *bufio.Writer
	out  chan frame
	done chan struct{}
	once sync.Once
}

/*
Reads a single WebSocket frame
*/
func (c *Conn) Read() (OpCode, string, error) {
	opcode, payload, err := readFrame(c.br)
	return opcode, string(payload), err
}

/*
Internal helper function to maintain a write loop
*/
func (c *Conn) writer() {
	for {
		select {
		case <-c.done:
			return
		case fr, ok := <-c.out:
			if !ok {
				return
			}

			err := c.conn.SetWriteDeadline(time.Now().Add(wTimeout))
			if err != nil {
				c.Close()
				return
			}
			err = writeFrame(c.bw, fr.code, fr.buff)
			if err != nil {
				c.Close()
				return
			}
		}
	}
}

/*
Writes bytes to a single WebSocket frame
*/
func (c *Conn) Write(code OpCode, payload []byte) error {
	t := time.NewTimer(wTimeout)
	defer t.Stop()

	select {
	case c.out <- newFrame(code, payload):
		return nil
	case <-c.done:
		return net.ErrClosed
	case <-t.C:
		return errors[6]
	}
}

func (c *Conn) TryWrite(code OpCode, payload []byte) error {
	select {
	case c.out <- newFrame(code, payload):
		return nil
	case <-c.done:
		return net.ErrClosed
	default:
		return errors[7]
	}
}

/*
Writes text to a single WebSocket frame
*/
func (c *Conn) WriteText(msg string) error {
	return c.Write(Text, []byte(msg))
}

func (c *Conn) TryWriteText(msg string) error {
	return c.TryWrite(Text, []byte(msg))
}

/*
Closes the hijacked net.Connection
*/
func (c *Conn) Close() error {
	var err error
	c.once.Do(func() {
		close(c.done)
		close(c.out)
		err = c.conn.Close()

	})
	return err
}

/*
Internal helper function to enforce correct tokens
*/
func validToken(s, token string) (string, bool) {
	switch {
	case s == "":
		return s, false
	case token == "*":
		return s, true
	default:
		s, token = strings.ToLower(s), strings.ToLower(token)
		if s == token {
			return s, true
		}
		return s, slices.ContainsFunc(strings.Split(s, ","), func(s string) bool {
			return strings.TrimSpace(s) == token
		})
	}
}

/*
Accepts and upgrades a valid WebSocket request
*/
func Accept(w http.ResponseWriter, r *http.Request) (*Conn, error) {
	if _, ok := validToken(r.Header.Get("Upgrade"), "websocket"); !ok {
		return nil, errors[0]
	}

	if _, ok := validToken(r.Header.Get("Connection"), "upgrade"); !ok {
		return nil, errors[1]
	}

	secKey, ok := validToken(r.Header.Get("Sec-WebSocket-Key"), "*")
	if !ok {
		return nil, errors[2]
	}

	h, ok := w.(http.Hijacker)
	if !ok {
		return nil, errors[3]
	}

	conn, rw, err := h.Hijack()
	if err != nil {
		return nil, err
	}
	defer func() {
		if conn == nil {
			return
		}
		conn.Close()
	}()

	br := rw.Reader
	bw := rw.Writer
	sha := sha1.New()
	sha.Write([]byte(secKey + guid))
	_, err = fmt.Fprintf(bw, acceptMessage, base64.StdEncoding.EncodeToString(sha.Sum(nil)))
	if err != nil {
		return nil, err
	}

	err = bw.Flush()
	if err != nil {
		return nil, err
	}

	// Happy days, connection succesfully upgraded
	c := &Conn{
		conn,
		br,
		bw,
		make(chan frame, 32),
		make(chan struct{}),
		sync.Once{},
	}
	go c.writer()

	// Null conn to avoid the defered close
	conn = nil

	return c, nil
}
