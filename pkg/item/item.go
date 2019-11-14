package item

import (
	"go.uber.org/atomic"
)

type Item struct {
	Id          int
	AccessCount *atomic.Int64
}

func (it *Item) ID() uint64 {
	return uint64(it.Id)
}
