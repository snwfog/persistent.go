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

// Cyclic doubly linked list
func NewDoublyLinkedList() *DoublyLinkedList {
  return &DoublyLinkedList{}
}

type DoublyLinkedList struct {
  head unsafe.Pointer
  tail unsafe.Pointer
  len  int32
}

func (l *DoublyLinkedList) Len() int {
  return int(atomic.LoadInt32(&l.len))
}

func (l *DoublyLinkedList) Head() *node {
  return (*node)(atomic.LoadPointer(&l.head))
}

func (l *DoublyLinkedList) Tail() *node {
  return (*node)(atomic.LoadPointer(&l.tail))
}

type node struct {
  value int
  next  unsafe.Pointer
  prev  unsafe.Pointer
}

func NewNode(id int) *node {
  return &node{value: id}
}

func (n *node) Next() *node {
  return (*node)(atomic.LoadPointer(&n.next))
}

func (n *node) Prev() *node {
  return (*node)(atomic.LoadPointer(&n.prev))
}

func (l *DoublyLinkedList) Append(el *node) (bool, error) {
  // tail, head := atomic.LoadPointer(&l.tail), atomic.LoadPointer(&l.head)
  head, tail := l.Head(), l.Tail()

  if head == nil && tail == head { // First el
    el.prev, el.next = unsafe.Pointer(el), unsafe.Pointer(el)

    // l.head, l.tail = n, n
    if !atomic.CompareAndSwapPointer(&l.head, unsafe.Pointer(head), unsafe.Pointer(el)) {
      return false, errors.New("retry")
    }

    if !atomic.CompareAndSwapPointer(&l.tail, unsafe.Pointer(tail), unsafe.Pointer(el)) {
      return false, errors.New("retry")
    }

    atomic.AddInt32(&l.len, 1)
    return true, nil
  }

  dt1 := l.Tail()
  dt2 := l.Tail().Next()
  use(dt1, dt2)

  if ok := insertAt(el, head); !ok {
    return false, errors.New("insert failed")
  }

  // l.tail = n
  if !atomic.CompareAndSwapPointer(&l.tail, unsafe.Pointer(tail), unsafe.Pointer(el)) {
    return false, inserterr
  }

  atomic.AddInt32(&l.len, 1)
  return true, nil
}

func (l *DoublyLinkedList) Delete(el *node) (bool, error) {
  if l.Head() == nil {
    return true, nil
  }

  for n, i := l.Head(), 0; i < l.Len(); i++ {
    if n.value == el.value {
      if ok := deleteAt(n); !ok {
        return false, deleteerr
      }
      atomic.AddInt32(&l.len, -1)
      return true, nil
    }

    n = n.Next()
  }

  return false, errors.New("value does not exists")
}

func (l *DoublyLinkedList) Contains(n *node) bool {
  head := l.Head()

  if head == nil {
    return false
  }

  for n, i := l.Head(), 0; i < l.Len(); i++ {
    // This case should not happen
    // if n == nil {
    // }
    if n.value == n.value {
      return true
    }

    n = n.Next()
  }

  return false
}

func (l *DoublyLinkedList) Iterator() *iterator {
  return &iterator{
    list: l,
    curr: l.Head(),
  }
}

func insertAt(n *node, dest *node) bool {
  prev := dest.Prev()
  // prev.next, n.prev = n, prev
  // n.next, dest.prev = dest, n
  n.prev, n.next = unsafe.Pointer(prev), unsafe.Pointer(dest)

  nprev := n.Prev()
  nnext := n.Next()
  use(nprev, nnext)

  if !atomic.CompareAndSwapPointer(&prev.next, unsafe.Pointer(dest), unsafe.Pointer(n)) {
    return false
  }

  pnext := prev.Next()

  if !atomic.CompareAndSwapPointer(&dest.prev, unsafe.Pointer(prev), unsafe.Pointer(n)) {
    return false
  }

  nprev2 := dest.Prev()
  use(pnext, nprev2)

  return true
}

func deleteAt(dest *node) bool {
  prev := dest.Prev()
  next := dest.Next()

  // prev.next = dest.next
  if !atomic.CompareAndSwapPointer(&prev.next, unsafe.Pointer(dest), unsafe.Pointer(dest.Next())) {
    return false
  }

  // dest.next.prev = prev
  if !atomic.CompareAndSwapPointer(&next.prev, unsafe.Pointer(dest), unsafe.Pointer(dest.Prev())) {
    return false
  }

  return true
}

func use(params ...interface{}) {
  for _, val := range params {
    _ = val
  }
}

// region Iterator
// WARN: NOT CONCURRENT SAFE!!
type iterator struct {
  curr *node
  pos  int
  list *DoublyLinkedList
}

func NewIterator(list *DoublyLinkedList) *iterator {
  return &iterator{
    list: list,
    curr: list.Head(),
  }
}

func (it *iterator) Next() (*node, bool) {
  if it.pos >= it.list.Len() {
    return nil, false
  }

  if it.curr == nil {
    return it.curr, false
  }

  n := it.curr
  it.pos = it.pos + 1
  it.curr = it.curr.Next()

  return n, true
}

// endregion
