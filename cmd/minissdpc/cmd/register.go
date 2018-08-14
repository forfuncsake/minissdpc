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

	"github.com/forfuncsake/minissdpc"
	"github.com/spf13/cobra"
)

// flags
var regType, regUSN, regServer, regLocation string

// registerCmd represents the register command
var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Used to register a new service for minissdp to advertise",

	Run: func(cmd *cobra.Command, args []string) {
		if regType == "" || regUSN == "" || regServer == "" || regLocation == "" {
			fmt.Fprintln(os.Stderr, "All fields must be provided to register a new service, see help for fields")
			os.Exit(3)
		}
		initClient()
		err := client.Connect()
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not connect to minissdpd: %v\n", err)
			os.Exit(2)
		}
		defer client.Close()

		err = client.RegisterService(minissdpc.Service{
			Type:     regType,
			USN:      regUSN,
			Server:   regServer,
			Location: regLocation,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not register new service: %v\n", err)
			os.Exit(2)
		}
		fmt.Printf("service %s successfully registered\n", regUSN)
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)

	registerCmd.Flags().StringVarP(&regType, "type", "t", "", "SSDP service/device type")
	registerCmd.Flags().StringVarP(&regUSN, "usn", "u", "", "SSDP unique service name")
	registerCmd.Flags().StringVarP(&regServer, "server", "s", "", "SSDP server identifier string")
	registerCmd.Flags().StringVarP(&regLocation, "location", "l", "", "URL of the service being advertised")
}
