package midifile

var Notes = map[rune]int{
	'c': 60, 'd': 62, 'e': 64, 'f': 65, 'g': 67, 'a': 69, 'b': 71,
}

func NotesToThings(s string) []Thing {
	var z []Thing
	octave := 0
	t := 0
	for _, r := range s {
		freq := -1
		switch r {
		case '_':
			octave, freq = -12, -1
		case '^':
			octave, freq = +12, -1
		case ' ':
			freq = 0
		case 'a', 'b', 'c', 'd', 'e', 'f', 'g':
			octave, freq = 0, octave+Notes[r]
		default:
			freq = -1
		}

		if freq >= 0 {
			z = append(z, Thing{
				When:  t,
				Track: 0,
				Act:   ON,
				X:     byte(freq),
				Y:     100,
			})
			z = append(z, Thing{
				When:  t + 400,
				Track: 0,
				Act:   ON,
				X:     byte(freq),
				Y:     0,
			})
			t += 500
		}
	}
	return z
}
