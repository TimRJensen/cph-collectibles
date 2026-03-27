package ws

import (
	"context"
)

type message struct {
	client *Conn
	opcode OpCode
	buff   []byte
}

type Hub struct {
	clients   map[*Conn]bool
	join      chan *Conn
	leave     chan *Conn
	broadcast chan *message
}

func (h *Hub) Run(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				for c := range h.clients {
					c.Close()
				}
				return
			case c := <-h.join:
				h.clients[c] = true
			case c := <-h.leave:
				if _, ok := h.clients[c]; !ok {
					continue
				}
				delete(h.clients, c)
				c.Close()
			case msg := <-h.broadcast:
				for c := range h.clients {
					if c == msg.client {
						continue
					}
					if err := c.TryWrite(msg.opcode, msg.buff); err != nil {
						delete(h.clients, c)
						c.Close()
					}
				}
			}
		}
	}()
}

func (h *Hub) Join(c *Conn) {
	h.join <- c
}

func (h *Hub) Leave(c *Conn) {
	h.leave <- c
}

func (h *Hub) Broadcast(c *Conn, opcode OpCode, msg string) {
	h.broadcast <- &message{c, opcode, []byte(msg)}
}

func NewHub() *Hub {
	return &Hub{
		make(map[*Conn]bool),
		make(chan *Conn),
		make(chan *Conn),
		make(chan *message, 32),
	}
}
