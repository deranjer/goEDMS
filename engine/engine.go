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
		fileName := filepath.Base((file))
		if filepath.Ext(file) != ".pdf" { //|| ".txt" || ".rtf" || ".doc" || ".docx" || ".tiff" || ".jpg" || ".jpeg" || ".odf" {
			Logger.Warn("Invalid file type: ", fileName)
			continue
		} else {
			fmt.Println("WOrking on current file:", fileName)
			//fullFilePath := filepath.Join(serverConfig.IngressPath, fileName)
			//fullFilePathAbs, _ := filepath.Abs(fullFilePath) //TODO error handle
			Logger.Debug("Working on current FULL PATH: ", file)
			file, result, err := pdf.Open(file)
			if err != nil {
				fmt.Println("Unable to open PDF", fileName)
				//Logger.Error("Unable to open PDF", file.Name())
				//TODO - send to OCR for processing
				continue
			}
			defer file.Close()
			var buf bytes.Buffer
			bytes, err := result.GetPlainText()
			if err != nil {
				fmt.Println("Unable to extract text from PDF", fileName)
				//Logger.Error("Unable to convert PDF to text", file.Name())
				//TODO - send to OCR for processing
				continue
			}
			buf.ReadFrom(bytes)
			fullText := buf.String()
			if fullText == "" {
				fmt.Println("text result is empty, send to OCR", fileName)
				//TODO - send to OCR for processing
			}
			fmt.Println(fullText)
			continue
		}

	}
}

func ocrFile(file os.FileInfo) {

}
