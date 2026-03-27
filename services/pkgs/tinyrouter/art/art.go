package art

import (
	"bytes"
	"fmt"
	"iter"
	"strings"
	"unsafe"
)

// kind identifies the concrete node type used in the adaptive radix tree.
// The values correspond to the node arities described in the ART paper:
//
//   - node4   : up to 4 children
//   - node16  : up to 16 children
//   - node48  : up to 48 children
//   - node256 : up to 256 children
//
// See: "The Adaptive Radix Tree: ARTful Indexing for Main-Memory Databases".
type kind = uint8

const (
	node4 kind = iota
	node16
	node48
	node256
)

// bucket wraps a value of type T and tracks whether it has been "locked".
//
// When locked is true, Set will not modify the stored value.
// This allows a Tree to be "locked" after construction so that its values
// cannot be accidentally overwritten (while still allowing lookups).
type bucket[T any] struct {
	v      T
	locked bool
}

func (b *bucket[T]) Lock() {
	b.locked = true
}

func (b *bucket[T]) Set(v T) {
	if !b.locked {
		b.v = v
	}
}

func (b *bucket[T]) Get() T {
	return b.v
}

// header holds metadata shared by all node types in the adaptive radix tree.
//
// The prefix/length pair represents the compressed path segment stored at
// this node, and children holds the count of non-nil child pointers.
type header struct {
	// The node's concrete kind (node4, node16, node48, node256).
	kind kind
	// The compressed key segment associated with this node.
	prefix []byte
	// The length of the compressed key segment.
	length int
	// Number of non-nil size attached to this node.
	size int
}

// matchPrefix compares this node's prefix against other starting at index idx.
// It returns the length of the common prefix.
func (h *header) matchPrefix(other []byte, idx int) int {
	m := min(max(len(other)-idx, 0), h.length)
	for i := range m {
		if other[i+idx] != h.prefix[i] {
			return i
		}
	}
	return m
}

type consumefn = func([]byte, int, int) int

// nodelike is the common interface implemented by all ART node types.
//
// It provides the operations needed by Tree to insert and search keys
// without knowing the concrete node kind.
type nodelike[T any] interface {
	meta() *header
	addChild(child nodelike[T], b byte, buff []byte) nodelike[T]
	insert(key []byte, depth int, bck *bucket[T]) nodelike[T]
	search(key []byte, depth int) *bucket[T]
	match(key []byte, wild byte, depth int, fn consumefn) *bucket[T]
}

// newNodelike constructs a new node of the given kind and attaches the
// provided bucket to it. The bucket may be nil for internal nodes that
// do not yet represent a terminal key.
func newNodelike[T any](which kind, bck *bucket[T]) nodelike[T] {
	switch which {
	case node4:
		return &n4[T]{
			header: header{
				node4,
				nil,
				0,
				0,
			},
			bucket: bck,
		}
	case node16:
		return &n16[T]{
			header: header{
				node16,
				nil,
				0,
				0,
			},
			bucket: bck,
		}
	case node48:
		return &n48[T]{
			header: header{
				node48,
				nil,
				0,
				0,
			},
			bucket: bck,
		}
	case node256:
		return &n256[T]{
			header: header{
				node256,
				nil,
				0,
				0,
			},
			bucket: bck,
		}
	default:
		// Should never happen
		return nil
	}
}

// nextNodelike finds the child of node reached by following edge b.
// It returns the child index and the child node, or (0, nil) if no child exists.
func nextNodelike[T any](node nodelike[T], b byte) (int, nodelike[T]) {
	switch n := node.(type) {
	case *n4[T]:
		for i := range n.size {
			if n.keys[i] == b {
				return i, n.children[i]
			}
		}
		return 0, nil
	case *n16[T]:
		for i := range n.size {
			if n.keys[i] == b {
				return i, n.children[i]
			}
		}
		return 0, nil
	case *n48[T]:
		if n.keys[b] != 0 {
			idx := int(n.keys[b]) - 1
			return idx, n.children[idx]
		}
		return 0, nil
	case *n256[T]:
		return int(b), n.children[b]
	default:
		return 0, nil
	}
}

