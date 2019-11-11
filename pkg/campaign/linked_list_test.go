package campaign

import (
  "sync"
  "testing"
  "time"
)

var (
  R   = 3                                     // Regions
  M   = 5000                                  // Number of items per map
  L   = 5000                                  // Number of items per list
  TH  = 20000                                 // Thread count
  DUR = time.Duration(100) * time.Millisecond // Campaign update frequency (ms)
)

type sig = uint

const TERM sig = iota

var Result1 int
// Benchmark iterate through maps and copy the value into a Result slice
func BenchmarkCampaignCopyMap(b *testing.B) {
  regionA, regionB, regionC := setupregions()
  // mu := sync.RWMutex{}

  b.ResetTimer()

  for i := 0; i < b.N; i++ {
    results := make(map[int]*Campaign, 3*M)
    for _, c := range regionA {
      results[c.id*R+0] = c
    }
    for _, c := range regionB {
      results[c.id*R+1] = c
    }
    for _, c := range regionC {
      results[c.id*R+2] = c
    }

    Result1 = Result1 + len(results)
  }

  // b.Logf("len: %d", len(results))
}

var Result2 int
// Persistent linked list
// Benchmark iterator through persistent linked list and copy the
// node item reference to a Result slice
func BenchmarkCampaignCopyLinkedList(b *testing.B) {
  regionA := NewCampaignLinkedList()
  regionB := NewCampaignLinkedList()
  regionC := NewCampaignLinkedList()

  for i := 0; i < L; i++ {
    c := &Campaign{id: i}
    _, _ = regionA.Insert(NewCampaignNode(c.id, c))
    _, _ = regionB.Insert(NewCampaignNode(c.id, c))
    _, _ = regionC.Insert(NewCampaignNode(c.id, c))
  }

  // b.Logf("len: %d, %d, %d", regionA.Len(), regionB.Len(), regionC.Len())
  itA := regionA.Iterator()
  itB := regionB.Iterator()
  itC := regionC.Iterator()

  b.ResetTimer()
  for i := 0; i < b.N; i++ {
    results := make([]*Campaign, 0, 3*L)
    for n, ok := itA.Next(); ok; n, ok = itA.Next() {
      results = append(results, n.GetCampaign())
    }

    for n, ok := itB.Next(); ok; n, ok = itB.Next() {
      results = append(results, n.GetCampaign())
    }

    for n, ok := itC.Next(); ok; n, ok = itC.Next() {
      results = append(results, n.GetCampaign())
    }

    Result2 = Result2 + len(results)
  }
  // b.Logf("len: %d", len(result))
}

var Result3 int
// Parallel copying campaign from regions into global result3
// Write access is synchronized with a mutex
func BenchmarkCampaignCopyParallelMapLock(b *testing.B) {
  regionA, regionB, regionC := setupregions()

  mu := sync.RWMutex{}
  for i := 0; i < M; i++ {
    regionA[i*R+0] = &Campaign{id: i*R + 0}
    regionB[i*R+1] = &Campaign{id: i*R + 1}
    regionC[i*R+2] = &Campaign{id: i*R + 2}
  }

  b.SetParallelism(TH)
  b.ResetTimer()
  b.RunParallel(func(pb *testing.PB) {
    for pb.Next() {
      results := make(map[int]*Campaign, 3*M)
      mu.RLock()
      for _, c := range regionA {
        results[c.id*R+0] = c
      }
      for _, c := range regionB {
        results[c.id*R+1] = c
      }
      for _, c := range regionC {
        results[c.id*R+2] = c
      }
      mu.RUnlock()

      mu.Lock()
      Result3 = Result3 + len(results)
      mu.Unlock()
    }
  })
}

var Result4 int
func BenchmarkCampaignCopyParallelList(b *testing.B) {
  regionA := NewCampaignLinkedList()
  regionB := NewCampaignLinkedList()
  regionC := NewCampaignLinkedList()

  for i := 0; i < L; i++ {
    c := &Campaign{id: i}
    _, _ = regionA.Insert(NewCampaignNode(c.id, c))
    _, _ = regionB.Insert(NewCampaignNode(c.id, c))
    _, _ = regionC.Insert(NewCampaignNode(c.id, c))
  }

  mu := sync.Mutex{}

  itA := regionA.Iterator()
  itB := regionB.Iterator()
  itC := regionC.Iterator()

  b.SetParallelism(TH)
  b.ResetTimer()
  b.RunParallel(func(pb *testing.PB) {
    for pb.Next() {
      results := make([]*Campaign, 0, L*R)

      for n, ok := itA.Next(); ok; n, ok = itA.Next() {
        results = append(results, n.GetCampaign())
      }

      for n, ok := itB.Next(); ok; n, ok = itB.Next() {
        results = append(results, n.GetCampaign())
      }

      for n, ok := itC.Next(); ok; n, ok = itC.Next() {
        results = append(results, n.GetCampaign())
      }

      mu.Lock()
      Result4 = Result4 + len(results)
      mu.Unlock()
    }
  })
}

