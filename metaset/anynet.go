package metaset

import (
	"github.com/unixpickle/anydiff/anyseq"
	"github.com/unixpickle/anynet/anys2s"
	"github.com/unixpickle/anynet/anysgd"
	"github.com/unixpickle/anyvec"
	"github.com/unixpickle/essentials"
)

// Seq creates a list of sequences where each sequence
// contains a PCM waveform.
//
// Time-steps in each sequence contain chunkSize samples.
// The last time-step is padded with 0's as needed.
func Seq(c anyvec.Creator, batch []*Sample, chunkSize int) ([][]anyvec.Vector, error) {
	var seqs [][]anyvec.Vector
	for _, sample := range batch {
		pcm, err := sample.Read()
		if err != nil {
			return nil, err
		}
		var seq []anyvec.Vector
		for i := 0; i < len(pcm); i += chunkSize {
			chunk := make([]float64, chunkSize)
			copy(chunk, pcm[i:])
			seq = append(seq, c.MakeVectorData(c.MakeNumericList(chunk)))
		}
		seqs = append(seqs, seq)
	}
	return seqs, nil
}

// Fetcher is an anysgd.Fetcher which produces
// *anys2s.Batch batches.
type Fetcher struct {
	Creator anyvec.Creator

	// Used to produce episodes.
	Set        Set
	NumClasses int
	NumSteps   int

	// Used for calls to Seq.
	ChunkSize int
}

// Fetch produces an *anys2s.Batch with random episodes.
// The s argument is only used to get the batch size.
//
// The input contains one sequence per sample.
// The output contains one sequence per batch, where each
// time-step is a one-hot vector label.
func (f *Fetcher) Fetch(s anysgd.SampleList) (anysgd.Batch, error) {
	var in [][]anyvec.Vector
	var out [][]anyvec.Vector
	for i := 0; i < s.Len(); i++ {
		batch, labels := f.Set.Episode(f.NumClasses, f.NumSteps)
		seq, err := Seq(f.Creator, batch, f.ChunkSize)
		if err != nil {
			return nil, essentials.AddCtx("fetch samples", err)
		}
		in = append(in, seq...)

		var outSeq []anyvec.Vector
		for _, label := range labels {
			oneHot := make([]float64, f.NumClasses)
			oneHot[label] = 1
			numList := f.Creator.MakeNumericList(oneHot)
			outSeq = append(outSeq, f.Creator.MakeVectorData(numList))
		}
		out = append(out, outSeq)
	}
	return &anys2s.Batch{
		Inputs:  anyseq.ConstSeqList(f.Creator, in),
		Outputs: anyseq.ConstSeqList(f.Creator, out),
	}, nil
}
