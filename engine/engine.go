package engine

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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

func (serverHandler *ServerHandler) ingressJobFunc(serverConfig config.ServerConfig, db *storm.DB, searchDB bleve.Index) {
	serverConfig, err := database.FetchConfigFromDB(db)
	if err != nil {
		Logger.Error("Error reading config from database: ", err)
	}
	Logger.Info("Starting Ingress Job on folder:", serverConfig.IngressPath)
	var documentPath []string
	err = filepath.Walk(serverConfig.IngressPath, func(path string, info os.FileInfo, err error) error {
		documentPath = append(documentPath, path)
		return nil
	})
	if err != nil {
		Logger.Error("Error reading files in from ingress!")
	}
	for _, filePath := range documentPath {
		serverHandler.ingressDocument(filePath, "ingress")
	}
	deleteEmptyIngressFolders(serverHandler.ServerConfig.IngressPath) //after ingress clean empty folders
}

func (serverHandler *ServerHandler) ingressDocument(filePath string, source string) {
	switch filepath.Ext(filePath) {
	case ".pdf":
		fullText, err := pdfProcessing(filePath)
		if err != nil {
			fullText, err = convertToImage(filePath)
			if err != nil {
				Logger.Error("OCR Processing failed on file: ", filePath, err)
				return
			}
		}
		serverHandler.addDocumentToDatabase(filePath, *fullText, source)

	case ".txt", ".rtf":
		textProcessing(filePath)
	case ".doc", ".docx", ".odf":
		wordDocProcessing(filePath)
	case ".tiff", ".jpg", ".jpeg", ".png":
		fullText, err := ocrProcessing(filePath)
		if err != nil {
			Logger.Error("OCR Processing failed on file: ", filePath, err)
			return
		}
		serverHandler.addDocumentToDatabase(filePath, *fullText, source)
	default:
		Logger.Warn("Invalid file type: ", filepath.Base((filePath)))
	}
}

func (serverHandler *ServerHandler) addDocumentToDatabase(filePath string, fullText string, source string) error {
	document, err := database.AddNewDocument(filePath, fullText, serverHandler.DB, serverHandler.SearchDB) //Adds everything but the URL, that is added afterwards
	if err != nil {
		fmt.Println("UNABLE TO ADD NEW DOCUMENT", document, err) //TODO: Handle document that we were unable to add
		return err
	}
	documentURL := "/document/view/" + document.ULID.String()
	serverHandler.Echo.File(documentURL, document.Path)                                                 //Generating a direct URL to document so it is live immediately after add
	_, err = database.UpdateDocumentField(document.ULID.String(), "URL", documentURL, serverHandler.DB) //updating the database with the new file location
	if err != nil {
		Logger.Error("Unable to update document field: Path ", err)
		return err
	}
	err = ingressCopyDocument(filePath, serverHandler.ServerConfig)
	if err != nil {
		Logger.Error("Error moving ingress file to new location! ", filePath, err)
		return err
	}
	if source == "ingress" { //if file was ingressed need to handle the original, if uploaded no problem
		err := ingressCleanup(filePath, *document, serverHandler.ServerConfig, serverHandler.DB)
		if err != nil {
			return err
		}
	}
	return nil
}

func deleteEmptyIngressFolders(path string) {
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.Readdirnames(1)
		if err == io.EOF {
			os.Remove(path)
			return nil
		}
		return err
	})
}

//DeleteFile deletes a folder (or file) and everything in that folder
func DeleteFile(filePath string) error {
	err := os.RemoveAll(filePath)
	if err != nil {
		Logger.Error("Error deleting File/Folder: ", err)
		return err
	}
	return nil
}

//DeleteDocumentFile deletes a file from the filesystem(database deletion handled in db)  //TODO Not sure if needed, might just use removeAll
/* func DeleteDocumentFile(filePath string) error {
	err := os.Remove(filePath)
	if err != nil {
		Logger.Error("Unable to delete file: ", err)
		return err
	}
	return nil
} */

