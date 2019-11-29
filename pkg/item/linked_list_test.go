package item

import (
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	ll := NewItemLinkedList()
	assert.Equal(t, 0, ll.Len())
	assert.NotNil(t, ll.headnode())
	assert.NotNil(t, ll.tailnode())
	assert.False(t, ll.headnode() == ll.tailnode())
}

func TestInsertUniqueOnly(t *testing.T) {
	ll := NewItemLinkedList()

	_, _ = ll.Insert(&Item{Id: 1})
	assert.Equal(t, 1, ll.Len())

	_, _ = ll.Insert(&Item{Id: 1})
	assert.Equal(t, 1, ll.Len())
}

func TestInsert1(t *testing.T) {
	ll := NewItemLinkedList()

	n1, n2 := &Item{Id: 1}, &Item{Id: 2}
	// t.Logf("%d, %d", n1.key, n2.key)

	_, _ = ll.Insert(n1)
	assert.Equal(t, 1, ll.Len())

	_, _ = ll.Insert(n2)
	assert.Equal(t, 2, ll.Len())
}

func TestInsert2(t *testing.T) {
	ll := NewItemLinkedList()
	n1, n2 := &Item{Id: 1}, &Item{Id: 2}
	n3 := &Item{Id: 3}

	_, _ = ll.Insert(n1)
	_, _ = ll.Insert(n2)
	_, _ = ll.Insert(n3)

	assert.Equal(t, 3, ll.Len())
}

func TestConcurrentInsert1(t *testing.T) {
	ll := NewItemLinkedList()
	p := runtime.NumCPU()
	n := 1 << 10

	// nodes := make([]*node, 0, n)
	// for i := 0; i < n; i++ {
	//   nodes = append(nodes, i)
	// }

	wg := sync.WaitGroup{}
	wg.Add(p)

	for i := 0; i < p; i += 1 {
		go func() {
			for j := 0; j < n; j++ {
				_, _ = ll.Insert(&Item{Id: j})
			}
			wg.Done()
		}()
	}

	wg.Wait()

	assert.Equal(t, n, ll.Len())

	it := ll.Iterator()
	var sum int
	for v, err := it.Next(); err == nil; v, err = it.Next() {
		sum += v.Id
	}

	assert.Equal(t, (n-1)*n/2, sum)
}

// var m map[int]bool
//
// func TestParallelMapInsert1(t *testing.T) {
//   m = make(map[int]bool)
//   mu := sync.Mutex{}
//   p := runtime.NumCPU()
//   n := 1 << 10
//
//   wg := sync.WaitGroup{}
//   wg.Add(p)
//
//   for i := 0; i < p; i += 1 {
//     go func() {
//       for j := 0; j < n; j++ {
//         mu.Lock()
//         m[j] = true
//         mu.Unlock()
//       }
//       wg.Done()
//     }()
//   }
//
//   wg.Wait()
//   assert.Equal(t, n, len(m))
//
//   t.Logf("%v", m)
// }

func TestConcurrentInsertDelete1(t *testing.T) {
	ll := NewItemLinkedList()
	n := 1 << 10
	workChan := make(chan int)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for j := 0; j < n; j++ {
			node := &Item{Id: j}
			_, _ = ll.Insert(node)

			if j%2 == 0 {
				workChan <- j
			}
		}

		close(workChan)
	}()

	go func() {
		for i := range workChan {
			node := &Item{Id: i}
			_, _ = ll.Delete(node)
		}

		wg.Done()
	}()

	wg.Wait()

	assert.Equal(t, n>>1, ll.Len())
}

func TestConcurrentInsertDeleteN1(t *testing.T) {
	n := 1 << 12
	pc := 100
	// pc := 20000

	ll := NewItemLinkedList()
	channel := make(chan int)
	doneChan := make(chan int)

	wg := sync.WaitGroup{}
	wg.Add(pc)

	// Producers
	for i := 0; i < pc; i++ {
		go func() {
			for j := 0; j < n; j++ {
				_, _ = ll.Insert(&Item{Id: j})
				// t.Logf("p[1] %d, hash %d", k, n.key)
				if j%2 == 0 {
					channel <- j
				}
			}

			wg.Done()
		}()
	}

	// Consumers
	go func() {
		for j := range channel {
			_, _ = ll.Delete(&Item{Id: j})
			// t.Logf("c[1] %d, hash %d, len: %d, ok: %v", k, n.key, ll.Len(), ok)
		}

		doneChan <- 1
	}()

	wg.Wait()
	close(doneChan)
	<-doneChan

	for n>>1 != ll.Len() {
	}
	assert.Equal(t, n>>1, ll.Len())

	it := ll.Iterator()

	var sum int
	for n, err := it.Next(); err == nil; n, err = it.Next() {
		sum += n.Id
		if n.Id%2 == 0 {
			t.Logf("value: %d!", n.Id)
		}
	}

	assert.Equal(t, n*n/4, sum)
	t.Logf("sum expected %d, actual %d", n*n/4, sum)
}

