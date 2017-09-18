// Copyright © 2017 Philipp Hug <philipp@hug.cx>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var (
	mapping         map[string]string = make(map[string]string)
	serverPort      int
	serverInterface string
	defaultRedirect string
	mappingFile     string
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the redirector",
	Long: `Starts a web server on the specified port to redirect all incoming requests
according to the rules set in the configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func load(fn string) {
	lines, err := readLines(fn)
	if err != nil {
		log.Fatalf("readLines: %s", err)
		return
	}
	for _, line := range lines {
		words := strings.Fields(line)
		if len(words) == 2 {
			mapping[words[0]] = words[1]
		} else {
			log.Print("Ignoring invalid line:", line)
		}
	}
}

func redirect(w http.ResponseWriter, r *http.Request) {
	var host, _, _ = net.SplitHostPort(r.Host)

	var red = mapping[host]

	if red != "" {
		http.Redirect(w, r, red, http.StatusMovedPermanently)
	} else {
		if defaultRedirect != "" {
			http.Redirect(w, r, defaultRedirect, http.StatusMovedPermanently)
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
		return
	}
}

func serve() {
	if mappingFile != "" {
		load(mappingFile)
	}
	http.HandleFunc("/", redirect)

	endpoint := net.JoinHostPort(serverInterface, strconv.Itoa(serverPort))
	fmt.Println("Listening on " + endpoint)
	err := http.ListenAndServe(endpoint, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func init() {
	RootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVarP(&serverInterface, "bind", "", "0.0.0.0", "interface to which the server will bind")
	serveCmd.Flags().IntVarP(&serverPort, "port", "p", 8080, "port on which the server will listen")
	serveCmd.Flags().StringVarP(&defaultRedirect, "default", "", "", "default URL to redirect to (e.g. https://www.example.com)")
	serveCmd.Flags().StringVarP(&mappingFile, "mappings", "", "", "mapping files with redirects")
}
