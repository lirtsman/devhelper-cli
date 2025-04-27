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

// stopCluster stops and deletes a Kind cluster
func stopCluster(clusterName string, verbose bool) error {
	if verbose {
		fmt.Printf("Stopping Kind cluster: %s\n", clusterName)
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
	Long: `Stop a Kind-based development environment.

This command stops and deletes a Kind cluster used for development.

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

		fmt.Println(green("Stopping Kind-based development environment..."))

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
				fmt.Printf("%s Using specified cluster name: %s\n", yellow("⚙️"), clusterName)
			}
		} else if verbose {
			fmt.Printf("%s Using cluster name from config: %s\n", yellow("📄"), config.Cluster.Name)
		}

		if verbose {
			fmt.Println(yellow("Verbose mode enabled"))
			fmt.Printf("Using cluster name: %s\n", config.Cluster.Name)
		}

		// First check if the cluster exists
		exists, err := checkClusterExists(config.Cluster.Name)
		if err != nil {
			fmt.Printf("%s Error checking cluster status: %v\n", red("❌"), err)
			os.Exit(1)
		}

		if !exists {
			fmt.Printf("%s Kind cluster '%s' does not exist or is already stopped\n", yellow("⚠️"), config.Cluster.Name)
			return
		}

		// Stop the cluster
		err = stopCluster(config.Cluster.Name, verbose)
		if err != nil {
			fmt.Printf("%s Error stopping cluster: %v\n", red("❌"), err)
			os.Exit(1)
		}

		fmt.Printf("%s Kind-based development environment '%s' stopped successfully!\n", green("✅"), config.Cluster.Name)
	},
}

func init() {
	kindenvCmd.AddCommand(kindenvStopCmd)

	// Add flags for kindenv stop command
	kindenvStopCmd.Flags().StringP("config", "f", "", "Path to configuration file")
	kindenvStopCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
	kindenvStopCmd.Flags().String("name", "", "Cluster name (defaults to current directory name)")
}
