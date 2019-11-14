package int

import (
	"testing"

	"github.com/snwfog/persistent.go/pkg"
)

func BenchmarkCyclicIterator(b *testing.B) {
	list := pkg.NewIntLinkedList()
	for i := 0; i < 10; i++ {
		n := i
		_, _ = list.Insert(pkg.NewIntNode(&n))
	}

	citr := list.CyclicIterator()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		citr.Next()
	}
}
