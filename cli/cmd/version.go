/*
Copyright (c) 2018 ceriath
This Package is part of "captainhook"
It is licensed under the MIT License
*/

package cmd

import (
	"fmt"

	"code.cerinuts.io/cerinuts/captainhook/server/server"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)

}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `All software has versions. This one is a semver`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s %s\n", ApplicationName, FullVersion)
		fmt.Println(GetVersion())
	},
}

// GetVersion gets the server version
func GetVersion() string {
	return RunRequest(server.ApplicationVersionPath, "GET")
}
