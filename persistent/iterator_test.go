package persistent

import (
  "testing"

  "github.com/stretchr/testify/assert"
)

func TestIterator1(t *testing.T) {
  dll := NewLinkedList()

  _, _ = dll.Insert(NewNode(1))
  _, _ = dll.Insert(NewNode(2))
  _, _ = dll.Insert(NewNode(3))

  it := dll.Iterator()

  _, ok := it.Next()
  assert.True(t, ok)

  _, ok = it.Next()
  assert.True(t, ok)

  _, ok = it.Next()
  assert.True(t, ok)
}

func TestIterator2(t *testing.T) {
  dll := NewLinkedList()
  _, _ = dll.Insert(NewNode(1))
  _, _ = dll.Insert(NewNode(2))
  _, _ = dll.Insert(NewNode(3))

  it := dll.Iterator()

  for i := 0; i < 3; i++ {
    _, ok := it.Next()
    assert.True(t, ok)
  }

  node, ok := it.Next()
  assert.False(t, ok)
  assert.Nil(t, node)

  node, ok = it.Next()
  assert.False(t, ok)
  assert.Nil(t, node)
}

func TestCyclicIterator1(t *testing.T) {
  dll := NewLinkedList()

  _, _ = dll.Insert(NewNode(1))
  _, _ = dll.Insert(NewNode(2))
  _, _ = dll.Insert(NewNode(3))

  it := dll.CyclicIterator()

  n1, ok := it.Next()
  assert.True(t, ok)

  n2, ok := it.Next()
  assert.True(t, ok)

  n3, ok := it.Next()
  assert.True(t, ok)

  n11, ok := it.Next()
  assert.True(t, ok)
  assert.Equal(t, n1, n11)

  n22, ok := it.Next()
  assert.True(t, ok)
  assert.Equal(t, n2, n22)

  n33, ok := it.Next()
  assert.True(t, ok)
  assert.Equal(t, n3, n33)

  n111, ok := it.Next()
  assert.True(t, ok)
  assert.Equal(t, n1, n111)
}
