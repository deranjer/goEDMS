package engine

import (
	"github.com/asdine/storm"
	database "github.com/deranjer/goEDMS/database"
	"github.com/robfig/cron/v3"
	"github.com/ziflex/lecho/v2"
)

//Logger is global since we will need it everywhere
var Logger *lecho.Logger

//InitializeSchedules starts all the cron jobs (currently just one)
func InitializeSchedules(db *storm.DB) {
	serverConfig := database.FetchConfigFromDB(db)
	c := cron.New(cron.WithLogger(cron.VerbosePrintfLogger(Logger)))
	var ingressJob cron.Job
	ingressJob = cron.FuncJob(func() { ingressJobFunc(serverConfig) })
	ingressJob = cron.NewChain(cron.SkipIfStillRunning(cron.DefaultLogger)).Then(ingressJob)
	serverConfig.IngressInterval = 1
	//c.AddJob(fmt.Sprintf("@every %dm", serverConfig.IngressInterval), ingressJob)
	c.AddJob("@every 5m", ingressJob)
	Logger.Infof("Adding Ingress Job that runs every %dm", serverConfig.IngressInterval)
	c.Start()

}
