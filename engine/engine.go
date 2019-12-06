package engine

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/deranjer/goEDMS/config"
	"github.com/ledongthuc/pdf"
)

func ingressJobFunc(serverConfig config.ServerConfig) {
	fmt.Println("Testing job func")
	//serverConfig := database.FetchConfigFromDB(db)
	Logger.Info("Starting Ingress Job on folder:", serverConfig.IngressPath)
	var documentPath []string
	err := filepath.Walk(serverConfig.IngressPath, func(path string, info os.FileInfo, err error) error {
		//fullFilePath := filepath.Join(os.Getwd(),
		fmt.Println("PATH", path)
		documentPath = append(documentPath, path)
		return nil
	})
	if err != nil {
		Logger.Error("Error reading files in from ingress!")
	}
	for _, file := range documentPath {
		switch filepath.Ext(file) {
		case ".pdf":
			pdfProcessing(file)
			fallthrough
		case ".txt", ".rtf":
			textProcessing(file)
			fallthrough
		case ".doc", ".docx", ".odf":
			wordDocProcessing(file)
			fallthrough
		case ".tiff", ".jpg", ".jpeg", ".png":
			imageProcessing(file)
			fallthrough
		default:
			Logger.Warn("Invalid file type: ", filepath.Base((file)))
		}
		continue
	}

}

func pdfProcessing(file string) {
	fileName := filepath.Base((file))
	fmt.Println("WOrking on current file:", fileName)
	Logger.Debug("Working on current FULL PATH: ", file)
	pdfFile, result, err := pdf.Open(file)
	if err != nil {
		fmt.Println("Unable to open PDF", fileName)
		//Logger.Error("Unable to open PDF", file.Name())
		//TODO - send to OCR for processing
		return
	}
	defer pdfFile.Close()
	var buf bytes.Buffer
	bytes, err := result.GetPlainText()
	if err != nil {
		fmt.Println("Unable to extract text from PDF", fileName)
		//Logger.Error("Unable to convert PDF to text", file.Name())
		//TODO - send to OCR for processing
		return
	}
	buf.ReadFrom(bytes)
	fullText := buf.String()
	if fullText == "" {
		fmt.Println("text result is empty, send to OCR", fileName)
		//TODO - send to OCR for processing
	}
	fmt.Println(fullText)
	return
}

func textProcessing(fileName string) {

}

func wordDocProcessing(fileName string) {

}

func imageProcessing(fileName string) {

}

func ocrFile(file os.FileInfo) {

}
