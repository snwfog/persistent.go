package persistent

import (
  "math/rand"
  "runtime"
  "sync"
  "sync/atomic"
  "testing"
  "time"

  "github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
  dll := NewLinkedList()
  assert.Equal(t, 0, dll.Len())
  assert.NotNil(t, dll.Head())
  assert.NotNil(t, dll.Tail())
  assert.False(t, dll.Head() == dll.Tail())
}

func TestInsertUniqueOnly(t *testing.T) {
  dll := NewLinkedList()

  n1, n2 := NewNode(1), NewNode(1)
  t.Logf("%d, %d", n1.key, n2.key)

  _, _ = dll.Insert(NewNode(1))
  assert.Equal(t, 1, dll.Len())

  _, _ = dll.Insert(NewNode(1))
  assert.Equal(t, 1, dll.Len())
}

func TestInsert1(t *testing.T) {
  dll := NewLinkedList()

  n1, n2 := NewNode(1), NewNode(2)
  t.Logf("%d, %d", n1.key, n2.key)

  _, _ = dll.Insert(n1)
  assert.Equal(t, 1, dll.Len())

  _, _ = dll.Insert(n2)
  assert.Equal(t, 2, dll.Len())
}

func TestInsert2(t *testing.T) {
  dll := NewLinkedList()
  n1, n2, n3 := NewNode(1), NewNode(2), NewNode(3)

  _, _ = dll.Insert(n1)
  _, _ = dll.Insert(n2)
  _, _ = dll.Insert(n3)

  assert.Equal(t, 3, dll.Len())

  assert.Equal(t, 1, dll.Head().Next().value)
  assert.Equal(t, 2, dll.Head().Next().Next().value)
  assert.Equal(t, 3, dll.Head().Next().Next().Next().value)
}

func TestConcurrentInsert1(t *testing.T) {
  dll := NewLinkedList()
  p := runtime.NumCPU()
  n := 1 << 16
  nodes := make([]*node, 0, n)
  for i := 0; i < n; i++ {
    nodes = append(nodes, NewNode(i))
  }

  wg := sync.WaitGroup{}
  wg.Add(p)

  for i := 0; i < p; i += 1 {
    go func() {
      for j := 0; j < n; j++ {
        _, _ = dll.Insert(nodes[j])
      }
      wg.Done()
    }()
  }

  wg.Wait()
  assert.Equal(t, n, dll.Len())

  s := 0
  count := (n - 1) * n / 2
  t.Logf("expected len: %d, actual len: %d", n, dll.Len())

  for n := dll.Head().Next(); n != dll.Tail(); n = n.Next() {
    s += n.value
    // t.Log(n.value)
    assert.NotNil(t, n)
  }

  assert.Equal(t, count, s)
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
  dll := NewLinkedList()
  n := 1 << 16
  workChan := make(chan int)

  wg := sync.WaitGroup{}
  wg.Add(1)

  go func() {
    for j := 0; j < n; j++ {
      node := NewNode(j)
      _, _ = dll.Insert(node)

      if j%2 == 0 {
        workChan <- j
      }
    }

    close(workChan)
  }()

  go func() {
    for i := range workChan {
      node := NewNode(i)
      _, _ = dll.Delete(node)
    }

    wg.Done()
  }()

  wg.Wait()

  assert.Equal(t, n>>1, dll.Len())
}

func TestParallelInsertDelete2(t *testing.T) {
  n := 1 << 10
  workChan := make(chan int)
  doneChan := make(chan int)
  dll := NewLinkedList()

  // Producers
  go func() {
    for i := 0; i < n; i++ {
      _, _ = dll.Insert(NewNode(i))

      if i%2 == 0 {
        workChan <- i
      }
    }

    close(workChan)
  }()

  go func() {
    for i := range workChan {
      _, _ = dll.Delete(NewNode(i))
    }

    doneChan <- 1
  }()

  go func() {
    for i := range workChan {
      _, _ = dll.Delete(NewNode(i))
    }

    doneChan <- 2
  }()

  <-doneChan
  <-doneChan

  assert.Equal(t, n>>1, dll.Len())
  it := dll.Iterator()

  var sum int
  for n, ok := it.Next(); ok; n, ok = it.Next() {
    sum += n.value
  }

  assert.Equal(t, n*n/4, sum)
  t.Logf("sum expected %d, actual %d", n*n/4, sum)
}

