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
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	yamlv3 "gopkg.in/yaml.v3"
)

// LocalEnvConfig represents the configuration for the local environment
type LocalEnvConfig struct {
	Components struct {
		Dapr          bool `yaml:"dapr"`
		Temporal      bool `yaml:"temporal"`
		DaprDashboard bool `yaml:"daprDashboard"`
	} `yaml:"components"`
	Paths struct {
		Podman   string `yaml:"podman"`
		Kind     string `yaml:"kind"`
		Dapr     string `yaml:"dapr"`
		Temporal string `yaml:"temporal"`
	} `yaml:"paths"`
	Temporal struct {
		Namespace string `yaml:"namespace"`
		UIPort    int    `yaml:"uiPort"`
		GRPCPort  int    `yaml:"grpcPort"`
	} `yaml:"temporal"`
	Dapr struct {
		DashboardPort int `yaml:"dashboardPort"`
		ZipkinPort    int `yaml:"zipkinPort"`
	} `yaml:"dapr"`
}

// initCmd represents the init command
var localenvInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize local development environment configuration",
	Long: `Initialize the local development environment by:
1. Validating required tools (Podman, Kind, Dapr CLI, Temporal CLI)
2. Creating a configuration file (localenv.yaml) in the current directory
3. Allowing you to customize which components to enable

This command should be run once before using other localenv commands.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Initializing local development environment...")

		force, _ := cmd.Flags().GetBool("force")
		verbose, _ := cmd.Flags().GetBool("verbose")

		// Check if localenv.yaml already exists
		configPath := "localenv.yaml"
		if _, err := os.Stat(configPath); err == nil && !force {
			fmt.Println("❌ Configuration file already exists. Use --force to overwrite.")
			return
		}

		// Initialize default configuration
		config := LocalEnvConfig{}
		config.Components.Dapr = true
		config.Components.Temporal = true
		config.Components.DaprDashboard = false // Disabled by default since it requires separate installation

		// Set default Temporal configuration
		config.Temporal.Namespace = "default"
		config.Temporal.UIPort = 8233
		config.Temporal.GRPCPort = 7233

		// Set default Dapr configuration
		config.Dapr.DashboardPort = 8080
		config.Dapr.ZipkinPort = 9411

		// Check for required tools and record their paths
		fmt.Println("\n=== Validating Required Tools ===")

		// Check Podman
		podmanPath, podmanErr := validateTool("podman", "--version", verbose)
		if podmanErr != nil {
			fmt.Printf("❌ Podman: %v\n", podmanErr)
			fmt.Println("   Please install Podman from: https://podman.io/getting-started/installation")
		} else {
			fmt.Printf("✅ Podman: Found at %s\n", podmanPath)
			config.Paths.Podman = podmanPath

			// Check if Podman can run containers
			cmd := exec.Command(podmanPath, "ps")
			if err := cmd.Run(); err != nil {
				fmt.Println("⚠️  Podman is installed but may not be configured correctly to run containers")
				fmt.Println("   Make sure you have proper permissions and Podman is configured correctly")
			}
		}

		// Check Kind
		kindPath, kindErr := validateTool("kind", "--version", verbose)
		if kindErr != nil {
			fmt.Printf("❌ Kind: %v\n", kindErr)
			fmt.Println("   Please install Kind from: https://kind.sigs.k8s.io/docs/user/quick-start/#installation")
		} else {
			fmt.Printf("✅ Kind: Found at %s\n", kindPath)
			config.Paths.Kind = kindPath

			// Check if Kind has any clusters configured
			clusterCmd := exec.Command(kindPath, "get", "clusters")
			output, err := clusterCmd.CombinedOutput()
			if err != nil || strings.TrimSpace(string(output)) == "" {
				fmt.Println("⚠️  Kind is installed but no clusters are configured")
				fmt.Println("   Note: Kubernetes functionality is not required for local development")
			} else {
				fmt.Printf("   Clusters found: %s\n", strings.ReplaceAll(string(output), "\n", ", "))
			}
		}

		// Check Dapr CLI
		daprPath, daprErr := validateTool("dapr", "--version", verbose)
		if daprErr != nil {
			fmt.Printf("❌ Dapr CLI: %v\n", daprErr)
			fmt.Println("   Please install Dapr CLI from: https://docs.dapr.io/getting-started/install-dapr-cli/")
			config.Components.Dapr = false
		} else {
			fmt.Printf("✅ Dapr CLI: Found at %s\n", daprPath)
			config.Paths.Dapr = daprPath

			// Check if Docker is available (for original Dapr support)
			dockerAvailable := isCommandAvailable("docker")

			// If Podman is available but Docker is not, add a note about using --container-runtime
			if podmanErr == nil && !dockerAvailable {
				fmt.Println("   Note: Dapr will be initialized with '--container-runtime podman' since Docker is not available")
				fmt.Println("   This will start Redis, Zipkin, placement, and scheduler containers using Podman")
			}

			// Check if Dapr is initialized
			listCmd := exec.Command(daprPath, "list")
			if err := listCmd.Run(); err != nil {
				fmt.Println("⚠️  Dapr CLI is installed but Dapr may not be initialized")
				fmt.Println("   Dapr will be initialized when you run 'devhelper-cli localenv start'")
			} else {
				fmt.Println("   Dapr is initialized and ready to use")
			}

			// Check if Dapr Dashboard is available
			dashboardCmd := exec.Command(daprPath, "dashboard", "--help")
			if err := dashboardCmd.Run(); err == nil {
				fmt.Println("✅ Dapr Dashboard: Available")
				fmt.Println("   Run with: 'dapr dashboard'")
				fmt.Println("   Or use 'devhelper-cli localenv start' with the Dapr Dashboard enabled")

				// Enable Dapr Dashboard by default if available
				config.Components.DaprDashboard = true
			} else {
				fmt.Println("ℹ️ Dapr Dashboard: Not available")
				fmt.Println("   Install with: 'dapr dashboard install'")
				config.Components.DaprDashboard = false
			}
		}

		// Check Temporal CLI
		temporalPath, temporalErr := validateTool("temporal", "--version", verbose)
		if temporalErr != nil {
			fmt.Printf("❌ Temporal CLI: %v\n", temporalErr)
			fmt.Println("   Please install Temporal CLI from: https://docs.temporal.io/cli#install")
			config.Components.Temporal = false
		} else {
			fmt.Printf("✅ Temporal CLI: Found at %s\n", temporalPath)
			config.Paths.Temporal = temporalPath
		}

		// Write configuration to file
		configData, err := yamlv3.Marshal(config)
		if err != nil {
			fmt.Printf("❌ Failed to generate configuration: %v\n", err)
			return
		}

		if err := os.WriteFile(configPath, configData, 0644); err != nil {
			fmt.Printf("❌ Failed to write configuration file: %v\n", err)
			return
		}

		fmt.Printf("\n✅ Configuration written to %s\n", configPath)
		fmt.Println("\nYou can now use 'devhelper-cli localenv start' to start the local environment")
		fmt.Println("or edit the configuration file to customize enabled components.")

		// Print a summary of enabled components
		fmt.Println("\n=== Enabled Components ===")
		if config.Components.Dapr {
			fmt.Println("✅ Dapr: Enabled")
		} else {
			fmt.Println("❌ Dapr: Disabled (install Dapr CLI to enable)")
		}

		if config.Components.DaprDashboard {
			fmt.Println("✅ Dapr Dashboard: Enabled")
		} else if config.Components.Dapr {
			fmt.Println("❌ Dapr Dashboard: Disabled (install with 'dapr dashboard install')")
		}

		if config.Components.Temporal {
			fmt.Println("✅ Temporal: Enabled")
		} else {
			fmt.Println("❌ Temporal: Disabled (install Temporal CLI to enable)")
		}
	},
}

func validateTool(name, versionFlag string, verbose bool) (string, error) {
	// Check if the tool is in PATH
	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("not found in PATH")
	}

	// Check if the tool works by running version command
	cmd := exec.Command(path, versionFlag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return path, fmt.Errorf("found but failed to run: %v", err)
	}

	if verbose {
		fmt.Printf("   %s version output: %s\n", name, strings.TrimSpace(string(output)))
	}

	return path, nil
}

func init() {
	localenvCmd.AddCommand(localenvInitCmd)

	// Add flags
	localenvInitCmd.Flags().BoolP("force", "f", false, "Force overwrite of existing configuration")
	localenvInitCmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
}
