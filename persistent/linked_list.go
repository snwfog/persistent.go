package persistent

import (
  "sync/atomic"
  "unsafe"

  "github.com/pkg/errors"
)

var (
  inserterr = errors.New("insert failed")
  // deleteerr = errors.New("delete failed")
  // freenode = unsafe.Pointer(new(int))
)

func NewLinkedList() *LinkedList {
  head := &sentinel{node{value: -2}}
  tail := &sentinel{node{value: -1}}

  head.next = unsafe.Pointer(tail)
  return &LinkedList{
    head: unsafe.Pointer(head),
    tail: unsafe.Pointer(tail),
  }
}

// region Node
func NewNode(id int) *node {
  return &node{
    value: id,
    key:   getKeyHash(id),
  }
}

type node struct {
  key   uint64
  value int
  next  unsafe.Pointer // What if GC runs?
}

func (n *node) Next() *node {
  return (*node)(atomic.LoadPointer(&n.next))
}

func (n *node) nextptr() unsafe.Pointer {
  return atomic.LoadPointer(&n.next)
}

type sentinel struct {
  node
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

func (l *LinkedList) Head() *node {
  return (*node)(atomic.LoadPointer(&l.head))
}

func (l *LinkedList) Tail() *node {
  return (*node)(atomic.LoadPointer(&l.tail))
}

func (l *LinkedList) Insert(el *node) (bool, error) {
  var left, right *node
  for {
    left, right = l.search(el.key)

    if right != l.Tail() && right.key == el.key {
      return false, inserterr
    }

    el.next = unsafe.Pointer(right)
    if atomic.CompareAndSwapPointer(&left.next, unsafe.Pointer(right), unsafe.Pointer(el)) {
      atomic.AddInt32(&l.len, 1)
      return true, nil
    }
  }
}

func (l *LinkedList) Delete(el *node) (bool, error) {
  var right *node
  var rightnext unsafe.Pointer

  for {
    _, right = l.search(el.key)
    if right == l.Tail() || right.key != el.key {
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

func (l *LinkedList) search(key uint64) (left, right *node) {
  var leftnext *node

  for {
    prev := l.Head()
    currptr := prev.nextptr()

    for {
      if !marked(currptr) {
        left = prev
        leftnext = (*node)(currptr)
      }

      prev = (*node)(unmark(currptr))
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

func (l *LinkedList) Contains(n *node) bool {
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
