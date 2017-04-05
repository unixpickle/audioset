package metaset

import (
	"compress/gzip"
	"errors"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/wav"
)

// Sample stores meta-data about an audio sample.
type Sample struct {
	Classes []string
	Path    string
}

// Read reads the sample as PCM data.
func (s *Sample) Read() ([]float64, error) {
	f, err := os.Open(s.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	unzip, err := gzip.NewReader(f)
	if err != nil {
		return nil, essentials.AddCtx("read "+s.Path, err)
	}
	sound, err := wav.ReadSound(unzip)
	if err != nil {
		return nil, essentials.AddCtx("read "+s.Path, err)
	}
	if err := unzip.Close(); err != nil {
		return nil, essentials.AddCtx("read "+s.Path, err)
	}
	samps := sound.Samples()
	ch := sound.Channels()
	if len(samps)%ch != 0 {
		return nil, errors.New("read " + s.Path + ": bad sample count")
	}

	// Quick and dirty channel mixing.
	res := make([]float64, len(samps)/ch)
	for i, x := range samps {
		res[i/ch] += float64(x) / float64(ch)
	}

	return res, nil
}
