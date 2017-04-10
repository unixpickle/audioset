package metaset

import (
	"math/rand"

	"github.com/unixpickle/audioset"
)

// Split splits the set into a training/evaluation set.
func Split(s audioset.Set, evalClasses []string) (training, eval audioset.Set) {
	evalSet := map[string]bool{}
	for _, c := range evalClasses {
		evalSet[c] = true
	}

	for _, sample := range s {
		var numEval, numTraining int
		for _, c := range sample.Classes {
			if evalSet[c] {
				numEval++
			} else {
				numTraining++
			}
		}
		if numEval != 0 && numTraining == 0 {
			eval = append(eval, sample)
		} else if numEval == 0 && numTraining != 0 {
			training = append(training, sample)
		}
	}

	return
}

// Episode generates a meta-learning episode.
//
// The episode will include up to numClasses classes.
// It will be at most numSteps timesteps.
//
// Labels are assigned in the range [0, numClasses).
func Episode(s audioset.Set, numClasses, numSteps int) (samples []*audioset.Sample,
	labels []int) {
	classes := []string{}
	byClass := map[string][]*audioset.Sample{}
	for _, sample := range s {
		for _, c := range sample.Classes {
			if byClass[c] == nil {
				classes = append(classes, c)
			}
			byClass[c] = append(byClass[c], sample)
		}
	}

	if numClasses > len(classes) {
		numClasses = len(classes)
	}

	for label, j := range rand.Perm(len(classes))[:numClasses] {
		class := classes[j]
		for _, sample := range byClass[class] {
			samples = append(samples, sample)
			labels = append(labels, label)
		}
	}

	for i := 0; i < len(samples); i++ {
		takeIdx := rand.Intn(len(samples)-i) + i
		samples[i], samples[takeIdx] = samples[takeIdx], samples[i]
		labels[i], labels[takeIdx] = labels[takeIdx], labels[i]
	}

	if len(samples) > numSteps {
		samples = samples[:numSteps]
		labels = labels[:numSteps]
	}

	return
}
