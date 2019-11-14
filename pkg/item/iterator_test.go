package item

import (
	"context"
	"sort"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"
)

func TestIterator1(t *testing.T) {
	ll := NewItemLinkedList()
	_, _ = ll.Insert(&Item{Id: 1})
	_, _ = ll.Insert(&Item{Id: 2})
	_, _ = ll.Insert(&Item{Id: 3})

	it := ll.Iterator()
	for n, ok := it.Next(); ok; n, ok = it.Next() {
		t.Log(*n)
	}

	it = ll.Iterator()
	n, ok := it.Next()
	assert.True(t, ok)
	assert.Equal(t, 1, n.Id)

	n, ok = it.Next()
	assert.True(t, ok)
	assert.Equal(t, 2, n.Id)

	n, ok = it.Next()
	assert.True(t, ok)
	assert.Equal(t, 3, n.Id)
}

func TestIterator2(t *testing.T) {
	ll := NewItemLinkedList()
	_, _ = ll.Insert(&Item{Id: 1})
	_, _ = ll.Insert(&Item{Id: 2})
	_, _ = ll.Insert(&Item{Id: 3})

	it := ll.Iterator()

	for i := 0; i < 3; i++ {
		n, ok := it.Next()
		assert.True(t, ok)
		assert.Equal(t, i+1, n.Id)
	}

	node, ok := it.Next()
	assert.False(t, ok)
	assert.Nil(t, node)

	node, ok = it.Next()
	assert.False(t, ok)
	assert.Nil(t, node)
}

func TestCyclicIterator1(t *testing.T) {
	dll := NewItemLinkedList()

	for i := 0; i < 3; i++ {
		item := &Item{Id: i, AccessCount: atomic.NewInt64(0)}
		_, _ = dll.Insert(item)
	}

	it := dll.CyclicIterator()

	n1, ok := it.Next()
	assert.True(t, ok)
	assert.Equal(t, 0, n1.Id)

	n2, ok := it.Next()
	assert.True(t, ok)
	assert.Equal(t, 1, n2.Id)

	n3, ok := it.Next()
	assert.True(t, ok)
	assert.Equal(t, 2, n3.Id)

	n11, ok := it.Next()
	assert.True(t, ok)
	assert.Equal(t, n1, n11)
	assert.Equal(t, 0, n11.Id)

	n22, ok := it.Next()
	assert.True(t, ok)
	assert.Equal(t, n2, n22)
	assert.Equal(t, 1, n22.Id)

	n33, ok := it.Next()
	assert.True(t, ok)
	assert.Equal(t, n3, n33)
	assert.Equal(t, 2, n33.Id)

	n111, ok := it.Next()
	assert.True(t, ok)
	assert.Equal(t, n1, n111)
	assert.Equal(t, 0, n111.Id)
}

func TestConcurrentIterator(t *testing.T) {
	ll := NewItemLinkedList()
	const M = 1000   // Number of workers
	const N = 100000 // Number of items

	items := make([]*Item, 0, N)
	for i := 0; i < N; i++ {
		items = append(items, &Item{Id: i})
	}

	for i := 0; i < N; i++ {
		_, _ = ll.Insert(items[i])
	}

	for i := 0; i < N; i++ {
		assert.True(t, ll.Contains(items[i]))
	}
	assert.Equal(t, N, ll.Len())

	var mu sync.Mutex
	ids := make([]int, 0, N)
	it := ll.Iterator()
	g, _ := errgroup.WithContext(context.Background())
	var worker [M]int
	for range worker {
		g.Go(func() error {
			for i := 0; i < N/M; i++ {
				item, ok := it.Next()
				if !ok {
					return nil
				}

				mu.Lock()
				ids = append(ids, item.Id)
				mu.Unlock()
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		t.Error(err)
	}

	assert.Len(t, ids, N)
	sort.Ints(ids)
	equals := true
	for i := 0; i < N; i++ {
		equals = equals && i == ids[i]
	}
}

func TestConcurrentCyclicIterator(t *testing.T) {
	list := NewItemLinkedList()
	cit := list.CyclicIterator()

	for i := 0; i < 10; i++ {
		item := &Item{Id: i, AccessCount: atomic.NewInt64(0)}
		_, _ = list.Insert(item)
	}

	runs := int64(10 * 1000 * 1000)
	N := atomic.NewInt64(runs)
	const M = 10
	g, _ := errgroup.WithContext(context.Background())

	var worker [M]int // Do cit with 10 workers
	for range worker {
		g.Go(func() error {
			for N.Load() > 0 {
				item, ok := cit.Next()
				// If we cannot retrieve an item, fail now; we should always return an item
				if !ok {
					t.FailNow()
				}

				if item == nil {
					panic("item nil")
				}

				item.AccessCount.Inc()
				N.Sub(1)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		t.FailNow()
	}

	e := 0.000001
	it := list.Iterator()
	for n, ok := it.Next(); ok; n, ok = it.Next() {
		percentusage := float64(n.AccessCount.Load()) / float64(runs)
		t.Log(percentusage)
		assert.True(t, percentusage >= (1/float64(M) - e))
		assert.True(t, percentusage <= (1/float64(M) + e))
	}
}
