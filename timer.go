package main

// Functions for timing metrics.

import (
	"expvar"
	"log"
	"time"
)

type timer struct {
	count    int
	time     float64
	interval time.Duration // avg is calculated at interval.
	v        *expvar.Float
}

// avg loops for ever and set v to the average time per count
// every interval.
func (t *timer) avg() {
	for {
		time.Sleep(t.interval)
		// Local copy of count so there is no
		// risk of div by 0 errors.
		n := t.count
		if n == 0 {
			t.v.Set(0)
		} else {
			t.v.Set(t.time / float64(n))
		}
		t.time = 0
		t.count = 0
	}
}

// track logs the time since start and increments the timer counters.
func (t *timer) track(start time.Time, m string) {
	dt := time.Since(start).Seconds()
	log.Printf("%s took %fs", m, dt)
	t.count++
	t.time += dt
}