// n256 corresponds to a 256-way node in the adaptive radix tree.
// It directly indexes children by the next key byte.
type n256[T any] struct {
	header
	children [256]nodelike[T]
	bucket   *bucket[T]
}

func (n *n256[T]) String() string {
	if n.bucket != nil {
		return fmt.Sprintf("Node256(prefix: %s, val: %v, children: %d)", n.prefix, n.bucket.v, n.size)
	}
	return fmt.Sprintf("Node256(prefix: %s, children: %d)", n.prefix, n.size)
}

func (n *n256[T]) meta() *header {
	return &n.header
}

func (n *n256[T]) full() bool {
	return n.size == 256
}

// addChild attaches child under the first byte of buff.
// If a child already exists for that byte, it is replaced.
func (n *n256[T]) addChild(child nodelike[T], b byte, buff []byte) nodelike[T] {
	if n.full() {
		// n256 is the highest arity; it cannot grow further.
		return n
	}

	// Update child
	h := child.meta()
	h.prefix = buff
	h.length = len(buff)

	// Add child
	if n.children[b] == nil {
		n.children[b] = child
		n.size++
	} else {
		n.children[b] = child
	}

	return n
}

// insert inserts bck into the subtree rooted at n256.
func (n *n256[T]) insert(buff []byte, depth int, bck *bucket[T]) nodelike[T] {
	if p := n.matchPrefix(buff, depth); p != n.length { // Prefix mismatch
		new := newNodelike[T](node4, nil).(*n4[T])
		new.prefix = buff[depth : depth+p]
		new.length = len(new.prefix)

		if depth+p == len(buff) { // path terminates here
			new.bucket = bck
			new.addChild(n, n.prefix[p], n.prefix[p+1:])
			return new
		}

		new.addChild(n, n.prefix[p], n.prefix[p+1:])
		new.addChild(newNodelike(node4, bck), buff[depth+p], buff[depth+p+1:])
		return new
	}

	depth += n.length
	if depth == len(buff) {
		if n.bucket != nil {
			n.bucket.Set(bck.v)
		} else {
			n.bucket = bck
		}
		return n
	}

	i, next := nextNodelike(n, buff[depth])
	if next != nil {
		n.children[i] = next.insert(buff, depth, bck)
		return n
	}

	return n.addChild(newNodelike(node4, bck), buff[depth], buff[depth+1:])
}

// search performs a lookup for buff starting at depth within n256.
func (n *n256[T]) search(buff []byte, depth int) *bucket[T] {
	if n.matchPrefix(buff, depth) != n.length {
		return nil
	}

	depth += n.length
	if depth == len(buff) {
		return n.bucket
	}

	_, child := nextNodelike(n, buff[depth])
	if child != nil {
		return child.search(buff, depth+1)
	}

	return nil
}

func (n *n256[T]) match(buff []byte, b byte, depth int, fn consumefn) *bucket[T] {
	if m := n.matchPrefix(buff, depth); m < n.length {
		if n.prefix[m] != b {
			return nil
		}
		depth += m
		depth += fn(buff, depth, len(buff)-depth)
	} else {
		depth += n.length
	}

	if depth >= len(buff) {
		return n.bucket
	}

	_, child := nextNodelike(n, buff[depth])
	if child != nil {
		if bck := child.match(buff, b, depth+1, fn); bck != nil {
			return bck
		}
	}

	_, child = nextNodelike(n, b)
	if child != nil {
		return child.match(buff, b, depth+fn(buff, depth, len(buff)-depth), fn)
	}

	return nil
}

// n48 corresponds to a 48-way node in the adaptive radix tree.
// It uses a 256-byte indirection table to map key bytes to up to 48 children.
type n48[T any] struct {
	header
	keys     [256]byte
	children [48]nodelike[T]
	bucket   *bucket[T]
}

