package engine

import (
	"fmt"

	"github.com/asdine/storm"
	database "github.com/deranjer/goEDMS/database"
	"github.com/robfig/cron/v3"
	"github.com/ziflex/lecho/v2"
)

//Logger is global since we will need it everywhere
var Logger *lecho.Logger

//InitialzeSchedules starts all the cron jobs (currently just one)
func InitializeSchedules(db *storm.DB) {
	serverConfig := database.FetchConfigFromDB(db)
	c := cron.New()

	c.AddJob(fmt.Sprintf("@every %dm", serverConfig.IngressInterval))
	c.Start()
}
