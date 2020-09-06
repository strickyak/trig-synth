package midifile

import (
	. "github.com/strickyak/trig-synth"
	"github.com/strickyak/trig-synth/mixer"
	. "github.com/strickyak/yak"

	"flag"
	"io"
	"math"
)

var ENVELOPE = flag.Bool("e", true, "Use -attack -decay -sustain -release envelope")
var ATTACK = flag.Float64("attack", 0.01, "Duration of attack, in seconds.")
var DECAY = flag.Float64("decay", 0.20, "Duration of decay, in seconds.")
var SUSTAIN = flag.Float64("sustain", 0.50, "Portion of sustain, zero to one.")
var RELEASE = flag.Float64("release", 0.01, "Duration of release, in seconds.")
var BASS = flag.Float64("bass", 0.03, "Bass boost per note.")
var BASS_CUTOFF = flag.Int("bass_cutoff", 50, "Midi note number where bass boost begins.")
var A = flag.Float64("a", 440, "Usually A == 440.")
var H2 = flag.Float64("h2", 0.0, "portion of 2nd harmonic")
var H3 = flag.Float64("h3", 0.0, "portion of 3rd harmonic")
var H4 = flag.Float64("h4", 0.0, "portion of 4th harmonic")
var H5 = flag.Float64("h5", 0.0, "portion of 5th harmonic")

func Read4str(r io.Reader) string {
	buf := make([]byte, 4)
	_, err := io.ReadFull(r, buf)
	if err == io.EOF {
		return ""
	}
	Check(err)
	return string(buf)
}

func Read4int(r io.Reader) int {
	buf := make([]byte, 4)
	_, err := io.ReadFull(r, buf)
	if err == io.EOF {
		return -1
	}
	Check(err)
	return int((uint(buf[0]) << 24) | (uint(buf[1]) << 16) | (uint(buf[2]) << 8) | uint(buf[3]))
}

func Int2(bb []byte) int {
	return int(uint(bb[0])<<8 | uint(bb[1]))
}
func Int3(bb []byte) int {
	return int(uint(bb[0])<<16 | uint(bb[1])<<8 | uint(bb[2]))
}

func IntV(bb []byte) (int, []byte) {
	var z uint
	for {
		var y byte
		y, bb = bb[0], bb[1:]
		z = (z << 7) | uint(y&127)
		if (y & 128) == 0 {
			return int(z), bb
		}
	}
}

func Preview(bb []byte) {
	s := F("Preview[%d]: ", len(bb))
	for i := 0; i < 30; i++ {
		if i >= len(bb) {
			break
		}
		s += F("%02x ", bb[i])
		if (i & 3) == 3 {
			s += " "
		}
	}
	println(s)
}

type Action int

const (
	ON Action = iota + 1
	OFF
	CONTROL
	PROGRAM
	F0
	F7
	FF
)

type Thing struct {
	When  int
	Track byte
	Act   Action
	X     byte
	Y     byte
	S     string
}

type Midifile struct {
	Chunks [][]Thing
}

func New(r io.Reader) Midifile {
	var chunks [][]Thing

	for {
		typ := Read4str(r)
		if typ == "" {
			break
		}
		n := Read4int(r)
		println("Type", typ, "Len", n)
		Must(n >= 0)

		bb := make([]byte, n)
		_, err := io.ReadFull(r, bb)
		Check(err)

		if typ == "MThd" && n == 6 {
			format := Int2(bb[0:2])
			ntrks := Int2(bb[2:4])
			division := Int2(bb[4:6])
			println("CHUNK: HEADER: format", format, "ntrks", ntrks, "division", division)
		}

		if typ == "MTrk" {
			println("CHUNK: TRACK: len=", len(bb))
			var chunk []Thing
			var bb0last byte
			var t int
			for len(bb) > 0 {
				Preview(bb)
				var deltaTime int
				deltaTime, bb = IntV(bb)
				t += deltaTime
				println("+++", deltaTime, "===", t)
				bb0 := bb[0]
				switch bb0 {
				case 0xF0, 0xF7:
					{
						var n int
						n, bb = IntV(bb[1:])
						var sysex string
						sysex, bb = string(bb[0:n]), bb[n:]
						println(F("SYSTEM EXCLUSIVE %02x [%d]: %q", bb0, n, sysex))
						act := F0
						if bb0 == 0xF7 {
							act = F7
						}
						note := Thing{
							When:  t,
							Track: 255,
							Act:   act,
							X:     0,
							Y:     0,
							S:     sysex,
						}
						chunk = append(chunk, note)
					}
				case 0xFF:
					{
						bb1 := bb[1]
						var n int
						n, bb = IntV(bb[2:])
						var text string
						text, bb = string(bb[:n]), bb[n:]
						println(F("FF Meta:%x [%d]: text: %q", bb1, n, text))
						note := Thing{
							When:  t,
							Track: 255,
							Act:   FF,
							X:     bb1,
							Y:     0,
							S:     text,
						}
						chunk = append(chunk, note)
					}
				default:
					{
						if bb0 < 128 {
							bb0 = bb0last
						} else {
							bb = bb[1:]
							bb0last = bb0
						}
						nybble := bb0 & 15

						if 0x80 <= bb0 && bb0 <= 0x8F {
							var x, y byte
							x, y, bb = bb[0], bb[1], bb[2:]
							println(F("@%12d <%d> OFF key=%d vel=%d", t, nybble, x, y))
							note := Thing{
								When:  t,
								Track: nybble,
								Act:   OFF,
								X:     x,
								Y:     y,
							}
							chunk = append(chunk, note)
						} else if 0x90 <= bb0 && bb0 <= 0x9F {
							var x, y byte
							x, y, bb = bb[0], bb[1], bb[2:]
							println(F("@%12d <%d> ON key=%d vel=%d", t, nybble, x, y))
							note := Thing{
								When:  t,
								Track: nybble,
								Act:   ON,
								X:     x,
								Y:     y,
							}
							chunk = append(chunk, note)
						} else if 0xB0 <= bb0 && bb0 <= 0xBF {
							var x, y byte
							x, y, bb = bb[0], bb[1], bb[2:]
							println(F("@%12d <%d> CONTROL controller=%d value=%d", t, nybble, x, y))
							note := Thing{
								When:  t,
								Track: nybble,
								Act:   CONTROL,
								X:     x,
								Y:     y,
							}
							chunk = append(chunk, note)
						} else if 0xC0 <= bb0 && bb0 <= 0xCF {
							var x byte
							x, bb = bb[0], bb[1:]
							println(F("@%12d <%d> PROGRAM #%d", t, nybble, x))
							note := Thing{
								When:  t,
								Track: nybble,
								Act:   PROGRAM,
								X:     x,
								Y:     0,
							}
							chunk = append(chunk, note)
						} else {
							//panic(F("bb0=0x%x", bb0))
						}
					}
				}
			}
			chunks = append(chunks, chunk)
		}
	}

	return Midifile{chunks}
}

