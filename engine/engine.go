package engine

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/asdine/storm"
	"github.com/blevesearch/bleve"
	"github.com/deranjer/goEDMS/config"
	"github.com/deranjer/goEDMS/database"
	"github.com/ledongthuc/pdf"
)

func ingressJobFunc(serverConfig config.ServerConfig, db *storm.DB, searchDB bleve.Index) {
	//fmt.Println("Testing job func")
	//serverConfig := database.FetchConfigFromDB(db)
	serverConfig, err := database.FetchConfigFromDB(db)
	if err != nil {
		Logger.Error("Error reading config from database: ", err)
	}
	var fullText string //The text extraction result whether it is OCR or other method
	Logger.Info("Starting Ingress Job on folder:", serverConfig.IngressPath)
	var documentPath []string
	err = filepath.Walk(serverConfig.IngressPath, func(path string, info os.FileInfo, err error) error {
		documentPath = append(documentPath, path)
		return nil
	})
	if err != nil {
		Logger.Error("Error reading files in from ingress!")
	}
	for _, file := range documentPath {
		switch filepath.Ext(file) {
		case ".pdf":
			fullText, err = pdfProcessing(file)
			if err != nil {
				fullText, err = ocrProcessing(file)
				if err != nil {
					Logger.Error("OCR Processing failed on file: ", file, err)
					continue
				}
			}
			document, err := database.AddNewDocument(file, fullText, db, searchDB)
			if err != nil {
				fmt.Println("UNABLE TO ADD NEW DOCUMENT", document, err)
			}
			err = ingressCopyDocument(file, serverConfig)
			if err != nil {
				Logger.Error("Error moving ingress file to new location! ", file, err)
			}
			ingressCleanup(file, document, serverConfig, db)

		case ".txt", ".rtf":
			textProcessing(file)
		case ".doc", ".docx", ".odf":
			wordDocProcessing(file)
		case ".tiff", ".jpg", ".jpeg", ".png":
			ocrProcessing(file)
		default:
			Logger.Warn("Invalid file type: ", filepath.Base((file)))
		}
		continue
	}

}

//DeleteFolder deletes a folder and everything in that folder
func DeleteFolder(folderName string) error {
	err := os.RemoveAll(folderName)
	if err != nil {
		Logger.Error("Error deleting Folder: ", err)
		return err
	}
	return nil
}

//DeleteDocumentFile deletes a file from the filesystem(database deletion handled in db)
func DeleteDocumentFile(fileName string) error {
	err := os.Remove(fileName)
	if err != nil {
		Logger.Error("Unable to delete file: ", err)
		return err
	}
	return nil
}

//ingressCopyDocument copies the document to document storage location
func ingressCopyDocument(fileName string, serverConfig config.ServerConfig) error {
	srcFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	newFileName := filepath.ToSlash(serverConfig.NewDocumentFolder + "/" + filepath.Base(fileName))
	err = ioutil.WriteFile(newFileName, srcFile, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

//ingressCleanup cleans up the ingress folder after we have handled the documents
func ingressCleanup(fileName string, document database.Document, serverConfig config.ServerConfig, db *storm.DB) error {
	if serverConfig.IngressDelete == true { //deleting the ingress files
		err := os.Remove(fileName)
		if err != nil {
			return err
		}
		return nil
	}
	newFile := serverConfig.IngressMoveFolder + "/" + filepath.Base(fileName) //Moving ingress files to another location
	newFile = filepath.FromSlash(newFile)
	err := os.Rename(fileName, newFile)
	if err != nil {
		return err
	}
	//_, err = database.UpdateDocumentField(document.ULID.String(), "Path", newFile, db) //updating the database with the new file location
	//if err != nil {
	//		Logger.Error("Unable to update document field: Path ", err)
	//	}
	return nil
}

func pdfProcessing(file string) (string, error) {
	fileName := filepath.Base((file))
	var fullText string
	Logger.Debug("Working on current file: ", fileName)
	pdfFile, result, err := pdf.Open(file)
	if err != nil {
		Logger.Error("Unable to open PDF", fileName)
		return fullText, err
	}
	defer pdfFile.Close()
	var buf bytes.Buffer
	bytes, err := result.GetPlainText()
	if err != nil {
		Logger.Error("Unable to convert PDF to text", fileName)
		return fullText, err
	}
	buf.ReadFrom(bytes)
	fullText = buf.String() //writing from the buffer to the string
	if fullText == "" {
		err = errors.New("PDF Text Result is empty")
		Logger.Info("PDF Text Result is empty, sending to OCR: ", fileName, err)
		return fullText, err
	}
	Logger.Info("Text processed from PDF without OCR: ", fileName)
	return fullText, nil
}

func textProcessing(fileName string) {

}

func wordDocProcessing(fileName string) {

}

func ocrProcessing(fileName string) (string, error) {
	var err error
	var output []byte
	var fullText string
	Logger.Info("Converting PDF To image for OCR", fileName)
	imageName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	imageName = filepath.Base(fmt.Sprint(imageName + ".png"))
	imageName = filepath.Join("temp", imageName)
	imageName, err = filepath.Abs(imageName)
	if err != nil {
		Logger.Error("Unable to edit absolute path string for temporary image for OCR: ", fileName, err)
	}
	err = os.MkdirAll(filepath.Dir(imageName), os.ModePerm)
	if err != nil {
		Logger.Error("Unable to create absolute path for temporary image for OCR (permissions?): ", filepath.Dir(imageName), err)
	}
	fileName = filepath.Clean(fileName)
	imageName = filepath.Clean(imageName)
	fmt.Println("Creating temp image for OCR AT: ", imageName)
	_, err = os.OpenFile(fileName, os.O_RDWR, 0755)
	if err != nil {
		fmt.Println("ERROR FILE ISSUE", err)
	}
	convertArgs := []string{"convert", "-density", "150", "-antialias", fileName, "-append", "-resize", "1024x", "-quality", "100", imageName}
	pdfConvertCmd := exec.Command("magick", convertArgs...)
	output, err = pdfConvertCmd.Output()
	if err != nil {
		Logger.Error("Unable to convert PDF Using Magick: ", fileName, err)
		return fullText, err
	}
	fmt.Println("Outputting image to ", imageName)
	//fmt.Println("Output from pdfConvertCmd ", string(output))
	cleanArgs := []string{"convert", imageName, "-auto-orient", "-deskew", "40%", "-despeckle", imageName}
	imageCleanCmd := exec.Command("magick", cleanArgs...)
	output, err = imageCleanCmd.Output()
	if err != nil {
		Logger.Error("Magick was unable to clean the image for some reason... skipping this file for now: ", fileName, err)
		return fullText, err
	}
	//fmt.Println("Output from imageCleanCmd", string(output))
	tesseractArgs := []string{imageName, "stdout"}
	tesseractCMD := exec.Command("tesseract", tesseractArgs...) //get the path to tesseract
	output, err = tesseractCMD.Output()
	if err != nil {
		Logger.Error("Tesseract encountered error when attempting to OCR image: ", fileName, err)
		return fullText, err
	}
	fullText = string(output)
	if fullText == "" {
		Logger.Error("OCR Result returned empty string... OCR'ing the document failed", fileName, err)
		return fullText, err
	}
	return fullText, nil
}
