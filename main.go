package main

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/ziflex/lecho/v2"

	config "github.com/deranjer/goEDMS/config"
	database "github.com/deranjer/goEDMS/database"
	engine "github.com/deranjer/goEDMS/engine"
)

//Logger is global since we will need it everywhere
var Logger *lecho.Logger

//injectGlobals injects all of our globals into their packages
func injectGlobals(logger *lecho.Logger) {
	Logger = logger
	database.Logger = Logger
	config.Logger = Logger
	engine.Logger = Logger
}

func main() {
	serverConfig, logger := config.SetupServer()
	injectGlobals(logger) //inject the logger into all of the packages
	db := database.SetupDatabase()
	searchDB, err := database.SetupSearchDB()
	if err != nil {
		Logger.Fatal("Unable to setup index database", err)
	}
	defer db.Close()
	defer searchDB.Close()
	database.WriteConfigToDB(serverConfig, db) //writing the config to the database
	e := echo.New()
	serverHandler := engine.ServerHandler{DB: db, SearchDB: searchDB, Echo: e, ServerConfig: serverConfig} //injecting the database into the handler for routes
	serverHandler.InitializeSchedules(db, searchDB)                                                        //initialize all the cron jobs
	e.Logger = Logger
	e.Use(lecho.Middleware(lecho.Config{
		Logger: logger}))
	e.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))
	e.Static("/", "public/built") //serving up the React Frontend
	log.Info("Logger enabled!!")
	//injecting database into the context so we can access it

	//Start the routes
	e.GET("/home", serverHandler.GetLatestDocuments)
	e.GET("/documents/filesystem", serverHandler.GetDocumentFileSystem)
	e.GET("/document/:id", serverHandler.GetDocument)
	e.GET("/folder/:folder", serverHandler.GetFolder)
	e.GET("/search/*", serverHandler.SearchDocuments)
	e.DELETE("/document/:id", serverHandler.DeleteDocument)
	e.PATCH("document/move/*", serverHandler.MoveDocuments)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", serverConfig.ListenAddrPort)))
}
