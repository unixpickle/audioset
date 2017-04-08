// Command find_bad finds corrupted audio samples in a
// directory of AudioSet samples.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
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

	for _, item := range listing {
		if strings.HasSuffix(item.Name(), ".wav.gz") ||
			strings.HasSuffix(item.Name(), ".wav") {
			path := filepath.Join(dirPath, item.Name())
			sample := metaset.Sample{Path: path}
			if _, err := sample.Read(); err != nil {
				fmt.Println(path)
			}
		}
	}
}
