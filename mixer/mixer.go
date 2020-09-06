package mixer

import (
	. "github.com/strickyak/trig-synth"
)

type Mixer struct {
	Inputs []Emitter
}

func (o *Mixer) Emit(out chan Volt) {
	numInputs := len(o.Inputs)
	gain := 1 / float64(numInputs)
	done := make([]bool, numInputs)
	ch := make([]chan Volt, numInputs)

	for i, e := range o.Inputs {
		ch[i] = make(chan Volt, BUFSIZ)
		go func(j int, x Emitter) {
			x.Emit(ch[j])
			close(ch[j])
		}(i, e)
	}

	numDone := 0
	for numDone < numInputs {
		var sum Volt
		for i, _ := range o.Inputs {
			if done[i] {
				continue
			}
			x, ok := <-ch[i]
			if !ok {
				numDone++
				done[i] = true
				continue
			}
			sum += x
		}
		out <- Volt(gain) * sum
	}
}
