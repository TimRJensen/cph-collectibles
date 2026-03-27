package art

import (
	"fmt"
	"maps"
	"testing"
)

func TestInsertSearch_Basics(t *testing.T) {
	var tr Tree[int]

	tr.Insert("foo", 1)
	tr.Insert("bar", 2)
	tr.Insert("baz", 3)

	tests := []struct {
		key  string
		want any
		ok   bool
	}{
		{"foo", 1, true},
		{"bar", 2, true},
		{"baz", 3, true},
		{"ba", nil, false},
		{"qux", nil, false},
	}

	for _, tt := range tests {
		got, ok := tr.Search(tt.key)
		if ok != tt.ok {
			t.Fatalf("Search(%q) ok=%v, want %v", tt.key, ok, tt.ok)
		}
		if ok && got != tt.want {
			t.Fatalf("Search(%q)=%v, want %v", tt.key, got, tt.want)
		}
	}
}

func TestInsertSearch_PrefixAndDivergence(t *testing.T) {
	var tr Tree[int]

	keys := []string{
		"foo",
		"foobar",
		"fo",
		"abcd",
		"abcdef",
		"abX",
		"abYz",
	}

	for i, k := range keys {
		tr.Insert(k, i)
	}

	for i, k := range keys {
		got, ok := tr.Search(k)
		if !ok {
			t.Fatalf("Search(%q) not found", k)
		}
		if got != i {
			t.Fatalf("Search(%q)=%v, want %v", k, got, i)
		}
	}

	// Some misses
	for _, miss := range []string{"f", "fooX", "abc", "ab", "abYZ"} {
		if _, ok := tr.Search(miss); ok {
			t.Fatalf("Search(%q) should be miss", miss)
		}
	}
}

func TestInsertSearch_Promotion(t *testing.T) {
	var tr Tree[int]

	// Force growth in root: "k<0..255>"
	for i := 1; i < 100; i++ { // enough to get past 4 and 16
		k := fmt.Sprintf("k%c", byte(i))
		tr.Insert(k, i)
	}

	for i := 1; i < 100; i++ {
		k := fmt.Sprintf("k%c", byte(i))
		got, ok := tr.Search(k)
		if !ok || got != i {
			t.Fatalf("Search(%q)=%v,%v, want %v,true", k, got, ok, i)
		}
	}
}

func TestAll_Basics(t *testing.T) {
	var tr Tree[int]

	tr.Insert("/foo", 1)
	tr.Insert("/bar", 2)
	tr.Insert("/baz/qux", 3)

	got := make(map[string]int)

	for k, v := range tr.All() {
		if _, ok := got[k]; ok {
			t.Fatalf("duplicate key from All(): %q", k)
		}
		got[k] = v
	}

	want := map[string]int{
		"/foo":     1,
		"/bar":     2,
		"/baz/qux": 3,
	}

	if len(got) != len(want) {
		t.Fatalf("All() returned %d items, want %d", len(got), len(want))
	}

	for k, v := range want {
		gotV, ok := got[k]
		if !ok {
			t.Fatalf("All() missing key %q", k)
		}
		if gotV != v {
			t.Fatalf("All()[%q]=%d, want %d", k, gotV, v)
		}
	}
}

func TestAll_AfterFreeze(t *testing.T) {
	var tr Tree[int]

	tr.Insert("/foo", 1)
	tr.Insert("/bar", 2)
	tr.Lock()

	// Insert after freeze should not affect values (depending on your semantics)
	tr.Insert("/foo", 99)
	tr.Insert("/baz", 3)
	got := maps.Collect(tr.All())

	// Depending on whether you allow new keys after Freeze or not,
	// adjust expectations. Assuming "no overwrite, but new keys allowed":
	if got["/foo"] != 1 {
		t.Fatalf("All() after Freeze(): /foo=%d, want 1", got["/foo"])
	}
}

func TestPrettyPrint(t *testing.T) {
	var tr Tree[int]

	// Some nice, router-ish keys
	tr.Insert("/foo", 1)
	tr.Insert("/foobar", 2)
	tr.Insert("/food", 3)
	tr.Insert("/bar", 4)
	tr.Insert("/baz/{qux}", 5)
	tr.Insert("/baz/{quux}", 6)
	tr.PrettyPrint()
}

var static = []string{
	"/foo/bar",
	"/foo/baz",
	"/foo/qux",

	"/foo/bar/baz",
	"/foo/bar/qux",

	"/api/v1/users",
	"/api/v1/posts",
	"/api/v1/comments",

	"/api/v2/users",
	"/api/v2/posts",

	"/assets/css/main",
	"/assets/js/app",
	"/assets/img/logo",

	"/health",
	"/metrics",
	"/status",
}

func BenchmarkSearch(b *testing.B) {
	tr := Tree[int]{}
	for i, key := range static {
		tr.Insert(key, i)
	}

	b.ResetTimer()
	for i := range b.N {
		tr.Search(static[i%len(static)])
	}
}

var dynamicKey = []string{
	"/foo/*/bar",
	"/foo/*/baz",
	"/foo/*/qux",

	"/foo/bar/*",
	"/foo/baz/*",

	"/api/*/users",
	"/api/*/posts",
	"/api/*/comments",

	"/api/v1/*",
	"/api/v2/*",

	"/assets/*/main",
	"/assets/*/app",
	"/assets/*/logo",

	"/user/*",
	"/user/*/profile",
	"/user/*/settings",
}

var dynamicQuery = []string{
	"/foo/123456789/bar",
	"/foo/123456789/baz",
	"/foo/123456789/qux",

	"/foo/bar/123456789",
	"/foo/baz/123456789",

	"/api/123456789/users",
	"/api/123456789/posts",
	"/api/123456789/comments",

	"/api/v1/123456789",
	"/api/v2/123456789",

	"/assets/123456789/main",
	"/assets/123456789/app",
	"/assets/123456789/logo",

	"/user/123456789",
	"/user/123456789/profile",
	"/user/123456789/settings",
}

func BenchmarkMatch(b *testing.B) {
	tr := Tree[int]{}
	for i, key := range dynamicKey {
		tr.Insert(key, i)
	}

	b.ResetTimer()
	for i := range b.N {
		k := dynamicQuery[i%len(dynamicQuery)]
		tr.MatchFn(k, '*', func(key []byte, read, remaining int) int {
			for i := range remaining {
				if key[read+i] == 0x2F {
					return i
				}
			}
			return remaining
		})
	}
}
