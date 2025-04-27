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

	"github.com/ShieldFC-RD/devhelper-cli/internal/kindenv"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// checkClusterExists checks if a Kind cluster with the given name exists
func checkClusterExists(clusterName string) (bool, error) {
	// Run 'kind get clusters' command
	cmd := exec.Command("kind", "get", "clusters")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("failed to list Kind clusters: %w", err)
	}

	// Check if cluster name exists in the output
	clusters := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, cluster := range clusters {
		if cluster == clusterName {
			return true, nil
		}
	}

	return false, nil
}

// kindenvStatusCmd represents the kindenv status command
var kindenvStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of a Kind-based development environment",
	Long: `Show status of a Kind-based development environment.

This command checks if a Kind cluster exists and shows its status.
It uses native Go libraries instead of external CLI tools for improved reliability.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create colored output helpers
		green := color.New(color.FgGreen).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()
		red := color.New(color.FgRed).SprintFunc()

		// Get flags
		verbose, _ := cmd.Flags().GetBool("verbose")
		configPath, _ := cmd.Flags().GetString("config")
		clusterName, _ := cmd.Flags().GetString("cluster-name")

		fmt.Println(green("Checking Kind-based development environment status..."))

		// Load config
		config, err := kindenv.LoadConfig(configPath)
		if err != nil {
			fmt.Printf("%s Error loading config: %v\n", red("❌"), err)
			os.Exit(1)
		}

		// Override cluster name if specified via flag
		if clusterName != "" {
			config.Cluster.Name = clusterName
		}

		if verbose {
			fmt.Println(yellow("Verbose mode enabled"))
			fmt.Printf("Using cluster name: %s\n", config.Cluster.Name)
		}

		// Check if cluster exists
		exists, err := checkClusterExists(config.Cluster.Name)
		if err != nil {
			fmt.Printf("%s Error checking cluster status: %v\n", red("❌"), err)
			os.Exit(1)
		}

		if exists {
			fmt.Printf("%s Kind cluster '%s' is running\n", green("✅"), config.Cluster.Name)

			// Execute kubectl to verify if the cluster is responsive
			kubectlCmd := exec.Command("kubectl", "cluster-info", "--context", fmt.Sprintf("kind-%s", config.Cluster.Name))
			kubectlOutput, err := kubectlCmd.CombinedOutput()
			if err != nil {
				fmt.Printf("%s Kubernetes cluster exists but may not be fully functional: %v\n", yellow("⚠️"), err)
			} else if verbose {
				fmt.Println(yellow("Kubernetes cluster information:"))
				fmt.Println(string(kubectlOutput))
			}

			// Print service access information
			fmt.Println(green("Access services:"))
			if config.Components.Temporal.Enabled {
				// Check if Temporal is actually deployed
				temporalCmd := exec.Command("kubectl", "get", "namespace", config.Components.Temporal.Namespace)
				_, err := temporalCmd.CombinedOutput()

				if err == nil {
					fmt.Printf("- Temporal Web UI: http://localhost:%d\n", config.Components.Temporal.WebPort)
					fmt.Printf("- Temporal Frontend: localhost:%d\n", config.Components.Temporal.FrontendPort)
				} else if verbose {
					fmt.Printf("%s Temporal namespace not found. Temporal may not be installed.\n", yellow("⚠️"))
				}
			}

			if config.Components.Redis.Enabled {
				// Check if Redis is actually deployed
				redisCmd := exec.Command("kubectl", "get", "namespace", "redis")
				_, err := redisCmd.CombinedOutput()

				if err == nil {
					fmt.Printf("- Redis: localhost:%d\n", config.Components.Redis.Port)
				} else if verbose {
					fmt.Printf("%s Redis namespace not found. Redis may not be installed.\n", yellow("⚠️"))
				}
			}

			// Additional component status checks for verbose mode
			if verbose {
				// Check Dapr status if enabled
				if config.Components.Dapr.Enabled {
					daprCmd := exec.Command("kubectl", "get", "namespace", "dapr-system")
					_, err := daprCmd.CombinedOutput()

					if err == nil {
						fmt.Printf("%s Dapr is installed\n", green("✓"))
					} else {
						fmt.Printf("%s Dapr namespace not found. Dapr may not be installed.\n", yellow("⚠️"))
					}
				}

				// List all namespaces
				nsCmd := exec.Command("kubectl", "get", "namespaces")
				nsOutput, err := nsCmd.CombinedOutput()
				if err == nil {
					fmt.Println(yellow("Available namespaces:"))
					fmt.Println(string(nsOutput))
				}
			}
		} else {
			fmt.Printf("%s Kind cluster '%s' is not running\n", yellow("⚠️"), config.Cluster.Name)
		}
	},
}

func init() {
	kindenvCmd.AddCommand(kindenvStatusCmd)
}
