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

	"github.com/spf13/cobra"
)

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "list services currently advertised by minissdpd",

	Run: func(cmd *cobra.Command, args []string) {
		initClient()
		err := client.Connect()
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not connect to minissdpd: %v\n", err)
			os.Exit(2)
		}
		defer client.Close()

		services, err := client.GetServicesAll()
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not list all services: %v\n", err)
			os.Exit(2)
		}

		printServices(services)
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)
}
