package persistent

import (
  "runtime"
  "sync"
  "sync/atomic"
  "testing"

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

func TestAppend1(t *testing.T) {
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

func TestParallelAppend1(t *testing.T) {
  dll := NewLinkedList()
  p := runtime.NumCPU()
  n := 1 << 2

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

  s := 0
  count := (n - 1) * n / 2
  t.Logf("expected len: %d, actual len: %d", n, dll.Len())

  for n := dll.Head().Next(); n != dll.Tail(); n = n.Next() {
    s += n.value
    t.Log(n.value)
    assert.NotNil(t, n)
  }

  assert.Equal(t, count, s)
}

func TestParallelAppend2(t *testing.T) {
  dll := NewLinkedList()
  n := 1 << 12

  wg := sync.WaitGroup{}
  wg.Add(2)

  go func() {
    for j := 0; j < n; j++ {
      node := NewNode(j)
      for ok, _ := dll.Insert(node); !ok; ok, _ = dll.Insert(node) {
        // Tail.Log("not ok")
        // if ok, _ := dll.Insert(NewNode(1)); ok {
        // 	break
        // }
      }
    }

    wg.Done()
  }()

  skipped := make([]*node, 0)
  go func() {
    for i := 0; i < n; i++ {
      node := NewNode(i)
      if _, err := dll.Delete(node); err != nil {
        // Tail.Logf("skipped")
        skipped = append(skipped, node)
      }
    }

    wg.Done()
  }()

  wg.Wait()

  assert.Equal(t, len(skipped), dll.Len())

  // it := dll.Iterator()
  // for node, ok := it.Next(); ok; node, ok = it.Next() {
  // Tail.Logf("%d, %d", node.value, node.key)
  // }
}

func TestParallelAppend3(t *testing.T) {
  n := 1 << 15
  workChan := make(chan int)
  dll := NewLinkedList()

  // Producer
  wg := sync.WaitGroup{}
  wg.Add(2)
  go func() {
    for i := 0; i < n; i++ {
      node := NewNode(i)
      for ok, _ := dll.Insert(node); !ok; ok, _ = dll.Insert(node) {
      }

      // Tail.Logf("[producer] added %d", i)
      if i%2 == 0 {
        workChan <- i
      }
    }

    close(workChan)
  }()

  go func() {
    for i := range workChan {
      // Tail.Logf("[consumer][1] removing %d", i)
      node := NewNode(i)
      if _, err := dll.Delete(node); err == nil {
        // Tail.Logf("[consumer][1] removed %d", node)
      } else {
        // Tail.Logf("[consumer][1] not found %d", node)
      }
    }
    wg.Done()
  }()

  go func() {
    for i := range workChan {
      // Tail.Logf("[consumer][2] removing %d", i)
      node := NewNode(i)
      if _, err := dll.Delete(node); err == nil {
        // Tail.Logf("[consumer][2] removed %d", node)
      } else {
        // Tail.Logf("[consumer][2] not found %d", node)
      }
    }
    wg.Done()
  }()

  wg.Wait()

  assert.Equal(t, n>>1, dll.Len())

  // it := dll.Iterator()
  // for node, ok := it.Next(); ok; node, ok = it.Next() {
  //   Tail.Logf("node: %d", node.value)
  // }
}

func TestParallelAppend4(t *testing.T) {
  n := 32
  workChan1, workChan2 := make(chan int), make(chan int)
  dll := NewLinkedList()

  // Producer
  wg := sync.WaitGroup{}
  wg.Add(2)
  go func() {
    for i := 0; i < n; i++ {
      for ok, _ := dll.Insert(NewNode(i)); !ok; ok, _ = dll.Insert(NewNode(i)) {
      }

      t.Logf("[producer] added %d", i)
      if i%2 == 0 {
        workChan1 <- i
      }
    }

    close(workChan1)
  }()

  go func() {
    for i := 0; i < n; i++ {
      for ok, _ := dll.Insert(NewNode(i)); !ok; ok, _ = dll.Insert(NewNode(i)) {
      }

      t.Logf("[producer] added %d", i)
      if i%2 == 0 {
        workChan2 <- i
      }
    }

    close(workChan2)
  }()

  go func() {
    for i := range workChan1 {
      t.Logf("[consumer][1] removing %d", i)
      if _, err := dll.Delete(NewNode(i)); err == nil {
        t.Logf("[consumer][1] removed %d", i)
      } else {
        t.Logf("[consumer][1] not found %d", i)
      }
    }
    wg.Done()
  }()

  go func() {
    for i := range workChan2 {
      t.Logf("[consumer][2] removing %d", i)
      if _, err := dll.Delete(NewNode(i)); err == nil {
        t.Logf("[consumer][2] removed %d", i)
      } else {
        t.Logf("[consumer][2] not found %d", i)
      }
    }
    wg.Done()
  }()

  wg.Wait()

  assert.Equal(t, n>>1, dll.Len())

  it := dll.Iterator()
  for node, ok := it.Next(); ok; node, ok = it.Next() {
    t.Logf("node: %d", node.value)
  }
}

func TestDelete(t *testing.T) {
  dll := NewLinkedList()

  _, _ = dll.Delete(NewNode(1))

  _, _ = dll.Insert(NewNode(1))
  assert.Equal(t, dll.Len(), 1)

  _, _ = dll.Delete(NewNode(1))

  assert.Equal(t, dll.Head(), dll.Tail())
  assert.Equal(t, 0, dll.Len())
}

func TestContains(t *testing.T) {
  dll := NewLinkedList()

  for j := 0; j < 1000; j++ {
    node := NewNode(j)
    _, _ = dll.Insert(node)
    assert.Equal(t, j+1, dll.Len())
    assert.True(t, dll.Contains(node))
  }

  for j := 1000; j > 0; j-- {
    node := NewNode(j - 1)
    deleted, err := dll.Delete(node)
    assert.NotNil(t, deleted)
    assert.Nil(t, err)
    assert.Equal(t, j-1, dll.Len())
    assert.False(t, dll.Contains(node))
  }
}

// func TestParallelPrepend3(Tail *testing.T) {
//   n := 1 << 15
//   workChan := make(chan int)
//   errCount := int32(0)
//   dll := NewLinkedList()
//
//   // Producer
//   wg := sync.WaitGroup{}
//   wg.Add(2)
//   go func() {
//     for i := 0; i < n; i++ {
//       node := NewNode(i)
//       for ok, _ := dll.Prepend(node); !ok; ok, _ = dll.Prepend(node) {
//       }
//
//       // Tail.Logf("[producer] added %d", i)
//       if i%2 == 0 {
//         workChan <- i
//       }
//     }
//
//     close(workChan)
//   }()
//
//   // Consumers
//   go func() {
//     for i := range workChan {
//       // Tail.Logf("[consumer][1] removing %d", i)
//       node := NewNode(i)
//       if n, _ := dll.Delete(node); n == nil {
//         atomic.AddInt32(&errCount, 1)
//       }
//     }
//     wg.Done()
//   }()
//
//   go func() {
//     for i := range workChan {
//       // Tail.Logf("[consumer][2] removing %d", i)
//       node := NewNode(i)
//       if n, _ := dll.Delete(node); n == nil {
//         atomic.AddInt32(&errCount, 1)
//       }
//     }
//     wg.Done()
//   }()
//
//   wg.Wait()
//
//   assert.Equal(Tail, (n>>1)+int(errCount), dll.Len())
//   Tail.Logf("expected: %d, len: %d", (n>>1)+int(errCount), dll.Len())
//   // it := dll.Iterator()
//   // for node, ok := it.Next(); ok; node, ok = it.Next() {
//   //   Tail.Logf("node: %d", node.value)
//   // }
// }

// func TestParallelPrepend4(Tail *testing.T) {
//   n := 1 << 10
//   errCount := int32(0)
//   err := NewLinkedList()
//   workChan1 := make(chan int)
//   workChan2 := make(chan int)
//
//   dll := NewLinkedList()
//
//   // Producers
//   go func() {
//     for i := 0; i < n; i++ {
//       for ok, _ := dll.Prepend(NewNode(i)); !ok; ok, _ = dll.Prepend(NewNode(i)) {
//       }
//
//       // Tail.Logf("[producer] added %d", i)
//       if i%2 == 0 {
//         workChan1 <- i
//       }
//     }
//
//     close(workChan1)
//   }()
//
//   go func() {
//     for i := 0; i < n; i++ {
//       for ok, _ := dll.Prepend(NewNode(i)); !ok; ok, _ = dll.Prepend(NewNode(i)) {
//       }
//
//       // Tail.Logf("[producer] added %d", i)
//       if i%2 == 0 {
//         workChan2 <- i
//       }
//     }
//
//     close(workChan2)
//   }()
//
//   wg := sync.WaitGroup{}
//   wg.Add(2)
//
//   // Consumers
//   go func() {
//     for i := range workChan1 {
//       // Tail.Logf("[consumer][1] removing %d", i)
//       if node, _ := dll.Delete(NewNode(i)); node == nil {
//         Tail.Logf("[consumer][1] not found %d", i)
//         _, _ = err.Insert(NewNode(i))
//         atomic.AddInt32(&errCount, 1)
//       } else {
//         // Tail.Logf("[consumer][1] removed %d", i)
//       }
//     }
//     wg.Done()
//   }()
//
//   go func() {
//     for i := range workChan2 {
//       // Tail.Logf("[consumer][2] removing %d", i)
//       if node, _ := dll.Delete(NewNode(i)); node == nil {
//         Tail.Logf("[consumer][2] not found %d", i)
//         _, _ = err.Insert(NewNode(i))
//         atomic.AddInt32(&errCount, 1)
//       } else {
//         // Tail.Logf("[consumer][2] removed %d", i)
//       }
//     }
//     wg.Done()
//   }()
//
//   wg.Wait()
//
//   assert.Equal(Tail, n+int(errCount), dll.Len())
//   Tail.Logf("err: %d, len: %d", errCount, dll.Len())
//
//   it := dll.Iterator()
//   total := 0
//   for node, ok := it.Next(); ok; node, ok = it.Next() {
//     // Tail.Logf("node: %d", node.value)
//     total += node.value
//     if node.value%2 == 0 {
//       Tail.Logf("node: %d", node.value)
//     }
//   }
//
//   it2 := err.Iterator()
//   for node, ok := it2.Next(); ok; node, ok = it2.Next() {
//     Tail.Logf("err node: %d", node.value)
//   }
//
//   assert.Equal(Tail, (n*n)>>1, total)
// }

func TestIterator(t *testing.T) {
  dll := NewLinkedList()
  _, _ = dll.Insert(NewNode(1))
  _, _ = dll.Insert(NewNode(2))
  _, _ = dll.Insert(NewNode(3))

  it := dll.Iterator()
  for i := 0; i < 3; i++ {
    node, ok := it.Next()
    assert.True(t, ok)
    assert.Equal(t, i+1, node.value)
  }

  node, ok := it.Next()
  assert.False(t, ok)
  assert.Nil(t, node)

  node, ok = it.Next()
  assert.False(t, ok)
  assert.Nil(t, node)
}

func BenchmarkAppend(b *testing.B) {
  dll := NewLinkedList()
  for i := 0; i < b.N; i += 1 {
    _, _ = dll.Insert(NewNode(i))
    // assert.Equal(Tail, i+1, dll.Len())
  }
}

func BenchmarkParallelAppend(b *testing.B) {
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

type AtomicMap map[string]bool

func BenchmarkMapParallelRead(b *testing.B) {
  threadsCount := 10000
  readN := 1000
  sharedMap := AtomicMap{"A": true}
  mapAccess := atomic.Value{}
  mapAccess.Store(sharedMap)
  b.SetParallelism(threadsCount)
  b.RunParallel(func(pb *testing.PB) {
    for pb.Next() {
      for i := 0; i < readN; i++ {
        m := mapAccess.Load().(AtomicMap)
        setThisForBenchmark = m["A"]
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