func TestParallelInsertDelete3(t *testing.T) {
  n := 1 << 10
  workChan := make(chan int)
  doneChan := make(chan int)
  dll := NewLinkedList()

  // Producers
  go func() {
    for i := 0; i < n; i++ {
      _, _ = dll.Insert(NewNode(i))

      if i%2 == 0 {
        workChan <- i
      }
    }

    close(workChan)
  }()

  go func() {
    for i := range workChan {
      _, _ = dll.Delete(NewNode(i))
    }

    doneChan <- 1
  }()

  go func() {
    for i := range workChan {
      _, _ = dll.Delete(NewNode(i))
    }

    doneChan <- 2
  }()

  <-doneChan
  <-doneChan

  assert.Equal(t, n>>1, dll.Len())
  it := dll.Iterator()

  var sum int
  for n, ok := it.Next(); ok; n, ok = it.Next() {
    sum += n.value
  }

  assert.Equal(t, n*n/4, sum)
  t.Logf("sum expected %d, actual %d", n*n/4, sum)
}

func TestDelete(t *testing.T) {
  dll := NewLinkedList()

  ok, err := dll.Delete(NewNode(1))
  assert.False(t, ok)
  assert.Nil(t, err)

  _, _ = dll.Insert(NewNode(1))
  assert.Equal(t, dll.Len(), 1)

  ok, err = dll.Delete(NewNode(1))
  assert.True(t, ok)
  assert.Nil(t, err)

  ok, err = dll.Delete(NewNode(1))
  assert.False(t, ok)
  assert.Nil(t, err)

  assert.Equal(t, 0, dll.Len())
}


func BenchmarkInsert(b *testing.B) {
  dll := NewLinkedList()
  for i := 0; i < b.N; i += 1 {
    _, _ = dll.Insert(NewNode(i))
    // assert.Equal(Tail, i+1, dll.Len())
  }
}

func BenchmarkParallelInsert(b *testing.B) {
  n := 1
  dll := NewLinkedList()
  b.RunParallel(func(pb *testing.PB) {
    for pb.Next() {
      for j := 0; j < n; j++ {
        for ok, _ := dll.Insert(NewNode(j)); !ok; ok, _ = dll.Insert(NewNode(j)) {
          // Tail.Log("not ok")
        }
      }
    }
  })
}

// func BenchmarkParallelPrepend(b *testing.B) {
//   n := 1
//   dll := NewLinkedList()
//   b.RunParallel(func(pb *testing.PB) {
//     for pb.Next() {
//       for j := 0; j < n; j++ {
//         for ok, _ := dll.Prepend(NewNode(j)); !ok; ok, _ = dll.Prepend(NewNode(j)) {
//           // Tail.Log("not ok")
//         }
//       }
//     }
//   })
// }

var setThisForBenchmark bool

func BenchmarkParallelRead(b *testing.B) {
  threadsCount := 10000
  readN := 1000
  dll := NewLinkedList()
  node := NewNode(1)
  _, _ = dll.Insert(NewNode(1))
  b.SetParallelism(threadsCount)
  b.RunParallel(func(pb *testing.PB) {
    for pb.Next() {
      for i := 0; i < readN; i++ {
        setThisForBenchmark = dll.Contains(node)
      }
    }
  })
}

func BenchmarkParallelUpdate(b *testing.B) {
  threadsCount := 10000
  readN := 1000
  dll := NewLinkedList()
  b.SetParallelism(threadsCount)
  b.RunParallel(func(pb *testing.PB) {
    for pb.Next() {
      for i := 0; i < readN; i++ {
        _, _ = dll.Insert(NewNode(1))
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
        setThisForBenchmark = m[1]
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
//         setThisForBenchmark = m["A"]
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
  dll := NewLinkedList()
  threadsCount := 10000
  // p := runtime.NumCPU()
  // n := b.N
  nodes := make([]*node, 0, nodeCount)
  for i := 0; i < nodeCount; i++ {
    nodes = append(nodes, NewNode(i))
  }

  // node := NewNode(rand.Int())
  b.SetParallelism(threadsCount)
  b.ResetTimer()
  b.RunParallel(func(pb *testing.PB) {
    for pb.Next() {
      for i := 0; i < nodeCount; i++ {
        _, _ = dll.Insert(nodes[i])
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
  dll := NewLinkedList()

  // Insert some nodes
  n := 1 << 10
  for i := 0; i < n; i++ {
    _, _ = dll.Insert(NewNode(i))
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
          _, _ = dll.Insert(NewNode(rand.Int()))
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
