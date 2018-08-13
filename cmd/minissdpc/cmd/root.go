// Copyright © 2018 Dave Russell <forfuncsake@gmail.com>
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
	"fmt"
	"os"

	"github.com/forfuncsake/minissdpd"
	"github.com/spf13/cobra"
)

var socket string
var client *minissdpd.Client

func initClient() {
	if client == nil {
		client = &minissdpd.Client{
			SocketPath: socket,
		}
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "minissdpc",
	Short: "A client to interact with minissdpd on its Unix socket",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&socket, "socket", minissdpd.DefaultSocket, "minissdpd's unix socket `path`")
}

func printServices(services []minissdpd.Service) {
	if len(services) == 0 {
		fmt.Println("No matching services returned")
		return
	}
	for _, s := range services {
		fmt.Printf("Type: %s\nUSN: %s\nLocation: %s\n\n", s.Type, s.USN, s.Location)
	}
}
