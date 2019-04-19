package persistent

import (
  "sync/atomic"
  "unsafe"

  "github.com/pkg/errors"
)

var (
  inserterr = errors.New("insert failed")
  deleteerr = errors.New("delete failed")
)

func NewLinkedList() *LinkedList {
  head := &sentinel{}
  tail := &sentinel{}
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
  key     uint64
  value   int
  deleted int32
  next    unsafe.Pointer
}

func (n *node) Next() *node {
  return (*node)(atomic.LoadPointer(&n.next))
}

func (n *node) marked() bool {
  return atomic.LoadInt32(&n.deleted) == 1
}

func (n *node) mark() {
  atomic.StoreInt32(&n.deleted, 0)
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
  var right, rightnext, left *node

  for {
    left, right = l.search(el.key)
    if right == l.Tail() || right.key != el.key {
      return false, nil // not deleted cause not found
    }

    rightnext = right.Next()
    if !rightnext.marked() {
      rightnext.mark()
      atomic.AddInt32(&l.len, -1)
      break
    }
  }

  if !atomic.CompareAndSwapPointer(&left.next, unsafe.Pointer(right), unsafe.Pointer(rightnext)) {
    _, _ = l.search(right.key) // cleanup
  }

  return true, nil
}

func (l *LinkedList) search(key uint64) (left, right *node) {
  var leftnext *node

  for {
    prev := l.Head()
    curr := prev.Next()

    for {
      if !curr.marked() {
        left = prev
        leftnext = curr
      }

      prev = curr
      if prev == l.Tail() {
        break
      }

      curr = curr.Next()
      if !curr.marked() && prev.key >= key {
        break
      }
    }

    right = prev
    if leftnext == right {
      if right != l.Tail() && right.Next().marked() {
        continue
      }

      return left, right
    }

    if atomic.CompareAndSwapPointer(&left.next, unsafe.Pointer(leftnext), unsafe.Pointer(right)) {
      if right != l.Tail() && right.Next().marked() {
        continue
      }

      return left, right
    }
  }
}

func (l *LinkedList) Contains(n *node) bool {
  head := l.Head()

  if head == nil {
    return false
  }

  _, right := l.search(n.key)
  return right != nil
}

func (l *LinkedList) Iterator() *iterator {
  return &iterator{
    list: l,
    curr: l.Head(),
  }
}

// endregion
