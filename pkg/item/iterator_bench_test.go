package item

import (
	"testing"

)

func BenchmarkCyclicIterator(b *testing.B) {
	list := NewItemLinkedList()
	for i := 0; i < 10; i++ {
		_, _ = list.Insert(&Item{Id: i})
	}

	citr := list.CyclicIterator()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		citr.Next()
	}
}
