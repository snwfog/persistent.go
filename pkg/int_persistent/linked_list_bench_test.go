package int_persistent

import (
  "math/rand"
  "sync"
  "sync/atomic"
  "testing"
  "time"
)

var contains bool

func BenchmarkParallelRead(b *testing.B) {
  threadsCount := 10000
  readN := 1000
  dll := NewIntLinkedList()
  node := NewBuiltinIntNode(1)
  _, _ = dll.Insert(NewBuiltinIntNode(1))
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
  dll := NewIntLinkedList()
  b.SetParallelism(threadsCount)
  b.RunParallel(func(pb *testing.PB) {
    for pb.Next() {
      for i := 0; i < readN; i++ {
        _, _ = dll.Insert(NewBuiltinIntNode(1))
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
  dll := NewIntLinkedList()
  threadsCount := 10000
  // p := runtime.NumCPU()
  // n := b.N
  nodes := make([]*IntNode, 0, nodeCount)
  for i := 0; i < nodeCount; i++ {
    nodes = append(nodes, NewBuiltinIntNode(i))
  }

  // node := NewBuiltinIntNode(rand.Int())
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
  dll := NewIntLinkedList()

  // Insert some nodes
  n := 1 << 10
  for i := 0; i < n; i++ {
    _, _ = dll.Insert(NewBuiltinIntNode(i))
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
          _, _ = dll.Insert(NewBuiltinIntNode(rand.Int()))
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
