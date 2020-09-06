package midifile

import (
	. "github.com/strickyak/trig-synth"
	. "github.com/strickyak/yak"
	"testing"
)

func TestC4(t *testing.T) {
	*TEMPO = 0.001
	*RATE = 8000
	one := OneChannel{
		[]Thing{
			Thing{
				When:  100,
				Track: 0,
				Act:   ON,
				X:     60,
				Y:     50,
			},
			Thing{
				When:  200,
				Track: 0,
				Act:   ON,
				X:     60,
				Y:     0,
			},
		},
	}
	volts := make(chan Volt, BUFSIZ)
	go func() {
		one.Emit(volts)
		println("closing")
		close(volts)
		println("closed")
	}()
	var out []float64
	for v := range volts {
		println("v", len(out), V(v))
		out = append(out, float64(v))
	}
	println("len", len(out))
}
