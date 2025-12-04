package memorystatistics

import (
	"fmt"
	"runtime"
	"strings"
	"sync"

	supfunc "github.com/av-belyakov/enricher_zabbix_information/internal/supportingfunctions"
)

var (
	once          sync.Once
	memCache      *MemoryCache      = nil
	memStatsCache *MemoryStatsCache = nil
)

func NewMemoryStats() *MemoryStatsCache {
	once.Do(func() {
		memStatsCache = new(MemoryStatsCache)
	})

	return memStatsCache
}

func GetMemoryStats() MemoryStatsCache {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	alloc := memStats.Alloc
	gcSys := memStats.GCSys
	heapSys := memStats.HeapSys
	heapAlloc := memStats.HeapAlloc
	numLiveObj := memStats.Mallocs - memStats.Frees
	returnedOS := memStats.HeapIdle - memStats.HeapReleased
	totalAlloc := memStats.TotalAlloc
	heapObjects := memStats.HeapObjects

	memStatsCache = NewMemoryStats()
	memStatsCache.Alloc = MemoryStatsValues{
		PointerUpOrDown: supfunc.GetPointerUpOrDown(memStatsCache.Alloc.Previous, alloc),
		Previous:        memStatsCache.Alloc.Current,
		Current:         alloc,
	}
	memStatsCache.HeapSys = MemoryStatsValues{
		PointerUpOrDown: supfunc.GetPointerUpOrDown(memStatsCache.HeapSys.Previous, heapSys),
		Previous:        memStatsCache.HeapSys.Current,
		Current:         heapSys,
	}
	memStatsCache.HeapAlloc = MemoryStatsValues{
		PointerUpOrDown: supfunc.GetPointerUpOrDown(memStatsCache.HeapAlloc.Previous, heapAlloc),
		Previous:        memStatsCache.HeapAlloc.Current,
		Current:         heapAlloc,
	}
	memStatsCache.TotalAlloc = MemoryStatsValues{
		PointerUpOrDown: supfunc.GetPointerUpOrDown(memStatsCache.TotalAlloc.Previous, totalAlloc),
		Previous:        memStatsCache.TotalAlloc.Current,
		Current:         totalAlloc,
	}
	memStatsCache.HeapObjects = MemoryStatsValues{
		PointerUpOrDown: supfunc.GetPointerUpOrDown(memStatsCache.HeapObjects.Previous, heapObjects),
		Previous:        memStatsCache.HeapObjects.Current,
		Current:         heapObjects,
	}
	memStatsCache.NumberLiveObjects = MemoryStatsValues{
		PointerUpOrDown: supfunc.GetPointerUpOrDown(memStatsCache.NumberLiveObjects.Previous, numLiveObj),
		Previous:        memStatsCache.NumberLiveObjects.Current,
		Current:         numLiveObj,
	}
	memStatsCache.CountMemoryReturned = MemoryStatsValues{
		PointerUpOrDown: supfunc.GetPointerUpOrDown(memStatsCache.CountMemoryReturned.Previous, returnedOS),
		Previous:        memStatsCache.CountMemoryReturned.Current,
		Current:         returnedOS,
	}
	memStatsCache.GarbagecollectorMemory = MemoryStatsValues{
		PointerUpOrDown: supfunc.GetPointerUpOrDown(memStatsCache.GarbagecollectorMemory.Previous, gcSys),
		Previous:        memStatsCache.GarbagecollectorMemory.Current,
		Current:         gcSys,
	}

	return *memStatsCache
}

func NewMemoryCache() *MemoryCache {
	once.Do(func() {
		memCache = new(MemoryCache)
	})

	return memCache
}

// PrintMemStats вывод информации по потребляемой памяти
func PrintMemStats() string {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	memCache = NewMemoryCache()

	alloc := memStats.Alloc
	gcSys := memStats.GCSys
	heapSys := memStats.HeapSys
	heapAlloc := memStats.HeapAlloc
	numLiveObj := memStats.Mallocs - memStats.Frees
	returnedOS := memStats.HeapIdle - memStats.HeapReleased
	totalAlloc := memStats.TotalAlloc
	heapObjects := memStats.HeapObjects

	str := strings.Builder{}

	str.WriteString(fmt.Sprintf("Allocated Memory: %v bytes %s\n", alloc, supfunc.GetPointerUpOrDown(memCache.Alloc, alloc)))
	str.WriteString(fmt.Sprintf("Total Allocated Memory: %v bytes %s\n", totalAlloc, supfunc.GetPointerUpOrDown(memCache.TotalAlloc, totalAlloc)))
	str.WriteString(fmt.Sprintf("Heap Alloc Memory: %v bytes %s\n", heapAlloc, supfunc.GetPointerUpOrDown(memCache.HeapAlloc, heapAlloc)))
	str.WriteString(fmt.Sprintf("Heap System Memory: %v bytes %s\n", heapSys, supfunc.GetPointerUpOrDown(memCache.HeapSys, heapSys)))
	str.WriteString(fmt.Sprintf("The number of allocated heap objects: %v bytes %s\n", heapObjects, supfunc.GetPointerUpOrDown(memCache.HeapObjects, heapObjects)))
	str.WriteString(fmt.Sprintf("The number of live objects: %v bytes %s\n", numLiveObj, supfunc.GetPointerUpOrDown(memCache.NumberLiveObjects, numLiveObj)))
	str.WriteString(fmt.Sprintf("Count memory that could be returned to the OS: %v bytes %s\n", returnedOS, supfunc.GetPointerUpOrDown(memCache.CountMemoryReturned, returnedOS)))
	str.WriteString(fmt.Sprintf("Garbage Collector Memory: %v bytes %s\n", gcSys, supfunc.GetPointerUpOrDown(memCache.GarbagecollectorMemory, gcSys)))

	memCache.Alloc = alloc
	memCache.HeapSys = heapSys
	memCache.HeapAlloc = heapAlloc
	memCache.TotalAlloc = totalAlloc
	memCache.HeapObjects = heapObjects
	memCache.NumberLiveObjects = numLiveObj
	memCache.CountMemoryReturned = returnedOS
	memCache.GarbagecollectorMemory = gcSys

	return str.String()
}
