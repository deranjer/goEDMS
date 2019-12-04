package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/labstack/gommon/log"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"github.com/ziflex/lecho/v2"
)

//Logger is global since we will need it everywhere
var Logger *lecho.Logger

//ServerConfig contains all of the server settings defined in the TOML file
type ServerConfig struct {
	StormID         int
	ListenAddrIP    string
	ListenAddrPort  string
	IngressPath     string
	DocumentPath    string
	WebUIPass       bool
	ClientUsername  string
	ClientPassword  string
	PushBulletToken string `json:"-"`
	UseReverseProxy bool
	BaseURL         string
	IngressInterval int
}

func defaultConfig() ServerConfig { //TODO: Do I even bother, if config fails most likely not worth continuing
	var ServerConfigDefault ServerConfig
	//Config.AppVersion
	zerolog.SetGlobalLevel(zerolog.WarnLevel)
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
	ingressDir := filepath.ToSlash(viper.GetString("serverConfig.DefaultIngressPath")) //Converting the string literal into a filepath
	ingressDirAbs, err := filepath.Abs(ingressDir)                                     //Converting to an absolute file path
	if err != nil {
		logger.Error("Failed creating absolute path for ingress directory", err)
	}
	serverConfigLive.IngressPath = ingressDirAbs
	logger.Infof("Logger is setup")
	serverConfigLive.ListenAddrPort = viper.GetString("serverConfig.ServerPort")
	serverConfigLive.ListenAddrIP = viper.GetString("serverConfig.ServerAddr")
	serverConfigLive.IngressInterval = viper.GetInt("scheduling.IngressInterval")
	return serverConfigLive, logger
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
	logger := lecho.New(
		os.Stdout,
		lecho.WithLevel(loglevel),
		lecho.WithFields(map[string]interface{}{"name": "lecho factory"}),
		lecho.WithTimestamp(),
		lecho.WithCaller(),
	)
	return logger
}
