package util

const (
	UintSize = 32 << (^uint(0) >> 32 & 1) // 32 or 64
	MaxInt   = 1<<(UintSize-1) - 1        // 1<<31 - 1 or 1<<63 - 1
)
