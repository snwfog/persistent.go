package item

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var Contains bool

func BenchmarkParallelRead(b *testing.B) {
	threadsCount := 10000
	readN := 1000
	dll := NewItemLinkedList()

	node := &Item{Id: 1}
	_, _ = dll.Insert(node)
	b.SetParallelism(threadsCount)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < readN; i++ {
				Contains = dll.Contains(node)
			}
		}
	})
}

func BenchmarkParallelUpdate(b *testing.B) {
	threadsCount := 10000
	readN := 1000
	dll := NewItemLinkedList()
	b.SetParallelism(threadsCount)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < readN; i++ {
				_, _ = dll.Insert(&Item{Id: i})
			}
		}
	})
}

type AtomicMap map[int]bool

func BenchmarkMapParallelRead(b *testing.B) {
	threadsCount := 10000
	readN := 1000
	sharedMap := AtomicMap{1: true}
	mapAccess := atomic.Value{}
	mapAccess.Store(sharedMap)
	b.SetParallelism(threadsCount)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < readN; i++ {
				m := mapAccess.Load().(AtomicMap)
				Contains = m[1]
			}
		}
	})
}

// func BenchmarkParallelMapUpdate(b *testing.B) {
//   threadsCount := 10000
//   readN := 1000
//   sharedMap := AtomicMap{"A": true}
//   mapAccess := atomic.Value{}
//   mapAccess.Store(sharedMap)
//   b.SetParallelism(threadsCount)
//   b.RunParallel(func(pb *testing.PB) {
//     for pb.Next() {
//       for i := 0; i < readN; i++ {
//         m := mapAccess.Load().(AtomicMap)
//         Contains = m["A"]
//       }
//     }
//   })
// }

func BenchmarkParallelInsert_1(b *testing.B)  { parallelInsert(b, 1<<1) }
func BenchmarkParallelInsert_10(b *testing.B) { parallelInsert(b, 1<<10) }
func BenchmarkParallelInsert_11(b *testing.B) { parallelInsert(b, 1<<11) }
func BenchmarkParallelInsert_12(b *testing.B) { parallelInsert(b, 1<<12) }
func BenchmarkParallelInsert_13(b *testing.B) { parallelInsert(b, 1<<13) }

func parallelInsert(b *testing.B, nodeCount int) {
	// b.Logf("%d", nodeCount)
	dll := NewItemLinkedList()
	threadsCount := 10000
	// p := runtime.NumCPU()
	// n := b.N
	items := make([]*Item, 0, nodeCount)
	for i := 0; i < nodeCount; i++ {
		items = append(items, &Item{Id: i})
	}

	// node := rand.Item()
	b.SetParallelism(threadsCount)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < nodeCount; i++ {
				_, _ = dll.Insert(items[i])
			}
		}
	})
}

func BenchmarkMapParallelInsert_1(b *testing.B)  { mapParallelInsert(b, 1<<1) }
func BenchmarkMapParallelInsert_10(b *testing.B) { mapParallelInsert(b, 1<<10) }
func BenchmarkMapParallelInsert_11(b *testing.B) { mapParallelInsert(b, 1<<11) }
func BenchmarkMapParallelInsert_12(b *testing.B) { mapParallelInsert(b, 1<<12) }
func BenchmarkMapParallelInsert_13(b *testing.B) { mapParallelInsert(b, 1<<13) }

func mapParallelInsert(b *testing.B, nodeCount int) {
	threadsCount := 10000
	mu := sync.Mutex{}
	sharedMap := AtomicMap{}
	b.SetParallelism(threadsCount)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < nodeCount; i++ {
				mu.Lock()
				sharedMap[i] = true
				mu.Unlock()
			}
		}
	})
}

func BenchmarkParallelTake(b *testing.B) {
	threadsCount := 10000
	dll := NewItemLinkedList()

	// Insert some nodes
	n := 1 << 10
	for i := 0; i < n; i++ {
		_, _ = dll.Insert(&Item{Id: i})
	}

	producers := 5
	doneChan := make(chan struct{})
	for i := 0; i < producers; i++ {
		go func() {
			interval := time.NewTicker(1 * time.Microsecond)
		L:
			for {
				select {
				case <-interval.C:
					_, _ = dll.Insert(&Item{Id: rand.Int()})
				case <-doneChan:
					break L
				}
			}
		}()
	}

	cit := dll.CyclicIterator()
	b.ResetTimer()

	take := 10
	b.SetParallelism(threadsCount)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < take; i++ {
				_, _ = cit.Next()
			}
		}
	})

	close(doneChan)
}

