/*
Copyright © 2023 Shield

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
package cmd

import (
	"fmt"
	"os"

	"github.com/ShieldFC-RD/devhelper-cli/internal/kindenv"
	"github.com/spf13/cobra"
)

// kindenvStopCmd represents the kindenv stop command
var kindenvStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a Kind-based development environment",
	Long: `Stop a Kind-based development environment.

This command stops and deletes a Kind cluster used for development.
It uses native Go libraries instead of external CLI tools for improved reliability.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get flags
		verbose, _ := cmd.Flags().GetBool("verbose")
		configPath, _ := cmd.Flags().GetString("config")

		// Load config file
		fmt.Println("Stopping Kind-based development environment...")

		// Load config
		config, err := kindenv.LoadConfig(configPath)
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Create KindEnv manager
		manager, err := kindenv.NewManager(config, verbose)
		if err != nil {
			fmt.Printf("Error creating manager: %v\n", err)
			os.Exit(1)
		}

		// Stop the cluster
		if err := manager.StopCluster(); err != nil {
			fmt.Printf("Error stopping cluster: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Kind-based development environment stopped successfully!")
	},
}

func init() {
	kindenvCmd.AddCommand(kindenvStopCmd)
}
