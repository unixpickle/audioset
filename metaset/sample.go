package metaset

import (
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/gonum/blas/blas64"
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

	data, err := readAndMix(source)
	if err != nil {
		source.Close()
		return nil, essentials.AddCtx("read "+s.Path, err)
	}
	if err := source.Close(); err != nil {
		return nil, essentials.AddCtx("read "+s.Path, err)
	}
	return data, nil
}

func readAndMix(r io.Reader) ([]float64, error) {
	sound, err := wav.ReadSound(r)
	if err != nil {
		return nil, err
	}

	samps := sound.Samples()
	numChan := sound.Channels()
	if len(samps)%numChan != 0 {
		return nil, errors.New("bad sample count")
	}

	// Quick and dirty channel mixing.
	res := make([]float64, len(samps)/numChan)
	if numChan == 1 {
		for i, x := range samps {
			res[i] = float64(x)
		}
	} else if numChan == 2 {
		for i := 0; i < len(samps); i += 2 {
			res[i>>1] = float64(samps[i]+samps[i+1]) / 2
		}
	} else {
		for ch := 0; ch < numChan; ch++ {
			for i := range res {
				res[i] += float64(samps[i*numChan+ch])
			}
		}
		blas64.Scal(len(res), 1/float64(numChan), blas64.Vector{Data: res, Inc: 1})
	}

	return res, nil
}
