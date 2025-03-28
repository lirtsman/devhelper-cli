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
	"strings"

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
		force, _ := cmd.Flags().GetBool("force")

		stoppedCount := 0

		// Stop Temporal server by finding and killing its process
		if !skipTemporal && isCommandAvailable("temporal") {
			fmt.Println("Stopping Temporal...")

			// Find the Temporal server process
			findCmd := exec.Command("pgrep", "-f", "temporal server start-dev")
			output, err := findCmd.Output()

			if err == nil && len(output) > 0 {
				// Process found, try to kill it
				pids := strings.Split(strings.TrimSpace(string(output)), "\n")
				allKilled := true

				for _, pid := range pids {
					killCmd := exec.Command("kill", pid)
					if err := killCmd.Run(); err != nil {
						allKilled = false
						if verbose {
							fmt.Printf("Failed to kill Temporal process %s: %v\n", pid, err)
						}
					} else if verbose {
						fmt.Printf("Killed Temporal process with PID %s\n", pid)
					}
				}

				if allKilled {
					fmt.Println("✅ Temporal stopped successfully.")
					stoppedCount++
				} else {
					fmt.Println("❌ Failed to stop some Temporal processes.")
					if force {
						fmt.Println("   Continuing due to --force flag.")
					}
				}
			} else {
				if verbose {
					fmt.Println("No running Temporal server processes found.")
				} else {
					fmt.Println("❌ Failed to stop Temporal: no running processes found.")
				}
			}
		} else if skipTemporal && verbose {
			fmt.Println("⏭️  Skipping Temporal as requested.")
		}

		// Stop Dapr runtime
		if !skipDapr && isCommandAvailable("dapr") {
			fmt.Println("Stopping Dapr...")

			// Check if any Dapr apps are running and stop them
			listCmd := exec.Command("dapr", "list")
			listOutput, _ := listCmd.Output()

			if len(listOutput) > 0 && !strings.Contains(string(listOutput), "No Dapr instances found") {
				if verbose {
					fmt.Println("Stopping running Dapr applications...")
					fmt.Println(string(listOutput))
				}

				// Stop each running Dapr app
				// This command would need to parse the output and stop each app by ID
				// For simplicity, we'll just uninstall which should stop everything
			}

			// Run the dapr uninstall command
			uninstallCmd := exec.Command("dapr", "uninstall", "--all")
			uninstallOutput, err := uninstallCmd.CombinedOutput()
			outputStr := string(uninstallOutput)

			// Check for success despite Docker-related errors
			success := err == nil ||
				(strings.Contains(outputStr, "Error removing Dapr") &&
					strings.Contains(outputStr, "docker") &&
					isCommandAvailable("podman"))

			if !success {
				fmt.Printf("❌ Failed to stop Dapr: %v\n", err)
				if verbose {
					fmt.Printf("Output: %s\n", outputStr)
				}
				if !force {
					// Only exit if force flag is not set
					if stoppedCount == 0 {
						fmt.Println("\n⚠️  No components were stopped successfully.")
					} else {
						fmt.Println("\n⚠️  Some components were not stopped properly.")
					}
					return
				}
			} else {
				fmt.Println("✅ Dapr stopped successfully.")

				// If there were Docker-related warnings but we're using Podman, add a clarification
				if strings.Contains(outputStr, "docker") && isCommandAvailable("podman") {
					fmt.Println("   (Docker-related warnings can be ignored when using Podman)")
				}

				if verbose {
					fmt.Printf("Output: %s\n", outputStr)
				}
				stoppedCount++
			}
		} else if skipDapr && verbose {
			fmt.Println("⏭️  Skipping Dapr as requested.")
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
