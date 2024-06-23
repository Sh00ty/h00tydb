package perf

import (
	"sync"

	"github.com/influxdata/tdigest"
)

type guardedTdigest struct {
	td tdigest.TDigest
	mu sync.Mutex
}

func (td *guardedTdigest) Add(x float64) {
	td.mu.Lock()
	td.td.Add(x, 1)
	td.mu.Unlock()
}

func (td *guardedTdigest) Percentile(percentiles []float64) []float64 {
	res := make([]float64, 0, len(percentiles))
	td.mu.Lock()
	for _, p := range percentiles {
		res = append(res, td.td.Quantile(p))
	}
	td.mu.Unlock()
	return res
}
