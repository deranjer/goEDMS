package engine

import (
	"fmt"
	"net/http"
	"net/url"
	"unicode"

	"github.com/asdine/storm"
	"github.com/blevesearch/bleve"
	"github.com/deranjer/goEDMS/database"
	"github.com/labstack/echo/v4"
)

//DBHandler will inject the database into routes
type DBHandler struct {
	DB       *storm.DB
	SearchDB bleve.Index
}

//DeleteDocument deletes a document from the database (and on disc and from bleve search)
func (dbHandle *DBHandler) DeleteDocument(context echo.Context) error {
	//ulid, err := strconv.Atoi(context.Param("id"))
	ulidStr := context.Param("id")
	document, _, _ := database.FetchDocument(ulidStr, dbHandle.DB)
	err := database.DeleteDocument(ulidStr, dbHandle.DB)
	if err != nil {
		Logger.Error("Unable to delete document from database: ", document.Name, err)
		return context.JSON(http.StatusNotFound, err)
	}
	err = DeleteDocumentFile(document.Path)
	if err != nil {
		Logger.Error("Unable to delete document from file system: ", document.Path, err)
		return context.JSON(http.StatusNotFound, err)
	}
	err = database.DeleteDocumentFromSearch(document, dbHandle.SearchDB)
	if err != nil {
		Logger.Error("Unable to delete document from bleve search: ", document.Path, err)
		return context.JSON(http.StatusNotFound, err)
	}
	return context.JSON(http.StatusOK, "Document Deleted")
}

//MoveDocuments will accept an API call from the frontend to move a document or documents
func (dbHandle *DBHandler) MoveDocuments(context echo.Context) error {
	var docIDs url.Values
	var newFolder string
	//var foundDocuments []database.Document
	docIDs = context.QueryParams()
	newFolder = docIDs.Get("folder")
	fmt.Println("newfolder: ", newFolder)
	fmt.Println("ID's: ", docIDs["id"])
	for _, docID := range docIDs["id"] { //fetching all the needed documents
		//document, httpStatus, err := database.FetchDocument(docID, dbHandle.DB)
		//if err != nil {
		//	Logger.Error("GetDocument API call failed (MoveDocuments): ", err)
		//	return context.JSON(httpStatus, err)
		//}
		//foundDocuments = append(foundDocuments, document)
		httpStatus, err := database.UpdateDocumentField(docID, "Folder", newFolder, dbHandle.DB)
		if err != nil {
			Logger.Error("GetDocument API call failed (MoveDocuments): ", err)
			return context.JSON(httpStatus, err)
		}
	}

	return context.JSON(http.StatusOK, "Ok")
}

//SearchDocuments will take the search terms and search all documents
func (dbHandle *DBHandler) SearchDocuments(context echo.Context) error {
	searchParams := context.QueryParams()
	searchTerm := searchParams.Get("term")
	if searchTerm == "" {
		return context.JSON(http.StatusNotFound, "Empty search term")
	}
	var phraseSearch bool
	var searchResults *bleve.SearchResult
	var err error
	for _, char := range searchTerm {
		if unicode.IsSpace(char) { //if there is a space in the result, do a phrase search
			Logger.Debug("Found space in search term, converting to phrase: ", searchTerm)
			phraseSearch = true
			searchResults, err = SearchPhrase(searchTerm, dbHandle.SearchDB)
			if err != nil {
				Logger.Error("Search failed: ", err)
				return context.JSON(http.StatusNotFound, err)
			}
		}
	}
	if !phraseSearch { //if no space found in search term
		Logger.Debug("Performing Single Term Search: ", searchTerm)
		searchResults, err = SearchSingleTerm(searchTerm, dbHandle.SearchDB)
	}
	documents, err := ParseSearchResults(searchResults, dbHandle.DB)
	if err != nil {
		Logger.Error("Unable to get documents from search: ", err)
		return context.JSON(http.StatusNotFound, err)
	}
	return context.JSON(http.StatusOK, documents)
}

//GetDocument will return a document by ULID
func (dbHandle *DBHandler) GetDocument(context echo.Context) error {
	ulidStr := context.Param("id")
	document, httpStatus, err := database.FetchDocument(ulidStr, dbHandle.DB)
	if err != nil {
		Logger.Error("GetDocument API call failed: ", err)
		return context.JSON(httpStatus, err)
	}
	return context.JSON(httpStatus, document)

}

//GetLatestDocuments gets the latest documents that were ingressed
func (dbHandle *DBHandler) GetLatestDocuments(context echo.Context) error {
	serverConfig, err := database.FetchConfigFromDB(dbHandle.DB)
	if err != nil {
		Logger.Error("Unable to pull config from database for GetLatestDocuments", err)
	}
	newDocuments, err := database.FetchNewestDocuments(serverConfig.FrontEndConfig.NewDocumentNumber, dbHandle.DB)
	if err != nil {
		Logger.Error("Can't find latest documents, might not have any: ", err)
		return err
	}
	return context.JSON(http.StatusOK, newDocuments)
}

//GetFolder fetches all the documents in the folder
func (dbHandle *DBHandler) GetFolder(context echo.Context) error {
	folderName := context.Param("folder")

	folderContents, err := database.FetchFolder(folderName, dbHandle.DB)
	if err != nil {
		Logger.Error("API GetFolder call failed: ", err)
		return err
	}
	return context.JSON(http.StatusOK, folderContents)

}
