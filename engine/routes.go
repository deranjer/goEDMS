package engine

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"unicode"

	"github.com/asdine/storm"
	"github.com/blevesearch/bleve"
	"github.com/deranjer/goEDMS/config"
	"github.com/deranjer/goEDMS/database"
	"github.com/labstack/echo/v4"
)

//ServerHandler will inject the variables needed into routes
type ServerHandler struct {
	DB           *storm.DB
	SearchDB     bleve.Index
	Echo         *echo.Echo
	ServerConfig config.ServerConfig
}

type Node struct {
	FullPath     string  `json:"path"`
	Name         string  `json:"name"`
	Size         int64   `json:"size"`
	DateModified string  `json:"dateModified"`
	Thumbnail    string  `json:"thumbnail"`
	IsDirectory  bool    `json:"isDirectory"`
	Children     []*Node `json:"items"`
	FileExt      string  `json:"fileExt"`
	ULID         string  `json:"ulid"`
	URL          string  `json:"documentURL"`
	Parent       *Node   `json:"-"`
}

//DeleteDocument deletes a document from the database (and on disc and from bleve search)
func (serverHandler *ServerHandler) DeleteDocument(context echo.Context) error {
	//ulid, err := strconv.Atoi(context.Param("id"))
	ulidStr := context.Param("id")
	document, _, _ := database.FetchDocument(ulidStr, serverHandler.DB)
	err := database.DeleteDocument(ulidStr, serverHandler.DB)
	if err != nil {
		Logger.Error("Unable to delete document from database: ", document.Name, err)
		return context.JSON(http.StatusNotFound, err)
	}
	err = DeleteDocumentFile(document.Path)
	if err != nil {
		Logger.Error("Unable to delete document from file system: ", document.Path, err)
		return context.JSON(http.StatusNotFound, err)
	}
	err = database.DeleteDocumentFromSearch(document, serverHandler.SearchDB)
	if err != nil {
		Logger.Error("Unable to delete document from bleve search: ", document.Path, err)
		return context.JSON(http.StatusNotFound, err)
	}
	return context.JSON(http.StatusOK, "Document Deleted")
}

//MoveDocuments will accept an API call from the frontend to move a document or documents
func (serverHandler *ServerHandler) MoveDocuments(context echo.Context) error {
	var docIDs url.Values
	var newFolder string
	//var foundDocuments []database.Document
	docIDs = context.QueryParams()
	newFolder = docIDs.Get("folder")
	fmt.Println("newfolder: ", newFolder)
	fmt.Println("ID's: ", docIDs["id"])
	for _, docID := range docIDs["id"] { //fetching all the needed documents
		//document, httpStatus, err := database.FetchDocument(docID, serverHandler.DB)
		//if err != nil {
		//	Logger.Error("GetDocument API call failed (MoveDocuments): ", err)
		//	return context.JSON(httpStatus, err)
		//}
		//foundDocuments = append(foundDocuments, document)
		httpStatus, err := database.UpdateDocumentField(docID, "Folder", newFolder, serverHandler.DB)
		if err != nil {
			Logger.Error("GetDocument API call failed (MoveDocuments): ", err)
			return context.JSON(httpStatus, err)
		}
	}

	return context.JSON(http.StatusOK, "Ok")
}

//SearchDocuments will take the search terms and search all documents
func (serverHandler *ServerHandler) SearchDocuments(context echo.Context) error {
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
			searchResults, err = SearchPhrase(searchTerm, serverHandler.SearchDB)
			if err != nil {
				Logger.Error("Search failed: ", err)
				return context.JSON(http.StatusNotFound, err)
			}
		}
	}
	if !phraseSearch { //if no space found in search term
		Logger.Debug("Performing Single Term Search: ", searchTerm)
		searchResults, err = SearchSingleTerm(searchTerm, serverHandler.SearchDB)
	}
	documents, err := ParseSearchResults(searchResults, serverHandler.DB)
	if err != nil {
		Logger.Error("Unable to get documents from search: ", err)
		return context.JSON(http.StatusNotFound, err)
	}
	return context.JSON(http.StatusOK, documents)
}

//GetDocument will return a document by ULID
func (serverHandler *ServerHandler) GetDocument(context echo.Context) error {
	ulidStr := context.Param("id")
	document, httpStatus, err := database.FetchDocument(ulidStr, serverHandler.DB)
	if err != nil {
		Logger.Error("GetDocument API call failed: ", err)
		return context.JSON(httpStatus, err)
	}
	return context.JSON(httpStatus, document)

}

//GetDocumentFileSystem will scan the document folder and get the complete tree to send to the frontend
func (serverHandler *ServerHandler) GetDocumentFileSystem(context echo.Context) error {
	fileSystemNodes, err := documentFileTree(serverHandler.ServerConfig.DocumentPath, serverHandler.DB)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", fileSystemNodes)
	return context.JSON(http.StatusOK, fileSystemNodes)

}

func documentFileTree(rootPath string, db *storm.DB) (result *Node, err error) {
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return
	}
	parents := make(map[string]*Node)
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		var document database.Document
		if !info.IsDir() {
			document, err = database.FetchDocumentFromPath(filepath.ToSlash(path), db)
			if err != nil {
				Logger.Error("Unable to fetch document: ", err, path)
			}
		}

		parents[path] = &Node{
			FullPath:     filepath.ToSlash(path),
			Name:         info.Name(),
			Size:         info.Size(),
			DateModified: info.ModTime().String(),
			Thumbnail:    "",
			FileExt:      filepath.Ext(path),
			ULID:         document.ULID.String(),
			URL:          document.URL,
			IsDirectory:  info.IsDir(),
			Children:     make([]*Node, 0),
		}
		return nil
	}
	if err = filepath.Walk(absRoot, walkFunc); err != nil {
		return
	}
	for path, node := range parents {
		parentPath := filepath.Dir(path)
		parent, exists := parents[parentPath]
		if !exists { // If a parent does not exist, this is the root.
			result = node
		} else {
			node.Parent = parent
			parent.Children = append(parent.Children, node)
		}
	}
	return
}

//GetLatestDocuments gets the latest documents that were ingressed
func (serverHandler *ServerHandler) GetLatestDocuments(context echo.Context) error {
	serverConfig, err := database.FetchConfigFromDB(serverHandler.DB)
	if err != nil {
		Logger.Error("Unable to pull config from database for GetLatestDocuments", err)
	}
	newDocuments, err := database.FetchNewestDocuments(serverConfig.FrontEndConfig.NewDocumentNumber, serverHandler.DB)
	if err != nil {
		Logger.Error("Can't find latest documents, might not have any: ", err)
		return err
	}
	return context.JSON(http.StatusOK, newDocuments)
}

//GetFolder fetches all the documents in the folder
func (serverHandler *ServerHandler) GetFolder(context echo.Context) error {
	folderName := context.Param("folder")

	folderContents, err := database.FetchFolder(folderName, serverHandler.DB)
	if err != nil {
		Logger.Error("API GetFolder call failed: ", err)
		return err
	}
	return context.JSON(http.StatusOK, folderContents)

}
