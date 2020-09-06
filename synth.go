package synth

import (
	"flag"
)

var MAIN = flag.String("main", "midifile", "main program to run")
var RATE = flag.Float64("rate", 44100, "sample rate in Hertz")
var TEMPO = flag.Float64("tempo", 0.001, "tempo in unknown units")
var CLIP = flag.Bool("clip", false, "clip if 5% over or under")

const BUFSIZ = 1000

type Volt float64

type Emitter interface {
	Emit(out chan Volt)
}
