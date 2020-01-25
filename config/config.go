package config

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	"github.com/labstack/gommon/log"
	"github.com/spf13/viper"
	"github.com/ziflex/lecho/v2"
)

//Logger is global since we will need it everywhere
var Logger *lecho.Logger

//ServerConfig contains all of the server settings defined in the TOML file
type ServerConfig struct {
	StormID              int `storm:"id"`
	ListenAddrIP         string
	ListenAddrPort       string
	IngressPath          string
	IngressDelete        bool
	IngressMoveFolder    string
	IngressPreserve      bool
	DocumentPath         string
	NewDocumentFolder    string //absolute path to new document folder
	NewDocumentFolderRel string //relative path to new document folder Needed for multiple levels deep.
	WebUIPass            bool
	ClientUsername       string
	ClientPassword       string
	PushBulletToken      string `json:"-"`
	MagickPath           string
	TesseractPath        string
	UseReverseProxy      bool
	BaseURL              string
	IngressInterval      int
	FrontEndConfig
}

//FrontEndConfig stores all of the frontend settings
type FrontEndConfig struct {
	NewDocumentNumber int
}

func defaultConfig() ServerConfig { //TODO: Do I even bother, if config fails most likely not worth continuing
	var ServerConfigDefault ServerConfig
	//Config.AppVersion
	//zerolog.SetGlobalLevel(zerolog.WarnLevel)
	ServerConfigDefault.DocumentPath = "documents"
	ServerConfigDefault.IngressPath = "ingress"
	ServerConfigDefault.WebUIPass = false
	ServerConfigDefault.UseReverseProxy = false
	return ServerConfigDefault
}

//SetupServer does the initial configuration
func SetupServer() (ServerConfig, *lecho.Logger) {
	var serverConfigLive ServerConfig
	viper.AddConfigPath("config/")
	viper.AddConfigPath(".")
	viper.SetConfigName("serverConfig")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %s \n", err))
	}
	logger := setupLogging()
	ingressDir := filepath.ToSlash(viper.GetString("ingress.IngressPath")) //Converting the string literal into a filepath
	ingressDirAbs, err := filepath.Abs(ingressDir)                         //Converting to an absolute file path
	if err != nil {
		logger.Error("Failed creating absolute path for ingress directory", err)
	}
	serverConfigLive.IngressPath = ingressDirAbs
	logger.Infof("Base Logger is setup, will switch to echo logger after config complete!")
	serverConfigLive.ListenAddrPort = viper.GetString("serverConfig.ServerPort")
	serverConfigLive.ListenAddrIP = viper.GetString("serverConfig.ServerAddr")
	serverConfigLive.IngressInterval = viper.GetInt("ingress.scheduling.IngressInterval")
	serverConfigLive.IngressPreserve = viper.GetBool("ingress.handling.PreserveDirStructure")
	serverConfigLive.IngressDelete = viper.GetBool("ingress.completed.IngressDeleteOnProcess")
	ingressMoveFolder := filepath.ToSlash(viper.GetString("ingress.completed.IngressMoveFolder"))
	ingressMoveFolderABS, err := filepath.Abs(ingressMoveFolder)
	if err != nil {
		logger.Error("Failed creating absolute path for ingress move folder", err)
	}
	serverConfigLive.IngressMoveFolder = ingressMoveFolderABS
	os.MkdirAll(ingressMoveFolderABS, os.ModePerm) //creating the directory for moving now
	fmt.Println("Ingress Interval: ", serverConfigLive.IngressInterval)
	documentPathRelative := filepath.ToSlash(viper.GetString("documentLibrary.DocumentFileSystemLocation"))
	serverConfigLive.DocumentPath, err = filepath.Abs(documentPathRelative)
	if err != nil {
		logger.Error("Failed creating absolute path for document library", err)
	}
	newDocumentPath := filepath.ToSlash(viper.GetString("documentLibrary.DefaultNewDocumentFolder"))
	serverConfigLive.NewDocumentFolderRel = newDocumentPath
	serverConfigLive.NewDocumentFolder = filepath.Join(serverConfigLive.DocumentPath, newDocumentPath)
	serverConfigLive.MagickPath, err = filepath.Abs(filepath.ToSlash(viper.GetString("ocr.MagickBin")))
	if err != nil {
		logger.Error("Failed creating absolute path for magick binary", err)
	}
	serverConfigLive.TesseractPath, err = filepath.Abs(filepath.ToSlash(viper.GetString("ocr.TesseractBin")))
	if err != nil {
		logger.Error("Failed creating absolute path for tesseract binary", err)
	}
	serverConfigLive.UseReverseProxy = viper.GetBool("reverseProxy.ProxyEnabled")
	serverConfigLive.BaseURL = viper.GetString("reverseProxy.BaseURL")
	os.MkdirAll(serverConfigLive.NewDocumentFolder, os.ModePerm)
	frontEndConfigLive := setupFrontEnd(serverConfigLive, logger)
	serverConfigLive.FrontEndConfig = frontEndConfigLive
	return serverConfigLive, logger
}

