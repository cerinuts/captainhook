package main

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"code.cerinuts.io/cerinuts/captainhook/server/server"
	"github.com/spf13/viper"
)

func main() {

	viper.SetConfigName("server")
	viper.AddConfigPath("/etc/cerinuts/captainhook")   // linux
	viper.AddConfigPath("$HOME/.cerinuts/captainhook") // windows
	viper.AddConfigPath(".")                           //fallback

	viper.SetDefault("Host", "127.0.0.1")
	viper.SetDefault("ExternalPort", 12840)
	viper.SetDefault("ExternalSSLPort", 12842)
	viper.SetDefault("InternalPort", 12841)
	viper.SetDefault("SSLCertificate", "")
	viper.SetDefault("SSLKey", "")
	viper.SetDefault("Debug", false)

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	if !viper.GetBool("Debug") {
		gin.SetMode(gin.ReleaseMode)
	}

	s := server.NewServer(viper.GetString("Host"), viper.GetString("ExternalPort"))
	s.Load()
	server.SetupSSLAPI(viper.GetString("Host"),
		viper.GetInt("ExternalPort"),
		viper.GetInt("ExternalSSLPort"),
		viper.GetInt("InternalPort"),
		s,
		viper.GetString("SSLCertificate"),
		viper.GetString("SSLKey"))
	s.Run()
}
