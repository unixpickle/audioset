package audioset

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"sort"
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

// Classes returns all the classes found in the dataset,
// sorted alphabetically.
func (s Set) Classes() []string {
	m := map[string]bool{}
	for _, sample := range s {
		for _, class := range sample.Classes {
			m[class] = true
		}
	}

	var res []string
	for class := range m {
		res = append(res, class)
	}
	sort.Strings(res)

	return res
}
