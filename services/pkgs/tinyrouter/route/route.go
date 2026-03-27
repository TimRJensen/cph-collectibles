package route

import (
	"bytes"
	"net/http"
)

type (
	Kind      uint8
	Key       string
	ParamMap  map[Key]string
	HandlerFn func(w http.ResponseWriter, r *http.Request, params ParamMap)
)

type Handlers interface {
	http.HandlerFunc | HandlerFn
}

const (
	Static Kind = iota
	Dynamic
)

type segment struct {
	buff []byte
	idx  int
	wild bool
}

func (s *segment) Bytes() []byte {
	return s.buff
}

func (s *segment) IsWild() bool {
	return s.wild
}

func parse(s []byte) ([]byte, []segment) {
	// find segments
	pos := make([]int, 32)
	n, m := len(s), 0
	for i, flag := 0, true; flag; i++ {
		switch s[i] {
		case 0x0D:
			if s[i+1] != 0x0A {
				// malformed, handle later
				return nil, nil
			}
			n = i
			flag = false
		case 0x20:
			if m == 0 {
				continue
			}
			n = i
			flag = false
		case 0x2F:
			pos[m] = i
			m++
		}
		flag = flag && i+1 < n
	}
	pos[m] = n

	segments := make([]segment, m)
	for i := range m {
		a, b := pos[i]+1, pos[i+1]
		if a == b {
			break
		}

		if s[a] == 0x7B {
			// wildcard
			if s[b-1] != 0x7D {
				// malformed, handle later
				return nil, nil
			}
			segments[i] = segment{s[a+1 : b-1], i, true}
		} else {
			// literal
			segments[i] = segment{s[a:b], i, false}
		}
	}

	buff := new(bytes.Buffer)
	buff.Write(s[:pos[0]])
	for i := range m {
		a, b := pos[i], pos[i+1]
		if a+1 == b {
			buff.Write(s[a:])
			break
		}

		switch s[a+1] {
		case 0x7B, 0x0D:
			buff.WriteByte(0x2F)
			buff.WriteByte(0x2A)
		default:
			buff.Write(s[a:b])
		}
	}

	return bytes.TrimSpace(buff.Bytes()), segments
}

type Route[T Handlers] struct {
	kind     Kind
	buff     []byte
	params   map[Key]string
	Segments []segment
	Handler  T
}

func (r Route[T]) String() string {
	return string(r.buff)
}

func (r Route[T]) Bytes() []byte {
	return r.buff
}

func (r Route[T]) Kind() Kind {
	return r.kind
}

func (r *Route[T]) Params() map[Key]string {
	return r.params
}

func (r *Route[T]) ParseParams(route string) map[Key]string {
	buff := []byte(route)
	_, segments := parse(buff)

	for _, s := range r.Segments {
		if s.idx >= len(segments) {
			break
		}

		if !s.wild {
			continue
		}
		r.params[Key(s.buff)] = string(segments[s.idx].buff)
	}
	return r.params
}

func New[T Handlers](route string, fn T) *Route[T] {
	buff := []byte(route)
	clean, segments := parse(buff)
	params := make(map[Key]string)
	for _, s := range segments {
		if !s.wild {
			continue
		}
		params[Key(s.buff)] = ""
	}

	if len(params) != 0 {
		return &Route[T]{
			Dynamic,
			clean,
			params,
			segments,
			fn,
		}
	}
	return &Route[T]{
		Static,
		clean,
		nil,
		nil,
		fn,
	}
}
