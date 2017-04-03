// Command class_split divides the set of classes into two
// separate sets, aiming to minimize the number of samples
// with tags from both sets.
//
// This is suitable for preparing a class-based data split
// for meta-learning or one-shot learning applications.
package main

import (
	"encoding/csv"
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"

	"github.com/unixpickle/essentials"
)

func main() {
	var csvPath string
	var numEval int
	var numIters int
	var trainOut string
	var evalOut string
	flag.StringVar(&csvPath, "data", "", "data CSV path")
	flag.IntVar(&numEval, "numeval", 50, "number of evaluation classes")
	flag.IntVar(&numIters, "iters", 10000, "number of hill-climbing steps")
	flag.StringVar(&trainOut, "trainout", "train.txt", "training list output file")
	flag.StringVar(&evalOut, "evalout", "eval.txt", "evaluation list output file")
	flag.Parse()

	if csvPath == "" {
		essentials.Die("Required flag: -data. See -help.")
	}

	labels := ReadLabels(csvPath)
	solution := RandSplit(numEval, labels)
	ndrop := solution.NumDrop(labels)
	log.Println("started with", ndrop, "dropped.")
	for i := 0; i < numIters; i++ {
		perm := solution.Mutate()
		d := perm.NumDrop(labels)
		if d < ndrop {
			ndrop = d
			solution = perm
			log.Printf("iter %d/%d: improved to %d dropped.", i, numIters, ndrop)
		}
	}

	log.Println("Writing", trainOut, "...")
	WriteLabels(trainOut, solution.KeysForVal(true))

	log.Println("Writing", evalOut, "...")
	WriteLabels(evalOut, solution.KeysForVal(false))
}

func ReadLabels(path string) [][]string {
	f, err := os.Open(path)
	if err != nil {
		essentials.Die(err)
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.Comment = '#'
	r.Comma = ' '
	rows, err := r.ReadAll()
	if err != nil {
		essentials.Die("read "+path+":", err)
	}
	var labels [][]string
	for _, row := range rows {
		labels = append(labels, strings.Split(row[len(row)-1], ","))
	}
	return labels
}

func WriteLabels(path string, labels []string) {
	data := strings.Join(labels, "\n") + "\n"
	if err := ioutil.WriteFile(path, []byte(data), 0644); err != nil {
		essentials.Die(err)
	}
}

// Split maps IDs to true if they are for training, or
// false if they are for evaluation.
type Split map[string]bool

// RandSplit creates a new, random split.
func RandSplit(numFalse int, list [][]string) Split {
	set := Split{}
	for _, sample := range list {
		for _, class := range sample {
			set[class] = true
		}
	}
	keys := set.KeysForVal(true)
	for _, j := range rand.Perm(len(keys))[:numFalse] {
		set[keys[j]] = false
	}
	return set
}

// Mutate swaps a true and a false key.
func (s Split) Mutate() Split {
	tvals := s.KeysForVal(true)
	fvals := s.KeysForVal(false)
	tidx := rand.Intn(len(tvals))
	fidx := rand.Intn(len(fvals))
	tvals[tidx], fvals[fidx] = fvals[fidx], tvals[tidx]
	res := Split{}
	for _, t := range tvals {
		res[t] = true
	}
	for _, f := range fvals {
		res[f] = false
	}
	return res
}

// NumDrop figures out how many samples would need to be
// removed in order to split the data.
func (s Split) NumDrop(list [][]string) int {
	var drop int
	for _, sample := range list {
		var numTrue, numFalse int
		for _, class := range sample {
			if s[class] {
				numTrue++
			} else {
				numFalse++
			}
		}
		if numTrue > 0 && numFalse > 0 {
			drop++
		}
	}
	return drop
}

func (s Split) KeysForVal(val bool) []string {
	var res []string
	for key, v := range s {
		if v == val {
			res = append(res, key)
		}
	}
	return res
}
