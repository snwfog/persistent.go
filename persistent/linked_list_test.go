package persistent

import (
  "runtime"
  "sync"
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

  // n1, n2 := NewNode(1), NewNode(1)
  // t.Logf("%d, %d", n1.key, n2.key)

  _, _ = dll.Insert(NewNode(1))
  assert.Equal(t, 1, dll.Len())

  _, _ = dll.Insert(NewNode(1))
  assert.Equal(t, 1, dll.Len())
}

func TestInsert1(t *testing.T) {
  dll := NewLinkedList()

  n1, n2 := NewNode(1), NewNode(2)
  // t.Logf("%d, %d", n1.key, n2.key)

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

func TestConcurrentInsertDeleteN1(t *testing.T) {
  n := 1 << 12
  pc := 100
  // pc := 20000

  dll := NewLinkedList()
  channel := make(chan int)
  doneChan := make(chan int)

  wg := sync.WaitGroup{}
  wg.Add(pc)

  // Producers
  for i := 0; i < pc; i++ {
    go func() {
      for j := 0; j < n; j++ {
        n := NewNode(j)
        _, _ = dll.Insert(n)
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
      n := NewNode(j)
      _, _ = dll.Delete(n)
      // t.Logf("c[1] %d, hash %d, len: %d, ok: %v", k, n.key, dll.Len(), ok)
    }

    doneChan <- 1
  }()

  wg.Wait()
  close(doneChan)
  <-doneChan

  for ; n>>1 != dll.Len(); {
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
