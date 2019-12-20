package database

import (
	"github.com/blevesearch/bleve"
	"os"
)

//SetupSearchDB sets up new bleve or opens existing
func SetupSearchDB() (bleve.Index, error) {
	mapping := bleve.NewIndexMapping()
	var index bleve.Index
	_, err := os.Stat("simpleEDMSIndex.bleve")
	if os.IsNotExist(err) {
		index, err = bleve.New("simpleEDMSIndex.bleve", mapping)
		if err != nil {
			return index, err
		}
	} else {
		index, err = bleve.Open("simpleEDMSIndex.bleve")
		if err != nil {
			return index, err
		}
	}
	return index, nil
}
