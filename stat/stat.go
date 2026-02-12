package stat

import (
	"sync"
	"sync/atomic"
	"time"
)

type Counter struct {
	value int64
}

func NewCounter() *Counter {
	return &Counter{}
}

func (c *Counter) Inc() {
	atomic.AddInt64(&c.value, 1)
}

func (c *Counter) Add(delta int64) {
	atomic.AddInt64(&c.value, delta)
}

func (c *Counter) Value() int64 {
	return atomic.LoadInt64(&c.value)
}

func (c *Counter) Reset() int64 {
	return atomic.SwapInt64(&c.value, 0)
}

type Gauge struct {
	value int64
}

func NewGauge() *Gauge {
	return &Gauge{}
}

func (g *Gauge) Set(v int64) {
	atomic.StoreInt64(&g.value, v)
}

func (g *Gauge) Inc() {
	atomic.AddInt64(&g.value, 1)
}

func (g *Gauge) Dec() {
	atomic.AddInt64(&g.value, -1)
}

func (g *Gauge) Add(delta int64) {
	atomic.AddInt64(&g.value, delta)
}

func (g *Gauge) Value() int64 {
	return atomic.LoadInt64(&g.value)
}

type Histogram struct {
	mu     sync.Mutex
	count  int64
	sum    float64
	min    float64
	max    float64
	values []float64
	maxLen int
}

func NewHistogram(maxLen int) *Histogram {
	if maxLen <= 0 {
		maxLen = 1000
	}
	return &Histogram{
		maxLen: maxLen,
		values: make([]float64, 0, maxLen),
	}
}

func (h *Histogram) Observe(v float64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.count++
	h.sum += v

	if h.count == 1 || v < h.min {
		h.min = v
	}
	if v > h.max {
		h.max = v
	}

	if len(h.values) < h.maxLen {
		h.values = append(h.values, v)
	}
}

func (h *Histogram) Count() int64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.count
}

func (h *Histogram) Sum() float64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.sum
}

func (h *Histogram) Min() float64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.min
}

func (h *Histogram) Max() float64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.max
}

func (h *Histogram) Mean() float64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.count == 0 {
		return 0
	}
	return h.sum / float64(h.count)
}

func (h *Histogram) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.count = 0
	h.sum = 0
	h.min = 0
	h.max = 0
	h.values = h.values[:0]
}

type Timer struct {
	histogram *Histogram
}

func NewTimer() *Timer {
	return &Timer{
		histogram: NewHistogram(1000),
	}
}

func (t *Timer) Time(f func()) time.Duration {
	start := time.Now()
	f()
	d := time.Since(start)
	t.histogram.Observe(float64(d.Nanoseconds()))
	return d
}

func (t *Timer) Observe(d time.Duration) {
	t.histogram.Observe(float64(d.Nanoseconds()))
}

func (t *Timer) Count() int64 {
	return t.histogram.Count()
}

func (t *Timer) Mean() time.Duration {
	return time.Duration(t.histogram.Mean())
}

func (t *Timer) Min() time.Duration {
	return time.Duration(t.histogram.Min())
}

func (t *Timer) Max() time.Duration {
	return time.Duration(t.histogram.Max())
}

type RateCounter struct {
	counter   *Counter
	startTime time.Time
}

func NewRateCounter() *RateCounter {
	return &RateCounter{
		counter:   NewCounter(),
		startTime: time.Now(),
	}
}

func (r *RateCounter) Inc() {
	r.counter.Inc()
}

func (r *RateCounter) Add(delta int64) {
	r.counter.Add(delta)
}

func (r *RateCounter) Count() int64 {
	return r.counter.Value()
}

func (r *RateCounter) Rate() float64 {
	elapsed := time.Since(r.startTime).Seconds()
	if elapsed == 0 {
		return 0
	}
	return float64(r.counter.Value()) / elapsed
}

func (r *RateCounter) Reset() {
	r.counter.Reset()
	r.startTime = time.Now()
}

type Registry struct {
	mu         sync.RWMutex
	counters   map[string]*Counter
	gauges     map[string]*Gauge
	histograms map[string]*Histogram
	timers     map[string]*Timer
}

func NewRegistry() *Registry {
	return &Registry{
		counters:   make(map[string]*Counter),
		gauges:     make(map[string]*Gauge),
		histograms: make(map[string]*Histogram),
		timers:     make(map[string]*Timer),
	}
}

func (r *Registry) Counter(name string) *Counter {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.counters[name]; ok {
		return c
	}
	c := NewCounter()
	r.counters[name] = c
	return c
}

func (r *Registry) Gauge(name string) *Gauge {
	r.mu.Lock()
	defer r.mu.Unlock()
	if g, ok := r.gauges[name]; ok {
		return g
	}
	g := NewGauge()
	r.gauges[name] = g
	return g
}

func (r *Registry) Histogram(name string) *Histogram {
	r.mu.Lock()
	defer r.mu.Unlock()
	if h, ok := r.histograms[name]; ok {
		return h
	}
	h := NewHistogram(1000)
	r.histograms[name] = h
	return h
}

func (r *Registry) Timer(name string) *Timer {
	r.mu.Lock()
	defer r.mu.Unlock()
	if t, ok := r.timers[name]; ok {
		return t
	}
	t := NewTimer()
	r.timers[name] = t
	return t
}

func (r *Registry) Snapshot() map[string]any {
	r.mu.RLock()
	defer r.mu.RUnlock()

	snapshot := make(map[string]any)

	for name, c := range r.counters {
		snapshot["counter."+name] = c.Value()
	}

	for name, g := range r.gauges {
		snapshot["gauge."+name] = g.Value()
	}

	for name, h := range r.histograms {
		snapshot["histogram."+name+".count"] = h.Count()
		snapshot["histogram."+name+".mean"] = h.Mean()
		snapshot["histogram."+name+".min"] = h.Min()
		snapshot["histogram."+name+".max"] = h.Max()
	}

	for name, t := range r.timers {
		snapshot["timer."+name+".count"] = t.Count()
		snapshot["timer."+name+".mean_ns"] = t.Mean().Nanoseconds()
		snapshot["timer."+name+".min_ns"] = t.Min().Nanoseconds()
		snapshot["timer."+name+".max_ns"] = t.Max().Nanoseconds()
	}

	return snapshot
}

var defaultRegistry = NewRegistry()

func Default() *Registry {
	return defaultRegistry
}

func CounterInc(name string) {
	defaultRegistry.Counter(name).Inc()
}

func CounterAdd(name string, delta int64) {
	defaultRegistry.Counter(name).Add(delta)
}

func GaugeSet(name string, v int64) {
	defaultRegistry.Gauge(name).Set(v)
}

func GaugeInc(name string) {
	defaultRegistry.Gauge(name).Inc()
}

func GaugeDec(name string) {
	defaultRegistry.Gauge(name).Dec()
}

func HistogramObserve(name string, v float64) {
	defaultRegistry.Histogram(name).Observe(v)
}

func TimerObserve(name string, d time.Duration) {
	defaultRegistry.Timer(name).Observe(d)
}

func Snapshot() map[string]any {
	return defaultRegistry.Snapshot()
}
