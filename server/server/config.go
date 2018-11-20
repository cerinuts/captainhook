package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// InitConfig initializes the config with default values
func InitConfig() {

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
	viper.SetDefault("Loglevel", "Warning")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}

	if !viper.GetBool("Debug") {
		gin.SetMode(gin.ReleaseMode)
	}
}
