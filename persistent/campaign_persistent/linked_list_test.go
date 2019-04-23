package campaign_persistent

import (
  "sync"
  "testing"
  "time"
)

var (
  R = 3                  // Regions
  N = 500                // Number of items per map
  M = 1000               // Number of items per list
  T = 20000              // Thread count
  D = time.Duration(100) // Campaign update frequency (ms)
)

var result1 map[int]*Campaign
var result2 [][]*Campaign
var result3 []map[int]*Campaign
var result4 [][]*Campaign

func BenchmarkCampaignCopyMap(b *testing.B) {
  regionA := make(map[int]*Campaign, N)
  regionB := make(map[int]*Campaign, N)
  regionC := make(map[int]*Campaign, N)

  mu := sync.RWMutex{}

  for i := 0; i < N; i++ {
    regionA[i*R+0] = &Campaign{id: i*R + 0}
    regionB[i*R+1] = &Campaign{id: i*R + 1}
    regionC[i*R+2] = &Campaign{id: i*R + 2}
  }

  b.ResetTimer()

  for i := 0; i < b.N; i++ {
    result1 = make(map[int]*Campaign)
    mu.RLock()
    for _, c := range regionA {
      result1[c.id] = c
    }
    for _, c := range regionB {
      result1[c.id] = c
    }
    for _, c := range regionC {
      result1[c.id] = c
    }
    mu.RUnlock()
  }
  // b.Logf("len: %d", len(result1))
}

// func TestCampaignCopyList(t *testing.T) {
//   regionA := NewCampaignLinkedList()
//   regionB := NewCampaignLinkedList()
//   regionC := NewCampaignLinkedList()
//
//   for i := 0; i < M; i++ {
//     c := &Campaign{id: i}
//     _, _ = regionA.Insert(NewCampaignNode(c.id, c))
//     _, _ = regionB.Insert(NewCampaignNode(c.id, c))
//     _, _ = regionC.Insert(NewCampaignNode(c.id, c))
//   }
//
//   itA := regionA.CyclicIterator()
//   itB := regionB.CyclicIterator()
//   itC := regionC.CyclicIterator()
//
//   t.Logf("len: %d, %d, %d", regionA.Len(), regionB.Len(), regionC.Len())
//
//   repeats := 10
//   result2 = make([][]*Campaign, 0, repeats)
//
//   for i := 0; i < repeats; i++ {
//     result := make([]*Campaign, 0, M)
//     j := 0
//     for n, ok := itA.Next(); ok && j < M; n, ok = itA.Next() {
//       result = append(result, n.GetCampaign())
//       j += 1
//     }
//
//     j = 0
//     for n, ok := itB.Next(); ok && j < M; n, ok = itB.Next() {
//       result = append(result, n.GetCampaign())
//       j += 1
//     }
//
//     j = 0
//     for n, ok := itC.Next(); ok && j < M; n, ok = itC.Next() {
//       result = append(result, n.GetCampaign())
//       j += 1
//     }
//
//     result2 = append(result2, result)
//   }
//
//   t.Logf("len: %d", len(result2))
//   assert.Equal(t, repeats, len(result2))
//
//   count := 0
//   for _, list := range result2 {
//     count += len(list)
//   }
//
//   assert.Equal(t, repeats*M*R, count)
// }

