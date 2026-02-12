package stat

import (
	"testing"
	"time"
)

func TestCounter(t *testing.T) {
	c := NewCounter()

	if c.Value() != 0 {
		t.Errorf("initial value = %d, want 0", c.Value())
	}

	c.Inc()
	if c.Value() != 1 {
		t.Errorf("after Inc() = %d, want 1", c.Value())
	}

	c.Add(5)
	if c.Value() != 6 {
		t.Errorf("after Add(5) = %d, want 6", c.Value())
	}

	v := c.Reset()
	if v != 6 {
		t.Errorf("Reset() returned %d, want 6", v)
	}
	if c.Value() != 0 {
		t.Errorf("after Reset() = %d, want 0", c.Value())
	}
}

func TestGauge(t *testing.T) {
	g := NewGauge()

	if g.Value() != 0 {
		t.Errorf("initial value = %d, want 0", g.Value())
	}

	g.Set(42)
	if g.Value() != 42 {
		t.Errorf("after Set(42) = %d, want 42", g.Value())
	}

	g.Inc()
	if g.Value() != 43 {
		t.Errorf("after Inc() = %d, want 43", g.Value())
	}

	g.Dec()
	if g.Value() != 42 {
		t.Errorf("after Dec() = %d, want 42", g.Value())
	}

	g.Add(-10)
	if g.Value() != 32 {
		t.Errorf("after Add(-10) = %d, want 32", g.Value())
	}
}

func TestHistogram(t *testing.T) {
	h := NewHistogram(100)

	h.Observe(10)
	h.Observe(20)
	h.Observe(30)

	if h.Count() != 3 {
		t.Errorf("Count() = %d, want 3", h.Count())
	}

	if h.Sum() != 60 {
		t.Errorf("Sum() = %f, want 60", h.Sum())
	}

	if h.Min() != 10 {
		t.Errorf("Min() = %f, want 10", h.Min())
	}

	if h.Max() != 30 {
		t.Errorf("Max() = %f, want 30", h.Max())
	}

	if h.Mean() != 20 {
		t.Errorf("Mean() = %f, want 20", h.Mean())
	}

	h.Reset()
	if h.Count() != 0 {
		t.Errorf("after Reset() Count() = %d, want 0", h.Count())
	}
}

func TestTimer(t *testing.T) {
	timer := NewTimer()

	timer.Time(func() {
		time.Sleep(10 * time.Millisecond)
	})

	timer.Observe(5 * time.Millisecond)

	if timer.Count() != 2 {
		t.Errorf("Count() = %d, want 2", timer.Count())
	}

	if timer.Min() < 5*time.Millisecond {
		t.Errorf("Min() = %v, expected >= 5ms", timer.Min())
	}
}

func TestRateCounter(t *testing.T) {
	r := NewRateCounter()

	r.Add(100)
	time.Sleep(100 * time.Millisecond)

	if r.Count() != 100 {
		t.Errorf("Count() = %d, want 100", r.Count())
	}

	rate := r.Rate()
	if rate < 500 || rate > 1500 {
		t.Errorf("Rate() = %f, expected around 1000/sec", rate)
	}
}

func TestRegistry(t *testing.T) {
	r := NewRegistry()

	r.Counter("requests").Inc()
	r.Counter("requests").Add(4)
	r.Gauge("connections").Set(10)
	r.Histogram("latency").Observe(100)
	r.Timer("response_time").Observe(50 * time.Millisecond)

	snapshot := r.Snapshot()

	if v, ok := snapshot["counter.requests"]; !ok || v.(int64) != 5 {
		t.Errorf("counter.requests = %v, want 5", v)
	}

	if v, ok := snapshot["gauge.connections"]; !ok || v.(int64) != 10 {
		t.Errorf("gauge.connections = %v, want 10", v)
	}

	if v, ok := snapshot["histogram.latency.count"]; !ok || v.(int64) != 1 {
		t.Errorf("histogram.latency.count = %v, want 1", v)
	}

	if v, ok := snapshot["timer.response_time.count"]; !ok || v.(int64) != 1 {
		t.Errorf("timer.response_time.count = %v, want 1", v)
	}
}

func TestDefaultRegistry(t *testing.T) {
	CounterInc("test_counter")
	CounterAdd("test_counter", 2)

	if Default().Counter("test_counter").Value() != 3 {
		t.Errorf("test_counter = %d, want 3", Default().Counter("test_counter").Value())
	}

	GaugeSet("test_gauge", 100)
	GaugeInc("test_gauge")
	GaugeDec("test_gauge")

	if Default().Gauge("test_gauge").Value() != 100 {
		t.Errorf("test_gauge = %d, want 100", Default().Gauge("test_gauge").Value())
	}

	HistogramObserve("test_histogram", 50)
	TimerObserve("test_timer", 100*time.Millisecond)

	snapshot := Snapshot()
	if _, ok := snapshot["counter.test_counter"]; !ok {
		t.Error("test_counter not found in snapshot")
	}
}
