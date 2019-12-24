package engine

import (
	"github.com/asdine/storm"
	"github.com/blevesearch/bleve"
	"github.com/deranjer/goEDMS/database"
)

//SearchPhrase is used to search using multiple words instead of just one word
func SearchPhrase(phrase string, index bleve.Index) (*bleve.SearchResult, error) {
	phraseQuery := bleve.NewMatchPhraseQuery(phrase)
	searchQuery := bleve.NewSearchRequest(phraseQuery)
	searchResultsQuery, err := index.Search(searchQuery)
	Logger.Debug("Search Results for phrase search: ", searchResultsQuery)
	if err != nil {
		Logger.Error("SearchPhrase failed with error: ", err)
		return nil, err
	}
	return searchResultsQuery, nil
}

//SearchSingleTerm searches for just one word
func SearchSingleTerm(term string, index bleve.Index) (*bleve.SearchResult, error) {
	query := bleve.NewMatchQuery(term)
	search := bleve.NewSearchRequest(query)
	searchResults, err := index.Search(search)
	Logger.Debug("Search Results for single term: ", searchResults)
	if err != nil {
		Logger.Error("SearchTerm failed with error: ", err)
		return nil, err
	}
	return searchResults, nil
}

//ParseSearchResults takes a result of any search and gets the document ID's, pulls the files from the database and sends them back
func ParseSearchResults(results *bleve.SearchResult, db *storm.DB) ([]database.Document, error) {
	var documentIDs []string
	searchHits := results.Hits
	for _, hit := range searchHits {
		Logger.Debug("Search found a hit: ", hit.ID)
		documentIDs = append(documentIDs, hit.ID)
	}
	documentResults, _, err := database.FetchDocuments(documentIDs, db)
	if err != nil {
		Logger.Error("Unable to fetch documents from database: ", err)
		return nil, err
	}
	return documentResults, nil
}
