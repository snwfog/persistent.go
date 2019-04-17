package persistent

import (
  "runtime"
  "sync"
  "sync/atomic"
  "testing"

  "github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
  dll := NewDoublyLinkedList()
  assert.Equal(t, 0, dll.Len())
  assert.Nil(t, dll.Head())
  assert.Nil(t, dll.Tail())
}

func TestAppend(t *testing.T) {
  dll := NewDoublyLinkedList()

  _, _ = dll.Append(NewNode(1))
  assert.Equal(t, 1, dll.Len())

  assert.Equal(t, true, dll.Head() == dll.Tail())
  assert.Equal(t, 1, dll.Head().value)

  _, _ = dll.Append(NewNode(1))
  assert.Equal(t, 2, dll.Len())
}

func TestDelete(t *testing.T) {
  dll := NewDoublyLinkedList()

  _, _ = dll.Delete(NewNode(1))

  _, _ = dll.Append(NewNode(1))
  assert.Equal(t, dll.Len(), 1)

  _, _ = dll.Delete(NewNode(1))

  assert.Equal(t, dll.Head(), dll.Tail())
  assert.Equal(t, dll.Len(), 0)
}

func TestContains(t *testing.T) {
  dll := NewDoublyLinkedList()

  for j := 0; j < 1000; j++ {
    node := NewNode(j)
    _, _ = dll.Append(node)
    assert.Equal(t, j+1, dll.Len())
    assert.True(t, dll.Contains(node))
  }

  for j := 1000; j > 0; j-- {
    node := NewNode(j - 1)
    ok, _ := dll.Delete(node)
    assert.True(t, ok)
    assert.Equal(t, j-1, dll.Len())
    assert.False(t, dll.Contains(node))
  }
}

func TestAppend1(t *testing.T) {
  dll := NewDoublyLinkedList()
  _, _ = dll.Append(NewNode(1))
  _, _ = dll.Append(NewNode(2))
  _, _ = dll.Append(NewNode(3))

  assert.Equal(t, 3, dll.Len())

  _, _ = dll.Delete(NewNode(2))

  assert.Equal(t, 2, dll.Len())

  assert.Equal(t, 1, dll.Head().value)
  assert.Equal(t, 3, dll.Head().Next().value)
  assert.Equal(t, 1, dll.Head().Next().Next().value)

  assert.Equal(t, true, dll.Head().Prev() == dll.Tail())
  assert.Equal(t, true, dll.Tail().Next() == dll.Head())
}

func TestAppend2(t *testing.T) {
  dll := NewDoublyLinkedList()
  _, _ = dll.Append(NewNode(1))
  _, _ = dll.Append(NewNode(2))
  _, _ = dll.Append(NewNode(3))

  _, _ = dll.Delete(NewNode(2))

  assert.Equal(t, 3, dll.Tail().value)
  assert.Equal(t, 1, dll.Tail().Prev().value)
  assert.Equal(t, 3, dll.Tail().Next().Next().value)
  assert.Equal(t, 2, dll.Len())
}

func TestParallelAppend1(t *testing.T) {
  dll := NewDoublyLinkedList()
  p := runtime.NumCPU()
  n := 1 << 5

  wg := sync.WaitGroup{}
  wg.Add(p)
  for i := 0; i < p; i += 1 {
    go func() {
      for j := 0; j < n; j++ {
        for ok, _ := dll.Append(NewNode(j)); !ok; ok, _ = dll.Append(NewNode(j)) {
          // t.Log("not ok")
          // if ok, _ := dll.Append(NewNode(1)); ok {
          // 	break
          // }
        }
      }

      wg.Done()
    }()
  }

  wg.Wait()
  assert.Equal(t, p*n, dll.Len())

  s := 0
  count := (n - 1) * n / 2 * p
  t.Logf("expected len: %d, actual len: %d", n*p, dll.Len())

  for n, i := dll.Head(), 0; i < dll.Len(); n, i = n.Next(), i+1 {
    s += n.value
    t.Log(n.value)
    assert.NotNil(t, n)
  }

  assert.Equal(t, count, s)
}

