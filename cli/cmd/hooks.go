package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/spf13/cobra"

	"code.cerinuts.io/cerinuts/captainhook/server/server"
)

func init() {
	rootCmd.AddCommand(hookCommand)
	hookCommand.AddCommand(addHookCommand)
	hookCommand.AddCommand(delHookCommand)
}

var hookCommand = &cobra.Command{
	Use:   "hook",
	Short: "Manage webhooks",
	Long:  `Manage the CaptainHook Webhooks`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var addHookCommand = &cobra.Command{
	Use:   "add",
	Short: "Add a new Hook",
	Long:  `Add a new CaptainHook Webhook`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Print("Not enough arguments (clientname, hookidentifier)")
			return
		}
		fmt.Print(addHook(args[0], args[1]))
	},
}

func addHook(clientname, hookIdentifier string) string {
	body := RunRequest(server.HookPath+"/"+clientname+"/"+hookIdentifier, "PUT")
	hook := new(server.Webhook)
	err := json.Unmarshal([]byte(body), &hook)
	if err != nil {
		alreadyExists := &server.ErrHookAlreadyExists{Identifier: hookIdentifier}
		if body == alreadyExists.Error() {
			return body
		}
		log.Print(err.Error())
		return "Could not read the server's answer"
	}

	return hook.URL
}

var delHookCommand = &cobra.Command{
	Use:   "del",
	Short: "Delete a Hook",
	Long:  `Delete a CaptainHook Webhook. Provide the full URL.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(delHook(args[0]))
	},
}

func delHook(URL string) string {
	return RunRequest(server.HookByUUIDPath+"/"+url.QueryEscape(URL), "DELETE")
}
