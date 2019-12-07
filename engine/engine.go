package engine

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/deranjer/goEDMS/config"
	"github.com/ledongthuc/pdf"
	"gopkg.in/gographics/imagick.v2/imagick"
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
			//ocrProcessing(file)
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
		ocrProcessing(file)
		return
	}
	defer pdfFile.Close()
	var buf bytes.Buffer
	bytes, err := result.GetPlainText()
	if err != nil {
		fmt.Println("Unable to extract text from PDF", fileName)
		//Logger.Error("Unable to convert PDF to text", file.Name())
		ocrProcessing(file)
		return
	}
	buf.ReadFrom(bytes)
	fullText := buf.String()
	if fullText == "" {
		fmt.Println("text result is empty, send to ocr", fileName)
		ocrProcessing(file)
		return
	}
	fmt.Println(fullText)
	return
}

func textProcessing(fileName string) {

}

func wordDocProcessing(fileName string) {

}

func ocrProcessing(fileName string) {
	//args := []string{}
	Logger.Info("Converting PDF To image for OCR", fileName)
	imageName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	imageName = fmt.Sprint(imageName + ".png")
	fmt.Println("Outputting image to ", imageName)
	imagick.Initialize()
	defer imagick.Terminate()

	mw := imagick.NewMagickWand()
	defer mw.Destroy()
	err := mw.SetResolution(300, 300) 
	if err != nil {
        return
    }


	err = mw.ReadImage(fileName)
	if err != nil {
		panic(err)
	}
	

	err = mw.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_FLATTEN) 
	if err != nil {
        return
    }

	err = mw.SetCompressionQuality(95);
	if err != nil {
        return 
    }

	err = mw.SetFormat("jpg")
	if err != nil {
		return
	}
	//mw.WriteImages()
	mw.WriteImage(imageName)


	

	/* cmd := exec.Command("tesseract", fileName, "out")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
	} */
	//fmt.Println("Result: " + out.String())
}

func ocrFile(file os.FileInfo) {

}
