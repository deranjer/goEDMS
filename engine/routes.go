package engine

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
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

/* type Node struct {
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
} */

type fileTreeStruct struct {
	ID          string   `json:"id"`
	ULIDStr     string   `json:"ulid"`
	Name        string   `json:"name"`
	Size        int64    `json:"size"`
	ModDate     string   `json:"modDate"`
	Openable    bool     `json:"openable"`
	ParentID    string   `json:"parentID"`
	IsDir       bool     `json:"isDir"`
	ChildrenIDs []string `json:"childrenIDs"`
	FullPath    string   `json:"fullPath"`
	FileURL     string   `json:"fileURL"`
}

//AddDocumentViewRoutes adds all of the current documents to an echo route
func (serverHandler *ServerHandler) AddDocumentViewRoutes() error {
	documents, err := database.FetchAllDocuments(serverHandler.DB)
	if err != nil {
		return err
	}
	for _, document := range *documents {
		documentURL := "/document/view/" + document.ULID.String()
		serverHandler.Echo.File(documentURL, document.Path)
	}
	return nil
}

//DeleteFile deletes a folder or file from the database (and all children if folder) (and on disc and from bleve search if document)
func (serverHandler *ServerHandler) DeleteFile(context echo.Context) error {
	var err error
	ulidStr := context.Param("id")
	path := context.Param("path")
	path = filepath.Join(serverHandler.ServerConfig.DocumentPath, path)
	path, err = filepath.Abs(path)
	if err != nil {
		return err
	}
	fmt.Println("PATH", path)
	if path == serverHandler.ServerConfig.DocumentPath { //TODO: IMPORTANT: Make this MUCH safer so we don't literally purge everything in root lol (side note, yes I did discover that the hard way)
		return err
	}

	return context.JSON(http.StatusOK, path)
	fileInfo, err := os.Stat(path)
	if err != nil {
		Logger.Error("Unable to get information for file: ", path, err)
		return context.JSON(http.StatusNotFound, err)
	}
	if fileInfo.IsDir() { //If a directory, just delete it and all children
		err = DeleteFile(path)
		if err != nil {
			Logger.Error("Unable to delete folder from document filesystem ", path, err)
			return context.JSON(http.StatusInternalServerError, err)
		}
		return context.JSON(http.StatusOK, "Folder Deleted")
	}
	document, _, err := database.FetchDocument(ulidStr, serverHandler.DB)
	if err != nil {
		Logger.Error("Unable to delete folder from document filesystem ", path, err)
		return context.JSON(http.StatusNotFound, err)
	}
	err = database.DeleteDocument(ulidStr, serverHandler.DB)
	if err != nil {
		Logger.Error("Unable to delete document from database: ", document.Name, err)
		return context.JSON(http.StatusNotFound, err)
	}
	err = DeleteFile(document.Path)
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

//UploadDocuments handles documents uploaded from the frontend
func (serverHandler *ServerHandler) UploadDocuments(context echo.Context) error {
	//context.Get()
	//formResult := context.FormValue("file")
	request := context.Request()
	uploadPath := request.FormValue("path")
	file, fileHeader, err := request.FormFile("file")
	if err != nil {
		fmt.Println("Problem finding file, ", err)
		return err
	}
	defer file.Close()
	path := filepath.ToSlash(serverHandler.ServerConfig.DocumentPath + "/" + uploadPath + fileHeader.Filename)
	body, err := ioutil.ReadAll(file) //get the file, write it to the filesystem
	err = ioutil.WriteFile(path, body, 0644)
	if err != nil {
		return err
	}
	serverHandler.ingressDocument(path, "upload") //ingress the document into the database
	return context.JSON(http.StatusOK, path)
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
	fileSystem, err := fileTree(serverHandler.ServerConfig.DocumentPath, serverHandler.DB)
	if err != nil {
		return err
	}
	//fileSystem := fileSystem{FolderTree: *folderTree, FileTree: *documents}
	return context.JSON(http.StatusOK, fileSystem)

}

func fileTree(rootPath string, db *storm.DB) (fileTree *[]fileTreeStruct, err error) {
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}
	var fullFileTree []fileTreeStruct
	var currentFile fileTreeStruct

	walkFunc := func(path string, info os.FileInfo, err error) error {
		newTime := time.Now()
		//fmt.Println("Generating based on time", newTime)

		if err != nil {
			return err
		}
		currentFile.Name = info.Name()
		currentFile.FullPath = path

		for _, fileElement := range fullFileTree { //Find the parentID
			if fileElement.FullPath == filepath.Dir(path) {
				currentFile.ParentID = fileElement.ID
			}
		}

		if info.IsDir() {
			ULID, err := database.CalculateUUID(newTime)
			//fmt.Println("New ULID for: ", path, ULID.String())
			if err != nil {
				return err
			}
			currentFile.ID = ULID.String() + filepath.Base(path) //TODO, should I store the entire filesystem layout?  Most likely yes?
			currentFile.IsDir = true
			currentFile.Openable = true
			childIDs, err := getChildrenIDs(path)
			if err != nil {
				return err
			}
			currentFile.ChildrenIDs = *childIDs
			/* 			if path == rootPath {
				fullFileTree = append(fullFileTree, currentFile)
				return nil
			} */
		} else { //for files process size, moddate, ulid
			currentFile.Size = info.Size()
			currentFile.Openable = true
			currentFile.IsDir = false
			currentFile.ModDate = info.ModTime().String()

			document, err := database.FetchDocumentFromPath(path, db)
			if err != nil {
				return err
			}
			currentFile.FileURL = document.URL
			currentFile.ID = document.ULID.String()
			currentFile.ULIDStr = document.ULID.String()
		}

		fullFileTree = append(fullFileTree, currentFile)
		return nil
	}
	err = filepath.Walk(absRoot, walkFunc)
	if err != nil {
		return nil, err
	}
	return &fullFileTree, nil
}

