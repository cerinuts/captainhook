/*
Copyright (c) 2018 ceriath
This Package is part of "captainhook"
It is licensed under the MIT License
*/

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

//ApplicationName is the name of the application
const ApplicationName = "CaptainHook CLI"

//VersionMajor 0 means in development, >1 ensures compatibility with each minor version, but breakes with new major version
const VersionMajor = "0"

//VersionMinor introduces changes that require a new version number. If the major version is 0, they are likely to break compatibility
const VersionMinor = "1"

//VersionPatch introduces changes that require a new version number. Follows semver specs.
const VersionPatch = "0"

//FullVersion contains the full version of this package in a printable string
const FullVersion = VersionMajor + "." + VersionMinor + "." + VersionPatch

var serverAddress string

// InternalError is any internal error
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

// RunRequest runs a request to the server to given path with http method
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