func hasAtLeastOneON(things []Thing) bool {
	for _, th := range things {
		if th.Act == ON {
			return true
		}
	}
	return false
}

func hasAtLeastOneONforTrack(things []Thing, track byte) bool {
	for _, th := range things {
		if th.Act == ON && th.Track == track {
			return true
		}
	}
	return false
}

func (o Midifile) Emit(out chan Volt) {
	var m mixer.Mixer
	for _, chunk := range o.Chunks {
		if hasAtLeastOneON(chunk) {
			c := &OneChunk{chunk}
			m.Inputs = append(m.Inputs, c)
		}
	}
	println(F("Midifile::Emit(): %d chunks have tones.", len(m.Inputs)))
	m.Emit(out)
}

type OneChunk struct {
	Chunk []Thing
}

func (o OneChunk) Emit(out chan Volt) {
	var m mixer.Mixer
	for i := byte(0); i < 16; i++ {
		if hasAtLeastOneONforTrack(o.Chunk, i) {
			m.Inputs = append(m.Inputs, OneTrack{
				Things:   o.Chunk,
				TrackNum: i,
			})
		}
	}
	m.Emit(out)
}

type OneTrack struct {
	Things   []Thing
	TrackNum byte
}

func (o OneTrack) Emit(out chan Volt) {
	things := o.Things
	var t float64 = -0.001 // time in seconds
	var freq float64       // Freq in Hz
	var vel float64        // Velocity
Loop:
	for _, th := range things {
		println("thing", V(th))
		switch th.Act {
		case ON, OFF:
			if th.Track != o.TrackNum {
				continue Loop
			}
			when := float64(th.When) * *TEMPO
			if t <= when {
				t = emitTone(t, when, freq, vel, out)
				freq = *A * math.Pow(2, (float64(th.X)-57)/12)
				if th.Act == OFF {
					vel = 0
				} else {
					vel = float64(th.Y)
					if *BASS > 0 && int(th.X) < *BASS_CUTOFF {
						vel += *BASS * float64(*BASS_CUTOFF-int(th.X)) * vel
						if vel > 127 {
							vel = 127
						}
					}
				}
			} else if t > when+0.001 {
				panic(V(t, when))
			}
		}
	}
}

func emitTone(t, when, freq, vel float64, out chan Volt) float64 {
	harmonicTotal := 1.0 + *H2 + *H3 + *H4 + *H5
	println("emitTone", t, when, freq, vel)
	tick := 1.0 / *RATE
	t0 := t

	if freq == 0 || vel == 0 {
		for t <= when {
			out <- 0
			t += tick
		}
		return t
	}

	omega := 2 * math.Pi * freq
	vel0to1 := vel / 128
	for t < when {
		var gain float64
		if *ENVELOPE {
			switch {
			case *ENVELOPE && t < t0+*ATTACK:
				{
					phi := (t - t0) * math.Pi / *ATTACK
					curve := 0.5 - math.Cos(phi)/2
					gain = curve * vel0to1
					//println("Attack curve", V(phi/math.Pi), V(curve), V(gain))
				}
			case *ENVELOPE && t < t0+*ATTACK+*DECAY:
				{
					phi := (t - (t0 + *ATTACK)) * math.Pi / *DECAY
					curve := 0.5 + math.Cos(phi)/2
					gain = vel0to1 - vel0to1*((1.0-curve)*(1.0-*SUSTAIN))
					//println("Decay curve", V(phi/math.Pi), V(curve), V(gain))
				}
			default:
				gain = *SUSTAIN * vel0to1
			}
			// Release can apply after any switch case.
			g0 := gain
			if *ENVELOPE && t > when-*RELEASE {
				phi := (when - t) * math.Pi / *RELEASE
				//println("Release curve when-t=", V(when), V(t), V(when-t), V(phi))
				curve := 0.5 - math.Cos(phi)/2
				gain = g0 * curve
				//println("Release curve", V(phi/math.Pi), V(curve), V(g0), V(gain))
			}
		} else {
			gain = *SUSTAIN * vel0to1 // Non-envelope case.
		}
		// volt := Volt(gain * math.Sin(t*omega))
		volt := Volt(math.Sin(t * omega))
		volt += Volt(math.Sin(2*t*omega) * *H2)
		volt += Volt(math.Sin(3*t*omega) * *H3)
		volt += Volt(math.Sin(4*t*omega) * *H4)
		volt += Volt(math.Sin(5*t*omega) * *H5)
		volt *= Volt(gain / harmonicTotal)
		out <- volt
		t += tick
	}
	return t
}