func (n *n48[T]) String() string {
	if n.bucket != nil {
		return fmt.Sprintf("Node48(prefix: %s, val: %v, children: %d)", n.prefix, n.bucket.v, n.size)
	}
	return fmt.Sprintf("Node48(prefix: %s, children: %d)", n.prefix, n.size)
}

func (n *n48[T]) meta() *header {
	return &n.header
}

func (n *n48[T]) full() bool {
	return n.size == 48
}

// grow promotes n48 to an n256, preserving all children and metadata.
func (n *n48[T]) grow() nodelike[T] {
	new := &n256[T]{
		header: n.header,
	}
	new.kind += 1

	for i := range 256 {
		if n.keys[i] == 0 {
			continue
		}
		new.children[i] = n.children[int(n.keys[i])-1]
	}

	return new
}

// addChild attaches child under the first byte of buff,
// promoting to n256 if necessary.
func (n *n48[T]) addChild(child nodelike[T], b byte, buff []byte) nodelike[T] {
	if n.full() {
		return n.grow().(*n256[T]).addChild(child, b, buff)
	}

	// Update child
	h := child.meta()
	h.prefix = buff
	h.length = len(buff)

	// Add child
	n.keys[b] = byte(n.size + 1) // Skip 0
	n.children[n.size] = child
	n.size++

	return n
}

// insert inserts bck into the subtree rooted at n48.
func (n *n48[T]) insert(buff []byte, depth int, bck *bucket[T]) nodelike[T] {
	if p := n.matchPrefix(buff, depth); p != n.length { // Prefix mismatch
		new := newNodelike[T](node4, nil).(*n4[T])
		new.prefix = buff[depth : depth+p]
		new.length = len(new.prefix)

		if depth+p == len(buff) { // path terminates here
			new.bucket = bck
			new.addChild(n, n.prefix[p], n.prefix[p+1:])
			return new
		}

		new.addChild(n, n.prefix[p], n.prefix[p+1:])
		new.addChild(newNodelike(node4, bck), buff[depth+p], buff[depth+p+1:])
		return new
	}

	depth += n.length
	if depth == len(buff) {
		if n.bucket != nil {
			n.bucket.Set(bck.v)
		} else {
			n.bucket = bck
		}
		return n
	}

	i, next := nextNodelike(n, buff[depth])
	if next != nil {
		n.children[i] = next.insert(buff, depth, bck)
		return n
	}

	return n.addChild(newNodelike(node4, bck), buff[depth], buff[depth+1:])
}

// search performs a lookup for buff starting at depth within n48.
func (n *n48[T]) search(buff []byte, depth int) *bucket[T] {
	if n.matchPrefix(buff, depth) != n.length {
		return nil
	}

	depth += n.length
	if depth == len(buff) {
		return n.bucket
	}

	_, child := nextNodelike(n, buff[depth])
	if child != nil {
		return child.search(buff, depth+1)
	}

	return nil
}

func (n *n48[T]) match(buff []byte, b byte, depth int, fn consumefn) *bucket[T] {
	if m := n.matchPrefix(buff, depth); m < n.length {
		if n.prefix[m] != b {
			return nil
		}
		depth += m
		depth += fn(buff, depth, len(buff)-depth)
	} else {
		depth += n.length
	}

	if depth >= len(buff) {
		return n.bucket
	}

	_, child := nextNodelike(n, buff[depth])
	if child != nil {
		if bck := child.match(buff, b, depth+1, fn); bck != nil {
			return bck
		}
	}

	_, child = nextNodelike(n, b)
	if child != nil {
		return child.match(buff, b, depth+fn(buff, depth, len(buff)-depth), fn)
	}

	return nil
}

// n16 corresponds to a 16-way node in the adaptive radix tree.
type n16[T any] struct {
	header
	keys     [16]byte
	children [16]nodelike[T]
	bucket   *bucket[T]
}

func (n *n16[T]) String() string {
	if n.bucket != nil {
		return fmt.Sprintf("Node16(prefix: %s, val: %v, children: %d)", n.prefix, n.bucket.v, n.size)
	}
	return fmt.Sprintf("Node16(prefix: %s, children: %d)", n.prefix, n.size)
}

