package tinyrouter

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TimRJensen/tinyrouter/route"
)

func TestClean(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"root", "/", "/"},
		{"dot", ".", "/"},
		{"empty", "", "/"},
		{"simple", "/foo", "/foo"},
		{"with_trailing_slash", "/foo/", "/foo"},
		{"no_leading_slash", "foo", "foo"}, // current behaviour
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clean(tt.in)
			if got != tt.want {
				t.Fatalf("clean(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestTinyRouterStatic(t *testing.T) {
	r := NewRouter()

	called := false
	r.Handle(http.MethodGet, "/", func(w http.ResponseWriter, req *http.Request, params route.ParamMap) {
		called = true
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("hello"))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if !called {
		t.Fatalf("handler was not called")
	}
	if rr.Code != http.StatusTeapot {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusTeapot)
	}
	if body := rr.Body.String(); body != "hello" {
		t.Fatalf("body = %q, want %q", body, "hello")
	}
}

func TestTinyRouterDynamic(t *testing.T) {
	r := NewRouter()
	cases := map[string][]string{
		"foo": {"/user/{test}/blog/{post}", "/user/foo/blog/bar"},
		"baz": {"/user/{user}/{blog}/{test}", "/user/foo/bar/baz"},
		"qux": {"/user/{test}", "/user/qux"},
	}
	called := false

	for _, tc := range cases {
		r.Handle(http.MethodGet, tc[0], func(w http.ResponseWriter, req *http.Request, params route.ParamMap) {
			called = true
			w.WriteHeader(http.StatusTeapot)
			_, _ = w.Write([]byte(params["test"]))
		})
	}

	for want, tc := range cases {
		req := httptest.NewRequest(http.MethodGet, tc[1], nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if !called {
			t.Fatalf("handler was not called")
		}
		if rr.Code != http.StatusTeapot {
			t.Fatalf("status code = %d, want %d", rr.Code, http.StatusTeapot)
		}
		if body := rr.Body.String(); body != want {
			t.Fatalf("body = %q, want %q", body, want)
		}
		called = false

	}

	req := httptest.NewRequest(http.MethodGet, "/user/foo/bar", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestTinyRouterStaticDynamic(t *testing.T) {
	r := NewRouter()
	called := false
	r.Handle(http.MethodGet, "/user/{user}/blog/{blog}", func(w http.ResponseWriter, req *http.Request, params route.ParamMap) {
		called = true
		fmt.Println(params)
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte(params["blog"]))
	})
	r.Handle(http.MethodGet, "/user/me", func(w http.ResponseWriter, req *http.Request, params route.ParamMap) {
		called = true
		fmt.Println(params)
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("baz"))
	})

	req := httptest.NewRequest(http.MethodGet, "/user/me", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if !called {
		t.Fatalf("handler was not called")
	}
	if rr.Code != http.StatusTeapot {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusTeapot)
	}
	if body := rr.Body.String(); body != "baz" {
		t.Fatalf("body = %q, want %q", body, "baz")
	}
}

func TestTinyRouter404(t *testing.T) {
	r := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/404", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusNotFound)
	}
}
