package campaign_persistent

import (
  "math/rand"
  "sync"
  "testing"
  "time"
)

var (
  N = 5000
  T = 20000
  D = time.Duration(100)
)

var result1 map[int]*Campaign
var result2 []*Campaign
var result3 []map[int]*Campaign
var result4 [][]*Campaign

func BenchmarkCampaignCopyMap(b *testing.B) {
  regionA := make(map[int]*Campaign, N)
  regionB := make(map[int]*Campaign, N)
  regionC := make(map[int]*Campaign, N)

  mu := sync.RWMutex{}

  for i := 0; i < N; i++ {
    regionA[i] = &Campaign{id: i}
    regionB[i] = &Campaign{id: i}
    regionC[i] = &Campaign{id: i}
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
//   for i := 0; i < N; i++ {
//     c := &Campaign{id: i}
//     _, _ = regionA.Insert(NewCampaignNode(c.id, c))
//     _, _ = regionB.Insert(NewCampaignNode(c.id, c))
//     _, _ = regionC.Insert(NewCampaignNode(c.id, c))
//   }
//
//   itA := regionA.Iterator()
//   itB := regionB.Iterator()
//   itC := regionC.Iterator()
//
//   t.Logf("len: %d, %d, %d", regionA.Len(), regionB.Len(), regionC.Len())
//
//   result2 = make([]*Campaign, 0)
//   for n, ok := itA.Next(); ok; n, ok = itA.Next() {
//     result2 = append(result2, n.GetCampaign())
//   }
//
//   for n, ok := itB.Next(); ok; n, ok = itB.Next() {
//     result2 = append(result2, n.GetCampaign())
//   }
//
//   for n, ok := itC.Next(); ok; n, ok = itC.Next() {
//     result2 = append(result2, n.GetCampaign())
//   }
//
//   t.Logf("len: %d", len(result2))
//   assert.Equal(t, N*3, len(result2))
// }

func BenchmarkCampaignCopyList(b *testing.B) {
  regionA := NewCampaignLinkedList()
  regionB := NewCampaignLinkedList()
  regionC := NewCampaignLinkedList()

  for i := 0; i < N; i++ {
    c := &Campaign{id: i}
    _, _ = regionA.Insert(NewCampaignNode(c.id, c))
    _, _ = regionB.Insert(NewCampaignNode(c.id, c))
    _, _ = regionC.Insert(NewCampaignNode(c.id, c))
  }

  // b.Logf("len: %d, %d, %d", regionA.Len(), regionB.Len(), regionC.Len())

  b.ResetTimer()
  for i := 0; i < b.N; i++ {
    result2 = make([]*Campaign, N*3)

    itA := regionA.Iterator()
    itB := regionB.Iterator()
    itC := regionC.Iterator()

    for n, ok := itA.Next(); ok; n, ok = itA.Next() {
      result2 = append(result2, n.GetCampaign())
    }

    for n, ok := itB.Next(); ok; n, ok = itB.Next() {
      result2 = append(result2, n.GetCampaign())
    }

    for n, ok := itC.Next(); ok; n, ok = itC.Next() {
      result2 = append(result2, n.GetCampaign())
    }
  }

  // b.Logf("len: %d", len(result2))
}

func BenchmarkCampaignCopyParallelMap(b *testing.B) {
  regionA := make(map[int]*Campaign, N)
  regionB := make(map[int]*Campaign, N)
  regionC := make(map[int]*Campaign, N)

  mu := sync.RWMutex{}

  for i := 0; i < N; i++ {
    regionA[i] = &Campaign{id: i}
    regionB[i] = &Campaign{id: i}
    regionC[i] = &Campaign{id: i}
  }

  result3 = make([]map[int]*Campaign, 0)
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

  for i := 0; i < N; i++ {
    c := &Campaign{id: i}
    _, _ = regionA.Insert(NewCampaignNode(c.id, c))
    _, _ = regionB.Insert(NewCampaignNode(c.id, c))
    _, _ = regionC.Insert(NewCampaignNode(c.id, c))
  }

  mu := sync.Mutex{}

  b.SetParallelism(T)
  b.ResetTimer()
  b.RunParallel(func(pb *testing.PB) {
    for pb.Next() {
      result := make([]*Campaign, N*3)

      itA := regionA.Iterator()
      itB := regionB.Iterator()
      itC := regionC.Iterator()

      for n, ok := itA.Next(); ok; n, ok = itA.Next() {
        result = append(result, n.GetCampaign())
      }

      for n, ok := itB.Next(); ok; n, ok = itB.Next() {
        result = append(result, n.GetCampaign())
      }

      for n, ok := itC.Next(); ok; n, ok = itC.Next() {
        result = append(result, n.GetCampaign())
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
    regionA[i] = &Campaign{id: i}
    regionB[i] = &Campaign{id: i}
    regionC[i] = &Campaign{id: i}
  }

  doneChan := make(chan int)
  flushChan := make(chan map[int]*Campaign, N*3)

  go func() {
    timer := time.NewTimer(D * time.Millisecond)
    for {
      select {
      case <-timer.C:
        id := rand.Int()
        c := &Campaign{id: id}
        mu.Lock()
        regionA[id] = c
        regionB[id] = c
        regionC[id] = c
        mu.Unlock()
      case <-doneChan:
        break
      }
    }
  }()

  go func() {
    for {
      select {
      case result := <-flushChan:
        result3 = append(result3, result)
      case <-doneChan:
        break
      }
    }
  }()

  result3 = make([]map[int]*Campaign, 0)
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

      flushChan <- result

      // mu.Lock()
      // result3 = append(result3, result)
      // mu.Unlock()
    }
  })

  doneChan <- 1
}

func BenchmarkCampaignCopyParallelListWithCampaignUpdate(b *testing.B) {
  regionA := NewCampaignLinkedList()
  regionB := NewCampaignLinkedList()
  regionC := NewCampaignLinkedList()

  for i := 0; i < N; i++ {
    c := &Campaign{id: i}
    _, _ = regionA.Insert(NewCampaignNode(c.id, c))
    _, _ = regionB.Insert(NewCampaignNode(c.id, c))
    _, _ = regionC.Insert(NewCampaignNode(c.id, c))
  }
  // mu := sync.Mutex{}

  doneChan := make(chan int)
  flushChan := make(chan []*Campaign, N*3)

  go func() {
    timer := time.NewTimer(D * time.Millisecond)
    for {
      select {
      case <-timer.C:
        id := rand.Int()
        c := &Campaign{id: id}
        _, _ = regionA.Insert(NewCampaignNode(c.id, c))
        _, _ = regionB.Insert(NewCampaignNode(c.id, c))
        _, _ = regionC.Insert(NewCampaignNode(c.id, c))
      case <-doneChan:
        break
      }
    }
  }()

  go func() {
    for {
      select {
      case result := <-flushChan:
        result4 = append(result4, result)
      case <-doneChan:
        break
      }
    }
  }()

  b.SetParallelism(T)
  b.ResetTimer()
  b.RunParallel(func(pb *testing.PB) {
    for pb.Next() {
      result := make([]*Campaign, N*3)

      itA := regionA.Iterator()
      itB := regionB.Iterator()
      itC := regionC.Iterator()

      for n, ok := itA.Next(); ok; n, ok = itA.Next() {
        result = append(result, n.GetCampaign())
      }

      for n, ok := itB.Next(); ok; n, ok = itB.Next() {
        result = append(result, n.GetCampaign())
      }

      for n, ok := itC.Next(); ok; n, ok = itC.Next() {
        result = append(result, n.GetCampaign())
      }

      flushChan <- result

      // mu.Lock()
      // result4 = append(result4, result)
      // mu.Unlock()
    }
  })
}
