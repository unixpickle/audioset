package audioset

import "math/rand"

const (
	augmentMinLen = 0.9
	augmentMaxLen = 1.1
	augmentNoise  = 0.005
)

// Augment performs data augmentation by slightly
// manipulating the audio stream.
func Augment(samples []float64) []float64 {
	res := make([]float64, 0, int(float64(len(samples))*augmentMaxLen+1))
	scale := augmentMinLen + rand.Float64()*(augmentMaxLen-augmentMinLen)
	inc := 1 / scale

	for softIdx := 0.0; int(softIdx+1) < len(samples); softIdx += inc {
		nextAmount := softIdx - float64(int(softIdx))
		mixed := samples[int(softIdx)]*(1-nextAmount) +
			samples[int(softIdx+1)]*nextAmount +
			rand.NormFloat64()*augmentNoise
		res = append(res, mixed)
	}

	return res
}
