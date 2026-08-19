package metrics

import (
	"slices"
	"time"

	"github.com/briheet/sen/internal/model"
)

type histogramDelta struct {
	at     time.Time
	counts []uint64
}

// histogramWindow aggregates monotonic runtime buckets over a bounded window.
type histogramWindow struct {
	buckets  []float64
	previous []uint64
	counts   []uint64
	samples  []histogramDelta
}

func (w *histogramWindow) Add(at time.Time, current *model.Histogram) {
	if current == nil || len(current.Buckets) != len(current.Counts)+1 {
		w.reset(nil)
		return
	}
	if !slices.Equal(w.buckets, current.Buckets) || len(w.previous) != len(current.Counts) {
		w.reset(current)
		return
	}
	delta := make([]uint64, len(current.Counts))
	for index, count := range current.Counts {
		if count < w.previous[index] {
			w.reset(current)
			return
		}
		delta[index] = count - w.previous[index]
		w.counts[index] += delta[index]
	}
	copy(w.previous, current.Counts)
	w.samples = append(w.samples, histogramDelta{at: at, counts: delta})
	w.expire(at.Add(-historyWindow))
}

func (w *histogramWindow) Histogram() *model.Histogram {
	if len(w.buckets) == 0 {
		return nil
	}
	return &model.Histogram{Counts: w.counts, Buckets: w.buckets}
}

func (w *histogramWindow) reset(current *model.Histogram) {
	w.samples = w.samples[:0]
	if current == nil {
		w.buckets = w.buckets[:0]
		w.previous = w.previous[:0]
		w.counts = w.counts[:0]
		return
	}
	w.buckets = append(w.buckets[:0], current.Buckets...)
	w.previous = append(w.previous[:0], current.Counts...)
	if cap(w.counts) < len(current.Counts) {
		w.counts = make([]uint64, len(current.Counts))
	} else {
		w.counts = w.counts[:len(current.Counts)]
		clear(w.counts)
	}
}

func (w *histogramWindow) expire(cutoff time.Time) {
	first := 0
	for first < len(w.samples) && w.samples[first].at.Before(cutoff) {
		for index, count := range w.samples[first].counts {
			w.counts[index] -= count
		}
		first++
	}
	if first > 0 {
		copy(w.samples, w.samples[first:])
		w.samples = w.samples[:len(w.samples)-first]
	}
}
