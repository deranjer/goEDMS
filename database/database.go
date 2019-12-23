package database

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/asdine/storm"
	"github.com/blevesearch/bleve"
	config "github.com/deranjer/goEDMS/config"
	"github.com/oklog/ulid/v2"
	"github.com/ziflex/lecho/v2"
)

//Document is all of the document information stored in the database
type Document struct {
	StormID      int `storm:"id,increment=100"` //all records start at 100 for the ID and go up
	Name         string
	Path         string //full path to the file
	IngressTime  time.Time
	Folder       string
	Hash         string
	ULID         ulid.ULID `storm:"index"` //Have a smaller (than hash) id that can be used in URL's, hopefully speed things up
	DocumentType string    //type of document (pdf, txt, etc)
	FullText     string
}

//Logger is global since we will need it everywhere
var Logger *lecho.Logger

//SetupDatabase initializes the storm/bbolt database
func SetupDatabase() (db *storm.DB) {
	db, err := storm.Open("goEDMS.db")
	if err != nil {
		Logger.Fatal("Unable to create/open database!", err)
	}
	return db
}

//FetchConfigFromDB pulls the server config from the database
func FetchConfigFromDB(db *storm.DB) (config.ServerConfig, error) {
	var serverConfig config.ServerConfig
	err := db.One("StormID", 1, &serverConfig)
	if err != nil {
		Logger.Fatal("Unable to fetch server config from db!", err)
		return serverConfig, err
	}
	return serverConfig, nil
}

//WriteConfigToDB writes the serverconfig to the database for later retrieval
func WriteConfigToDB(serverConfig config.ServerConfig, db *storm.DB) {
	serverConfig.StormID = 1 //config will be stored in bucket 1
	fmt.Printf("%+v\n", serverConfig)
	err := db.Save(&serverConfig)
	if err != nil {
		Logger.Error("Unable to write server config to database!", err)
	}
}

//AddNewDocument adds a new document to the database
func AddNewDocument(fileName string, fullText string, db *storm.DB, searchDB bleve.Index) (Document, error) {
	serverConfig, err := FetchConfigFromDB(db)
	if err != nil {
		fmt.Println("Can't get the goddamn config")
	}
	var newDocument Document
	fileHash, err := calculateHash(fileName)
	if err != nil {
		return newDocument, err
	}
	duplicate := checkDuplicateDocument(fileHash, fileName, db)
	if duplicate {
		err = errors.New("Duplicate document found on import (Hash collision) ! " + fileName)
		Logger.Error(err)
		return newDocument, err //TODO return actual error
	}
	newTime := time.Now()
	newULID, err := calculateUUID(newTime)
	if err != nil {
		return newDocument, err
	}
	newDocument.Name = filepath.Base(fileName)
	newDocument.Path = fileName
	newDocument.Hash = fileHash
	newDocument.IngressTime = newTime
	newDocument.ULID = newULID
	documentFolder := serverConfig.DocumentPath + "/" + serverConfig.IngressMoveFolder
	newDocument.Folder = documentFolder
	newDocument.DocumentType = filepath.Ext(fileName)
	newDocument.FullText = fullText

	searchDB.Index(newDocument.ULID.String(), newDocument.FullText) //adding to bleve using the ULID as the ID and the fulltext TODO: Perhaps add entire struct this will give more search options
	if err != nil {
		Logger.Error("Unable to index Document in Bleve Search", newDocument.Name, err)
	}
	err = db.Save(&newDocument) //Writing it in document bucket
	if err != nil {
		Logger.Fatal("Unable to write document to bucket!", err)
	}
	return newDocument, err
}

//FetchNewestDocuments fetches the documents that were added last
func FetchNewestDocuments(numberOf int, db *storm.DB) ([]Document, error) {
	var newestDocuments []Document
	err := db.AllByIndex("StormID", &newestDocuments, storm.Limit(numberOf), storm.Reverse())
	//err := db.Find("StormID", &newestDocuments, storm.Limit(numberOf), storm.Reverse()) //getting it from the last added
	if err != nil {
		Logger.Error("Unable to find the latest documents: ", err)
		return newestDocuments, err
	}
	return newestDocuments, nil
}

