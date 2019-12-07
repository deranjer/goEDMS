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
	injectGlobals(logger)
	db := database.SetupDatabase()
	defer db.Close()
	database.WriteConfigToDB(serverConfig, db)
	engine.InitializeSchedules(db)

	e := echo.New()
	e.Logger = Logger
	e.Use(lecho.Middleware(lecho.Config{
		Logger: logger}))
	e.Static("/", "public/built")
	log.Info("Logger enabled!!")
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", serverConfig.ListenAddrPort)))
}
