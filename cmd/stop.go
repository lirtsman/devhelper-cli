/*
Copyright © 2023 ShieldDev

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
	"os/exec"

	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop local development environment",
	Long: `Stop the local development environment components.

This command will stop all running components in the local development
environment, including:
- Dapr runtime
- Temporal server
- Related processes`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Stopping local development environment...")

		verbose, _ := cmd.Flags().GetBool("verbose")
		skipDapr, _ := cmd.Flags().GetBool("skip-dapr")
		skipTemporal, _ := cmd.Flags().GetBool("skip-temporal")

		// Components to stop - in reverse order from how we start them
		components := []struct {
			Name      string
			Command   string
			Args      []string
			SkipFlag  bool
			Available bool
		}{
			{
				Name:      "Temporal",
				Command:   "temporal",
				Args:      []string{"server", "stop-dev"},
				SkipFlag:  skipTemporal,
				Available: isCommandAvailable("temporal"),
			},
			{
				Name:      "Dapr",
				Command:   "dapr",
				Args:      []string{"uninstall", "--all"},
				SkipFlag:  skipDapr,
				Available: isCommandAvailable("dapr"),
			},
		}

		stoppedCount := 0

		for _, comp := range components {
			if comp.SkipFlag {
				if verbose {
					fmt.Printf("⏭️  Skipping '%s' as requested.\n", comp.Name)
				}
				continue
			}

			if !comp.Available {
				if verbose {
					fmt.Printf("⚠️  '%s' command not found, skipping.\n", comp.Name)
				}
				continue
			}

			fmt.Printf("Stopping %s...\n", comp.Name)

			// Run the stop command
			cmd := exec.Command(comp.Command, comp.Args...)
			output, err := cmd.CombinedOutput()

			if err != nil {
				fmt.Printf("❌ Failed to stop %s: %v\n", comp.Name, err)
				if verbose {
					fmt.Printf("Output: %s\n", string(output))
				}
				continue
			}

			if verbose {
				fmt.Printf("✅ %s stopped successfully.\n", comp.Name)
				fmt.Printf("Output: %s\n", string(output))
			} else {
				fmt.Printf("✅ %s stopped successfully.\n", comp.Name)
			}

			stoppedCount++
		}

		if stoppedCount > 0 {
			fmt.Println("\n✅ Local development environment has been stopped.")
		} else {
			fmt.Println("\n⚠️  No components were stopped. They may not be running or were not found.")
		}
	},
}

func init() {
	localenvCmd.AddCommand(stopCmd)

	// Add flags specific to the stop command
	stopCmd.Flags().Bool("skip-dapr", false, "Skip stopping Dapr runtime")
	stopCmd.Flags().Bool("skip-temporal", false, "Skip stopping Temporal server")
	stopCmd.Flags().Bool("force", false, "Force stop all components even if errors occur")
}
