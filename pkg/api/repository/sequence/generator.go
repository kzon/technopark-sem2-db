package sequence

import "sync/atomic"

type Generator struct {
	current int32
}

func NewGenerator() Generator {
	return Generator{}
}

func (s *Generator) Next(count int) []int {
	result := make([]int, 0, count)
	for i := 0; i < count; i++ {
		result = append(result, int(atomic.AddInt32(&s.current, 1)))
	}
	return result
}
