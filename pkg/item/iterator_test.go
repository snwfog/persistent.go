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
	for n, err := it.Next(); err == nil; n, err = it.Next() {
		t.Log(*n)
	}

	it = ll.Iterator()
	n, err := it.Next()
	assert.NoError(t, err)
	assert.Equal(t, 1, n.Id)

	n, err = it.Next()
	assert.NoError(t, err)
	assert.Equal(t, 2, n.Id)

	n, err = it.Next()
	assert.NoError(t, err)
	assert.Equal(t, 3, n.Id)
}

func TestIterator2(t *testing.T) {
	ll := NewItemLinkedList()
	_, _ = ll.Insert(&Item{Id: 1})
	_, _ = ll.Insert(&Item{Id: 2})
	_, _ = ll.Insert(&Item{Id: 3})

	it := ll.Iterator()

	for i := 0; i < 3; i++ {
		n, err := it.Next()
		assert.NoError(t, err)
		assert.Equal(t, i+1, n.Id)
	}

	node, err := it.Next()
	assert.Error(t, err)
	assert.Nil(t, node)

	node, err = it.Next()
	assert.Error(t, err)
	assert.Nil(t, node)
}

func TestIteratorCyclic1(t *testing.T) {
	ll := NewItemLinkedList()

	for i := 0; i < 3; i++ {
		item := &Item{Id: i, AccessCount: atomic.NewInt64(0)}
		_, _ = ll.Insert(item)
	}

	it := ll.CyclicIterator()

	n1, err := it.Next()
	assert.NoError(t, err)
	assert.Equal(t, 0, n1.Id)

	n2, err := it.Next()
	assert.NoError(t, err)
	assert.Equal(t, 1, n2.Id)

	n3, err := it.Next()
	assert.NoError(t, err)
	assert.Equal(t, 2, n3.Id)

	n11, err := it.Next()
	assert.NoError(t, err)
	assert.Equal(t, n1, n11)
	assert.Equal(t, 0, n11.Id)

	n22, err := it.Next()
	assert.NoError(t, err)
	assert.Equal(t, n2, n22)
	assert.Equal(t, 1, n22.Id)

	n33, err := it.Next()
	assert.NoError(t, err)
	assert.Equal(t, n3, n33)
	assert.Equal(t, 2, n33.Id)

	n111, err := it.Next()
	assert.NoError(t, err)
	assert.Equal(t, n1, n111)
	assert.Equal(t, 0, n111.Id)
}

// Create N items, and insert them into ll
// Check all N items are in ll
// Create one iterator, and create M workers
// Concurrently access iterator, by calling `Next`
// and collect ids of item, assert that we were able to get
// all items from the list into the slices
func TestIteratorConcurrent(t *testing.T) {
	ll := NewItemLinkedList()
	const M = 10000 // Number of workers
	const N = 10000 // Number of items

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

	// Extra, to test randomly remove element,
	// enable this rogue channel that would delete elements
	// and cause the test case to fail
	rogue := make(chan struct{})
	// go func() {
	// L:
	// 	for {
	// 		select {
	// 		case <-time.Tick(time.Microsecond):
	// 			_, err := ll.Delete(items[rand.Intn(ll.Len())])
	// 			assert.NoError(t, err)
	// 		case <-rogue:
	// 			t.Log("rogue done")
	// 			break L
	// 		}
	// 	}
	// }()

	var mu sync.Mutex
	ids := make([]int, 0, N)
	it := ll.Iterator()
	g, _ := errgroup.WithContext(context.Background())
	var worker [M]int
	for range worker {
		g.Go(func() error {
			for i := 0; i < N/M; i++ {
				item, err := it.Next()
				if err != nil {
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

	close(rogue)
	assert.Len(t, ids, N, "len expected %d, actual %d", N, len(ids))
	sort.Ints(ids)
	equals := true
	for i := 0; i < N; i++ {
		equals = equals && i == ids[i]
	}
}

func TestIteratorCyclicConcurrent(t *testing.T) {
	ll := NewItemLinkedList()
	cit := ll.CyclicIterator()

	for i := 0; i < 10; i++ {
		item := &Item{Id: i, AccessCount: atomic.NewInt64(0)}
		_, _ = ll.Insert(item)
	}

	runs := int64(ll.Len() * 1000 * 1000)
	N := atomic.NewInt64(runs)
	const M = 10
	g, _ := errgroup.WithContext(context.Background())

	var worker [M]int // Do cit with 10 workers
	for range worker {
		g.Go(func() error {
			for N.Load() > 0 {
				item, err := cit.Next()
				// Fail if we cannot retrieve an item; we should always return an item
				if err != nil {
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
	it := ll.Iterator()
	for n, err := it.Next(); err == nil; n, err = it.Next() {
		percentusage := float64(n.AccessCount.Load()) / float64(runs)
		t.Log(percentusage)
		assert.True(t, percentusage >= (1/float64(M) - e))
		assert.True(t, percentusage <= (1/float64(M) + e))
	}
}
