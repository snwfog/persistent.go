// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/cheekybits/genny

package item

import (
	"errors"
	"fmt"
	"strconv"
	"sync/atomic"
	"unsafe"

	"github.com/dchest/siphash"
	"github.com/snwfog/persistent.go/pkg/identify"
)

var (
	inserterr = errors.New("insert failed")
)

func NewItemLinkedList() *linkedlist {
	head := &sentinel{node{}}
	tail := &sentinel{node{}}

	head.next = unsafe.Pointer(tail)
	return &linkedlist{
		head: unsafe.Pointer(head),
		tail: unsafe.Pointer(tail),
	}
}

// region Node
func NewItemNode(valueptr *Item) *node {
	return &node{
		valueptr: unsafe.Pointer(valueptr),
		key:      getKeyHash(valueptr),
	}
}

type node struct {
	key      uint64
	valueptr unsafe.Pointer // TODO: use uintptr
	next     unsafe.Pointer // What if GC runs?
}

func (n *node) Next() *node {
	return (*node)(atomic.LoadPointer(&n.next))
}

func (n *node) GetItem() *Item {
	return (*Item)(atomic.LoadPointer(&n.valueptr))
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

func (l *linkedlist) Head() *node {
	return (*node)(atomic.LoadPointer(&l.head))
}

func (l *linkedlist) Tail() *node {
	return (*node)(atomic.LoadPointer(&l.tail))
}

func (l *linkedlist) Insert(v *node) (bool, error) {
	var left, right *node
	for {
		left, right = l.search(v.key)

		if right != l.Tail() && right.key == v.key {
			return false, inserterr
		}

		v.next = unsafe.Pointer(right)
		if atomic.CompareAndSwapPointer(&left.next, unsafe.Pointer(right), unsafe.Pointer(v)) {
			atomic.AddInt32(&l.len, 1)
			return true, nil
		}
	}
}

func (l *linkedlist) Upsert(v *node) (bool, error) {
	_, right := l.search(v.key)
	if right != l.Tail() && right.key == v.key {
		atomic.StorePointer(&right.valueptr, v.valueptr)
		return true, nil
	}

	return l.Insert(v)
}

func (l *linkedlist) Delete(v *node) (bool, error) {
	var right *node
	var rightnext unsafe.Pointer

	for {
		_, right = l.search(v.key)
		if right == l.Tail() || right.key != v.key {
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

	// if !atomic.CompareAndSwapPointer(&left.next, unsafe.Pointer(right), rightnext) {
	//   _, _ = l.search(right.key) // cleanup
	// }

	return true, nil
}

// search returns left and right such
// - left and right are unmarked node
// - right is immediate successor of left
// - left.key < key <= right.key
func (l *linkedlist) search(key uint64) (left, right *node) {
	var leftnext *node

	for {
		curr := l.Head()
		next := curr.nextptr()

		// find left and right node
		for {
			// if current node is not logically deleting
			if !deletemarked(next) {
				left = curr
				leftnext = (*node)(next)
			}

			curr = (*node)(deleteunmark(next))

			if curr == l.Tail() {
				break
			}

			next = curr.nextptr()
			if !deletemarked(next) && curr.key >= key {
				break
			}
		}

		// check if left right are adjacent
		right = curr
		if leftnext == right {
			if right != l.Tail() && deletemarked(right.nextptr()) {
				continue
			}

			return left, right
		}

		if atomic.CompareAndSwapPointer(&left.next, unsafe.Pointer(leftnext), unsafe.Pointer(right)) {
			if right != l.Tail() && deletemarked(right.nextptr()) {
				continue
			}

			return left, right
		}
	}
}

func (l *linkedlist) Contains(n *node) bool {
	_, right := l.search(n.key)
	if right == l.Tail() || right.key != n.key {
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
	curr *node
	list *linkedlist
}

func NewIterator(list *linkedlist) *iterator {
	return &iterator{
		list: list,
		curr: list.Head(),
	}
}

func (it *iterator) Next() (*Item, bool) {
	n, ok := it.nextnode()
	if !ok {
		return nil, ok
	}

	return n.GetItem(), ok
}

func (it *iterator) nextnode() (*node, bool) {
	if it.curr == it.list.Head() {
		it.curr = it.curr.Next()
	}

	for {
		if it.curr == it.list.Tail() {
			return nil, false
		}

		n, nextptr := it.curr, it.curr.nextptr()
		it.curr = (*node)(deleteunmark(nextptr))

		if deletemarked(nextptr) {
			continue
		}

		return n, true
	}
}

// endregion

// region CyclicIterator
type cycliciterator struct {
	iterator
}

func NewCyclicIterator(list *linkedlist) *cycliciterator {
	return &cycliciterator{
		*NewIterator(list),
	}
}

func (it *cycliciterator) Next() (*Item, bool) {
	v, ok := it.iterator.Next()
	if !ok {
		it.curr = it.list.Head()
		return it.iterator.Next()
	}

	return v, ok
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

func getKeyHash(v interface{}) uint64 {
	if isnil(v) {
		panic("v cannot be nil")
	}

	if itf, ok := v.(identify.Identify); ok {
		return itf.ID()
	}

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
	default:
		panic(fmt.Errorf("unsupported v type %T", v))
	}
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

func isnil(itf interface{}) bool {
	return itf == nil || (reflect.ItemOf(itf).Kind() == reflect.Ptr && reflect.ItemOf(itf).IsNil())
}