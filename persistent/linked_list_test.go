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
}

func TestConcurrentInsert1(t *testing.T) {
  dll := NewLinkedList()
  p := runtime.NumCPU()
  n := 1 << 10

  // nodes := make([]*node, 0, n)
  // for i := 0; i < n; i++ {
  //   nodes = append(nodes, NewNode(i))
  // }

  wg := sync.WaitGroup{}
  wg.Add(p)

  for i := 0; i < p; i += 1 {
    go func() {
      for j := 0; j < n; j++ {
        _, _ = dll.Insert(NewNode(j))
      }
      wg.Done()
    }()
  }

  wg.Wait()

  assert.Equal(t, n, dll.Len())

  it := dll.Iterator()
  var sum int
  for n, ok := it.Next(); ok; n, ok = it.Next() {
    sum += n.value
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
  dll := NewLinkedList()
  n := 1 << 10
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

func TestParallelInsertDeleteN1(t *testing.T) {
  n := 1 << 10
  pc := 20000

  channel := make(chan int)

  doneChan := make(chan int)
  dll := NewLinkedList()
  wg := sync.WaitGroup{}
  wg.Add(pc)

  // Producers
  for i := 0; i < pc; i++ {
    go func(k int) {
      for j := 0; j < n; j++ {
        n := NewNode(j)
        _, _ = dll.Insert(n)
        // t.Logf("p[1] %d, hash %d", k, n.key)
        if j%2 == 0 {
          channel <- j
        }
      }

      wg.Done()
    }(i)
  }

  // Consumers
  go func() {
    for j := range channel {
      n := NewNode(j)
      _, _ = dll.Delete(n)
      // t.Logf("c[1] %d, hash %d, len: %d, ok: %v", k, n.key, dll.Len(), ok)
    }

    doneChan <- 1
  }()

  wg.Wait()
  close(doneChan)
  <-doneChan

  assert.Equal(t, n>>1, dll.Len())
  it := dll.Iterator()

  var sum int
  for n, ok := it.Next(); ok; n, ok = it.Next() {
    sum += n.value
    if n.value%2 == 0 {
      t.Logf("value: %d!", n.value)
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
  dll := NewLinkedList()

  // Producers
  go func() {
    for j := 0; j < n; j++ {
      n := NewNode(j)
      _, _ = dll.Insert(n)
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
        n := NewNode(j)
        _, _ = dll.Delete(n)
        // t.Logf("c[1] %d, hash %d, len: %d, ok: %v", k, n.key, dll.Len(), ok)
      }

      doneChan <- 1
    }(i)
  }

  for i := 0; i < pc; i++ {
    <-doneChan
  }

  assert.Equal(t, n>>1, dll.Len())
  it := dll.Iterator()

  var sum int
  for n, ok := it.Next(); ok; n, ok = it.Next() {
    sum += n.value
    if n.value%2 == 0 {
      t.Logf("value: %d!", n.value)
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
  dll := NewLinkedList()

  // Producers
  for i := 0; i < pc; i++ {
    go func(k int) {
      for j := 0; j < n; j++ {
        n := NewNode(j)
        _, _ = dll.Insert(n)
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
        n := NewNode(j)
        _, _ = dll.Delete(n)
        // t.Logf("c[1] %d, hash %d, len: %d, ok: %v", k, n.key, dll.Len(), ok)
      }

      doneChan <- 1
    }(i)
  }

  for i := 0; i < pc; i++ {
    <-doneChan
  }

  assert.Equal(t, n>>1, dll.Len())
  it := dll.Iterator()

  var sum int
  for n, ok := it.Next(); ok; n, ok = it.Next() {
    sum += n.value
    if n.value%2 == 0 {
      t.Logf("value: %d!", n.value)
    }
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

func TestConcurrentDelete1(t *testing.T) {
  dll := NewLinkedList()
  p := runtime.NumCPU()
  n := 1 << 10

  for i := 0; i < n; i++ {
    _, _ = dll.Insert(NewNode(i))
  }

  wg := sync.WaitGroup{}
  wg.Add(p)

  for i := 0; i < p; i += 1 {
    go func() {
      for j := 0; j < n; j++ {
        _, _ = dll.Delete(NewNode(j))
      }
      wg.Done()
    }()
  }

  wg.Wait()

  assert.Equal(t, 0, dll.Len())

  it := dll.Iterator()
  var sum int
  for n, ok := it.Next(); ok; n, ok = it.Next() {
    sum += n.value
  }

  assert.Equal(t, 0, sum)
}

var contains bool

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
        contains = dll.Contains(node)
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
        contains = m[1]
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
//         contains = m["A"]
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
