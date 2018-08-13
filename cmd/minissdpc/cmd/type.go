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

// typeCmd represents the type command
var typeCmd = &cobra.Command{
	Use:   "type",
	Short: "filter services that match a type string",

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Fprintln(os.Stderr, "a single type filter string must be provided")
			os.Exit(3)
		}
		initClient()
		err := client.Connect()
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not connect to minissdpd: %v\n", err)
			os.Exit(2)
		}
		defer client.Close()

		services, err := client.GetServicesByType(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not get services by type: %v\n", err)
			os.Exit(2)
		}

		printServices(services)
		os.Exit(0)
	},
}

func init() {
	lsCmd.AddCommand(typeCmd)
}
