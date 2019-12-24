package main

import (
	"fmt"

	"github.com/labstack/echo/v4"
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
	database.WriteConfigToDB(serverConfig, db) //Write the config back to the database
	searchDB, err := database.SetupSearchDB()
	if err != nil {
		Logger.Fatal("Unable to setup index database", err)
	}
	defer db.Close()
	defer searchDB.Close()
	database.WriteConfigToDB(serverConfig, db) //writing the config to the database
	engine.InitializeSchedules(db, searchDB)   //initialize all the cron jobs
	e := echo.New()
	dbHandle := engine.DBHandler{DB: db, SearchDB: searchDB} //injecting the database into the handler for routes
	e.Logger = Logger
	e.Use(lecho.Middleware(lecho.Config{
		Logger: logger}))
	e.Static("/", "public/built") //serving up the React Frontend
	log.Info("Logger enabled!!")
	//injecting database into the context so we can access it

	//Start the routes
	e.GET("/home", dbHandle.GetLatestDocuments)
	e.GET("/document/:id", dbHandle.GetDocument)
	e.GET("/folder/:folder", dbHandle.GetFolder)
	e.GET("/search/*", dbHandle.SearchDocuments)
	e.DELETE("/document/:id", dbHandle.DeleteDocument)
	e.PATCH("document/move/*", dbHandle.MoveDocuments)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", serverConfig.ListenAddrPort)))
}
