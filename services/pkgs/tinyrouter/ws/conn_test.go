package ws

import (
	"bufio"
	"bytes"
	"net"
	"net/http"
	"testing"
)

// --- Mock hijacker because httptest.ResponseRecorder doesn't support it ---

type mockHijacker struct {
	http.ResponseWriter
	conn *mockConn
	rw   *bufio.ReadWriter
}

type mockConn struct {
	net.Conn
	buf *bytes.Buffer
}

func newMockResponse() (*mockHijacker, *mockConn) {
	buf := &bytes.Buffer{}
	mc := &mockConn{buf: buf}
	rw := bufio.NewReadWriter(bufio.NewReader(buf), bufio.NewWriter(buf))
	return &mockHijacker{conn: mc, rw: rw}, mc
}

func (m *mockHijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return m.conn, m.rw, nil
}

func (mc *mockConn) Write(p []byte) (int, error) {
	return mc.buf.Write(p)
}

func (mc *mockConn) Read(p []byte) (int, error) {
	return mc.buf.Read(p)
}

func (mc *mockConn) Close() error { return nil }

// --- Helper to build a masked client frame ---

func buildClientFrame(opcode byte, payload []byte) []byte {
	maskKey := [4]byte{1, 2, 3, 4}

	var header []byte
	b1 := byte(0x80) | (opcode & 0x0F)
	header = append(header, b1)

	n := len(payload)
	if n < 126 {
		header = append(header, byte(0x80|n)) // MASK bit + len
	} else if n <= 0xFFFF {
		header = append(header, 0x80|126)
		header = append(header, byte(n>>8), byte(n))
	} else {
		header = append(header, 0x80|127)
		l := uint64(n)
		for i := 7; i >= 0; i-- {
			header = append(header, byte(l>>(8*uint(i))))
		}
	}

	header = append(header, maskKey[:]...)

	masked := make([]byte, n)
	for i, b := range payload {
		masked[i] = b ^ maskKey[i%4]
	}

	return append(header, masked...)
}

// --- Tests ---

func TestAcceptSuccess(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==") // RFC example

	w, conn := newMockResponse()

	ws, err := Accept(w, req)
	if err != nil {
		t.Fatalf("Accept failed: %v", err)
	}
	if ws == nil {
		t.Fatalf("Accept returned nil Conn")
	}

	// Verify handshake output
	resp := conn.buf.String()
	if !bytes.Contains([]byte(resp), []byte("101 Switching Protocols")) {
		t.Errorf("response missing status line: %s", resp)
	}
	if !bytes.Contains([]byte(resp), []byte("Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=")) {
		t.Errorf("wrong accept key: %s", resp)
	}
}

func TestMissingUpgrade(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-WebSocket-Key", "x")

	w, _ := newMockResponse()

	_, err := Accept(w, req)
	if err == nil {
		t.Fatalf("expected error for missing Upgrade header")
	}
}

func TestMissingConnection(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Key", "x")

	w, _ := newMockResponse()

	_, err := Accept(w, req)
	if err == nil {
		t.Fatalf("expected error for missing Connection header")
	}
}

func TestMissingKey(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")

	w, _ := newMockResponse()

	_, err := Accept(w, req)
	if err == nil {
		t.Fatalf("expected error for missing Sec-WebSocket-Key")
	}
}

func TestConnectionTokenParsing(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "keep-alive, Upgrade")
	req.Header.Set("Sec-WebSocket-Key", "x")

	w, _ := newMockResponse()

	_, err := Accept(w, req)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestReadFrameSimpleText(t *testing.T) {
	payload := []byte("hello")
	frame := buildClientFrame(0x1, payload) // text

	r := bufio.NewReader(bytes.NewReader(frame))
	opcode, data, err := readFrame(r)
	if err != nil {
		t.Fatalf("readFrame error: %v", err)
	}
	if opcode != 0x1 {
		t.Fatalf("expected opcode 0x1, got 0x%x", opcode)
	}
	if string(data) != "hello" {
		t.Fatalf("expected payload 'hello', got %q", string(data))
	}
}

func TestReadFrameExtended126(t *testing.T) {
	payload := bytes.Repeat([]byte("a"), 130) // >125, <65536
	frame := buildClientFrame(0x1, payload)

	r := bufio.NewReader(bytes.NewReader(frame))
	opcode, data, err := readFrame(r)
	if err != nil {
		t.Fatalf("readFrame error: %v", err)
	}
	if opcode != 0x1 {
		t.Fatalf("expected opcode 0x1, got 0x%x", opcode)
	}
	if len(data) != len(payload) {
		t.Fatalf("expected len %d, got %d", len(payload), len(data))
	}
}

func TestReadFrameUnmaskedError(t *testing.T) {
	// Same as buildClientFrame but without MASK bit set
	payload := []byte("hi")
	var f []byte
	f = append(f, 0x81)               // FIN + text
	f = append(f, byte(len(payload))) // MASK bit 0, len
	f = append(f, payload...)

	r := bufio.NewReader(bytes.NewReader(f))
	_, _, err := readFrame(r)
	if err == nil {
		t.Fatalf("expected error for unmasked frame, got nil")
	}
}

func TestWriteText(t *testing.T) {
	buf := &bytes.Buffer{}
	w := bufio.NewWriter(buf)

	err := writeFrame(w, 0x1, []byte("hello"))
	if err != nil {
		t.Fatalf("writeFrame error: %v", err)
	}

	data := buf.Bytes()

	// Byte 0: FIN + opcode
	if data[0] != 0x81 {
		t.Fatalf("expected 0x81, got 0x%x", data[0])
	}

	// Byte 1: MASK=0 + length=5
	if data[1] != 5 {
		t.Fatalf("expected length 5, got %d", data[1])
	}

	if string(data[2:]) != "hello" {
		t.Fatalf("expected 'hello', got %q", string(data[2:]))
	}
}

func TestWriteFrame126(t *testing.T) {
	payload := bytes.Repeat([]byte("a"), 200)

	buf := &bytes.Buffer{}
	w := bufio.NewWriter(buf)

	err := writeFrame(w, 0x1, payload)
	if err != nil {
		t.Fatalf("writeFrame error: %v", err)
	}

	data := buf.Bytes()

	if data[1] != 126 {
		t.Fatalf("expected 126 length marker, got %d", data[1])
	}

	length := int(data[2])<<8 | int(data[3])
	if length != len(payload) {
		t.Fatalf("wrong length: expected %d, got %d", len(payload), length)
	}
}