var Result5 int

func BenchmarkCampaignCopyParallelMapLockWithCampaignInsert(b *testing.B) {
  regionA, regionB, regionC := setupregions()
  done := make(chan sig)

  mu := sync.RWMutex{}
  // Every `DUR` interval, take the `mu` lock
  // an add new campaigns into all regions
  go func() {
    i := L + 1
    for {
      select {
      case <-time.Tick(DUR):
        id := i
        c1, c2, c3 := &Campaign{id: id*R + 0}, &Campaign{id: id*R + 1}, &Campaign{id: id*R + 2}
        mu.Lock()
        regionA[id*R+0] = c1
        regionB[id*R+1] = c2
        regionC[id*R+2] = c3
        mu.Unlock()
        i += 1
      case <-done:
        break
      }
    }
  }()

  b.SetParallelism(TH)
  b.ResetTimer()
  b.RunParallel(func(pb *testing.PB) {
    for pb.Next() {
      results := make(map[int]*Campaign)
      mu.RLock()
      for _, c := range regionA {
        results[c.id] = c
      }
      for _, c := range regionB {
        results[c.id] = c
      }
      for _, c := range regionC {
        results[c.id] = c
      }
      mu.RUnlock()

      mu.Lock()
      Result5 = Result5 + len(results)
      mu.Unlock()
    }
  })

  done <- TERM
}

var Result6 int

func BenchmarkCampaignCopyParallelListWithCampaignInsert(b *testing.B) {
  regionA := NewCampaignLinkedList()
  regionB := NewCampaignLinkedList()
  regionC := NewCampaignLinkedList()

  for i := 0; i < L; i++ {
    c := &Campaign{id: i * R}
    _, _ = regionA.Insert(NewCampaignNode(c.id, c))
    _, _ = regionB.Insert(NewCampaignNode(c.id, c))
    _, _ = regionC.Insert(NewCampaignNode(c.id, c))
  }

  mu := sync.Mutex{}
  done := make(chan sig)

  go func() {
    i := L + 1
    for {
      select {
      case <-time.Tick(DUR):
        id := i
        c := &Campaign{id: id}
        _, _ = regionA.Insert(NewCampaignNode(c.id*R+0, c))
        _, _ = regionB.Insert(NewCampaignNode(c.id*R+1, c))
        _, _ = regionC.Insert(NewCampaignNode(c.id*R+2, c))
        i += 1
      case <-done:
        break
      }
    }
  }()

  itA := regionA.Iterator()
  itB := regionB.Iterator()
  itC := regionC.Iterator()

  b.SetParallelism(TH)
  b.ResetTimer()
  b.RunParallel(func(pb *testing.PB) {
    for pb.Next() {
      results := make([]*Campaign, 0, L*R)

      for n, ok := itA.Next(); ok; n, ok = itA.Next() {
        results = append(results, n.GetCampaign())
      }
      for n, ok := itB.Next(); ok; n, ok = itB.Next() {
        results = append(results, n.GetCampaign())
      }
      for n, ok := itC.Next(); ok; n, ok = itC.Next() {
        results = append(results, n.GetCampaign())
      }

      mu.Lock()
      Result6 = Result6 + len(results)
      mu.Unlock()
    }
  })

  done <- TERM
}

// region Private
func setupregions() (map[int]*Campaign, map[int]*Campaign, map[int]*Campaign) {
  regionA := make(map[int]*Campaign, M)
  regionB := make(map[int]*Campaign, M)
  regionC := make(map[int]*Campaign, M)

  for i := 0; i < M; i++ {
    regionA[i*R+0] = &Campaign{id: i*R + 0}
    regionB[i*R+1] = &Campaign{id: i*R + 1}
    regionC[i*R+2] = &Campaign{id: i*R + 2}
  }

  return regionA, regionB, regionC
}

// endregion
