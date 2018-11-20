/*
Copyright (c) 2018 ceriath
This Package is part of "captainhook"
It is licensed under the MIT License
*/

package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"

	"code.cerinuts.io/cerinuts/captainhook/server/server"
)

func init() {
	rootCmd.AddCommand(clientCommand)
	clientCommand.AddCommand(addClientCommand)
	clientCommand.AddCommand(delClientCommand)
	clientCommand.AddCommand(listClientCommand)
	clientCommand.AddCommand(regenClientCommand)
}

var clientCommand = &cobra.Command{
	Use:   "client",
	Short: "Manage clients",
	Long:  `Manage the CaptainHook clients`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var addClientCommand = &cobra.Command{
	Use:   "add",
	Short: "Add a new Client",
	Long:  `Add a new CaptainHook client`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(addClient(args[0]))
	},
}

func addClient(clientname string) string {
	return RunRequest(server.ClientPath+"/"+clientname, "POST")
}

var delClientCommand = &cobra.Command{
	Use:   "del",
	Short: "Delete a Client",
	Long:  `Delete a CaptainHook client`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(delClient(args[0]))
	},
}

func delClient(clientname string) string {
	return RunRequest(server.ClientPath+"/"+clientname, "DELETE")
}

var listClientCommand = &cobra.Command{
	Use:   "list",
	Short: "List all Client",
	Long:  `List all CaptainHook client`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(listClients())
	},
}

func listClients() string {
	body := RunRequest(server.ClientPath, "GET")
	clients := make([]*server.Client, 0)
	err := json.Unmarshal([]byte(body), &clients)
	if err != nil {
		log.Print(err.Error())
		return "Could not read the server's answer"
	}

	res := ""
	for _, c := range clients {
		res = res + fmt.Sprintf("Name: %s, Hooks: %d, LastAction: %s\n", c.Name, len(c.Hooks), c.LastAction.Format(time.RFC822))
	}
	return res
}

var regenClientCommand = &cobra.Command{
	Use:   "regen",
	Short: "Generate a new secret for a Client",
	Long:  `Generates a new secret for the client. The old one will be invalid afterwards!`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(regenSecret(args[0]))
	},
}

func regenSecret(clientname string) string {
	return RunRequest(server.ClientPath+"/"+clientname, "PATCH")
}