func setupFrontEnd(serverConfigLive ServerConfig, logger *lecho.Logger) FrontEndConfig {
	var frontEndConfigLive FrontEndConfig
	var frontEndURL string
	frontEndConfigLive.NewDocumentNumber = viper.GetInt("frontend.NewDocumentNumber") //number of new documents to display //TODO: maybe not using this...
	if serverConfigLive.UseReverseProxy {                                             //if using a proxy set the proxy URL
		frontEndURL = serverConfigLive.BaseURL
	} else { //If NOT using a proxy determine the IP URL
		if serverConfigLive.ListenAddrIP == "" { //If no IP listed attempt to discover the default IP addr
			ipAddr, err := getDefaultIP(logger)
			if err != nil {
				logger.Error("WARNING! Unable to determine default IP, frontend-config.js may need to be manually modified for goEDMS to work! ", err)
				frontEndURL = fmt.Sprintf("http://%s:%s", serverConfigLive.ListenAddrIP, serverConfigLive.ListenAddrPort)
			} else {
				frontEndURL = fmt.Sprintf("http://%s:%s", *ipAddr, serverConfigLive.ListenAddrPort)
			}
		} else { //If IP addr listed then just use that in the IP URL
			frontEndURL = fmt.Sprintf("http://%s:%s", serverConfigLive.ListenAddrIP, serverConfigLive.ListenAddrPort)
		}

	}
	var frontEndJS = fmt.Sprintf(`window['runConfig'] = { 
		apiUrl: "%s"
	}`, frontEndURL) //Creating the react API file so the frontend will connect with the backend
	err := ioutil.WriteFile("public/built/frontend-config.js", []byte(frontEndJS), 0644)
	if err != nil {
		logger.Fatal("Error writing frontend config to public/built/frontend-config.js", err)
	}
	return frontEndConfigLive
}

func getDefaultIP(logger *lecho.Logger) (*string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80") //attempting to determine the default IP by connecting out
	if err != nil {
		logger.Error("Error discovering Local IP! Either network connection error (outbound connection is used to determine default IP) or error determining default IP!", err)
		return nil, err
	}
	defer conn.Close()
	localaddr := conn.LocalAddr().(*net.UDPAddr).IP.String()
	logger.Info("Local IP Determined: ", localaddr)
	return &localaddr, nil
}

func setupLogging() *lecho.Logger {
	logLevelString := viper.GetString("logging.Level")
	var loglevel log.Lvl
	switch logLevelString { //Options = Debug 0, Info 1, Warn 2, Error 3, Fatal 4, Panic 5
	case "Off", "off":
		loglevel = log.OFF
	case "Error", "error":
		loglevel = log.ERROR
	case "Warn", "warn":
		loglevel = log.WARN
	case "Info", "info":
		loglevel = log.INFO
	case "Debug", "debug":
		loglevel = log.DEBUG
	default:
		loglevel = log.WARN
	}
	var logWriter *os.File
	logOutput := viper.GetString("logging.OutputPath")
	if logOutput == "file" {
		logPath, err := filepath.Abs(filepath.ToSlash(viper.GetString("logging.LogFileLocation")))
		if err != nil {
			fmt.Println("Unable to create log file path: ", err)
			logPath = "output.log"
		}
		logFile, err := os.Create(logPath)
		if err != nil {
			fmt.Println("Unable to create log file: ", err)
			return nil
		}
		logWriter = logFile
		fmt.Println("Logging to file: ", logPath)
	} else { //TODO: this technically catches EVERYTHING that doesn't say "file"
		logWriter = os.Stdout
		fmt.Println("Will be logging to stdout...")
	}

	logger := lecho.New(
		logWriter,
		lecho.WithLevel(loglevel),
		lecho.WithTimestamp(),
	)
	return logger
}
