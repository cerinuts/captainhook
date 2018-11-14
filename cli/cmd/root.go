package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

const ApplicationName = "CaptainHook CLI"
const VersionMajor = "0"
const VersionMinor = "1"
const VersionPatch = "0"
const FullVersion = VersionMajor + "." + VersionMinor + "." + VersionPatch

var serverAddress string

type InternalError struct {
	Message string `json:"message"`
}

func init() {
	rootCmd.Flags().StringVarP(&serverAddress, "url", "u", "http://localhost:12841", "The full url of the internal CaptainHook server")
}

var rootCmd = &cobra.Command{
	Use:   "captainhook",
	Short: "CLI Application for captainhook server",
	Long:  `Command line application to manage a local captainhook server`,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			panic("Error running help")
		}

	},
}

// Execute starts the CLI
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func RunRequest(path, method string) string {
	u, err := url.Parse(serverAddress + path)

	if err != nil {
		nerr := errors.New("Invalid server-url given: " + serverAddress + " : " + err.Error())
		log.Print(nerr)
		return nerr.Error()
	}

	req := &http.Request{
		Method: method,
		URL:    u,
	}

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Print(err.Error())
		return err.Error()
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var intError InternalError
		body, err := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(body, &intError)
		if err != nil {
			nerr := errors.New("Server responded with " + strconv.Itoa(resp.StatusCode))
			log.Print(nerr)
			return nerr.Error()
		}
		return intError.Message
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(err.Error())
		return "Error reading body"
	}

	res := string(body)
	if res == "" {
		res = "Success."
	}

	return res
}
