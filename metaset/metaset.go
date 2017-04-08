package metaset

import (
	"encoding/csv"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/unixpickle/essentials"
)

// Set is a set of audio samples.
type Set []*Sample

// ReadSet reads a dataset by matching filenames in a
// directory with records from a CSV file.
func ReadSet(dir, csvFile string) (Set, error) {
	f, err := os.Open(csvFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comment = '#'
	r.Comma = ' '
	r.FieldsPerRecord = 4
	rows, err := r.ReadAll()
	if err != nil {
		essentials.Die("read "+csvFile+":", err)
	}

	var set Set
	for _, row := range rows {
		id := strings.Trim(row[0], ",")
		start := strings.Trim(row[1], ",")
		filename := filepath.Join(dir, id+"_"+start+".wav.gz")
		if _, err := os.Stat(filename); err != nil {
			filename = filepath.Join(dir, id+"_"+start+".wav")
			if _, err := os.Stat(filename); err != nil {
				continue
			}
		}
		set = append(set, &Sample{
			Classes: strings.Split(row[3], ","),
			Path:    filename,
		})
	}

	return set, nil
}

// Split splits the set into a training/evaluation set.
func (s Set) Split(evalClasses []string) (training, eval Set) {
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
func (s Set) Episode(numClasses, numSteps int) (samples []*Sample, labels []int) {
	classes := []string{}
	byClass := map[string][]*Sample{}
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