func BenchmarkCampaignCopyList(b *testing.B) {
  regionA := NewCampaignLinkedList()
  regionB := NewCampaignLinkedList()
  regionC := NewCampaignLinkedList()

  for i := 0; i < M; i++ {
    c := &Campaign{id: i}
    _, _ = regionA.Insert(NewCampaignNode(c.id, c))
    _, _ = regionB.Insert(NewCampaignNode(c.id, c))
    _, _ = regionC.Insert(NewCampaignNode(c.id, c))
  }

  // b.Logf("len: %d, %d, %d", regionA.Len(), regionB.Len(), regionC.Len())
  result2 = make([][]*Campaign, 0, b.N)
  itA := regionA.CyclicIterator()
  itB := regionB.CyclicIterator()
  itC := regionC.CyclicIterator()

  b.ResetTimer()
  for i := 0; i < b.N; i++ {
    result := make([]*Campaign, 0, M*R)
    j := 0

    for n, ok := itA.Next(); ok && j < M; n, ok = itA.Next() {
      result = append(result, n.GetCampaign())
      j += 1
    }

    j = 0
    for n, ok := itB.Next(); ok && j < M; n, ok = itB.Next() {
      result = append(result, n.GetCampaign())
      j += 1
    }

    j = 0
    for n, ok := itC.Next(); ok && j < M; n, ok = itC.Next() {
      result = append(result, n.GetCampaign())
      j += 1
    }

    result2 = append(result2, result)
  }
  // b.Logf("len: %d", len(result2))
}

func BenchmarkCampaignCopyParallelMap(b *testing.B) {
  regionA := make(map[int]*Campaign, N)
  regionB := make(map[int]*Campaign, N)
  regionC := make(map[int]*Campaign, N)

  mu := sync.RWMutex{}

  for i := 0; i < N; i++ {
    regionA[i*R+0] = &Campaign{id: i*R + 0}
    regionB[i*R+1] = &Campaign{id: i*R + 1}
    regionC[i*R+2] = &Campaign{id: i*R + 2}
  }

  result3 = make([]map[int]*Campaign, 0, b.N)
  b.SetParallelism(T)
  b.ResetTimer()
  b.RunParallel(func(pb *testing.PB) {
    for pb.Next() {
      result := make(map[int]*Campaign)
      mu.RLock()
      for _, c := range regionA {
        result[c.id] = c
      }
      for _, c := range regionB {
        result[c.id] = c
      }
      for _, c := range regionC {
        result[c.id] = c
      }
      mu.RUnlock()

      mu.Lock()
      result3 = append(result3, result)
      mu.Unlock()
    }
  })
}

func BenchmarkCampaignCopyParallelList(b *testing.B) {
  regionA := NewCampaignLinkedList()
  regionB := NewCampaignLinkedList()
  regionC := NewCampaignLinkedList()

  for i := 0; i < M; i++ {
    c := &Campaign{id: i}
    _, _ = regionA.Insert(NewCampaignNode(c.id, c))
    _, _ = regionB.Insert(NewCampaignNode(c.id, c))
    _, _ = regionC.Insert(NewCampaignNode(c.id, c))
  }

  mu := sync.Mutex{}
  result4 = make([][]*Campaign, 0, b.N)

  itA := regionA.CyclicIterator()
  itB := regionB.CyclicIterator()
  itC := regionC.CyclicIterator()

  b.SetParallelism(T)
  b.ResetTimer()
  b.RunParallel(func(pb *testing.PB) {
    for pb.Next() {
      result := make([]*Campaign, 0, M*R)

      j := 0
      for n, ok := itA.Next(); ok && j < M; n, ok = itA.Next() {
        result = append(result, n.GetCampaign())
        j += 1
      }

      j = 0
      for n, ok := itB.Next(); ok && j < M; n, ok = itB.Next() {
        result = append(result, n.GetCampaign())
        j += 1
      }

      j = 0
      for n, ok := itC.Next(); ok && j < M; n, ok = itC.Next() {
        result = append(result, n.GetCampaign())
        j += 1
      }

      mu.Lock()
      result4 = append(result4, result)
      mu.Unlock()
    }
  })
}

