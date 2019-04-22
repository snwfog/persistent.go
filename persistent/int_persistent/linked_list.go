// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/cheekybits/genny

package int_persistent

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync/atomic"
	"unsafe"

	"github.com/dchest/siphash"
)

var (
	inserterr = errors.New("insert failed")
	// deleteerr = errors.New("delete failed")
	// freenode = unsafe.Pointer(new(int))
)

type ()

func NewIntLinkedList() *LinkedList {
	// head := &sentinel{IntNode{valueptr: -2}}
	// tail := &sentinel{IntNode{valueptr: -1}}
	head := &sentinel{IntNode{}}
	tail := &sentinel{IntNode{}}

	head.next = unsafe.Pointer(tail)
	return &LinkedList{
		head: unsafe.Pointer(head),
		tail: unsafe.Pointer(tail),
	}
}

// region Node
func NewIntNode(key interface{}, valueptr unsafe.Pointer) *IntNode {
	return &IntNode{
		valueptr: valueptr,
		key:      getKeyHash(key),
	}
}

func NewBuiltinIntNode(value int) *IntNode {
	return NewIntNode(value, unsafe.Pointer(&value))
}

type IntNode struct {
	key      uint64
	valueptr unsafe.Pointer
	next     unsafe.Pointer // What if GC runs?
}

func (n *IntNode) Next() *IntNode {
	return (*IntNode)(atomic.LoadPointer(&n.next))
}

func (n *IntNode) GetInt() *int {
	return (*int)(atomic.LoadPointer(&n.valueptr))
}

func (n *IntNode) GetBuiltinInt() int {
	return *(n.GetInt())
}

func (n *IntNode) nextptr() unsafe.Pointer {
	return atomic.LoadPointer(&n.next)
}

type sentinel struct {
	IntNode
}

// endregion

// region LinkedList
type LinkedList struct {
	head unsafe.Pointer
	tail unsafe.Pointer

	len int32
}

func (l *LinkedList) Len() int {
	return int(atomic.LoadInt32(&l.len))
}

func (l *LinkedList) Head() *IntNode {
	return (*IntNode)(atomic.LoadPointer(&l.head))
}

func (l *LinkedList) Tail() *IntNode {
	return (*IntNode)(atomic.LoadPointer(&l.tail))
}

func (l *LinkedList) Insert(v *IntNode) (bool, error) {
	var left, right *IntNode
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

func (l *LinkedList) Upsert(v *IntNode) (bool, error) {
	_, right := l.search(v.key)
	if right != l.Tail() && right.key == v.key {
		atomic.StorePointer(&right.valueptr, v.valueptr)
		return true, nil
	}

	return l.Insert(v)
}

func (l *LinkedList) Delete(v *IntNode) (bool, error) {
	var right *IntNode
	var rightnext unsafe.Pointer

	for {
		_, right = l.search(v.key)
		if right == l.Tail() || right.key != v.key {
			return false, nil // not deleted cause not found
		}

		rightnext = right.nextptr()
		if !marked(rightnext) {
			if atomic.CompareAndSwapPointer(&right.next, rightnext, mark(rightnext)) {
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

func (l *LinkedList) search(key uint64) (left, right *IntNode) {
	var leftnext *IntNode

	for {
		prev := l.Head()
		currptr := prev.nextptr()

		for {
			if !marked(currptr) {
				left = prev
				leftnext = (*IntNode)(currptr)
			}

			prev = (*IntNode)(unmark(currptr))
			if prev == l.Tail() {
				break
			}

			currptr = prev.nextptr()
			if !marked(currptr) && prev.key >= key {
				break
			}
		}

		right = prev
		if leftnext == right {
			if right != l.Tail() && marked(right.nextptr()) {
				continue
			}

			return left, right
		}

		if atomic.CompareAndSwapPointer(&left.next, unsafe.Pointer(leftnext), unsafe.Pointer(right)) {
			if right != l.Tail() && marked(right.nextptr()) {
				continue
			}

			return left, right
		}
	}
}

func (l *LinkedList) Contains(n *IntNode) bool {
	_, right := l.search(n.key)
	if right == l.Tail() || right.key != n.key {
		return false
	}

	return true
}

func (l *LinkedList) Iterator() *iterator {
	return NewIterator(l)
}

func (l *LinkedList) CyclicIterator() *cyclicIterator {
	return NewCyclicIterator(l)
}

// endregion

// region Iterators
type iterator struct {
	curr *IntNode
	list *LinkedList
}

func NewIterator(list *LinkedList) *iterator {
	return &iterator{
		list: list,
		curr: list.Head(),
	}
}

func (it *iterator) Next() (*IntNode, bool) {
	if it.curr == it.list.Head() {
		it.curr = it.curr.Next()
	}

	for {
		if it.curr == it.list.Tail() {
			return nil, false
		}

		n, nextptr := it.curr, it.curr.nextptr()
		it.curr = (*IntNode)(unmark(nextptr))

		if marked(nextptr) {
			continue
		}

		return n, true
	}
}

// endregion

// region CyclicIterator
type cyclicIterator struct {
	iterator
}

func NewCyclicIterator(list *LinkedList) *cyclicIterator {
	return &cyclicIterator{
		*NewIterator(list),
	}
}

func (it *cyclicIterator) Next() (*IntNode, bool) {
	node, ok := it.iterator.Next()

	if !ok {
		it.curr = it.list.Head()
		return it.Next()
	}

	return node, ok
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

func getKeyHash(key interface{}) uint64 {
	switch x := key.(type) {
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
	panic(fmt.Errorf("unsupported key type %T", key))
}

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

func use(params ...interface{}) {
	for _, val := range params {
		_ = val
	}
}

func marked(ptr unsafe.Pointer) bool {
	return (uintptr(ptr) & 0x1) > 0
}

func mark(ptr unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(uintptr(ptr) | 0x1)
}

func unmark(ptr unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(uintptr(ptr) &^ 0x1)
}
