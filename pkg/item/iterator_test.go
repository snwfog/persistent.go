package item

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"

	"github.com/snwfog/persistent.go"
)

func TestCyclicIterator1(t *testing.T) {
	dll := main.NewItemLinkedList()

	for i := 0; i < 3; i++ {
		item := &main.Item{Id: i, AccessCount: atomic.NewInt64(0)}
		_, _ = dll.Insert(main.NewItemNode(item))
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

func TestConcurrentCyclicIterator(t *testing.T) {
	list := main.NewItemLinkedList()
	cit := list.CyclicIterator()
	for i := 0; i < 10; i++ {
		item := &main.Item{Id: i, AccessCount: atomic.NewInt64(0)}
		_, _ = list.Insert(main.NewItemNode(item))
	}

	runs := int64(1000 * 1000 * 1000)
	N := atomic.NewInt64(runs)
	g, _ := errgroup.WithContext(context.Background())

	var worker [10]int // Do cit with 10 workers
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

	it := list.Iterator()
	for n, ok := it.Next(); ok; n, ok = it.Next() {
		percentusage := float64(n.AccessCount.Load()) / float64(N.Load())
		t.Log(percentusage)
	}
}
