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
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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
		skipDaprDashboard, _ := cmd.Flags().GetBool("skip-dapr-dashboard")
		force, _ := cmd.Flags().GetBool("force")
		configPath, _ := cmd.Flags().GetString("config")

		// If no config path is provided, look for localenv.yaml in current directory
		if configPath == "" {
			configPath = "localenv.yaml"
		}

		// Load configuration if available
		config := LocalEnvConfig{}
		configLoaded := false

		// Check if config file exists
		if _, err := os.Stat(configPath); err == nil {
			// Read and parse configuration
			configData, err := os.ReadFile(configPath)
			if err == nil {
				err = yaml.Unmarshal(configData, &config)
				if err == nil {
					configLoaded = true
					fmt.Printf("✅ Loaded configuration from %s\n", configPath)
				} else if verbose {
					fmt.Printf("⚠️ Failed to parse configuration: %v\n", err)
				}
			} else if verbose {
				fmt.Printf("⚠️ Failed to read configuration: %v\n", err)
			}
		} else if verbose {
			fmt.Printf("⚠️ Configuration file not found at %s\n", configPath)
			fmt.Println("   Run 'shielddev-cli localenv init' to create a configuration")
		}

		// Determine which components to stop based on config and flags
		stopDapr := !skipDapr
		stopTemporal := !skipTemporal
		stopDaprDashboard := !skipDaprDashboard

		if configLoaded {
			// If config is loaded, only stop enabled components (unless explicitly skipped)
			if !config.Components.Dapr {
				stopDapr = false
			}
			if !config.Components.Temporal {
				stopTemporal = false
			}
			if !config.Components.DaprDashboard {
				stopDaprDashboard = false
			}
		}

		stoppedCount := 0

		// Stop Dapr Dashboard by finding and killing its process
		if stopDaprDashboard && isCommandAvailable("dapr") && isDaprDashboardAvailable() {
			fmt.Println("Stopping Dapr Dashboard...")

			// Try to find the dashboard process
			found := false

			// Try using pgrep first (more reliable on macOS and Linux)
			pgrepCmd := exec.Command("pgrep", "-f", "dapr dashboard")
			pgrepOutput, pgrepErr := pgrepCmd.Output()

			if pgrepErr == nil && len(pgrepOutput) > 0 {
				// Process found with pgrep
				pids := strings.Split(strings.TrimSpace(string(pgrepOutput)), "\n")
				allKilled := true
				found = true

				for _, pid := range pids {
					killCmd := exec.Command("kill", pid)
					if err := killCmd.Run(); err != nil {
						allKilled = false
						if verbose {
							fmt.Printf("Failed to kill Dapr Dashboard process %s: %v\n", pid, err)
						}
					} else if verbose {
						fmt.Printf("Killed Dapr Dashboard process with PID %s\n", pid)
					}
				}

				if allKilled {
					fmt.Println("✅ Dapr Dashboard stopped successfully.")
					stoppedCount++
				} else {
					fmt.Println("❌ Failed to stop some Dapr Dashboard processes.")
					if force {
						fmt.Println("   Continuing due to --force flag.")
					}
				}
			} else {
				// Try using ps as fallback
				psCmd := exec.Command("ps", "-ef")
				psOutput, psErr := psCmd.Output()

				if psErr == nil {
					lines := strings.Split(string(psOutput), "\n")
					var dashboardPids []string

					for _, line := range lines {
						if strings.Contains(line, "dapr dashboard") && !strings.Contains(line, "grep") {
							// Parse the line to extract PID
							fields := strings.Fields(line)
							if len(fields) > 1 {
								dashboardPids = append(dashboardPids, fields[1])
							}
						}
					}

					if len(dashboardPids) > 0 {
						found = true
						allKilled := true

						for _, pid := range dashboardPids {
							killCmd := exec.Command("kill", pid)
							if err := killCmd.Run(); err != nil {
								allKilled = false
								if verbose {
									fmt.Printf("Failed to kill Dapr Dashboard process %s: %v\n", pid, err)
								}
							} else if verbose {
								fmt.Printf("Killed Dapr Dashboard process with PID %s\n", pid)
							}
						}

						if allKilled {
							fmt.Println("✅ Dapr Dashboard stopped successfully.")
							stoppedCount++
						} else {
							fmt.Println("❌ Failed to stop some Dapr Dashboard processes.")
							if force {
								fmt.Println("   Continuing due to --force flag.")
							}
						}
					}
				}
			}

			if !found {
				if verbose {
					fmt.Println("No running Dapr Dashboard processes found.")
				} else {
					fmt.Println("❌ No running Dapr Dashboard processes found.")
				}
			}
		} else if !stopDaprDashboard && verbose {
			fmt.Println("⏭️  Skipping Dapr Dashboard (disabled in config or by flag).")
		}

		// Stop Temporal server by finding and killing its process
		if stopTemporal && isCommandAvailable("temporal") {
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
		} else if !stopTemporal && verbose {
			fmt.Println("⏭️  Skipping Temporal (disabled in config or by flag).")
		}

		// Stop Dapr runtime
		if stopDapr && isCommandAvailable("dapr") {
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
			uninstallCmd := exec.Command("dapr", "uninstall", "--all", "--container-runtime", "podman")
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
		} else if !stopDapr && verbose {
			fmt.Println("⏭️  Skipping Dapr (disabled in config or by flag).")
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
	stopCmd.Flags().Bool("skip-dapr-dashboard", false, "Skip stopping Dapr Dashboard")
	stopCmd.Flags().Bool("force", false, "Force stop all components even if errors occur")
	stopCmd.Flags().StringP("config", "c", "", "Path to environment configuration file (default: localenv.yaml)")
}
