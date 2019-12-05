package database

import (
	"fmt"

	"github.com/asdine/storm"
	config "github.com/deranjer/goEDMS/config"
	"github.com/ziflex/lecho/v2"
)

//Logger is global since we will need it everywhere
var Logger *lecho.Logger

//SetupDatabase initializes the storm/bbolt database
func SetupDatabase() (db *storm.DB) {
	db, err := storm.Open("goEDMS.db")
	if err != nil {
		Logger.Fatal("Unable to create/open database!", err)
	}
	return db
}

//FetchConfigFromDB pulls the server config from the database
func FetchConfigFromDB(db *storm.DB) config.ServerConfig {
	var serverConfig config.ServerConfig
	err := db.One("StormID", 1, &serverConfig)
	if err != nil {
		Logger.Fatal("Unable to fetch server config from db!", err)
	}
	return serverConfig
}

//WriteConfigToDB writes the serverconfig to the database for later retrieval
func WriteConfigToDB(serverConfig config.ServerConfig, db *storm.DB) {
	serverConfig.StormID = 1 //config will be stored in bucket 1
	fmt.Printf("%+v\n", serverConfig)
	err := db.Save(&serverConfig)
	if err != nil {
		Logger.Error("Unable to write server config to database!", err)
	}
}