func (n *n16[T]) meta() *header {
	return &n.header
}

func (n *n16[T]) full() bool {
	return n.size == 16
}

// grow promotes n16 to an n48, preserving all children and metadata.
func (n *n16[T]) grow() nodelike[T] {
	// Next kind
	new := &n48[T]{
		header: n.header,
	}
	new.kind += 1

	// Grab children
	for i := 0; i < n.size; i++ {
		c := n.children[i]
		new.keys[n.keys[i]] = byte(i + 1)
		new.children[i] = c
	}

	return new
}

// addChild attaches child under the first byte of buff,
// promoting to n48 if necessary.
func (n *n16[T]) addChild(child nodelike[T], b byte, buff []byte) nodelike[T] {
	if n.full() {
		return n.grow().(*n48[T]).addChild(child, b, buff)
	}

	// Update child
	h := child.meta()
	h.prefix = buff
	h.length = len(buff)

	// Shift children & add child
	i := 0
	for i = 0; i < n.size; i++ {
		if b > n.keys[i] {
			continue
		}
		n.keys[i+1], n.children[i+1] = n.keys[i], n.children[i]
		break
	}
	n.keys[i] = b
	n.children[i] = child
	n.size++

	return n
}

// insert inserts bck into the subtree rooted at n16.
func (n *n16[T]) insert(buff []byte, depth int, bck *bucket[T]) nodelike[T] {
	if p := n.matchPrefix(buff, depth); p != n.length { // Prefix mismatch
		new := newNodelike[T](node4, nil).(*n4[T])
		new.prefix = buff[depth : depth+p]
		new.length = len(new.prefix)

		if depth+p == len(buff) { // path terminates here
			new.bucket = bck
			new.addChild(n, n.prefix[p], n.prefix[p+1:])
			return new
		}

		new.addChild(n, n.prefix[p], n.prefix[p+1:])
		new.addChild(newNodelike(node4, bck), buff[depth+p], buff[depth+p+1:])
		return new
	}

	depth += n.length
	if depth == len(buff) {
		if n.bucket != nil {
			n.bucket.Set(bck.v)
		} else {
			n.bucket = bck
		}
		return n
	}

	i, next := nextNodelike(n, buff[depth])
	if next != nil {
		n.children[i] = next.insert(buff, depth+1, bck)
		return n
	}

	return n.addChild(newNodelike(node4, bck), buff[depth], buff[depth+1:])
}

// search performs a lookup for buff starting at depth within n16.
func (n *n16[T]) search(buff []byte, depth int) *bucket[T] {
	if n.matchPrefix(buff, depth) != n.length {
		return nil
	}

	depth += n.length
	if depth == len(buff) {
		return n.bucket
	}

	_, child := nextNodelike(n, buff[depth])
	if child != nil {
		return child.search(buff, depth+1)
	}

	return nil
}

func (n *n16[T]) match(buff []byte, b byte, depth int, fn consumefn) *bucket[T] {
	if m := n.matchPrefix(buff, depth); m < n.length {
		if n.prefix[m] != b {
			return nil
		}
		depth += m
		depth += fn(buff, depth, len(buff)-depth)
	} else {
		depth += n.length
	}

	if depth >= len(buff) {
		return n.bucket
	}

	_, child := nextNodelike(n, buff[depth])
	if child != nil {
		if bck := child.match(buff, b, depth+1, fn); bck != nil {
			return bck
		}
	}

	_, child = nextNodelike(n, b)
	if child != nil {
		return child.match(buff, b, depth+fn(buff, depth, len(buff)-depth), fn)
	}

	return nil
}

// n4 corresponds to a 4-way node in the adaptive radix tree.
type n4[T any] struct {
	header
	keys     [4]byte
	children [4]nodelike[T]
	bucket   *bucket[T]
}

func (n *n4[T]) String() string {
	if n.bucket != nil {
		return fmt.Sprintf("Node4(prefix: %s, val: %v, children: %d)", n.prefix, n.bucket.v, n.size)
	}
	return fmt.Sprintf("Node4(prefix: %s, children: %d)", n.prefix, n.size)
}

