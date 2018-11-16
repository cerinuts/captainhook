package main

import (
	"code.cerinuts.io/cerinuts/captainhook/server/server"
	"github.com/spf13/viper"
)

func main() {
	server.InitConfig()
	server.InitLogger()

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
