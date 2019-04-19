package persistent

import (
  "sync/atomic"
  "unsafe"

  "github.com/pkg/errors"
)

var (
  appenderr  = errors.New("append failed")
  prependerr = errors.New("prepend failed")
  deleteerr  = errors.New("delete failed")
)

func NewLinkedList() *LinkedList {
  // g := &node{}
  // g.next = unsafe.Pointer(g)
  // g.prev = unsafe.Pointer(g)
  head := &sentinel{}
  tail := &sentinel{}
  head.next = tail
  return &LinkedList{
    head: head,
    tail: tail,
  }
}

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

func (l *LinkedList) TailN() *node {
  // return (*node)(atomic.LoadPointer(&l.tail))
  // return l.Head().Prev()
  n := l.Head()

  if n == nil {
    return nil
  }

  prev := n
  for n := n.Next(); n != nil; prev, n = n, n.Next() {
    // Nothing
  }

  return prev
}

func (l *LinkedList) h() *node {
  return (*node)(atomic.LoadPointer(&l.head))
}

func (l *LinkedList) t() *node {
  return (*node)(atomic.LoadPointer(&l.tail))
}

type node struct {
  key     uint64
  value   int
  deleted int32
  next    unsafe.Pointer
  // prev  unsafe.Pointer
}

type sentinel struct {
  node
}

func NewNode(id int) *node {
  return &node{
    value: id,
    key:   getKeyHash(id),
  }
}

func (n *node) Next() *node {
  return (*node)(atomic.LoadPointer(&n.next))
}

// func (n *node) Prev() *node {
//   return (*node)(atomic.LoadPointer(&n.prev))
// }

func (l *LinkedList) Prepend(el *node) (bool, error) {
  head := l.Head()
  if head != nil {
    el.next = unsafe.Pointer(head)
  }

  if !atomic.CompareAndSwapPointer(&l.head, unsafe.Pointer(head), unsafe.Pointer(el)) {
    return false, prependerr
  }

  atomic.AddInt32(&l.len, 1)
  return true, nil
}

func (l *LinkedList) Append(el *node) (bool, error) {
  // WARN: This is questionable, caller should handle nil
  // if el == nil {
  //   return true, nil
  // }

  // tail, head := atomic.LoadPointer(&l.tail), atomic.LoadPointer(&l.head)
  // head, tail := l.Head(), l.TailN()
  head := l.Head()
  if head == nil { // First el
    // el.prev, el.next = unsafe.Pointer(el), unsafe.Pointer(el)

    // l.head, l.tail = n, n
    if !atomic.CompareAndSwapPointer(&l.head, unsafe.Pointer(head), unsafe.Pointer(el)) {
      return false, appenderr
    }

    // if !atomic.CompareAndSwapPointer(&l.tail, unsafe.Pointer(tail), unsafe.Pointer(el)) {
    //   return false, appenderr
    // }

    atomic.AddInt32(&l.len, 1)
    return true, nil
  }

  // dt1 := l.TailN()
  // dt2 := l.TailN().Next()
  // use(dt1, dt2)

  tail := l.TailN()
  if tail == nil {
    return false, appenderr
  }

  if ok := insertAt(el, tail, tail.Next()); !ok {
    return false, appenderr
  }

  // l.tail = n
  // if !atomic.CompareAndSwapPointer(&l.tail, unsafe.Pointer(tail), unsafe.Pointer(el)) {
  //   return false, appenderr
  // }

  atomic.AddInt32(&l.len, 1)
  return true, nil
}

func (l *LinkedList) Delete(el *node) (*node, error) {
  // WARN: This is questionable, caller should handle nil
  // if el == nil {
  //   return true, nil
  // }

  head := l.Head()
  if head == nil {
    return nil, deleteerr
  }

  prev, node, ok := l.find(head, el.key)

  if !ok {
    return nil, nil // Node not found, but no error either
  }

  // Removing head
  if prev == nil {
    if !atomic.CompareAndSwapPointer(&l.head, unsafe.Pointer(head), unsafe.Pointer(head.Next())) {
      return nil, deleteerr
    }

    atomic.AddInt32(&l.len, -1)
    return head, nil
  }

  // prev.next = dest.next
  if !atomic.CompareAndSwapPointer(&prev.next, unsafe.Pointer(node), unsafe.Pointer(node.Next())) {
    return nil, deleteerr
  }
  // if ok := deleteAt(node, prev); !ok {
  //   return nil, deleteerr
  // }

  // if !atomic.CompareAndSwapPointer(&l.head, unsafe.Pointer(head), unsafe.Pointer(el.Next())) {
  //   return false, deleteerr
  // }

  atomic.AddInt32(&l.len, -1)
  return node, nil
}

func (l *LinkedList) find(head *node, key uint64) (prev, dest *node, ok bool) {
  if head == nil {
    return nil, nil, false
  }

  // Head is value
  if head.key == key {
    return nil, head, true
  }

  for prev, n := head, head.Next(); n != nil; prev, n = n, n.Next() {
    if n.key == key {
      return prev, n, true
    }
  }
  return nil, nil, false
}

func (n *node) marked() bool {
  return atomic.LoadInt32(&n.deleted) == 1
}

func (l *LinkedList) search(key uint64) (left, right *node) {
  var leftnext *node

  for {
    prev := l.h()
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

    right = curr
    if leftnext == right {
      
    }
  }
}

func marked(n *node) bool {
  return atomic.LoadInt32(&n.deleted) == 1
}

func (l *LinkedList) Contains(n *node) bool {
  head := l.Head()

  if head == nil {
    return false
  }

  _, _, ok := l.find(head, n.key)
  return ok
}

func (l *LinkedList) Iterator() *iterator {
  return &iterator{
    list: l,
    curr: l.Head(),
  }
}

// func (l *LinkedList) Head() *node {
//   return (*node)(atomic.LoadPointer(&l.head))
// }

func insertAt(n, prev, dest *node) bool {
  // prev.next, n.prev = n, prev
  // n.next, dest.prev = dest, n
  // n.prev = unsafe.Pointer(prev)
  n.next = unsafe.Pointer(dest)

  // nprev := n.Prev()
  // nnext := n.Next()
  // use(nprev, nnext)

  if !atomic.CompareAndSwapPointer(&prev.next, unsafe.Pointer(dest), unsafe.Pointer(n)) {
    return false
  }

  // if !atomic.CompareAndSwapPointer(&dest.prev, unsafe.Pointer(prev), unsafe.Pointer(n)) {
  //   return false
  // }

  // pnext := prev.Next()
  // nprev2 := dest.Prev()
  // use(pnext, nprev2)

  return true
}

func deleteAt(dest, prev *node) bool {
  // prev, next := dest.Prev(), dest.Next()

  // dest.next.prev = prev
  // if !atomic.CompareAndSwapPointer(&next.prev, unsafe.Pointer(dest), unsafe.Pointer(dest.Prev())) {
  //   return false
  // }

  // prev.next = dest.next
  if !atomic.CompareAndSwapPointer(&prev.next, unsafe.Pointer(dest), unsafe.Pointer(dest.Next())) {
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
  list *LinkedList
}

func NewIterator(list *LinkedList) *iterator {
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

// region CyclicIterator
// type cyclicIterator struct {
//   it *iterator
// }
//
// func NewCyclicIterator
// endregion
