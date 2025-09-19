// pkg/levelmeter.go
package pkg

import (
	"math"
	"sync"

	"github.com/faiface/beep"
)

type LevelMeter struct {
	Src        beep.Streamer
	SampleRate beep.SampleRate

	mu  sync.RWMutex
	rms float64 // smoothed 0..~1
	ema float64 // internal state
	a   float64 // smoothing (0..1), e.g. 0.2
}

func NewLevelMeter(src beep.Streamer, sr beep.SampleRate) *LevelMeter {
	return &LevelMeter{Src: src, SampleRate: sr, a: 0.2}
}

func (m *LevelMeter) Stream(buf [][2]float64) (n int, ok bool) {
	n, ok = m.Src.Stream(buf)
	if n <= 0 {
		return n, ok
	}

	// RMS across both channels, normalized-ish
	var sum float64
	for i := 0; i < n; i++ {
		l, r := buf[i][0], buf[i][1]
		sum += l*l + r*r
	}
	rms := math.Sqrt(sum / (2 * float64(n)))

	// Exponential moving average
	m.mu.Lock()
	m.ema = m.ema*(1-m.a) + rms*m.a
	m.rms = m.ema
	m.mu.Unlock()
	return n, ok
}

func (m *LevelMeter) Err() error { return m.Src.Err() }

// Level returns a smoothed 0..1-ish loudness.
func (m *LevelMeter) Level() float64 {
	m.mu.RLock()
	v := m.rms
	m.mu.RUnlock()
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
