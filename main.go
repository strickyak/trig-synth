// +build main

/*
trig-synth: Software synthesizer with crude midifile player (in golang).

Only understands Key Down and Key Up events in Midi.  Use -tempo to fix the tempo.

For help with flags:

    go run main.go -help

== DEMO USING -main=midifile ==

    go run main.go -main=midifile -tempo=0.001 -h3=0.4  -h5=0.2 ./widor_toccata_\(c\)shattuck.mid | paplay --rate=44100 --channels=1 --format=s16le --raw /dev/stdin

    go run main.go -main=midifile -tempo=0.004 -h3=0.8  -h5=0.5  Tetris.mid | paplay --rate=44100 --channels=1 --format=s16le --raw /dev/stdin

== DEMO USING -main=notes ==

    go run main.go -main=notes 'cdefgab^c' | paplay --rate=44100 --channels=1 --format=s16le --raw /dev/stdin

== DEMO USING -main=KilOscillator  ==

    # Inspired by YouTube: LOOK MUM NO COMPUTER: The 1000 Osciallator Megadrone is Complete! The KiloDrone is ALIVE
    go run main.go -main=KilOscillator -tempo=0.03 -kgain=0.01 -n=1000 -clip > /tmp/KilOscillator.raw
    paplay --rate=44100 --channels=1 --format=s16le --raw /tmp/KilOscillator.raw
*/
package main

import (
	. "github.com/strickyak/trig-synth"
	"github.com/strickyak/trig-synth/experimental"
	"github.com/strickyak/trig-synth/midifile"
	"github.com/strickyak/trig-synth/mixer"
	. "github.com/strickyak/yak"

	"bufio"
	"flag"
	"io"
	"log"
	"os"
)

func main() {
	flag.Parse()
	volts := make(chan Volt, BUFSIZ)
	var emitter Emitter
	w := bufio.NewWriterSize(os.Stdout, 1024)

	switch *MAIN {
	case "KilOscillator":
		emitter = experimental.NewKilOscillator()
	case "notes":
		emitter = midifile.OneChunk{midifile.NotesToThings(flag.Args()[0])}
		{
			var mixer mixer.Mixer
			for _, a := range flag.Args() {
				mixer.Inputs = append(mixer.Inputs, midifile.OneChunk{midifile.NotesToThings(a)})
			}
			emitter = &mixer
		}
	case "midifile":
		{
			var mixer mixer.Mixer
			for _, a := range flag.Args() {
				r, err := os.Open(a)
				Check(err)
				defer r.Close()
				mixer.Inputs = append(mixer.Inputs, midifile.New(r))
			}
			emitter = &mixer
		}
	default:
		panic(*MAIN)
	}

	go func() {
		emitter.Emit(volts)
		close(volts)
	}()

	var clipped int
	for v := range volts {
		if *CLIP {
			if v < -1.05 {
				v = -1.05
				clipped++
			}
			if v > 1.05 {
				v = 1.05
				clipped++
			}
		}
		Must(v >= -1.05)
		Must(v <= 1.05)
		u := uint(int(v * 32700))
		putbb(w, byte(u), byte(u>>8))
	}
	w.Flush()
	if clipped > 0 {
		log.Printf("Clipped %d samples", clipped)
	}
}

func putbb(w io.Writer, bb ...byte) {
	cc, err := w.Write(bb)
	Check(err)
	Must(cc == len(bb))
}
