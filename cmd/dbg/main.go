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
	"github.com/spf13/pflag"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {

	//There aren't any flags yet, just arguments
	var flags = pflag.NewFlagSet("", pflag.ExitOnError)
	flags.Parse(os.Args)
	args := flags.Args()

	if len(args) == 2 && args[1] == "backends" {
		backends()
	} else if len(args) == 3 && args[1] == "backends" && args[2] == "list" {
		backendsList()
	} else if len(args) == 4 && args[1] == "backends" && args[2] == "get" {
		backendsGet(args[3])
	} else if len(args) == 2 && args[1] == "general" {
		general()
	} else if len(args) == 2 && args[1] == "conf" {
		readNginxConf()
	} else if len(args) == 1 {
		info()
	} else {
		fmt.Println("Unknown command.")
	}
}

func info() {
	fmt.Println(`dbg is a tool for quickly inspecting the state of the nginx instance.
Subcommands:

- dbg backends             Output the dynamic backend information as JSON.
- dbg backends list        Just list the names of all the backends.
- dbg backends get <NAME>  Output the backend information only for the backend that has this name.
- dbg general              Output the other dynamic information as JSON.
- dbg conf                 Dump the contents of /etc/nginx/nginx.conf`)
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
