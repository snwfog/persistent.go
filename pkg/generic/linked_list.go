package generic

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync/atomic"
	"unsafe"

	"github.com/cheekybits/genny/generic"
	"github.com/dchest/siphash"

	"github.com/snwfog/persistent.go/pkg/identify"
	"github.com/snwfog/persistent.go/pkg/util"
)

var (
	inserterr = errors.New("insert failed")
)

type Value generic.Type

func NewValueLinkedList() *linkedlist {
	head := &sentinel{node{}}
	tail := &sentinel{node{}}

	head.next = unsafe.Pointer(tail)
	return &linkedlist{
		head: unsafe.Pointer(head),
		tail: unsafe.Pointer(tail),
	}
}

// region Node
func newnode(valueptr *Value) *node {
	return &node{
		valueptr: unsafe.Pointer(valueptr),
		key:      getid(valueptr),
	}
}

type node struct {
	key      uint64
	valueptr unsafe.Pointer // TODO: use uintptr
	next     unsafe.Pointer // What if GC runs?
}

func (n *node) value() *Value {
	return (*Value)(atomic.LoadPointer(&n.valueptr))
}

func (n *node) nextnode() *node {
	return (*node)(atomic.LoadPointer(&n.next))
}

func (n *node) nextptr() unsafe.Pointer {
	return atomic.LoadPointer(&n.next)
}

type sentinel struct {
	node
}

// endregion

// region linkedlist
type linkedlist struct {
	head unsafe.Pointer
	tail unsafe.Pointer

	len int32
}

func (l *linkedlist) Len() int {
	return int(atomic.LoadInt32(&l.len))
}

func (l *linkedlist) headnode() *node {
	return (*node)(atomic.LoadPointer(&l.head))
}

func (l *linkedlist) tailnode() *node {
	return (*node)(atomic.LoadPointer(&l.tail))
}

func (l *linkedlist) Insert(v *Value) (bool, error) {
	n := newnode(v)
	var left, right *node
	for {
		left, right = l.search(n.key)

		if right != l.tailnode() && right.key == n.key {
			return false, inserterr
		}

		n.next = unsafe.Pointer(right)
		if atomic.CompareAndSwapPointer(&left.next, unsafe.Pointer(right), unsafe.Pointer(n)) {
			atomic.AddInt32(&l.len, 1)
			return true, nil
		}
	}
}

func (l *linkedlist) Upsert(v *Value) (bool, error) {
	n := newnode(v)
	_, right := l.search(n.key)
	if right != l.tailnode() && right.key == n.key {
		atomic.StorePointer(&right.valueptr, n.valueptr)
		return true, nil
	}

	return l.Insert(v)
}

func (l *linkedlist) Delete(v *Value) (bool, error) {
	var right, left *node
	var rightnext unsafe.Pointer

	key := getid(v)
	for {
		left, right = l.search(key)
		if right == l.tailnode() || right.key != key {
			return false, nil // not deleted cause not found
		}

		rightnext = right.nextptr()
		if !deletemarked(rightnext) {
			if atomic.CompareAndSwapPointer(&right.next, rightnext, deletemark(rightnext)) {
				atomic.AddInt32(&l.len, -1)
				break
			}
		}
	}

	// Cleanup run
	if !atomic.CompareAndSwapPointer(&left.next, unsafe.Pointer(right), rightnext) {
	  _, _ = l.search(right.key)
	}

	return true, nil
}

// search returns left and right such
// - left and right are unmarked node
// - right is immediate successor of left
// - left.key < key <= right.key
// - - remember: we do not allow duplicate on insert
func (l *linkedlist) search(key uint64) (left, right *node) {
	var leftnext *node

	for {
		curr := l.headnode()
		nextptr := curr.nextptr()

		// find left and right node
		for {
			// if current node is not logically deleting
			// remember: node.next serve 2 purposes,
			// 1. unmark(nextptr) -> gives next node
			// 2. deletemarked(nextptr) -> means current
			// node is logically removed from the linkedlist
			if !deletemarked(nextptr) {
				left = curr
				leftnext = (*node)(nextptr)
			}

			curr = (*node)(deleteunmark(nextptr))
			if curr == l.tailnode() {
				break
			}

			// if curr is not deleted, we found the right node
			nextptr = curr.nextptr()
			if !deletemarked(nextptr) && curr.key >= key {
				break
			}
		}

		// check if left right are adjacent
		right = curr
		if leftnext == right {
			if right != l.tailnode() && deletemarked(right.nextptr()) {
				// oopsie, someone just logically deleted curr node,
				// restart
				continue
			}

			return left, right
		}

		// cleanup, set left.next to be right node
		if atomic.CompareAndSwapPointer(&left.next, unsafe.Pointer(leftnext), unsafe.Pointer(right)) {
			// oopsie, someone just logically deleted curr node,
			// restart
			if right != l.tailnode() && deletemarked(right.nextptr()) {
				continue
			}

			return left, right
		}
	}
}

func (l *linkedlist) Contains(v *Value) bool {
	key := getid(v)
	_, right := l.search(key)
	if right == l.tailnode() || right.key != key {
		return false
	}

	return true
}

