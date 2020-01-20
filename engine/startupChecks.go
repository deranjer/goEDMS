package engine

import (
	"fmt"
	"os"

	"github.com/deranjer/goEDMS/config"
	"github.com/deranjer/goEDMS/database"
)

//StartupChecks performs all the checks to make sure everything works
func (serverHandler *ServerHandler) StartupChecks() error {
	serverConfig, err := database.FetchConfigFromDB(serverHandler.DB)
	if err != nil {
		Logger.Error("Error fetching config:", err)
		return err
	}
	magickChecks(serverConfig)
	return nil
}

func magickChecks(serverConfig config.ServerConfig) error {
	magickInfo, err := os.Stat(serverConfig.MagickPath)
	if err != nil {
		Logger.Fatal("Err finding Magick executable (required):", err)
		return err
	}
	if magickInfo.IsDir() {
		Logger.Fatal("Magick path ends in a directory, not executable:", err)
		return err
	}
	fmt.Println("Perms: ", magickInfo.Mode())
	if magickInfo.Mode() == 0111 {
		fmt.Println("Mode is exectuable?", magickInfo.Mode())
	}
	return nil
}