func TestConcurrentInsertDelete1N(t *testing.T) {
	n := 1 << 10
	pc := 100

	channels := make([]chan int, 0, pc)
	for i := 0; i < pc; i++ {
		channels = append(channels, make(chan int))
	}

	doneChan := make(chan int)
	ll := NewItemLinkedList()

	// Producers
	go func() {
		for j := 0; j < n; j++ {
			_, _ = ll.Insert(&Item{Id: j})
			// t.Logf("p[1] %d, hash %d", k, n.key)
			if j%2 == 0 {
				for k := 0; k < pc; k++ {
					channels[k] <- j
				}
			}
		}

		for k := 0; k < pc; k++ {
			close(channels[k])
		}
	}()

	// Consumers
	for i := 0; i < pc; i++ {
		go func(k int) {
			for j := range channels[k] {
				_, _ = ll.Delete(&Item{Id: j})
				// t.Logf("c[1] %d, hash %d, len: %d, ok: %v", k, n.key, ll.Len(), ok)
			}

			doneChan <- 1
		}(i)
	}

	for i := 0; i < pc; i++ {
		<-doneChan
	}

	assert.Equal(t, n>>1, ll.Len())
	it := ll.Iterator()

	var sum int
	for n, err := it.Next(); err == nil; n, err = it.Next() {
		sum += n.Id
		if n.Id%2 == 0 {
			t.Logf("value: %d!", n.Id)
		}
	}

	assert.Equal(t, n*n/4, sum)
	t.Logf("sum expected %d, actual %d", n*n/4, sum)
}

func TestConcurrentInsertDeleteNN(t *testing.T) {
	n := 1 << 10
	pc := 100

	channels := make([]chan int, 0, pc)
	for i := 0; i < pc; i++ {
		channels = append(channels, make(chan int))
	}

	doneChan := make(chan int)
	ll := NewItemLinkedList()

	// Producers
	for i := 0; i < pc; i++ {
		go func(k int) {
			for j := 0; j < n; j++ {
				_, _ = ll.Insert(&Item{Id: j})
				// t.Logf("p[1] %d, hash %d", k, n.key)
				if j%2 == 0 {
					channels[k] <- j
				}
			}

			close(channels[k])
		}(i)
	}

	// Consumers
	for i := 0; i < pc; i++ {
		go func(k int) {
			for j := range channels[k] {
				_, _ = ll.Delete(&Item{Id: j})
				// t.Logf("c[1] %d, hash %d, len: %d, ok: %v", k, n.key, ll.Len(), ok)
			}

			doneChan <- 1
		}(i)
	}

	for i := 0; i < pc; i++ {
		<-doneChan
	}

	assert.Equal(t, n>>1, ll.Len())
	it := ll.Iterator()

	var sum int
	for n, err := it.Next(); err == nil; n, err = it.Next() {
		sum += n.Id
		if n.Id%2 == 0 {
			t.Logf("value: %d!", n.Id)
		}
	}

	assert.Equal(t, n*n/4, sum)
	t.Logf("sum expected %d, actual %d", n*n/4, sum)
}

func TestUpsert(t *testing.T) {
	ll := NewItemLinkedList()
	a := 1
	ok, err := ll.Delete(&Item{Id: a})
	assert.False(t, ok)
	assert.Nil(t, err)

	_, _ = ll.Insert(&Item{Id: a})
	assert.Equal(t, ll.Len(), 1)

	ok, err = ll.Upsert(&Item{Id: a})
	assert.True(t, ok)
	assert.Nil(t, err)
	assert.Equal(t, ll.Len(), 1)

	ok, err = ll.Upsert(&Item{Id: a})
	assert.True(t, ok)
	assert.Nil(t, err)
	assert.Equal(t, ll.Len(), 1)

	ok, err = ll.Delete(&Item{Id: a})
	assert.True(t, ok)
	assert.Nil(t, err)
	assert.Equal(t, ll.Len(), 0)
}

func TestDelete(t *testing.T) {
	ll := NewItemLinkedList()

	ok, err := ll.Delete(&Item{Id: 1})
	assert.False(t, ok)
	assert.Nil(t, err)

	_, _ = ll.Insert(&Item{Id: 1})
	assert.Equal(t, ll.Len(), 1)

	ok, err = ll.Delete(&Item{Id: 1})
	assert.True(t, ok)
	assert.Nil(t, err)

	ok, err = ll.Delete(&Item{Id: 1})
	assert.False(t, ok)
	assert.Nil(t, err)

	assert.Equal(t, 0, ll.Len())
}

func TestConcurrentDelete1(t *testing.T) {
	ll := NewItemLinkedList()
	p := runtime.NumCPU()
	n := 1 << 10

	for i := 0; i < n; i++ {
		_, _ = ll.Insert(&Item{Id: i})
	}

	wg := sync.WaitGroup{}
	wg.Add(p)

	for i := 0; i < p; i += 1 {
		go func() {
			for j := 0; j < n; j++ {
				_, _ = ll.Delete(&Item{Id: j})
			}
			wg.Done()
		}()
	}

	wg.Wait()

	assert.Equal(t, 0, ll.Len())

	it := ll.Iterator()
	var sum int
	for n, err := it.Next(); err == nil; n, err = it.Next() {
		sum += n.Id
	}

	assert.Equal(t, 0, sum)
}

// region Plumbing
func TestMarkDeleteTest1(t *testing.T) {
	ll := NewItemLinkedList()
	item := &Item{Id: 1}
	_, _ = ll.Insert(item)
	t.Logf("%+v", ll.headnode().nextptr())
	// set n to be deleted

	n := (*node)(ll.headnode().next)
	n.next = deletemark(n.next)
	assert.False(t, ll.Contains(item))
}

// endregion
