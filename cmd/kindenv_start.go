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

// kindenvStartCmd represents the kindenv start command
var kindenvStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a Kind-based development environment",
	Long: `Start a Kind-based development environment for Shield applications.

This command creates a Kind cluster if it doesn't exist and installs required components:
- Temporal
- Redis
- Dapr
- cert-manager

It also sets up the Kubernetes context and creates required namespaces.

This implementation uses native Go libraries instead of external CLI tools for improved reliability.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get flags
		verbose, _ := cmd.Flags().GetBool("verbose")
		configPath, _ := cmd.Flags().GetString("config")
		useAwsEcr, _ := cmd.Flags().GetBool("use-aws-ecr")
		skipTemporal, _ := cmd.Flags().GetBool("skip-temporal")
		skipDapr, _ := cmd.Flags().GetBool("skip-dapr")
		skipRedis, _ := cmd.Flags().GetBool("skip-redis")
		sequential, _ := cmd.Flags().GetBool("sequential")

		// Enhanced debugging
		debugMode, _ := cmd.Flags().GetBool("debug")

		// Load config file
		fmt.Println("Setting up Kind-based development environment...")

		if debugMode {
			fmt.Println("Debug: Loading config from", configPath)
		}

		// Load config
		config, err := kindenv.LoadConfig(configPath)
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Debug output config
		if debugMode {
			fmt.Printf("Debug: Configuration loaded. Components enabled:\n")
			fmt.Printf("  - Temporal: %v\n", config.Components.Temporal.Enabled)
			fmt.Printf("  - Redis: %v\n", config.Components.Redis.Enabled)
			fmt.Printf("  - Dapr: %v\n", config.Components.Dapr.Enabled)
			fmt.Printf("  - AWS ECR: %v\n", config.Images.UseAwsEcr)
		}

		// Override config with command line args
		if useAwsEcr {
			config.Images.UseAwsEcr = true
			if debugMode {
				fmt.Println("Debug: Enabling AWS ECR from command line")
			}
		}

		// Set component enabled/disabled based on skip flags
		if skipTemporal {
			config.Components.Temporal.Enabled = false
			if debugMode {
				fmt.Println("Debug: Disabling Temporal from command line")
			}
		}

		if skipDapr {
			config.Components.Dapr.Enabled = false
			if debugMode {
				fmt.Println("Debug: Disabling Dapr from command line")
			}
		}

		if skipRedis {
			config.Components.Redis.Enabled = false
			if debugMode {
				fmt.Println("Debug: Disabling Redis from command line")
			}
		}

		// Create KindEnv manager
		if debugMode {
			fmt.Println("Debug: Creating KindEnv manager")
		}
		manager, err := kindenv.NewManager(config, verbose || debugMode)
		if err != nil {
			fmt.Printf("Error creating manager: %v\n", err)
			os.Exit(1)
		}

		// Start the cluster
		if debugMode {
			fmt.Println("Debug: Starting cluster and installing components...")
			if sequential {
				fmt.Println("Debug: Using sequential installation mode")
			}
		}
		if err := manager.StartCluster(sequential); err != nil {
			fmt.Printf("Error starting cluster or installing components: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Kind-based development environment setup complete!")
		fmt.Println("Access services:")
		if config.Components.Temporal.Enabled {
			fmt.Printf("- Temporal Web UI: http://localhost:%d\n", config.Components.Temporal.WebPort)
			fmt.Printf("- Temporal Frontend: localhost:%d\n", config.Components.Temporal.FrontendPort)
		}
		if config.Components.Redis.Enabled {
			fmt.Printf("- Redis: localhost:%d\n", config.Components.Redis.Port)
		}
	},
}

func init() {
	kindenvCmd.AddCommand(kindenvStartCmd)

	// Add flags for kindenv start command
	kindenvStartCmd.Flags().Bool("skip-temporal", false, "Skip installing Temporal")
	kindenvStartCmd.Flags().Bool("skip-dapr", false, "Skip installing Dapr")
	kindenvStartCmd.Flags().Bool("skip-redis", false, "Skip installing Redis")
	kindenvStartCmd.Flags().Bool("skip-cert-manager", false, "Skip installing cert-manager")
	kindenvStartCmd.Flags().String("operator-namespace", "temporal-worker-operator-system",
		"Namespace for Temporal worker operator")
	kindenvStartCmd.Flags().Bool("use-aws-ecr", false, "Use AWS ECR for pulling images")
	kindenvStartCmd.Flags().Bool("debug", false, "Enable additional debug output")
	kindenvStartCmd.Flags().Bool("sequential", false, "Install components one at a time in sequence, waiting for each to complete")
}
