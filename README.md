# trig-synth
Software synthesizer with crude midifile player (in golang)

Only understands Key Down and Key Up events in Midi.  Use -tempo to fix the tempo.

There are many midi files that this does not understand, or only plays a few random notes.
But some simple files sound pretty good.

For help with flags:

    go run main.go -help

## DEMO USING -main=midifile ##

    go run main.go -main=midifile -tempo=0.001 -h3=0.4  -h5=0.2 ./widor_toccata_\(c\)shattuck.mid | paplay --rate=44100 --channels=1 --format=s16le --raw /dev/stdin

    go run main.go -main=midifile -tempo=0.004 -h3=0.8  -h5=0.5  Tetris.mid | paplay --rate=44100 --channels=1 --format=s16le --raw /dev/stdin

## DEMO USING -main=notes ##

    go run main.go -main=notes 'cdefgab^c' | paplay --rate=44100 --channels=1 --format=s16le --raw /dev/stdin

## DEMO USING -main=KilOscillator  ##

    # Inspired by YouTube: LOOK MUM NO COMPUTER: The 1000 Osciallator Megadrone is Complete! The KiloDrone is ALIVE
    go run main.go -main=KilOscillator -tempo=0.03 -kgain=0.01 -n=1000 -clip > /tmp/KilOscillator.raw
    paplay --rate=44100 --channels=1 --format=s16le --raw /tmp/KilOscillator.raw