func getChildrenIDs(rootPath string) (*[]string, error) {
	results, err := ioutil.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}
	var childIDs []string
	for _, result := range results {
		childIDs = append(childIDs, result.Name())
	}
	return &childIDs, nil

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

//CreateFolder creates a folder in the document tree
func (serverHandler *ServerHandler) CreateFolder(context echo.Context) error {
	params := context.QueryParams()
	folderName := params.Get("folder")
	folderPath := params.Get("path")
	fullFolder := filepath.Join(folderPath, folderName)
	fullFolder = filepath.Join(serverHandler.ServerConfig.DocumentPath, fullFolder)
	fullFolder = filepath.Clean(fullFolder)
	fmt.Println("fullfolder: ", fullFolder, " folderName: ", folderName, "Path: ", folderPath)
	err := os.Mkdir(fullFolder, os.ModePerm)
	if err != nil {
		Logger.Error("Unable to create directory: ", err)
		return err
	}
	serverHandler.GetDocumentFileSystem(context)
	return context.JSON(http.StatusOK, fullFolder)
}

//TODO: for a different react frontend that requires a nested JSON structure, also used for recreating dir structure in ingress
/* func folderTree(rootPath string) (folderTree *[]folderTreeStruct, err error) {
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}

	var fullFolderTree []folderTreeStruct
	var currentFolder folderTreeStruct
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			currentFolder.ID = info.Name()
			currentFolder.Name = info.Name()
			currentFolder.IsDir = true
			currentFolder.Openable = true
			childIDs, err := getChildrenIDs(path)
			if err != nil {
				return err
			}
			currentFolder.ChildrenIDs = *childIDs
			if path == rootPath {
				fullFolderTree = append(fullFolderTree, currentFolder)
				return nil
			}
			getDir := filepath.Dir(path)
			currentFolder.ParentID = filepath.Base(getDir) //purging the end folder
			fullFolderTree = append(fullFolderTree, currentFolder)
		}
		return nil
	}
	err = filepath.Walk(absRoot, walkFunc)
	if err != nil {
		return nil, err
	}
	return &fullFolderTree, nil
} */

/* func documentFileTree(rootPath string, db *storm.DB) (result *Node, err error) {
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}
	parents := make(map[string]*Node)
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		var document database.Document
		if !info.IsDir() {
			document, err = database.FetchDocumentFromPath(path, db)
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
} */