func (l *linkedlist) Iterator() *iterator {
	return NewIterator(l)
}

func (l *linkedlist) CyclicIterator() *cycliciterator {
	return NewCyclicIterator(l)
}

// endregion

// region Iterator
type iterator struct {
	curr unsafe.Pointer
	list *linkedlist
}

func NewIterator(list *linkedlist) *iterator {
	return &iterator{
		curr: unsafe.Pointer(list.headnode()),
		list: list,
	}
}

func (it *iterator) Next() (*Value, bool) {
	n, ok := it.nextnode()
	if !ok {
		return nil, ok
	}

	return n.value(), ok
}

func (it *iterator) nextnode() (*node, bool) {
	for {
		curr := (*node)(it.curr)
		if curr == it.list.tailnode() {
			return nil, false
		}

		next := curr.nextptr()
		nextnode := (*node)(next)
		if atomic.CompareAndSwapPointer(&it.curr, unsafe.Pointer(curr), next) {
			if nextnode == it.list.tailnode() {
				return nil, false
			}

			if deletemarked(next) {
				continue
			}

			return nextnode, true
		}
	}
}

// endregion

// region CyclicIterator
type cycliciterator struct {
	curr unsafe.Pointer
	list *linkedlist
}

func NewCyclicIterator(list *linkedlist) *cycliciterator {
	return &cycliciterator{
		curr: unsafe.Pointer(list.headnode()),
		list: list,
	}
}

func (it *cycliciterator) Next() (*Value, bool) {
	n, ok := it.nextnode()
	if !ok {
		return nil, ok
	}

	return n.value(), ok
}

func (it *cycliciterator) nextnode() (*node, bool) {
	for {
		curr := (*node)(it.curr)
		if curr == it.list.tailnode() {
			return nil, false
		}

		next := curr.nextptr()
		nextnode := (*node)(next)
		if atomic.CompareAndSwapPointer(&it.curr, unsafe.Pointer(curr), next) {
			if nextnode == it.list.tailnode() {
				head := unsafe.Pointer(it.list.headnode())
				// Try to swap to head
				// We don't care if its sucesssful or not
				// If its successful, next run will grab the next node
				// If it fails, means some thread already swapped, we rerun to grab next
				atomic.CompareAndSwapPointer(&it.curr, unsafe.Pointer(curr), head)
			}

			if deletemarked(next) {
				continue
			}

			return nextnode, true
		}
	}
}

// endregion

// region Utilities
const (
	// intSizeBytes is the size in byte of an int or uint value.
	intSizeBytes = strconv.IntSize >> 3

	// generated by splitting the md5 sum of "hashmap"
	sipHashKey1 = 0xdda7806a4847ec61
	sipHashKey2 = 0xb5940c2623a5aabd
)

func getid(v interface{}) uint64 {
	if util.IsNil(v) {
		panic("v cannot be nil")
	}

	if itf, ok := v.(identify.Identify); ok {
		return itf.Identity()
	}

	panic("must implement Identity")
}

func getKeyHash(v interface{}) uint64 {
	// if util.IsNil(v) {
	// 	panic("v cannot be nil")
	// }

	// TODO: Support buildtin types
	switch x := v.(type) {
	case string:
		return getStringHash(x)
	case []byte:
		return siphash.Hash(sipHashKey1, sipHashKey2, x)
	case int:
		return getUintptrHash(uintptr(x))
	case int8:
		return getUintptrHash(uintptr(x))
	case int16:
		return getUintptrHash(uintptr(x))
	case int32:
		return getUintptrHash(uintptr(x))
	case int64:
		return getUintptrHash(uintptr(x))
	case uint:
		return getUintptrHash(uintptr(x))
	case uint8:
		return getUintptrHash(uintptr(x))
	case uint16:
		return getUintptrHash(uintptr(x))
	case uint32:
		return getUintptrHash(uintptr(x))
	case uint64:
		return getUintptrHash(uintptr(x))
	case uintptr:
		return getUintptrHash(x)
	}

	panic(fmt.Errorf("unsupported v type %T", v))
}

// Verify endianess
// https://github.com/tensorflow/tensorflow/blob/master/tensorflow/go/tensor.go#L498

// zero copy from string to []byte
// https://mina86.com/2017/golang-string-and-bytes/
func getStringHash(s string) uint64 {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	buf := *(*[]byte)(unsafe.Pointer(&bh))
	return siphash.Hash(sipHashKey1, sipHashKey2, buf)
}

func getUintptrHash(num uintptr) uint64 {
	bh := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&num)),
		Len:  intSizeBytes,
		Cap:  intSizeBytes,
	}
	buf := *(*[]byte)(unsafe.Pointer(&bh))
	return siphash.Hash(sipHashKey1, sipHashKey2, buf)
}

func deletemarked(ptr unsafe.Pointer) bool {
	return (uintptr(ptr) & 0x1) > 0
}

func deletemark(ptr unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(uintptr(ptr) | 0x1)
}

func deleteunmark(ptr unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(uintptr(ptr) &^ 0x1)
}

// endregion
