package int

import (
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/snwfog/persistent.go/pkg"
)

func TestCreate(t *testing.T) {
	ll := pkg.NewIntLinkedList()
	assert.Equal(t, 0, ll.Len())
	assert.NotNil(t, ll.Head())
	assert.NotNil(t, ll.Tail())
	assert.False(t, ll.Head() == ll.Tail())
}

func TestInsertUniqueOnly(t *testing.T) {
	ll := pkg.NewIntLinkedList()

	// n1, n2 := NewIntNode(1), NewIntNode(1)
	// t.Logf("%d, %d", n1.key, n2.key)

	a := 1
	_, _ = ll.Insert(pkg.NewIntNode(&a))
	assert.Equal(t, 1, ll.Len())

	b := 2
	_, _ = ll.Insert(pkg.NewIntNode(&b))
	assert.Equal(t, 1, ll.Len())
}

func TestInsert1(t *testing.T) {
	ll := pkg.NewIntLinkedList()

	a, b := 1, 2
	n1, n2 := pkg.NewIntNode(&a), pkg.NewIntNode(&b)
	// t.Logf("%d, %d", n1.key, n2.key)

	_, _ = ll.Insert(n1)
	assert.Equal(t, 1, ll.Len())

	_, _ = ll.Insert(n2)
	assert.Equal(t, 2, ll.Len())
}

func TestInsert2(t *testing.T) {
	ll := pkg.NewIntLinkedList()
	a, b, c := 1, 2, 3
	n1, n2, n3 := pkg.NewIntNode(&a), pkg.NewIntNode(&b), pkg.NewIntNode(&c)

	_, _ = ll.Insert(n1)
	_, _ = ll.Insert(n2)
	_, _ = ll.Insert(n3)

	assert.Equal(t, 3, ll.Len())
}

func TestConcurrentInsert1(t *testing.T) {
	ll := pkg.NewIntLinkedList()
	p := runtime.NumCPU()
	n := 1 << 10

	// nodes := make([]*node, 0, n)
	// for i := 0; i < n; i++ {
	//   nodes = append(nodes, NewIntNode(i))
	// }

	wg := sync.WaitGroup{}
	wg.Add(p)

	for i := 0; i < p; i += 1 {
		go func() {
			for j := 0; j < n; j++ {
				a := j
				_, _ = ll.Insert(pkg.NewIntNode(&a))
			}
			wg.Done()
		}()
	}

	wg.Wait()

	assert.Equal(t, n, ll.Len())

	it := ll.Iterator()
	var sum int
	for v, ok := it.Next(); ok; v, ok = it.Next() {
		sum += *v
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
	ll := pkg.NewIntLinkedList()
	n := 1 << 10
	workChan := make(chan int)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for j := 0; j < n; j++ {
			a := j
			node := pkg.NewIntNode(&a)
			_, _ = ll.Insert(node)

			if j%2 == 0 {
				workChan <- j
			}
		}

		close(workChan)
	}()

	go func() {
		for i := range workChan {
			a := i
			node := pkg.NewIntNode(&a)
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

	ll := pkg.NewIntLinkedList()
	channel := make(chan int)
	doneChan := make(chan int)

	wg := sync.WaitGroup{}
	wg.Add(pc)

	// Producers
	for i := 0; i < pc; i++ {
		go func() {
			for j := 0; j < n; j++ {
				a := j
				_, _ = ll.Insert(pkg.NewIntNode(&a))
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
			a := j
			_, _ = ll.Delete(pkg.NewIntNode(&a))
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
	for n, ok := it.Next(); ok; n, ok = it.Next() {
		sum += *n
		if (*n)%2 == 0 {
			t.Logf("value: %d!", n)
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
	ll := pkg.NewIntLinkedList()

	// Producers
	go func() {
		for j := 0; j < n; j++ {
			a := j
			_, _ = ll.Insert(pkg.NewIntNode(&a))
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
				a := j
				_, _ = ll.Delete(pkg.NewIntNode(&a))
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
	for n, ok := it.Next(); ok; n, ok = it.Next() {
		sum += *n
		if (*n)%2 == 0 {
			t.Logf("value: %d!", n)
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
	ll := pkg.NewIntLinkedList()

	// Producers
	for i := 0; i < pc; i++ {
		go func(k int) {
			for j := 0; j < n; j++ {
				a := j
				_, _ = ll.Insert(pkg.NewIntNode(&a))
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
				a := j
				_, _ = ll.Delete(pkg.NewIntNode(&a))
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
	for n, ok := it.Next(); ok; n, ok = it.Next() {
		sum += *n
		if (*n)%2 == 0 {
			t.Logf("value: %d!", n)
		}
	}

	assert.Equal(t, n*n/4, sum)
	t.Logf("sum expected %d, actual %d", n*n/4, sum)
}

func TestUpsert(t *testing.T) {
	ll := pkg.NewIntLinkedList()
	a := 1
	ok, err := ll.Delete(pkg.NewIntNode(&a))
	assert.False(t, ok)
	assert.Nil(t, err)

	_, _ = ll.Insert(pkg.NewIntNode(&a))
	assert.Equal(t, ll.Len(), 1)

	ok, err = ll.Upsert(pkg.NewIntNode(&a))
	assert.True(t, ok)
	assert.Nil(t, err)
	assert.Equal(t, ll.Len(), 1)

	ok, err = ll.Upsert(pkg.NewIntNode(&a))
	assert.True(t, ok)
	assert.Nil(t, err)
	assert.Equal(t, ll.Len(), 1)

	ok, err = ll.Delete(pkg.NewIntNode(&a))
	assert.True(t, ok)
	assert.Nil(t, err)
	assert.Equal(t, ll.Len(), 0)
}

func TestDelete(t *testing.T) {
	ll := pkg.NewIntLinkedList()

	a := 1
	ok, err := ll.Delete(pkg.NewIntNode(&a))
	assert.False(t, ok)
	assert.Nil(t, err)

	_, _ = ll.Insert(pkg.NewIntNode(&a))
	assert.Equal(t, ll.Len(), 1)

	ok, err = ll.Delete(pkg.NewIntNode(&a))
	assert.True(t, ok)
	assert.Nil(t, err)

	ok, err = ll.Delete(pkg.NewIntNode(&a))
	assert.False(t, ok)
	assert.Nil(t, err)

	assert.Equal(t, 0, ll.Len())
}

func TestConcurrentDelete1(t *testing.T) {
	ll := pkg.NewIntLinkedList()
	p := runtime.NumCPU()
	n := 1 << 10

	for i := 0; i < n; i++ {
		a := i
		_, _ = ll.Insert(pkg.NewIntNode(&a))
	}

	wg := sync.WaitGroup{}
	wg.Add(p)

	for i := 0; i < p; i += 1 {
		go func() {
			for j := 0; j < n; j++ {
				a := j
				_, _ = ll.Delete(pkg.NewIntNode(&a))
			}
			wg.Done()
		}()
	}

	wg.Wait()

	assert.Equal(t, 0, ll.Len())

	it := ll.Iterator()
	var sum int
	for n, ok := it.Next(); ok; n, ok = it.Next() {
		sum += *n
	}

	assert.Equal(t, 0, sum)
}

// region Plumbing
func TestMarkDeleteTest1(t *testing.T) {
	ll := pkg.NewIntLinkedList()
	a := 1
	n := pkg.NewIntNode(&a)
	_, _ = ll.Insert(n)
	t.Logf("%+v", ll.Head().nextptr())
	// set n to be deleted
	n.next = pkg.deletemark(n.next)
	assert.False(t, ll.Contains(n))
}

// endregion