//FetchDocuments fetches an array of documents //TODO: Not fucking needed?
func FetchDocuments(docULIDSt []string, db *storm.DB) ([]Document, int, error) {
	var foundDocuments []Document
	var tempDocument Document
	//var foundULIDs []ulid.ULID
	for _, ulidStr := range docULIDSt {
		docULID, err := ulid.Parse(ulidStr)
		if err != nil {
			Logger.Error("Failed to parse UILD: ", ulidStr, err)
			return foundDocuments, http.StatusNotFound, err
		}
		//foundULIDs = append(foundULIDs, newULID)
		err = db.One("ULID", docULID, &tempDocument)
		if err != nil {
			Logger.Error("Unable to find the requested document: ", err)
			return foundDocuments, http.StatusNotFound, err
		}
		foundDocuments = append(foundDocuments, tempDocument)
	}
	return foundDocuments, http.StatusOK, nil

}

//UpdateDocumentField updates a single field in a document
func UpdateDocumentField(docULIDSt string, field string, newValue interface{}, db *storm.DB) (int, error) {
	var newDocument Document
	docULID, err := ulid.Parse(docULIDSt)
	if err != nil {
		Logger.Error("Unable to parse ULID String to convert to ID: ", err)
		return http.StatusNotFound, err
	}
	err = db.One("ULID", docULID, &newDocument)
	if err != nil {
		Logger.Error("Unable to find document with ID: ", docULID, err)
	}
	stormIDDoc := newDocument.StormID
	err = db.UpdateField(&Document{StormID: stormIDDoc}, field, newValue)
	if err != nil {
		Logger.Error("Unable to update document in db: ID: ", docULID, err)
		return http.StatusNotFound, err
	}
	return http.StatusOK, nil

}

//FetchDocument fetches the requested document by ULID
func FetchDocument(docULIDSt string, db *storm.DB) (Document, int, error) {
	var foundDocument Document
	fmt.Println("UUID STRING: ", docULIDSt)
	docULID, err := ulid.Parse(docULIDSt) //convert string into ULID
	if err != nil {
		Logger.Error("Unable to parse ULID String to convert to ID: ", err)
		return foundDocument, http.StatusNotFound, err
	}
	err = db.One("ULID", docULID, &foundDocument)
	if err != nil {
		Logger.Error("Unable to find the requested document: ", err)
		return foundDocument, http.StatusNotFound, err
	}
	return foundDocument, http.StatusOK, nil
}

//FetchFolder grabs all of the documents contained in a folder
func FetchFolder(folderName string, db *storm.DB) ([]Document, error) {
	var folderContents []Document
	err := db.Find("Folder", folderName, &folderContents) //TODO limit this?
	if err != nil {
		Logger.Error("Unable to find the requested folder: ", err)
		return folderContents, err
	}
	return folderContents, nil
}

//DeleteDocument fetches the requested document by ULID
func DeleteDocument(docULIDSt string, db *storm.DB) error {
	deleteDocument, _, err := FetchDocument(docULIDSt, db)
	if err != nil {
		Logger.Error("Unable to fetch document for deletion: ", err)
	}
	err = db.DeleteStruct(&deleteDocument)
	if err != nil {
		Logger.Error("Unable to delete requested document: ", err)
		return err
	}
	return nil
}

//DeleteDocumentFromSearch deletes everything in the search engine
func DeleteDocumentFromSearch(deleteDocument Document, searchDB bleve.Index) error {
	err := searchDB.Delete(deleteDocument.ULID.String()) //Delete everything tied to the ULID in bleve
	if err != nil {
		Logger.Error("Unable to delete document index in Bleve Search", deleteDocument.Name, err)
		return err
	}
	return nil
}

func checkDuplicateDocument(fileHash string, fileName string, db *storm.DB) bool { //TODO: Check for duplicates before you do a shit ton of processing, why wasn't this obvious?
	var document Document
	err := db.One("Hash", fileHash, &document)
	if err != nil {
		Logger.Info("No record found, assume no duplicate hash", err)
		return false
	}
	Logger.Info("Duplicate document found on import (Hash collision) !" + fileName + " With documentTHISDOCUMENT!!!!: " + document.Name)
	return true
}

//calculate a UUID for the incoming file
func calculateUUID(time time.Time) (ulid.ULID, error) {
	entropy := ulid.Monotonic(rand.New(rand.NewSource(time.UnixNano())), 0)
	newULID, err := ulid.New(ulid.Timestamp(time), entropy)
	return newULID, err
}

//calculate the hash of the incoming file
func calculateHash(fileName string) (string, error) {
	var fileHash string
	file, err := os.Open(fileName)
	if err != nil {
		return fileHash, err
	}
	defer file.Close()
	hash := md5.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return fileHash, err
	}
	fileHash = fmt.Sprintf("%x", hash.Sum(nil))
	return fileHash, nil
}
