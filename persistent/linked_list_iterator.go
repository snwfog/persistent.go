package persistent

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
