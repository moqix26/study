package main

import (
	"fmt"
	"math"
	"sync"
	"time"
)

type Bucket struct {
	mu       sync.Mutex
	capacity float64
	tokens   float64
	rate     float64
	last     time.Time
}

func NewBucket(capacity, rate float64, now time.Time) *Bucket {
	if capacity <= 0 || rate <= 0 {
		panic("capacity and rate must be positive")
	}
	return &Bucket{
		capacity: capacity,
		tokens:   capacity,
		rate:     rate,
		last:     now,
	}
}

func (b *Bucket) AllowN(now time.Time, n float64) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if n <= 0 || n > b.capacity {
		return false
	}
	elapsed := now.Sub(b.last).Seconds()
	if elapsed < 0 {
		elapsed = 0
	}
	b.tokens = math.Min(b.capacity, b.tokens+elapsed*b.rate)
	b.last = now
	if b.tokens < n {
		return false
	}
	b.tokens -= n
	return true
}

func must(name string, got, want bool) {
	fmt.Printf("%-28s got=%v want=%v\n", name, got, want)
	if got != want {
		panic(name)
	}
}

func main() {
	t0 := time.Unix(0, 0)
	b := NewBucket(5, 2, t0)

	must("initial burst of five", b.AllowN(t0, 5), true)
	must("sixth request rejected", b.AllowN(t0, 1), false)
	must("one token after 500ms", b.AllowN(t0.Add(500*time.Millisecond), 1), true)
	must("immediate next rejected", b.AllowN(t0.Add(500*time.Millisecond), 1), false)
	must("four tokens after two sec", b.AllowN(t0.Add(2500*time.Millisecond), 4), true)

	fmt.Println("token bucket invariants hold")
}
