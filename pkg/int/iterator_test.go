package int

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/snwfog/persistent.go/pkg"
)

func TestIterator1(t *testing.T) {
	ll := pkg.NewIntLinkedList()

	a, b, c := 1, 2, 3
	_, _ = ll.Insert(pkg.NewIntNode(&a))
	_, _ = ll.Insert(pkg.NewIntNode(&b))
	_, _ = ll.Insert(pkg.NewIntNode(&c))

	it := ll.Iterator()
	for n, ok := it.Next(); ok; n, ok = it.Next() {
		t.Log(*n)
	}

	it = ll.Iterator()
	var n *int
	n, ok := it.Next()
	assert.True(t, ok)
	assert.Equal(t, 1, *n)

	n, ok = it.Next()
	assert.True(t, ok)
	assert.Equal(t, 2, n)

	n, ok = it.Next()
	assert.True(t, ok)
	assert.Equal(t, 3, n)
}

func TestIterator2(t *testing.T) {
	ll := pkg.NewIntLinkedList()
	a, b, c := 1, 2, 3
	_, _ = ll.Insert(pkg.NewIntNode(&a))
	_, _ = ll.Insert(pkg.NewIntNode(&b))
	_, _ = ll.Insert(pkg.NewIntNode(&c))

	it := ll.Iterator()

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
	ll := pkg.NewIntLinkedList()

	a, b, c := 1, 2, 3
	_, _ = ll.Insert(pkg.NewIntNode(&a))
	_, _ = ll.Insert(pkg.NewIntNode(&b))
	_, _ = ll.Insert(pkg.NewIntNode(&c))

	it := ll.CyclicIterator()

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
