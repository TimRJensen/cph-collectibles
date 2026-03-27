package route

import (
	"bytes"
	"net/http"
	"testing"
)

func TestParseSegments(t *testing.T) {
	cases := []string{
		"GET /a/{foo}/b/{bar}\r\n",
		"GET /a/{foo}/b/{bar} HTTP/1",
	}

	want := []segment{
		{[]byte("a"), 0, false},
		{[]byte("foo"), 0, true},
		{[]byte("b"), 0, false},
		{[]byte("bar"), 0, true},
	}
	for _, tc := range cases {
		_, got := parse([]byte(tc))

		if len(got) != len(want) {
			t.Fatalf("length mismatch; got %d, want %d", len(got), len(want))
		}

		for i := range got {
			if !bytes.Equal(want[i].buff, got[i].buff) {
				t.Fatalf("value mismatch; got %s, want %s", got[i].buff, want[i].buff)
			}
		}
	}
}

func TestParseClean(t *testing.T) {
	cases := []string{
		"GET /a/{foo}/b/{bar}\r\n",
		"GET /a/{foo}/b/{bar} HTTP/1",
	}

	want := "GET /a/*/b/*"
	for _, tc := range cases {
		got, _ := parse([]byte(tc))

		if !bytes.Equal([]byte(want), got) {
			t.Fatalf("value mismatch; got %s, want %s", got, want)
		}
	}
}

func TestParseParams(t *testing.T) {
	r := New[HandlerFn]("GET /a/{foo}/b/{bar}\r\n", func(http.ResponseWriter, *http.Request, ParamMap) {})

	if r.Kind() != Dynamic {
		t.Fatalf("value mismatch; got %v, want %v", r.Kind(), Dynamic)
	}

	r.ParseParams("GET /a/foo/b/bar HTTP/1.1")
	if len(r.params) != 2 {
		t.Fatalf("value mismatch; got %v, want %v", len(r.params), 2)
	}

	for k, v := range r.params {
		if string(k) != v {
			t.Fatalf("value mismatch; got %v, want %s", k, v)
		}
	}
}
