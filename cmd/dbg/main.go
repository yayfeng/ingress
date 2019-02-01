/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "dbg",
		Short: "dbg is a tool for quickly inspecting the state of the nginx instance",
	}

	backendsCmd := &cobra.Command{
		Use:   "backends",
		Short: "Output the dynamic backend information as a JSON array",
		Run: func(cmd *cobra.Command, args []string) {
			backends()
		},
	}
	rootCmd.AddCommand(backendsCmd)

	backendsListCmd := &cobra.Command{
		Use:   "list-backends",
		Short: "Output a newline-separated list of the backend names",
		Run: func(cmd *cobra.Command, args []string) {
			backendsList()
		},
	}
	rootCmd.AddCommand(backendsListCmd)

	backendsGetCmd := &cobra.Command{
		Use:   "get-backend",
		Short: "Output the backend information only for the backend that has this name",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			backendsGet(args[0])
		},
	}
	rootCmd.AddCommand(backendsGetCmd)

	generalCmd := &cobra.Command{
		Use:   "general",
		Short: "Output the general dynamic lua state",
		Run: func(cmd *cobra.Command, args []string) {
			general()
		},
	}
	rootCmd.AddCommand(generalCmd)

	confCmd := &cobra.Command{
		Use:   "conf",
		Short: "Dump the contents of /etc/nginx/nginx.conf",
		Run: func(cmd *cobra.Command, args []string) {
			readNginxConf()
		},
	}
	rootCmd.AddCommand(confCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func backends() {
	//Get the backend information from the nginx instance
	body, requestErr := makeRequest("http://127.0.0.1:18080/configuration/backends")
	if requestErr != nil {
		fmt.Println(requestErr)
		return
	}

	//Indent the returned JSON
	var prettyBuffer bytes.Buffer
	indentErr := json.Indent(&prettyBuffer, body, "", "  ")
	if indentErr != nil {
		fmt.Println(indentErr)
		return
	}

	//Print the pretty JSON
	fmt.Println(string(prettyBuffer.Bytes()))
}

func backendsList() {
	body, requestErr := makeRequest("http://127.0.0.1:18080/configuration/backends")
	if requestErr != nil {
		fmt.Println(requestErr)
		return
	}

	//Read the array of backends from the returned JSON
	var f interface{}
	unmarshalErr := json.Unmarshal(body, &f)
	if unmarshalErr != nil {
		fmt.Println(unmarshalErr)
		return
	}
	backends := f.([]interface{})

	//Just list the names of the backends
	for _, backendi := range backends {
		backend := backendi.(map[string]interface{})
		fmt.Println(backend["name"].(string))
	}
}

func backendsGet(name string) {
	body, requestErr := makeRequest("http://127.0.0.1:18080/configuration/backends")
	if requestErr != nil {
		fmt.Println(requestErr)
		return
	}

	//Read the array of backends from the returned JSON
	var f interface{}
	unmarshalErr := json.Unmarshal(body, &f)
	if unmarshalErr != nil {
		fmt.Println(unmarshalErr)
		return
	}
	backends := f.([]interface{})

	//Search for a backend by name and output its config if found
	for _, backendi := range backends {
		backend := backendi.(map[string]interface{})
		if backend["name"].(string) == name {
			printed, _ := json.MarshalIndent(backend, "", "  ")
			fmt.Println(string(printed))
			return
		}
	}
	fmt.Println("A backend of this name was not found.")
}

func general() {
	//Get the other (general) information from the nginx instance
	body, requestErr := makeRequest("http://127.0.0.1:18080/configuration/general")
	if requestErr != nil {
		fmt.Println(requestErr)
		return
	}

	//Indent the returned JSON
	var prettyBuffer bytes.Buffer
	indentErr := json.Indent(&prettyBuffer, body, "", "  ")
	if indentErr != nil {
		fmt.Println(indentErr)
		return
	}

	//Print the pretty JSON
	fmt.Println(string(prettyBuffer.Bytes()))
}

func readNginxConf() {
	confFile, err := os.Open("/etc/nginx/nginx.conf")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer confFile.Close()
	contents, _ := ioutil.ReadAll(confFile)
	fmt.Print(string(contents))
}

func makeRequest(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
