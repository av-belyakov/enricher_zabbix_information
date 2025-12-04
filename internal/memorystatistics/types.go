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

type MemoryStatsCache struct {
	Alloc                  MemoryStatsValues
	HeapSys                MemoryStatsValues
	HeapAlloc              MemoryStatsValues
	TotalAlloc             MemoryStatsValues
	HeapObjects            MemoryStatsValues
	NumberLiveObjects      MemoryStatsValues
	CountMemoryReturned    MemoryStatsValues
	GarbagecollectorMemory MemoryStatsValues
}

type MemoryStatsValues struct {
	Previous        uint64
	Current         uint64
	PointerUpOrDown string
}
