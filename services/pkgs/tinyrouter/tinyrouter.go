package tinyrouter

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/TimRJensen/tinyrouter/art"
	"github.com/TimRJensen/tinyrouter/route"
	"github.com/TimRJensen/tinyrouter/ws"
)

func clean(raw string) string {
	raw = path.Clean(raw)
	if raw == "." {
		return "/"
	}
	return raw
}

func bestRoute[T route.HandlerFn](t art.Tree[*route.Route[T]], route string) (*route.Route[T], bool) {
	r, ok := t.MatchFn(route, 0x2A, func(key []byte, read, remaining int) int {
		for i := range remaining {
			if key[read+i] == 0x2F {
				return i
			}
		}
		return remaining
	})

	switch {
	case !ok:
		return nil, false
	default:
		return r, true
	}
}

type TinyRouter[T route.HandlerFn] struct {
	tree       art.Tree[*route.Route[T]]
	err        route.HandlerFn
	hub        *ws.Hub
	wsHandlers map[ws.OpCode]func(*ws.Hub, string)
	wsJoin     func(*ws.Conn)
	wsLeave    func(*ws.Conn)
}

func (tr *TinyRouter[T]) wsHeartbeat(c *ws.Conn, pong chan int64, ctx context.Context) {
	go func() {
		tick := time.NewTicker(30 * time.Second)
		defer tick.Stop()
		last := time.Now().UnixNano()

		for {
			select {
			case <-ctx.Done():
				return
			case n, ok := <-pong:
				if ok {
					last = n
				}
			case <-tick.C:
				if time.Since(time.Unix(0, last)) > 60*time.Second {
					tr.hub.Leave(c)
					return
				}
				if err := c.TryWrite(ws.Ping, []byte("heartbeat")); err != nil {
					tr.hub.Leave(c)
					return
				}
			}
		}
	}()
}

func (tr *TinyRouter[T]) wsHandler(c *ws.Conn) {
	tr.hub.Join(c)
	if tr.wsJoin != nil {
		tr.wsJoin(c)
	}
	defer func() {
		tr.hub.Leave(c)
		if tr.wsLeave != nil {
			tr.wsLeave(c)
		}
	}()

	pong := make(chan int64, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tr.wsHeartbeat(c, pong, ctx)

	for {
		opcode, msg, err := c.Read()
		if err != nil {
			log.Printf("read error: %v", err)
			return
		}

		switch opcode {
		case ws.Pong:
			pong <- time.Now().UnixNano()
			if fn, ok := tr.wsHandlers[ws.Pong]; ok {
				fn(tr.hub, msg)
			}
		case ws.Text:
			tr.hub.Broadcast(c, ws.Text, msg)
			if fn, ok := tr.wsHandlers[ws.Text]; ok {
				fn(tr.hub, msg)
			}
		case ws.Ping:
			if err := c.TryWrite(ws.Pong, []byte(msg)); err != nil {
				tr.hub.Leave(c)
				return
			}
			if fn, ok := tr.wsHandlers[ws.Ping]; ok {
				fn(tr.hub, msg)
			}
		case ws.Close:
			if err := c.TryWrite(ws.Close, []byte(msg)); err != nil {
				tr.hub.Leave(c)
			}
			if fn, ok := tr.wsHandlers[ws.Close]; ok {
				fn(tr.hub, msg)
			}
			return
		}
	}
}

func (tr *TinyRouter[T]) Listen(addr string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tr.hub.Run(ctx)

	srv := http.Server{
		Addr:    addr,
		Handler: tr,
	}

	go func() {
		log.Printf("HTTP server listening on: %s\n", addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
		log.Println("Stopped serving new connections.")
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	log.Println("Graceful shutdown complete.")
}

func (tr *TinyRouter[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// try websocket
	c, err := ws.Accept(w, r)
	if err == nil {
		tr.wsHandler(c)
		return
	}
	// build key
	pattern := clean(r.RequestURI)
	buff := new(bytes.Buffer)
	buff.WriteString(r.Method)
	buff.WriteByte(0x20)
	buff.WriteString(pattern)
	// search
	best, ok := bestRoute(tr.tree, buff.String())
	if !ok {
		if tr.err != nil {
			tr.err(w, r, nil)
		} else {
			http.Error(w, "nothing here", http.StatusNotFound)
		}
		return
	}

	if best.Kind() == route.Static {
		best.Handler(w, r, best.Params())
	} else {
		best.Handler(w, r, best.ParseParams(buff.String()))
	}
}

func (tr *TinyRouter[T]) Handle(method, pattern string, fn T) {
	// build key
	pattern = clean(pattern)
	buff := new(bytes.Buffer)
	buff.WriteString(method)
	buff.WriteByte(0x20)
	buff.WriteString(pattern)
	// insert
	new := route.New(buff.String(), fn)
	tr.tree.Insert(new.String(), new)
}

func (tr *TinyRouter[T]) Error(fn route.HandlerFn) {
	tr.err = fn
}

func (tr *TinyRouter[T]) WebSocketJoin(fn func(*ws.Conn)) {
	tr.wsJoin = fn
}

func (tr *TinyRouter[T]) WebSocketLeave(fn func(*ws.Conn)) {
	tr.wsLeave = fn
}

func (tr *TinyRouter[T]) WebSocketReceive(opcode ws.OpCode, fn func(*ws.Hub, string)) {
	tr.wsHandlers[opcode] = fn
}

func (tr *TinyRouter[T]) WebSocketPing(fn func(*ws.Hub, string)) {
	tr.wsHandlers[ws.Ping] = fn
}

func (tr *TinyRouter[T]) WebSocketPong(fn func(*ws.Hub, string)) {
	tr.wsHandlers[ws.Pong] = fn
}

func (tr *TinyRouter[T]) WebSocketBroadcast(opcode ws.OpCode, msg string) {
	tr.hub.Broadcast(nil, opcode, msg)
}

func NewRouter[T route.HandlerFn]() *TinyRouter[T] {
	return &TinyRouter[T]{
		hub:        ws.NewHub(),
		wsHandlers: make(map[ws.OpCode]func(*ws.Hub, string)),
	}
}
