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

	"github.com/ShieldFC-RD/devhelper-cli/internal/kindenv"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)


// stopCluster completely stops and deletes a Kind cluster, removing all deployed components
func stopCluster(clusterName string, verbose bool) error {
	if verbose {
		fmt.Printf("Stopping and DELETING Kind cluster: %s\n", clusterName)
		fmt.Println("This will remove all deployed components and data!")
	}

	// Delete the cluster using 'kind delete cluster' command
	cmd := exec.Command("kind", "delete", "cluster", "--name", clusterName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop cluster: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// kindenvStopCmd represents the kindenv stop command
var kindenvStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a Kind-based development environment",
	Long: `Stop and DELETE a Kind-based development environment.

This command completely stops and DELETES the Kind cluster used for development.
All deployed services, applications, and data within the cluster will be removed.

By default, the command will use the cluster name from kindenv.yaml.
You can override this with the --name flag.

It uses native Go libraries instead of external CLI tools for improved reliability.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create colored output helpers
		green := color.New(color.FgGreen).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()
		red := color.New(color.FgRed).SprintFunc()

		// Get flags
		verbose, _ := cmd.Flags().GetBool("verbose")
		configPath, _ := cmd.Flags().GetString("config")
		clusterName, _ := cmd.Flags().GetString("name")

		fmt.Println(green("Stopping and DELETING Kind-based development environment..."))

		// Load config
		config, err := kindenv.LoadConfig(configPath)
		if err != nil {
			fmt.Printf("%s Error loading config: %v\n", red("❌"), err)
			os.Exit(1)
		}

		// Override cluster name only if explicitly provided with --name flag
		if cmd.Flags().Changed("name") {
			config.Cluster.Name = clusterName
			if verbose {
				fmt.Printf("%s Using specified cluster name for deletion: %s\n", yellow("⚙️"), clusterName)
			}
		} else if verbose {
			fmt.Printf("%s Using cluster name from config for deletion: %s\n", yellow("📄"), config.Cluster.Name)
		}

		if verbose {
			fmt.Println(yellow("Verbose mode enabled"))
			fmt.Printf("Target cluster for deletion: %s\n", config.Cluster.Name)
		}

		// First check if the cluster exists
		exists, err := checkClusterExists(config.Cluster.Name)
		if err != nil {
			fmt.Printf("%s Error checking cluster status: %v\n", red("❌"), err)
			os.Exit(1)
		}

		if !exists {
			fmt.Printf("%s Kind cluster '%s' does not exist or has already been deleted\n", yellow("⚠️"), config.Cluster.Name)
			return
		}

		// Clean up MySQL if enabled
		if config.Components.MySQL.Enabled {
			if verbose {
				fmt.Println(yellow("Cleaning up MySQL resources..."))
			}

			// Uninstall MySQL Helm release
			_, err = executeCommand("helm", "uninstall", "mysql", "--namespace", config.Components.MySQL.Namespace, "--ignore-not-found")
			if err != nil {
				if verbose {
					fmt.Printf("%s Warning: Failed to uninstall MySQL Helm release: %v\n", yellow("⚠️"), err)
				}
			} else if verbose {
				fmt.Printf("%s MySQL Helm release uninstalled\n", green("✅"))
			}

			// Clean up PersistentVolumeClaims if persistence was enabled
			if config.Components.MySQL.Persistence.Enabled {
				if verbose {
					fmt.Println(yellow("Cleaning up MySQL PersistentVolumeClaims..."))
				}
				_, err = executeCommand("kubectl", "delete", "pvc", "-n", config.Components.MySQL.Namespace, "--all", "--ignore-not-found")
				if err != nil {
					if verbose {
						fmt.Printf("%s Warning: Failed to delete MySQL PVCs: %v\n", yellow("⚠️"), err)
					}
				} else if verbose {
					fmt.Printf("%s MySQL PersistentVolumeClaims cleaned up\n", green("✅"))
				}
			}

			// Delete MySQL namespace
			_, err = executeCommand("kubectl", "delete", "namespace", config.Components.MySQL.Namespace, "--ignore-not-found")
			if err != nil {
				if verbose {
					fmt.Printf("%s Warning: Failed to delete MySQL namespace: %v\n", yellow("⚠️"), err)
				}
			} else if verbose {
				fmt.Printf("%s MySQL namespace deleted\n", green("✅"))
			}
		}

		// Stop the cluster
		err = stopCluster(config.Cluster.Name, verbose)
		if err != nil {
			fmt.Printf("%s Error stopping cluster: %v\n", red("❌"), err)
			os.Exit(1)
		}

		fmt.Printf("%s Kind-based development environment '%s' has been completely stopped and DELETED!\n", green("✅"), config.Cluster.Name)
		fmt.Println(yellow("All deployed services, applications, and their data have been removed."))
		fmt.Println(yellow("To recreate the environment, use: devhelper-cli kindenv start"))
	},
}

func init() {
	kindenvCmd.AddCommand(kindenvStopCmd)

	// Add flags for kindenv stop command
	kindenvStopCmd.Flags().StringP("config", "f", "", "Path to configuration file")
	kindenvStopCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
	kindenvStopCmd.Flags().String("name", "", "Cluster name (defaults to current directory name)")
}
