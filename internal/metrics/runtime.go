package metrics

import (
	"runtime"
)

func (c *collector) CollectRuntimeMetrics() {
	var (
		memstats   runtime.MemStats
		goroutines = runtime.NumGoroutine()
	)

	runtime.ReadMemStats(&memstats)

	c.mGoroutines.Set(float64(goroutines))
	c.mHeapAlloc.Set(float64(memstats.HeapAlloc))
	c.mSys.Set(float64(memstats.Sys))
	c.mNumGCs.Set(float64(memstats.NumGC))
	c.mPauseTotalNano.Set(float64(memstats.PauseTotalNs))
}