var m AtomicMap
var val int

func BenchmarkMapParallelTake(b *testing.B) {
	threadsCount := 10000
	m = AtomicMap{}
	mu := sync.Mutex{}

	// Insert some nodes
	n := 1 << 10
	for i := 0; i < n; i++ {
		m[i] = true
	}

	producers := 5
	doneChan := make(chan struct{})
	for i := 0; i < producers; i++ {
		go func() {
			interval := time.NewTicker(1 * time.Microsecond)
		L:
			for {
				select {
				case <-interval.C:
					mu.Lock()
					m[rand.Int()] = true
					mu.Unlock()
				case <-doneChan:
					break L
				}
			}
		}()
	}

	b.ResetTimer()

	take := 10
	b.SetParallelism(threadsCount)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < take; i++ {
				mu.Lock()
				for key := range m {
					val = key
					break
				}
				mu.Unlock()
			}
		}
	})

	close(doneChan)
}

var (
	R   = 3                                     // Regions
	M   = 5000                                  // Number of items per map
	L   = 5000                                  // Number of items per list
	TH  = 20000                                 // Thread count
	DUR = time.Duration(100) * time.Millisecond // Item update frequency (ms)
)

type sig = uint

const TERM sig = iota

var Result1 int

// Benchmark iterate through maps and copy the value into a Result slice
func BenchmarkItemCopyMap(b *testing.B) {
	regionA, regionB, regionC := setupregions()
	// mu := sync.RWMutex{}

	b.ResetTimer()

	results := make([]*Item, 0, 3*M)
	for i := 0; i < b.N; i++ {
		for j := 0; j < M; j++ {
			results = append(results, regionA[j])
			results = append(results, regionB[j])
			results = append(results, regionC[j])
		}

		Result1 = Result1 + len(results)
		results = results[:0] // Zero out slice
	}

	// b.Logf("len: %d", len(results))
}

var Result2 int

// Persistent linked list
// Benchmark iterator through persistent linked list and copy the
// node item reference to a Result slice
func BenchmarkItemCopyLinkedList(b *testing.B) {
	regionA := NewItemLinkedList()
	regionB := NewItemLinkedList()
	regionC := NewItemLinkedList()

	for i := 0; i < L; i++ {
		c1 := &Item{Id: i*R + 0}
		c2 := &Item{Id: i*R + 1}
		c3 := &Item{Id: i*R + 2}
		_, _ = regionA.Insert(c1)
		_, _ = regionB.Insert(c2)
		_, _ = regionC.Insert(c3)
	}

	// b.Logf("len: %d, %d, %d", regionA.Len(), regionB.Len(), regionC.Len())
	itA := regionA.Iterator()
	itB := regionB.Iterator()
	itC := regionC.Iterator()

	b.ResetTimer()
	results := make([]*Item, 0, 3*L)
	for i := 0; i < b.N; i++ {
		for n, err := itA.Next(); err == nil; n, err = itA.Next() {
			results = append(results, n)
		}

		for n, err := itB.Next(); err == nil; n, err = itB.Next() {
			results = append(results, n)
		}

		for n, err := itC.Next(); err == nil; n, err = itC.Next() {
			results = append(results, n)
		}

		Result2 = Result2 + len(results)
		results = results[:0]
	}
	// b.Logf("len: %d", len(result))
}

var Result3 int

// Parallel copying item from regions into global result3
// Write access is synchronized with a mutex
func BenchmarkItemCopyParallelMapLock(b *testing.B) {
	regionA, regionB, regionC := setupregions()
	mu := sync.RWMutex{}
	// b.SetParallelism(TH)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		results := make([]*Item, 0, 3*M)
		for pb.Next() {
			mu.RLock()
			for j := 0; j < M; j++ {
				results = append(results, regionA[j])
				results = append(results, regionB[j])
				results = append(results, regionC[j])
			}
			mu.RUnlock()

			mu.Lock()
			Result3 = Result3 + len(results)
			mu.Unlock()
			results = results[:0]
		}
	})
}

var Result4 int

