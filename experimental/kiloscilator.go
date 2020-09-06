package experimental

// Inspired by YouTube: LOOK MUM NO COMPUTER: The 1000 Osciallator Megadrone is Complete! The KiloDrone is ALIVE

// go run main.go -main=KilOscillator -tempo=0.03 -kgain=0.01 -nosc=1000 -clip > _

// paplay --rate=44100 --channels=1 --format=s16le --raw _

import (
	"flag"
	"math"
	"math/rand"

	. "github.com/strickyak/trig-synth"
)

var KRAMP = flag.Float64("kramp", 0.05, "For KilOscillator: ramp up/down time, seconds")
var KGAIN = flag.Float64("kgain", 0.1, "For KilOscillator: individual oscillator gain")
var N = flag.Int("nosc", 1000, "For KilOscillator: number of oscillators")

type KilOscillator struct {
	Oscs []*Osc
}

type Osc struct {
	BaseFreq float64
	Freq     float64
	Gain     float64
	Phase    float64
}

const LO_FREQ = 20
const HI_FREQ = 1000
const Tau = 2 * math.Pi

func Interpolate(way, min, max float64) float64 {
	return (1.0-way)*min + way*max
}

func NewKilOscillator() *KilOscillator {
	z := &KilOscillator{}
	z.Oscs = make([]*Osc, *N)
	for i := range z.Oscs {
		z.Oscs[i] = &Osc{
			BaseFreq: Interpolate(rand.Float64(), LO_FREQ, HI_FREQ),
			Gain:     *KGAIN,
			Phase:    rand.Float64() * Tau,
		}
	}
	return z
}

func (o *Osc) Step() Volt {
	deltaPhase := Tau * o.Freq / *RATE
	o.Phase += deltaPhase
	if o.Phase > Tau {
		o.Phase -= Tau // Stay in 0..Tau
	}
	return Volt(o.Gain * math.Cos(o.Phase))
}

// Ramp up & ramp down beginning & ending of emission by multiplying by this gain.
func ramp(i, steps int) float64 {
	r := int(float64(*RATE) * (*KRAMP)) // ramp duration in steps.
	if i < r {
		return 0.5 - math.Cos(math.Pi*(float64(i)/float64(r)))/2.0
	}
	if i > steps-r {
		return 0.5 - math.Cos(math.Pi*(float64(steps-i)/float64(r)))/2.0
	}
	return 1.0
}

func (k *KilOscillator) Emit(out chan Volt) {
	seconds := 1.0 / *TEMPO
	steps := int(*RATE * seconds)

	for i := 0; i < steps; i++ {
		way := float64(i) / float64(steps)

		var v Volt
		for _, o := range k.Oscs {
			// Poke a new Freq.
			o.Freq = Interpolate(way, o.BaseFreq, HI_FREQ)

			v += o.Step() // This will adjust phase.
		}
		out <- Volt(ramp(i, steps)) * v
	}
}