//ingressCopyDocument copies the document to document storage location
func ingressCopyDocument(filePath string, serverConfig config.ServerConfig) error {
	srcFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	var newFilePath string
	if serverConfig.IngressPreserve == false { //if we are not saving the folder structure just read each file in with new path
		newFilePath = filepath.ToSlash(serverConfig.NewDocumentFolder + "/" + filepath.Base(filePath))
	} else { //If we ARE preserving ingress structure, create a new full path by creating a relative path and joining it to the
		basePath := serverConfig.IngressPath
		newFileNameRoot := serverConfig.DocumentPath
		relativePath, err := filepath.Rel(basePath, filePath)
		if err != nil {
			return err
		}
		newFilePath = filepath.Join(newFileNameRoot, relativePath)
		os.MkdirAll(filepath.Dir(newFilePath), os.ModePerm) //creating the directory structure so we can write the file: TODO: not sure if ioutil.WriteFile does this for us?  Don't think so.
	}
	err = ioutil.WriteFile(newFilePath, srcFile, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

//ingressCleanup cleans up the ingress folder after we have handled the documents //TODO: Maybe ALSO preserve folder structure from ingress folder here as well?
func ingressCleanup(fileName string, document database.Document, serverConfig config.ServerConfig, db *storm.DB) error {
	if serverConfig.IngressDelete == true { //deleting the ingress files
		err := os.Remove(fileName)
		if err != nil {
			return err
		}
		return nil
	}
	newFile := filepath.FromSlash(serverConfig.IngressMoveFolder + "/" + filepath.Base(fileName)) //Moving ingress files to another location
	err := os.Rename(fileName, newFile)
	if err != nil {
		return err
	}
	return nil
}

func pdfProcessing(file string) (*string, error) {
	fileName := filepath.Base((file))
	var fullText string
	Logger.Debug("Working on current file: ", fileName)
	pdfFile, result, err := pdf.Open(file)
	if err != nil {
		Logger.Error("Unable to open PDF", fileName)
		return nil, err
	}
	defer pdfFile.Close()
	var buf bytes.Buffer
	bytes, err := result.GetPlainText()
	if err != nil {
		Logger.Error("Unable to convert PDF to text", fileName)
		return nil, err
	}
	buf.ReadFrom(bytes)
	fullText = buf.String() //writing from the buffer to the string
	if fullText == "" {
		err = errors.New("PDF Text Result is empty")
		Logger.Info("PDF Text Result is empty, sending to OCR: ", fileName, err)
		return nil, err
	}
	Logger.Info("Text processed from PDF without OCR: ", fileName)
	return &fullText, nil
}

func textProcessing(fileName string) {

}

func wordDocProcessing(fileName string) {

}

func convertToImage(fileName string) (*string, error) {
	var err error
	Logger.Info("Converting PDF To image for OCR", fileName)
	imageName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	imageName = filepath.Base(fmt.Sprint(imageName + ".png"))
	imageName = filepath.Join("temp", imageName)
	imageName, err = filepath.Abs(imageName)
	if err != nil {
		Logger.Error("Unable to edit absolute path string for temporary image for OCR: ", fileName, err)
		return nil, err
	}
	err = os.MkdirAll(filepath.Dir(imageName), os.ModePerm)
	if err != nil {
		Logger.Error("Unable to create absolute path for temporary image for OCR (permissions?): ", filepath.Dir(imageName), err)
		return nil, err
	}
	fileName = filepath.Clean(fileName)
	imageName = filepath.Clean(imageName)
	Logger.Info("Creating temp image for OCR at: ", imageName)
	_, err = os.OpenFile(fileName, os.O_RDWR, 0755) //TODO: Perhaps OS.stat would be enough of a test
	if err != nil {
		fmt.Println("ERROR FILE ISSUE", err)
		return nil, err
	}
	convertArgs := []string{"convert", "-density", "150", "-antialias", fileName, "-append", "-resize", "1024x", "-quality", "100", imageName}
	pdfConvertCmd := exec.Command("magick", convertArgs...)
	output, err := pdfConvertCmd.Output()
	if err != nil {
		Logger.Error("Unable to convert PDF Using Magick: ", fileName, err)
		return nil, err
	}
	fmt.Println("Outputting image to ", imageName)
	Logger.Debug("Output from pdfConvertCmd ", string(output))
	cleanArgs := []string{"convert", imageName, "-auto-orient", "-deskew", "40%", "-despeckle", imageName} //cleaning the resulting image
	imageCleanCmd := exec.Command("magick", cleanArgs...)
	output, err = imageCleanCmd.Output()
	if err != nil {
		Logger.Error("Magick was unable to clean the image for some reason... skipping this file for now: ", fileName, err)
		return nil, err
	}
	fullText, err := ocrProcessing(imageName)
	if err != nil {
		return nil, err
	}
	return fullText, nil
}

func ocrProcessing(imageName string) (*string, error) {
	var fullText string
	//fmt.Println("Output from imageCleanCmd", string(output))
	tesseractArgs := []string{imageName, "stdout"}
	tesseractCMD := exec.Command("tesseract", tesseractArgs...) //get the path to tesseract
	output, err := tesseractCMD.Output()
	if err != nil {
		Logger.Error("Tesseract encountered error when attempting to OCR image: ", imageName, err)
		return nil, err
	}
	fullText = string(output)
	if fullText == "" {
		Logger.Error("OCR Result returned empty string... OCR'ing the document failed", imageName, err)
		return nil, err
	}
	return &fullText, nil
}
