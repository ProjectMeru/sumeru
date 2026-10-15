// Package metrics provides lightweight Prometheus-style counters and histograms.
package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	mu        sync.Mutex
	counters  = map[string]*uint64{}
	histSum   = map[string]*uint64{} // microseconds
	histCount = map[string]*uint64{}
)

func counter(name string) *uint64 {
	mu.Lock()
	defer mu.Unlock()
	if p, ok := counters[name]; ok {
		return p
	}
	var v uint64
	counters[name] = &v
	return &v
}

func hist(name string) (*uint64, *uint64) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := histSum[name]; !ok {
		var s, c uint64
		histSum[name] = &s
		histCount[name] = &c
	}
	return histSum[name], histCount[name]
}

// Inc increments a named counter.
func Inc(name string) {
	atomic.AddUint64(counter(name), 1)
}

// ObserveDuration records a duration sample in microseconds for histogram-like sums.
func ObserveDuration(name string, d time.Duration) {
	s, c := hist(name)
	atomic.AddUint64(s, uint64(d.Microseconds()))
	atomic.AddUint64(c, 1)
}

// Handler writes Prometheus text exposition format.
func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	mu.Lock()
	names := make([]string, 0, len(counters))
	for n := range counters {
		names = append(names, n)
	}
	sort.Strings(names)
	var b strings.Builder
	for _, n := range names {
		fmt.Fprintf(&b, "# TYPE %s counter\n%s %d\n", n, n, atomic.LoadUint64(counters[n]))
	}
	hnames := make([]string, 0, len(histSum))
	for n := range histSum {
		hnames = append(hnames, n)
	}
	sort.Strings(hnames)
	for _, n := range hnames {
		sum := atomic.LoadUint64(histSum[n])
		cnt := atomic.LoadUint64(histCount[n])
		fmt.Fprintf(&b, "# TYPE %s summary\n%s_sum %f\n%s_count %d\n", n, n, float64(sum)/1e6, n, cnt)
	}
	mu.Unlock()
	_, _ = w.Write([]byte(b.String()))
}