func (n *n4[T]) meta() *header {
	return &n.header
}

func (n *n4[T]) full() bool {
	return n.size == 4
}

// grow promotes n4 to an n16, preserving all children and metadata.
func (n *n4[T]) grow() nodelike[T] {
	// Next kind
	new := &n16[T]{
		header: n.header,
	}
	new.kind += 1

	// Grab keys & children
	for i := 0; i < n.size; i++ {
		new.keys[i] = n.keys[i]
		new.children[i] = n.children[i]
	}

	return new
}

// addChild attaches child under the first byte of buff,
// promoting to n16 if necessary.
func (n *n4[T]) addChild(child nodelike[T], b byte, buff []byte) nodelike[T] {
	if n.full() {
		return n.grow().(*n16[T]).addChild(child, b, buff)
	}

	// Update child
	h := child.meta()
	h.prefix = buff
	h.length = len(h.prefix)

	// Shift children & add child
	i := 0
	for i = 0; i < n.size; i++ {
		if b > n.keys[i] {
			continue
		}
		n.keys[i+1], n.children[i+1] = n.keys[i], n.children[i]
		break
	}
	n.keys[i] = b
	n.children[i] = child
	n.size++

	return n
}

// insert inserts bck into the subtree rooted at n4.
func (n *n4[T]) insert(buff []byte, depth int, bck *bucket[T]) nodelike[T] {
	if p := n.matchPrefix(buff, depth); p != n.length { // Prefix mismatch
		new := newNodelike[T](node4, nil).(*n4[T])
		new.prefix = buff[depth : depth+p]
		new.length = len(new.prefix)

		if depth+p == len(buff) { // path terminates here
			new.bucket = bck
			new.addChild(n, n.prefix[p], n.prefix[p+1:])
			return new
		}

		new.addChild(n, n.prefix[p], n.prefix[p+1:])
		new.addChild(newNodelike(node4, bck), buff[depth+p], buff[depth+p+1:])
		return new
	}

	depth += n.length
	if depth == len(buff) {
		if n.bucket != nil {
			n.bucket.Set(bck.v)
		} else {
			n.bucket = bck
		}
		return n
	}

	i, next := nextNodelike(n, buff[depth])
	if next != nil {
		n.children[i] = next.insert(buff, depth+1, bck)
		return n
	}

	return n.addChild(newNodelike(node4, bck), buff[depth], buff[depth+1:])
}

// search performs a lookup for buff starting at depth within n4.
func (n *n4[T]) search(buff []byte, depth int) *bucket[T] {
	if n.matchPrefix(buff, depth) != n.length {
		return nil
	}

	depth += n.length
	if depth == len(buff) {
		return n.bucket
	}

	_, child := nextNodelike(n, buff[depth])
	if child != nil {
		return child.search(buff, depth+1)
	}

	return nil
}

func (n *n4[T]) match(buff []byte, b byte, depth int, fn consumefn) *bucket[T] {
	if m := n.matchPrefix(buff, depth); m < n.length {
		if n.prefix[m] != b {
			return nil
		}
		depth += m
		depth += fn(buff, depth, len(buff)-depth)
	} else {
		depth += n.length
	}

	if depth >= len(buff) {
		return n.bucket
	}

	_, child := nextNodelike(n, buff[depth])
	if child != nil {
		if bck := child.match(buff, b, depth+1, fn); bck != nil {
			return bck
		}
	}

	_, child = nextNodelike(n, b)
	if child != nil {
		return child.match(buff, b, depth+fn(buff, depth, len(buff)-depth), fn)
	}

	return nil
}

