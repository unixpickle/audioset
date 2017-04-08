// Command find_bad finds corrupted audio samples in a
// directory of AudioSet samples.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/unixpickle/audioset/metaset"
	"github.com/unixpickle/essentials"
)

func main() {
	var dirPath string
	flag.StringVar(&dirPath, "dir", "", "path to sample directory")
	flag.Parse()

	if dirPath == "" {
		essentials.Die("Required flag: -dir. See -help.")
	}

	listing, err := ioutil.ReadDir(dirPath)
	if err != nil {
		essentials.Die(err)
	}

	var r RollingAvg
	for _, item := range listing {
		if strings.HasSuffix(item.Name(), ".wav.gz") ||
			strings.HasSuffix(item.Name(), ".wav") {
			path := filepath.Join(dirPath, item.Name())
			sample := metaset.Sample{Path: path}
			data, err := sample.Read()
			if err != nil {
				fmt.Println()
				fmt.Fprintln(os.Stderr, err)
			} else {
				for _, sample := range data {
					r.Add(sample)
				}
				fmt.Printf("variance=%f     \r", r.Variance())
			}
		}
	}
	fmt.Println()
}

type RollingAvg struct {
	Sum       float64
	SquareSum float64
	N         int
}

func (r *RollingAvg) Add(x float64) {
	r.N++
	r.Sum += x
	r.SquareSum += x * x
}

func (r *RollingAvg) Variance() float64 {
	div := 1 / float64(r.N)
	return div*r.SquareSum - math.Pow(div*r.Sum, 2)
}