func BenchmarkItemCopyParallelList(b *testing.B) {
	regionA := NewItemLinkedList()
	regionB := NewItemLinkedList()
	regionC := NewItemLinkedList()

	for i := 0; i < L; i++ {
		c1 := &Item{Id: i*R + 0}
		c2 := &Item{Id: i*R + 1}
		c3 := &Item{Id: i*R + 2}
		_, _ = regionA.Insert(c1)
		_, _ = regionB.Insert(c2)
		_, _ = regionC.Insert(c3)
	}

	mu := sync.Mutex{}

	itA := regionA.Iterator()
	itB := regionB.Iterator()
	itC := regionC.Iterator()

	// b.SetParallelism(TH)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		results := make([]*Item, 0, L*R)
		for pb.Next() {
			for n, err := itA.Next(); err == nil; n, err = itA.Next() {
				results = append(results, n)
			}

			for n, err := itB.Next(); err == nil; n, err = itB.Next() {
				results = append(results, n)
			}

			for n, err := itC.Next(); err == nil; n, err = itC.Next() {
				results = append(results, n)
			}

			mu.Lock()
			Result4 = Result4 + len(results)
			mu.Unlock()

			results = results[:0]
		}
	})
}

var Result5 int

func BenchmarkItemCopyParallelMapLockWithItemInsert(b *testing.B) {
	regionA, regionB, regionC := setupregions()
	done := make(chan sig)

	mu := sync.RWMutex{}
	// Every `DUR` interval, take the `mu` lock
	// an add new Items into all regions
	go func() {
		i := L + 1
		for {
			select {
			case <-time.Tick(DUR):
				id := i
				c1, c2, c3 := &Item{Id: id*R + 0}, &Item{Id: id*R + 1}, &Item{Id: id*R + 2}
				mu.Lock()
				regionA = append(regionA, c1)
				regionB = append(regionB, c2)
				regionC = append(regionC, c3)
				mu.Unlock()
				i += 1
			case <-done:
				break
			}
		}
	}()

	// b.SetParallelism(TH)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		results := make([]*Item, 0, 3*M)
		for pb.Next() {
			mu.RLock()
			for j := 0; j < M; j++ {
				results = append(results, regionA[j])
				results = append(results, regionB[j])
				results = append(results, regionC[j])
			}
			mu.RUnlock()

			mu.Lock()
			Result5 = Result5 + len(results)
			mu.Unlock()
			results = results[:0]
		}
	})

	done <- TERM
}

var Result6 int

func BenchmarkItemCopyParallelListWithItemInsert(b *testing.B) {
	regionA := NewItemLinkedList()
	regionB := NewItemLinkedList()
	regionC := NewItemLinkedList()

	for i := 0; i < L; i++ {
		c1 := &Item{Id: i*R + 0}
		c2 := &Item{Id: i*R + 1}
		c3 := &Item{Id: i*R + 2}
		_, _ = regionA.Insert(c1)
		_, _ = regionB.Insert(c2)
		_, _ = regionC.Insert(c3)
	}

	mu := sync.Mutex{}
	done := make(chan sig)

	go func() {
		i := L + 1
		for {
			select {
			case <-time.Tick(DUR):
				c1 := &Item{Id: i*R + 0}
				c2 := &Item{Id: i*R + 1}
				c3 := &Item{Id: i*R + 2}
				_, _ = regionA.Insert(c1)
				_, _ = regionB.Insert(c2)
				_, _ = regionC.Insert(c3)
				i += 1
			case <-done:
				break
			}
		}
	}()

	itA := regionA.Iterator()
	itB := regionB.Iterator()
	itC := regionC.Iterator()

	// b.SetParallelism(TH)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		results := make([]*Item, 0, L*R)
		for pb.Next() {
			for n, err := itA.Next(); err == nil; n, err = itA.Next() {
				results = append(results, n)
			}
			for n, err := itB.Next(); err == nil; n, err = itB.Next() {
				results = append(results, n)
			}
			for n, err := itC.Next(); err == nil; n, err = itC.Next() {
				results = append(results, n)
			}

			mu.Lock()
			Result6 = Result6 + len(results)
			mu.Unlock()

			results = results[:0]
		}
	})

	done <- TERM
}

// region Private
func setupregions() ([]*Item, []*Item, []*Item) {
	regionA := make([]*Item, 0, M)
	regionB := make([]*Item, 0, M)
	regionC := make([]*Item, 0, M)

	for i := 0; i < M; i++ {
		regionA = append(regionA, &Item{Id: i*R + 0})
		regionB = append(regionB, &Item{Id: i*R + 1})
		regionC = append(regionC, &Item{Id: i*R + 2})
	}

	return regionA, regionB, regionC
}

// endregion