func unsafeCast(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// Tree is a generic adaptive radix tree (ART) keyed by strings and storing
// values of type T.
//
// Keys are treated as byte slices and compressed into path segments
// (prefixes) as described in the ART paper. Internal nodes may also carry
// values when a key terminates exactly at that node's path.
type Tree[T any] struct {
	root   nodelike[T]
	locked bool
}

// Insert stores value v under key k in the tree.
// If a value already exists for k, it will be overwritten unless the
// bucket has been locked (see Lock).
//
// Insert is safe to call multiple times before Lock, and may also be
// called after Lock to insert new keys, but existing values cannot be
// changed once locked.
func (t *Tree[T]) Insert(k string, v T) {
	if t.root == nil {
		t.root = &n4[T]{
			header: header{
				node4,
				unsafeCast(k),
				len(k),
				0,
			},
			bucket: &bucket[T]{v, t.locked},
		}
		return
	}
	t.root = t.root.insert(unsafeCast(k), 0, &bucket[T]{v, t.locked})
}

// Search performs an exact lookup of key k in the tree.
//
// The lookup is byte-for-byte: The key must match a stored key exactly
// for a value to be returned.
//
// Search returns the associated value and true if an exact match is found.
// Otherwise it returns the zero value of T and false.
func (t *Tree[T]) Search(k string) (T, bool) {
	var zero T
	if t.root == nil {
		return zero, false
	}

	if bck := t.root.search(unsafeCast(k), 0); bck == nil {
		return zero, false
	} else {
		return bck.Get(), true
	}
}

// MatchFn performs a lookup of key k in the tree, allowing the caller to define
// custom wildcard matching semantics via fn.
//
// The tree itself is byte-oriented and has no understanding of wildcards,
// path segments, or higher-level structure. Instead:
//
//   - wildcard specifies a byte value that may appear in node prefixes or as
//     an edge label and is treated specially during matching.
//   - fn is responsible for determining how many bytes of the key are consumed
//     when a wildcard is encountered.
//
// The callback fn is invoked with:
//   - key: the given k
//   - read: the current index into k where wildcard matching begins
//   - remainder: the number of bytes remaining in k starting at read
//
// fn must return the number of bytes to consume from k. This value is added to
// the traversal depth before continuing the search.
//
// A return value less than zero or greater than remainder results in undefined
// behavior. fn must not mutate k and must be safe for repeated calls during a
// single lookup.
//
// MatchFn returns the value associated with the best matching key, if any.
func (t *Tree[T]) MatchFn(k string, wildcard byte, fn consumefn) (T, bool) {
	var zero T
	if t.root == nil {
		return zero, false
	}

	if bck := t.root.match(unsafeCast(k), wildcard, 0, fn); bck == nil {
		return zero, false
	} else {
		return bck.Get(), true
	}
}

func consume(stop byte) consumefn {
	return func(key []byte, read, remainder int) int {
		for i := range remainder {
			if key[read+i] == stop {
				return i
			}
		}
		return remainder
	}
}

// Match performs a lookup of key k in the tree using a simple wildcard
// matching rule.
//
// wildcard specifies the byte value that represents a wildcard in stored
// keys. When a wildcard is encountered during traversal, matching consumes
// bytes from k until the stop byte is reached.
//
// stop specifies the byte that terminates wildcard consumption. For example,
// using wildcard='*' and stop='/' causes '*' to match a single path segment.
//
// Match is a convenience wrapper around MatchFn and is suitable for common
// use cases such as path-based routing.
//
// For more complex wildcard semantics, use MatchFn directly.
func (t *Tree[T]) Match(k string, wildcard, stop byte) (T, bool) {
	var zero T
	if t.root == nil {
		return zero, false
	}

	if bck := t.root.match(unsafeCast(k), byte(wildcard), 0, consume(stop)); bck == nil {
		return zero, false
	} else {
		return bck.Get(), true
	}
}

func walk[T any](n nodelike[T], prefix []byte, fn func([]byte, *bucket[T]) bool) bool {
	// Extend current prefix with this node's prefix
	curr := append(prefix, n.meta().prefix...)

	buff := bytes.Buffer{}
	buff.Write(prefix)
	switch n := n.(type) {
	case *n4[T]:
		buff.Write(n.prefix)
		if n.bucket != nil {
			if !fn(buff.Bytes(), n.bucket) {
				return false
			}
		}
		for i, c := range n.children {
			if c == nil {
				continue
			}
			buff.WriteByte(n.keys[i])
			if !walk(c, buff.Bytes(), fn) {
				return false
			}
			buff.Truncate(buff.Len() - 1)
		}
		return true
	case *n16[T]:
		if n.bucket != nil {
			if !fn(curr, n.bucket) {
				return false
			}
		}
		for i, c := range n.children {
			if c == nil {
				continue
			}
			buff.WriteByte(n.keys[i])
			if !walk(c, buff.Bytes(), fn) {
				return false
			}
			buff.Truncate(buff.Len() - 1)
		}
		return true
	case *n48[T]:
		if n.bucket != nil {
			if !fn(curr, n.bucket) {
				return false
			}
		}
		for b := range 256 {
			k := n.keys[byte(b)]
			if k == 0 {
				continue
			}
			buff.WriteByte(k)
			if !walk(n.children[int(k)-1], buff.Bytes(), fn) {
				return false
			}
			buff.Truncate(buff.Len() - 1)
		}
		return true
	case *n256[T]:
		if n.bucket != nil {
			if !fn(curr, n.bucket) {
				return false
			}
		}
		for b := range 256 {
			if n.children[b] == nil {
				continue
			}

			buff.WriteByte(byte(b))
			if !walk(n.children[b], buff.Bytes(), fn) {
				return false
			}
			buff.Truncate(buff.Len() - 1)
		}
		return true
	default:
		// Should never happen
		return true
	}
}

// All returns an iterator over all key/value pairs stored in the tree,
// in lexicographic order of their keys.
//
// It uses Go's iterator pattern from the iter package:
//
//	for k, v := range tree.All() {
//	    // use k, v
//	}
func (t *Tree[T]) All() iter.Seq2[string, T] {
	return func(yield func(k string, v T) bool) {
		if t.root == nil {
			return
		}

		walk(t.root, make([]byte, 0), func(k []byte, bck *bucket[T]) bool {
			return yield(string(k), bck.Get())
		})
	}
}

// Lock makes all existing and future buckets in the tree immutable.
//
// After Lock is called, Insert can still add new keys, but cannot
// overwrite values stored before locking. This is useful for finalizing
// a tree (e.g., a routing table) after configuration.
func (t *Tree[T]) Lock() {
	t.locked = true
	walk(t.root, make([]byte, 0), func(k []byte, bck *bucket[T]) bool {
		bck.locked = t.locked
		return true
	})
}

func prettyPrint[T any](n nodelike[T], depth int) {
	indent := strings.Repeat("  ", depth)

	switch n := n.(type) {
	case *n4[T]:
		fmt.Printf("%s%v\n", indent, n)
		for i := range n.size {
			fmt.Printf("%s  ├─[%s]→\n", indent, string(n.keys[i]))
			prettyPrint(n.children[i], depth+2)
		}
	case *n16[T]:
		fmt.Printf("%s%v\n", indent, n)
		for i := range n.size {
			fmt.Printf("%s  ├─[%s]→\n", indent, string(n.keys[i]))
			prettyPrint(n.children[i], depth+2)
		}
	case *n48[T]:
		fmt.Printf("%s%v\n", indent, n)
		for b := range 256 {
			k := n.keys[byte(b)]
			if k == 0 {
				continue
			}
			fmt.Printf("%s  ├─[%s]→\n", indent, string(rune(b)))
			prettyPrint(n.children[int(k)-1], depth+2)
		}
	case *n256[T]:
		fmt.Printf("%s%v\n", indent, n)
		for b := range 256 {
			if n.children[b] == nil {
				continue
			}
			fmt.Printf("%s  ├─[%s]→\n", indent, string(rune(b)))
			prettyPrint(n.children[b], depth+2)
		}
	default:
		// Should never happen
	}
}

// PrettyPrint prints a human-readable representation of the entire tree
// to stdout. It is intended for debugging and visualization only.
func (t *Tree[T]) PrettyPrint() {
	if t.root == nil {
		fmt.Println("<empty tree>")
		return
	}
	prettyPrint(t.root, 0)
}
