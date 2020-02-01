package engine

import (
	"fmt"
	"github.com/asdine/storm"
	"github.com/blevesearch/bleve"
	database "github.com/deranjer/goEDMS/database"
	"github.com/robfig/cron/v3"
	"github.com/ziflex/lecho/v2"
)

//Logger is global since we will need it everywhere
var Logger *lecho.Logger

//InitializeSchedules starts all the cron jobs (currently just one)
func (serverHandler *ServerHandler) InitializeSchedules(db *storm.DB, searchDB bleve.Index) {
	serverConfig, err := database.FetchConfigFromDB(db)
	if err != nil {
		fmt.Println("Error reading db when initializing")
	}
	c := cron.New(cron.WithLogger(cron.VerbosePrintfLogger(Logger)))
	var ingressJob cron.Job
	ingressJob = cron.FuncJob(func() { serverHandler.ingressJobFunc(serverConfig, db, searchDB) })
	ingressJob = cron.NewChain(cron.SkipIfStillRunning(cron.DefaultLogger)).Then(ingressJob) //ensure we don't kick off another if old one is still running
	c.AddJob(fmt.Sprintf("@every %dm", serverConfig.IngressInterval), ingressJob)
	//c.AddJob("@every 1m", ingressJob)
	Logger.Infof("Adding Ingress Job that runs every %dm", serverConfig.IngressInterval)
	c.Start()
}
