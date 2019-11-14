package item

import (
	"sync"
	"testing"
	"time"

	"github.com/snwfog/persistent.go"
)

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

	results := make([]*main.Item, 0, 3*M)
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
	regionA := main.NewItemLinkedList()
	regionB := main.NewItemLinkedList()
	regionC := main.NewItemLinkedList()

	for i := 0; i < L; i++ {
		c1 := &main.Item{Id: i*R+0}
		c2 := &main.Item{Id: i*R+1}
		c3 := &main.Item{Id: i*R+2}
		_, _ = regionA.Insert(main.NewItemNode(c1))
		_, _ = regionB.Insert(main.NewItemNode(c2))
		_, _ = regionC.Insert(main.NewItemNode(c3))
	}

	// b.Logf("len: %d, %d, %d", regionA.Len(), regionB.Len(), regionC.Len())
	itA := regionA.Iterator()
	itB := regionB.Iterator()
	itC := regionC.Iterator()

	b.ResetTimer()
	results := make([]*main.Item, 0, 3*L)
	for i := 0; i < b.N; i++ {
		for n, ok := itA.Next(); ok; n, ok = itA.Next() {
			results = append(results, n)
		}

		for n, ok := itB.Next(); ok; n, ok = itB.Next() {
			results = append(results, n)
		}

		for n, ok := itC.Next(); ok; n, ok = itC.Next() {
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
		results := make([]*main.Item, 0, 3*M)
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
	regionA := main.NewItemLinkedList()
	regionB := main.NewItemLinkedList()
	regionC := main.NewItemLinkedList()

	for i := 0; i < L; i++ {
		c1 := &main.Item{Id: i*R+0}
		c2 := &main.Item{Id: i*R+1}
		c3 := &main.Item{Id: i*R+2}
		_, _ = regionA.Insert(main.NewItemNode(c1))
		_, _ = regionB.Insert(main.NewItemNode(c2))
		_, _ = regionC.Insert(main.NewItemNode(c3))
	}

	mu := sync.Mutex{}

	itA := regionA.Iterator()
	itB := regionB.Iterator()
	itC := regionC.Iterator()

	// b.SetParallelism(TH)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		results := make([]*main.Item, 0, L*R)
		for pb.Next() {
			for n, ok := itA.Next(); ok; n, ok = itA.Next() {
				results = append(results, n)
			}

			for n, ok := itB.Next(); ok; n, ok = itB.Next() {
				results = append(results, n)
			}

			for n, ok := itC.Next(); ok; n, ok = itC.Next() {
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
				c1, c2, c3 := &main.Item{Id: id*R + 0}, &main.Item{Id: id*R + 1}, &main.Item{Id: id*R + 2}
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
		results := make([]*main.Item, 0, 3*M)
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
	regionA := main.NewItemLinkedList()
	regionB := main.NewItemLinkedList()
	regionC := main.NewItemLinkedList()

	for i := 0; i < L; i++ {
		c1 := &main.Item{Id: i*R+0}
		c2 := &main.Item{Id: i*R+1}
		c3 := &main.Item{Id: i*R+2}
		_, _ = regionA.Insert(main.NewItemNode(c1))
		_, _ = regionB.Insert(main.NewItemNode(c2))
		_, _ = regionC.Insert(main.NewItemNode(c3))
	}

	mu := sync.Mutex{}
	done := make(chan sig)

	go func() {
		i := L + 1
		for {
			select {
			case <-time.Tick(DUR):
				c1 := &main.Item{Id: i*R+0}
				c2 := &main.Item{Id: i*R+1}
				c3 := &main.Item{Id: i*R+2}
				_, _ = regionA.Insert(main.NewItemNode(c1))
				_, _ = regionB.Insert(main.NewItemNode(c2))
				_, _ = regionC.Insert(main.NewItemNode(c3))
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
		results := make([]*main.Item, 0, L*R)
		for pb.Next() {
			for n, ok := itA.Next(); ok; n, ok = itA.Next() {
				results = append(results, n)
			}
			for n, ok := itB.Next(); ok; n, ok = itB.Next() {
				results = append(results, n)
			}
			for n, ok := itC.Next(); ok; n, ok = itC.Next() {
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
func setupregions() ([]*main.Item, []*main.Item, []*main.Item) {
	regionA := make([]*main.Item, 0, M)
	regionB := make([]*main.Item, 0, M)
	regionC := make([]*main.Item, 0, M)

	for i := 0; i < M; i++ {
		regionA = append(regionA, &main.Item{Id: i*R + 0})
		regionB = append(regionB, &main.Item{Id: i*R + 1})
		regionC = append(regionC, &main.Item{Id: i*R + 2})
	}

	return regionA, regionB, regionC
}

// endregion