func TestParallelAppend2(t *testing.T) {
  dll := NewDoublyLinkedList()
  n := 1 << 8

  wg := sync.WaitGroup{}
  wg.Add(2)

  go func() {
    for j := 0; j < n; j++ {
      for ok, _ := dll.Append(NewNode(j)); !ok; ok, _ = dll.Append(NewNode(j)) {
        // t.Log("not ok")
        // if ok, _ := dll.Append(NewNode(1)); ok {
        // 	break
        // }
      }
    }

    wg.Done()
  }()

  skipped := make([]*node, 0)
  go func() {
    for j := 0; j < n; j++ {
      node := NewNode(j)
      if dll.Contains(node) {
        for ok, _ := dll.Delete(node); !ok; ok, _ = dll.Delete(node) {
        }
        continue
      }

      skipped = append(skipped, node)
    }

    wg.Done()
  }()

  wg.Wait()

  assert.Equal(t, len(skipped), dll.Len())
}

func TestParallelAppend3(t *testing.T) {
  n := 1 << 8
  workChan := make(chan int)
  dll := NewDoublyLinkedList()

  // Producer
  wg := sync.WaitGroup{}
  wg.Add(2)
  go func() {
    for i := 0; i < n; i++ {
      for ok, _ := dll.Append(NewNode(i)); !ok; ok, _ = dll.Append(NewNode(i)) {
      }

      t.Logf("[producer] added %d", i)
      if i%2 == 0 {
        workChan <- i
      }
    }

    close(workChan)
  }()

  go func() {
    for i := range workChan {
      t.Logf("[consumer][1] removing %d", i)
      if ok, _ := dll.Delete(NewNode(i)); ok {
        t.Logf("[consumer][1] removed %d", i)
      } else {
        t.Logf("[consumer][1] not found %d", i)
      }
    }
    wg.Done()
  }()

  go func() {
    for i := range workChan {
      t.Logf("[consumer][2] removing %d", i)
      if ok, _ := dll.Delete(NewNode(i)); ok {
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

func TestParallelAppend4(t *testing.T) {
  n := 32
  workChan1, workChan2 := make(chan int), make(chan int)
  dll := NewDoublyLinkedList()

  // Producer
  wg := sync.WaitGroup{}
  wg.Add(2)
  go func() {
    for i := 0; i < n; i++ {
      for ok, _ := dll.Append(NewNode(i)); !ok; ok, _ = dll.Append(NewNode(i)) {
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
      for ok, _ := dll.Append(NewNode(i)); !ok; ok, _ = dll.Append(NewNode(i)) {
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
      if ok, _ := dll.Delete(NewNode(i)); ok {
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
      if ok, _ := dll.Delete(NewNode(i)); ok {
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

func TestIterator(t *testing.T) {
  dll := NewDoublyLinkedList()
  _, _ = dll.Append(NewNode(1))
  _, _ = dll.Append(NewNode(2))
  _, _ = dll.Append(NewNode(3))

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
  dll := NewDoublyLinkedList()
  for i := 0; i < b.N; i += 1 {
    _, _ = dll.Append(NewNode(i))
    // assert.Equal(t, i+1, dll.Len())
  }
}

func BenchmarkParallelAppend(b *testing.B) {
  n := 1
  dll := NewDoublyLinkedList()
  b.RunParallel(func(pb *testing.PB) {
    for pb.Next() {
      for j := 0; j < n; j++ {
        for ok, _ := dll.Append(NewNode(j)); !ok; ok, _ = dll.Append(NewNode(j)) {
          // t.Log("not ok")
        }
      }
    }
  })
}

var setThisForBenchmark bool

func BenchmarkParallelRead(b *testing.B) {
  threadsCount := 1000
  readN := 1000
  dll := NewDoublyLinkedList()
  node := NewNode(1)
  _, _ = dll.Append(NewNode(1))
  b.SetParallelism(threadsCount)
  b.RunParallel(func(pb *testing.PB) {
    for pb.Next() {
      for i := 0; i < readN; i++ {
        setThisForBenchmark = dll.Contains(node)
      }
    }
  })
}

type AtomicMap map[string]bool

func BenchmarkMapParallelRead(b *testing.B) {
  threadsCount := 1000
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
