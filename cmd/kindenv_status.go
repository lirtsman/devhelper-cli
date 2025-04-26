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

// kindenvStatusCmd represents the kindenv status command
var kindenvStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of a Kind-based development environment",
	Long: `Show status of a Kind-based development environment.

This command checks if a Kind cluster exists and shows its status.
It uses native Go libraries instead of external CLI tools for improved reliability.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get flags
		verbose, _ := cmd.Flags().GetBool("verbose")
		configPath, _ := cmd.Flags().GetString("config")

		// Load config file
		fmt.Println("Checking Kind-based development environment status...")

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

		// Get cluster status
		exists, err := manager.GetClusterStatus()
		if err != nil {
			fmt.Printf("Error checking cluster status: %v\n", err)
			os.Exit(1)
		}

		if exists {
			fmt.Printf("Kind cluster '%s' is running\n", config.Cluster.Name)
			fmt.Println("Access services:")
			if config.Components.Temporal.Enabled {
				fmt.Printf("- Temporal Web UI: http://localhost:%d\n", config.Components.Temporal.WebPort)
				fmt.Printf("- Temporal Frontend: localhost:%d\n", config.Components.Temporal.FrontendPort)
			}
			if config.Components.Redis.Enabled {
				fmt.Printf("- Redis: localhost:%d\n", config.Components.Redis.Port)
			}
		} else {
			fmt.Printf("Kind cluster '%s' is not running\n", config.Cluster.Name)
		}
	},
}

func init() {
	kindenvCmd.AddCommand(kindenvStatusCmd)
}
