package memorystatistics

type MemoryCache struct {
	Alloc                  uint64
	HeapSys                uint64
	HeapAlloc              uint64
	TotalAlloc             uint64
	HeapObjects            uint64
	NumberLiveObjects      uint64
	CountMemoryReturned    uint64
	GarbagecollectorMemory uint64
}
