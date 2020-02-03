package database

import (
	"os"
	"path/filepath"

	"github.com/blevesearch/bleve"
)

//SetupSearchDB sets up new bleve or opens existing
func SetupSearchDB() (bleve.Index, error) {
	mapping := bleve.NewIndexMapping()
	var index bleve.Index
	_, err := os.Stat(filepath.Clean("databases/simpleEDMSIndex.bleve"))
	if os.IsNotExist(err) {
		index, err = bleve.New(filepath.Clean("databases/simpleEDMSIndex.bleve"), mapping)
		if err != nil {
			return index, err
		}
	} else {
		index, err = bleve.Open("databases/simpleEDMSIndex.bleve")
		if err != nil {
			return index, err
		}
	}
	return index, nil
}
