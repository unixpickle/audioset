package metaset

import (
	"github.com/unixpickle/anydiff"
	"github.com/unixpickle/anydiff/anyseq"
	"github.com/unixpickle/anynet"
	"github.com/unixpickle/anynet/anys2s"
	"github.com/unixpickle/anynet/anysgd"
	"github.com/unixpickle/anyvec"
	"github.com/unixpickle/anyvec/anyvec64"
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

// Batch is a training batch used by a Trainer.
type Batch struct {
	// Samples contains audio sample data, as produced by the
	// Seq() function.
	Samples anyseq.Seq

	// Labels contains one sequence per episode, where each
	// time-step is a one-hot label vector.
	//
	// This is considered constant.
	// Back-propagation through the labels is not guaranteed
	// to work properly.
	Labels anyseq.Seq
}

// Trainer is an anysgd.Fetcher and anysgd.Gradienter for
// training meta-learning models on AudioSet data.
type Trainer struct {
	Creator anyvec.Creator

	// FeatureFunc turns audio samples into a packed list of
	// features for use by LearnerFunc.
	FeatureFunc func(samples anyseq.Seq) anydiff.Res

	// LearnerFunc turns meta-training episodes into output
	// predictions.
	//
	// The inputs are concatenated <feature, label> vectors,
	// where features were computed by FeatureFunc.
	LearnerFunc func(episodes anyseq.Seq) anyseq.Seq

	// Parameters for gradient computation.
	Params []*anydiff.Var

	// Used to produce episodes.
	Set        Set
	NumClasses int
	NumSteps   int

	// Used to produce sample sequences.
	ChunkSize int

	// Average indicate whether or not costs should be
	// averaged (rather than summed).
	Average bool

	// LastCost is updated at every call to Gradient with
	// the latest batch cost.
	LastCost anyvec.Numeric
}

// Fetch produces a *Batch with random episodes.
// The s argument is only used to get the batch size.
func (t *Trainer) Fetch(s anysgd.SampleList) (anysgd.Batch, error) {
	c := anyvec64.DefaultCreator{}
	var in [][]anyvec.Vector
	var out [][]anyvec.Vector
	for i := 0; i < s.Len(); i++ {
		batch, labels := t.Set.Episode(t.NumClasses, t.NumSteps)
		seq, err := Seq(c, batch, t.ChunkSize)
		if err != nil {
			return nil, essentials.AddCtx("fetch samples", err)
		}
		in = append(in, seq...)

		var outSeq []anyvec.Vector
		for _, label := range labels {
			oneHot := make([]float64, t.NumClasses)
			oneHot[label] = 1
			outSeq = append(outSeq, c.MakeVectorData(oneHot))
		}
		out = append(out, outSeq)
	}
	return &Batch{
		Samples: convertSeq(t.Creator, anyseq.ConstSeqList(c, in)),
		Labels:  convertSeq(t.Creator, anyseq.ConstSeqList(c, out)),
	}, nil
}

// TotalCost computes the total cost for the *Batch.
func (t *Trainer) TotalCost(b anysgd.Batch) anydiff.Res {
	batch := b.(*Batch)
	features := t.FeatureFunc(batch.Samples)
	return anydiff.Pool(features, func(features anydiff.Res) anydiff.Res {
		epSeq := episodeSeq(features, batch.Labels)
		tr := &anys2s.Trainer{
			Func: func(s anyseq.Seq) anyseq.Seq {
				return t.LearnerFunc(epSeq)
			},
			Cost:    anynet.DotCost{},
			Average: t.Average,
		}
		return tr.TotalCost(&anys2s.Batch{Outputs: batch.Labels})
	})
}

// Gradient computes the gradient for the batch's cost.
// It also sets t.LastCost to the numerical value of the
// total cost.
//
// The b argument must be a *Batch.
func (t *Trainer) Gradient(b anysgd.Batch) anydiff.Grad {
	grad, lc := anysgd.CosterGrad(t, b, t.Params)
	t.LastCost = lc
	return grad
}

// episodeSeq creates an episode sequence by joining
// feature vectors with the labels from the previous
// timesteps.
func episodeSeq(features anydiff.Res, labelSeq anyseq.Seq) anyseq.Seq {
	labels := anyseq.SeparateSeqs(labelSeq.Output())

	var numSamples int
	for _, list := range labels {
		numSamples += len(list)
	}
	featureSize := features.Output().Len() / numSamples

	var featureOffset int
	var episodes [][]anydiff.Res
	for _, labelSeq := range labels {
		var episode []anydiff.Res
		for i := range labelSeq {
			var lastLabel anyvec.Vector
			if i == 0 {
				lastLabel = labelSeq[0].Creator().MakeVector(labelSeq[0].Len())
			} else {
				lastLabel = labelSeq[i-1]
			}
			labelRes := anydiff.NewConst(lastLabel)

			feature := anydiff.Slice(features, featureOffset, featureOffset+featureSize)
			featureOffset += featureSize
			episode = append(episode, anydiff.Concat(feature, labelRes))
		}
		episodes = append(episodes, episode)
	}

	var resBatches []*anyseq.ResBatch
	for t, batch := range labelSeq.Output() {
		var reses []anydiff.Res
		for _, episode := range episodes {
			if len(episode) > t {
				reses = append(reses, episode[t])
			}
		}
		resBatches = append(resBatches, &anyseq.ResBatch{
			Packed:  anydiff.Concat(reses...),
			Present: batch.Present,
		})
	}

	return anyseq.ResSeq(labelSeq.Creator(), resBatches)
}

func convertSeq(c anyvec.Creator, inSeqs anyseq.Seq) anyseq.Seq {
	batches := []*anyseq.Batch{}
	for _, batch := range inSeqs.Output() {
		vec := c.MakeVectorData(c.MakeNumericList(batch.Packed.Data().([]float64)))
		batches = append(batches, &anyseq.Batch{
			Packed:  vec,
			Present: batch.Present,
		})
	}
	return anyseq.ConstSeq(c, batches)
}
