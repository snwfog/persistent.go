package persistent

// region Iterator
type iterator struct {
  curr *node
  list *LinkedList
}

func NewIterator(list *LinkedList) *iterator {
  return &iterator{
    list: list,
    curr: list.Head(),
  }
}

func (it *iterator) Next() (*node, bool) {
  if it.curr == it.list.Head() {
    it.curr = it.curr.Next()
  }

  for {
    if it.curr == it.list.Tail() {
      return nil, false
    }

    n, nextptr := it.curr, it.curr.nextptr()
    it.curr = (*node)(unmark(nextptr))

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

func (it *cyclicIterator) Next() (*node, bool) {
  node, ok := it.iterator.Next()

  if !ok {
    it.curr = it.list.Head()
    return it.Next()
  }

  return node, ok
}

// endregion