func BenchmarkCampaignCopyParallelMapWithCampaignUpdate(b *testing.B) {
  regionA := make(map[int]*Campaign, N)
  regionB := make(map[int]*Campaign, N)
  regionC := make(map[int]*Campaign, N)

  mu := sync.RWMutex{}

  for i := 0; i < N; i++ {
    regionA[i*R+0] = &Campaign{id: i*R + 0}
    regionB[i*R+1] = &Campaign{id: i*R + 1}
    regionC[i*R+2] = &Campaign{id: i*R + 2}
  }

  doneChan := make(chan int)
  // flushChan := make(chan map[int]*Campaign, b.N)

  go func() {
    timer := time.NewTimer(D * time.Millisecond)
    i := M + 1
    for {
      select {
      case <-timer.C:
        id := i
        i += 1
        c1, c2, c3 := &Campaign{id: id*R + 0}, &Campaign{id: id*R + 1}, &Campaign{id: id*R + 2}
        mu.Lock()
        regionA[id] = c1
        regionB[id] = c2
        regionC[id] = c3
        mu.Unlock()
      case <-doneChan:
        break
      }
    }
  }()

  // go func() {
  //   for {
  //     select {
  //     case result := <-flushChan:
  //       result3 = append(result3, result)
  //     case <-doneChan:
  //       break
  //     }
  //   }
  // }()

  result3 = make([]map[int]*Campaign, 0, b.N)
  b.SetParallelism(T)
  b.ResetTimer()
  b.RunParallel(func(pb *testing.PB) {
    for pb.Next() {
      result := make(map[int]*Campaign)
      mu.RLock()
      for _, c := range regionA {
        result[c.id] = c
      }
      for _, c := range regionB {
        result[c.id] = c
      }
      for _, c := range regionC {
        result[c.id] = c
      }
      mu.RUnlock()

      // flushChan <- result

      mu.Lock()
      result3 = append(result3, result)
      mu.Unlock()
    }
  })

  doneChan <- 1
}

func BenchmarkCampaignCopyParallelListWithCampaignUpdate(b *testing.B) {
  regionA := NewCampaignLinkedList()
  regionB := NewCampaignLinkedList()
  regionC := NewCampaignLinkedList()

  for i := 0; i < M; i++ {
    c := &Campaign{id: i * R}
    _, _ = regionA.Insert(NewCampaignNode(c.id, c))
    _, _ = regionB.Insert(NewCampaignNode(c.id, c))
    _, _ = regionC.Insert(NewCampaignNode(c.id, c))
  }

  mu := sync.Mutex{}

  doneChan := make(chan int)

  // flushChan := make(chan []*Campaign, b.N)
  result4 = make([][]*Campaign, 0, b.N)

  go func() {
    timer := time.NewTimer(D * time.Millisecond)
    i := M + 1
    for {
      select {
      case <-timer.C:
        id := i
        i += 1
        c := &Campaign{id: id}
        _, _ = regionA.Insert(NewCampaignNode(c.id, c))
        _, _ = regionB.Insert(NewCampaignNode(c.id, c))
        _, _ = regionC.Insert(NewCampaignNode(c.id, c))
      case <-doneChan:
        break
      }
    }
  }()

  // go func() {
  //   for {
  //     select {
  //     case result := <-flushChan:
  //       result4 = append(result4, result)
  //     case <-doneChan:
  //       break
  //     }
  //   }
  // }()

  itA := regionA.CyclicIterator()
  itB := regionB.CyclicIterator()
  itC := regionC.CyclicIterator()

  b.SetParallelism(T)
  b.ResetTimer()
  b.RunParallel(func(pb *testing.PB) {
    for pb.Next() {
      result := make([]*Campaign, 0, M*R)

      j := 0
      for n, ok := itA.Next(); ok && j < M; n, ok = itA.Next() {
        result = append(result, n.GetCampaign())
        j += 1
      }

      j = 0
      for n, ok := itB.Next(); ok && j < M; n, ok = itB.Next() {
        result = append(result, n.GetCampaign())
        j += 1
      }

      j = 0
      for n, ok := itC.Next(); ok && j < M; n, ok = itC.Next() {
        result = append(result, n.GetCampaign())
        j += 1
      }

      // flushChan <- result

      mu.Lock()
      result4 = append(result4, result)
      mu.Unlock()
    }
  })

  doneChan <- 1
}
