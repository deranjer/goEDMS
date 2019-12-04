module github.com/deranjer/goEDMS

go 1.13

require (
	github.com/asdine/storm v2.1.2+incompatible
	github.com/labstack/echo v3.3.10+incompatible
	github.com/labstack/echo/v4 v4.1.10
	github.com/labstack/gommon v0.3.0
	github.com/robfig/cron v1.2.0
	github.com/robfig/cron/v3 v3.0.0
	github.com/rs/zerolog v1.17.2
	github.com/sirupsen/logrus v1.3.0
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/viper v1.5.0
	github.com/ziflex/lecho v1.2.0
	github.com/ziflex/lecho/v2 v2.0.0
)

replace github.com/deranjer/goEDMS => ../goEDMS
