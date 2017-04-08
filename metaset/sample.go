package metaset

import (
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"

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

	var source io.ReadCloser
	if filepath.Ext(s.Path) == ".gz" {
		defer f.Close()
		source, err = gzip.NewReader(f)
		if err != nil {
			return nil, essentials.AddCtx("read "+s.Path, err)
		}
	} else {
		source = f
	}

	sound, err := wav.ReadSound(source)
	if err != nil {
		return nil, essentials.AddCtx("read "+s.Path, err)
	}

	if err := source.Close(); err != nil {
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
