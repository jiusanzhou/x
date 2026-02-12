package x

import (
	"testing"
	"time"
)

func TestRealClock_Now(t *testing.T) {
	clock := RealClock{}
	before := time.Now()
	now := clock.Now()
	after := time.Now()

	if now.Before(before) || now.After(after) {
		t.Errorf("RealClock.Now() = %v, expected between %v and %v", now, before, after)
	}
}

func TestRealClock_Since(t *testing.T) {
	clock := RealClock{}
	start := time.Now()
	time.Sleep(10 * time.Millisecond)
	elapsed := clock.Since(start)

	if elapsed < 10*time.Millisecond {
		t.Errorf("RealClock.Since() = %v, expected >= 10ms", elapsed)
	}
}

func TestRealClock_After(t *testing.T) {
	clock := RealClock{}
	start := time.Now()
	<-clock.After(10 * time.Millisecond)
	elapsed := time.Since(start)

	if elapsed < 10*time.Millisecond {
		t.Errorf("RealClock.After() elapsed = %v, expected >= 10ms", elapsed)
	}
}

func TestRealClock_Sleep(t *testing.T) {
	clock := RealClock{}
	start := time.Now()
	clock.Sleep(10 * time.Millisecond)
	elapsed := time.Since(start)

	if elapsed < 10*time.Millisecond {
		t.Errorf("RealClock.Sleep() elapsed = %v, expected >= 10ms", elapsed)
	}
}

func TestRealClock_NewTimer(t *testing.T) {
	clock := RealClock{}
	timer := clock.NewTimer(10 * time.Millisecond)

	start := time.Now()
	<-timer.C()
	elapsed := time.Since(start)

	if elapsed < 10*time.Millisecond {
		t.Errorf("RealClock.NewTimer() elapsed = %v, expected >= 10ms", elapsed)
	}
}

func TestRealClock_NewTimer_Stop(t *testing.T) {
	clock := RealClock{}
	timer := clock.NewTimer(1 * time.Hour)

	if !timer.Stop() {
		t.Error("Timer.Stop() should return true when timer hasn't fired")
	}
}

func TestRealClock_NewTimer_Reset(t *testing.T) {
	clock := RealClock{}
	timer := clock.NewTimer(1 * time.Hour)
	timer.Stop()

	timer.Reset(10 * time.Millisecond)
	start := time.Now()
	<-timer.C()
	elapsed := time.Since(start)

	if elapsed < 10*time.Millisecond {
		t.Errorf("Timer.Reset() elapsed = %v, expected >= 10ms", elapsed)
	}
}

func TestRealClock_NewTicker(t *testing.T) {
	clock := RealClock{}
	ticker := clock.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	start := time.Now()
	<-ticker.C()
	elapsed := time.Since(start)

	if elapsed < 10*time.Millisecond {
		t.Errorf("RealClock.NewTicker() first tick = %v, expected >= 10ms", elapsed)
	}
}

func TestRealClock_Tick(t *testing.T) {
	clock := RealClock{}
	ch := clock.Tick(10 * time.Millisecond)

	start := time.Now()
	<-ch
	elapsed := time.Since(start)

	if elapsed < 10*time.Millisecond {
		t.Errorf("RealClock.Tick() first tick = %v, expected >= 10ms", elapsed)
	}
}

func TestRealClock_AfterFunc(t *testing.T) {
	clock := RealClock{}
	done := make(chan struct{})

	start := time.Now()
	timer := clock.AfterFunc(10*time.Millisecond, func() {
		close(done)
	})
	defer timer.Stop()

	<-done
	elapsed := time.Since(start)

	if elapsed < 10*time.Millisecond {
		t.Errorf("RealClock.AfterFunc() elapsed = %v, expected >= 10ms", elapsed)
	}
}

func TestRealClock_ImplementsWithTicker(t *testing.T) {
	var _ WithTicker = RealClock{}
}
