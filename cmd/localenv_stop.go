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
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	yamlv3 "gopkg.in/yaml.v3"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop local development environment",
	Long: `Stop the local development environment components including:
- Dapr runtime
- Temporal server
- OpenSearch
- Related services and containers`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Stopping local development environment...")

		verbose, _ := cmd.Flags().GetBool("verbose")
		skipDapr, _ := cmd.Flags().GetBool("skip-dapr")
		skipTemporal, _ := cmd.Flags().GetBool("skip-temporal")
		skipDaprDashboard, _ := cmd.Flags().GetBool("skip-dapr-dashboard")
		skipOpenSearch, _ := cmd.Flags().GetBool("skip-opensearch")
		force, _ := cmd.Flags().GetBool("force")
		cleanLogs, _ := cmd.Flags().GetBool("clean-logs")
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
				err = yamlv3.Unmarshal(configData, &config)
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
		}

		// Determine which components to stop based on config and flags
		stopDapr := !skipDapr
		stopTemporal := !skipTemporal
		stopDaprDashboard := !skipDaprDashboard

		if configLoaded {
			// If config is loaded, only stop enabled components (unless explicitly skipped)
			if !config.Components.Dapr.Enabled {
				stopDapr = false
			}
			if !config.Components.Temporal.Enabled {
				stopTemporal = false
			}
			if !config.Components.Dapr.Dashboard {
				stopDaprDashboard = false
			}
		}

		stoppedCount := 0

		// Stop Dapr Dashboard by finding and killing its process
		if stopDaprDashboard && isCommandAvailable("dapr") && isDaprDashboardAvailable() {
			fmt.Println("Stopping Dapr Dashboard...")

			// Try multiple methods to find dashboard processes
			dashboardPids := []string{}
			found := false

			// Method 1: Try using pgrep first (more reliable on macOS and Linux)
			pgrepCmd := exec.Command("pgrep", "-f", "dapr dashboard")
			pgrepOutput, pgrepErr := pgrepCmd.Output()

			if pgrepErr == nil && len(pgrepOutput) > 0 {
				// Process found with pgrep
				pids := strings.Split(strings.TrimSpace(string(pgrepOutput)), "\n")
				for _, pid := range pids {
					if pid != "" {
						dashboardPids = append(dashboardPids, pid)
						found = true
					}
				}
			}

			// Method 2: Try using ps (works on most Unix systems)
			psCmd := exec.Command("ps", "-ef")
			psOutput, psErr := psCmd.Output()

			if psErr == nil {
				lines := strings.Split(string(psOutput), "\n")
				for _, line := range lines {
					if strings.Contains(line, "dapr dashboard") && !strings.Contains(line, "grep") {
						// Parse the line to extract PID
						fields := strings.Fields(line)
						if len(fields) > 1 {
							// Check if we already have this PID
							pidExists := false
							for _, existingPid := range dashboardPids {
								if existingPid == fields[1] {
									pidExists = true
									break
								}
							}
							if !pidExists {
								dashboardPids = append(dashboardPids, fields[1])
								found = true
							}
						}
					}
				}
			}

			// Method 3: Try using lsof to find processes using the dashboard port
			if configLoaded && config.Components.Dapr.DashboardPort > 0 {
				portStr := fmt.Sprintf("%d", config.Components.Dapr.DashboardPort)
				lsofCmd := exec.Command("lsof", "-i", fmt.Sprintf(":%s", portStr))
				lsofOutput, lsofErr := lsofCmd.Output()

				if lsofErr == nil && len(lsofOutput) > 0 {
					lines := strings.Split(string(lsofOutput), "\n")
					for _, line := range lines {
						if strings.Contains(line, "LISTEN") {
							fields := strings.Fields(line)
							if len(fields) > 1 {
								pidExists := false
								for _, existingPid := range dashboardPids {
									if existingPid == fields[1] {
										pidExists = true
										break
									}
								}
								if !pidExists {
									dashboardPids = append(dashboardPids, fields[1])
									found = true
								}
							}
						}
					}
				}
			}

			if found && len(dashboardPids) > 0 {
				allKilled := true

				if verbose {
					fmt.Printf("Found %d Dapr Dashboard processes: %s\n", len(dashboardPids), strings.Join(dashboardPids, ", "))
				}

				for _, pid := range dashboardPids {
					// First try a gentle termination with SIGTERM
					killCmd := exec.Command("kill", pid)
					killErr := killCmd.Run()

					if killErr != nil && force {
						// If that fails and we're forcing, try SIGKILL
						killCmd = exec.Command("kill", "-9", pid)
						killErr = killCmd.Run()
					}

					if killErr != nil {
						allKilled = false
						if verbose {
							fmt.Printf("Failed to kill Dapr Dashboard process %s: %v\n", pid, killErr)
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

			// Get Temporal port configuration
			temporalUIPort := 8233   // Default UI port
			temporalGRPCPort := 7233 // Default GRPC port

			// Load port values from config if available
			if configLoaded {
				if config.Components.Temporal.UIPort != 0 {
					temporalUIPort = config.Components.Temporal.UIPort
				}
				if config.Components.Temporal.GRPCPort != 0 {
					temporalGRPCPort = config.Components.Temporal.GRPCPort
				}
			}

			// Keep track of whether we successfully stopped the server
			temporalStopped := false

			// Try multiple methods to find and stop Temporal processes

			// Method 1: Find the Temporal server process by name
			findCmd := exec.Command("pgrep", "-f", "temporal server start-dev")
			output, err := findCmd.Output()

			if err == nil && len(output) > 0 {
				// Process found, try to kill it
				pids := strings.Split(strings.TrimSpace(string(output)), "\n")
				allKilled := true

				if verbose {
					fmt.Printf("Found %d Temporal server processes: %s\n", len(pids), strings.Join(pids, ", "))
				}

				for _, pid := range pids {
					fmt.Printf("Stopping Temporal server process (PID: %s)...\n", pid)

					// First try graceful termination with SIGTERM
					killCmd := exec.Command("kill", pid)
					killErr := killCmd.Run()

					if killErr != nil && force {
						// If that fails and force flag is set, try SIGKILL
						fmt.Println("  Attempting forceful termination with SIGKILL...")
						killCmd = exec.Command("kill", "-9", pid)
						killErr = killCmd.Run()
					}

					if killErr != nil {
						allKilled = false
						if verbose {
							fmt.Printf("Failed to kill Temporal process %s: %v\n", pid, killErr)
						}
					} else if verbose {
						fmt.Printf("Killed Temporal process with PID %s\n", pid)
					}
				}

				if allKilled {
					temporalStopped = true
				} else {
					fmt.Println("❌ Failed to stop some Temporal processes.")
					if force {
						fmt.Println("   Continuing due to --force flag.")
					}
				}
			} else if verbose {
				fmt.Println("No Temporal server processes found by name search.")
			}

			// Method 2: Find processes using the Temporal UI port
			uiPortCmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", temporalUIPort), "-t")
			uiPortOutput, _ := uiPortCmd.Output()

			if len(uiPortOutput) > 0 {
				pids := strings.Split(strings.TrimSpace(string(uiPortOutput)), "\n")

				if verbose {
					fmt.Printf("Found %d processes using Temporal UI port %d: %s\n",
						len(pids), temporalUIPort, strings.Join(pids, ", "))
				}

				for _, pid := range pids {
					fmt.Printf("Stopping process using Temporal UI port %d (PID: %s)...\n", temporalUIPort, pid)

					// Try to kill the process, with force if requested
					killCmd := exec.Command("kill", pid)
					killErr := killCmd.Run()

					if killErr != nil && force {
						killCmd = exec.Command("kill", "-9", pid)
						killErr = killCmd.Run()
					}

					if killErr == nil {
						temporalStopped = true
					}
				}
			}

			// Method 3: Find processes using the Temporal GRPC port
			grpcPortCmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", temporalGRPCPort), "-t")
			grpcPortOutput, _ := grpcPortCmd.Output()

			if len(grpcPortOutput) > 0 {
				pids := strings.Split(strings.TrimSpace(string(grpcPortOutput)), "\n")

				if verbose {
					fmt.Printf("Found %d processes using Temporal GRPC port %d: %s\n",
						len(pids), temporalGRPCPort, strings.Join(pids, ", "))
				}

				for _, pid := range pids {
					fmt.Printf("Stopping process using Temporal GRPC port %d (PID: %s)...\n", temporalGRPCPort, pid)

					// Try to kill the process, with force if requested
					killCmd := exec.Command("kill", pid)
					killErr := killCmd.Run()

					if killErr != nil && force {
						killCmd = exec.Command("kill", "-9", pid)
						killErr = killCmd.Run()
					}

					if killErr == nil {
						temporalStopped = true
					}
				}
			}

			// Verify ports are actually free
			time.Sleep(2 * time.Second)
			uiPortInUse := isPortInUse(temporalUIPort)
			grpcPortInUse := isPortInUse(temporalGRPCPort)

			if uiPortInUse || grpcPortInUse {
				if uiPortInUse {
					fmt.Printf("❌ Temporal UI port %d is still in use\n", temporalUIPort)
					fmt.Printf("   Try manually killing the process: lsof -i :%d -t | xargs kill -9\n", temporalUIPort)
				}

				if grpcPortInUse {
					fmt.Printf("❌ Temporal GRPC port %d is still in use\n", temporalGRPCPort)
					fmt.Printf("   Try manually killing the process: lsof -i :%d -t | xargs kill -9\n", temporalGRPCPort)
				}

				if force {
					fmt.Println("   Continuing due to --force flag.")
				}
			}

			if temporalStopped {
				fmt.Println("✅ Temporal stopped successfully.")

				// Clean up Temporal server logs
				logsDir := filepath.Join(os.Getenv("HOME"), ".logs", "devhelper-cli")
				logFilePath := filepath.Join(logsDir, "temporal-server.log")

				if _, err := os.Stat(logFilePath); err == nil {
					// Log file exists, clean it up
					if cleanLogs {
						if err := os.Remove(logFilePath); err != nil {
							fmt.Printf("⚠️ Failed to remove log file: %v\n", err)
						} else {
							fmt.Printf("✅ Removed Temporal server log file: %s\n", logFilePath)
						}
					} else {
						fmt.Printf("ℹ️ Temporal server logs are available at: %s\n", logFilePath)
						fmt.Println("   Use --clean-logs flag to remove logs when stopping")
					}
				}

				stoppedCount++
			} else {
				fmt.Println("❌ Failed to find or stop Temporal server processes.")
				if verbose {
					fmt.Println("   If Temporal is still running, you can try:")
					fmt.Printf("   lsof -i :%d -t | xargs kill -9\n", temporalUIPort)
					fmt.Printf("   lsof -i :%d -t | xargs kill -9\n", temporalGRPCPort)
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

		// Stop OpenSearch if enabled
		if !skipOpenSearch && (!configLoaded || config.Components.OpenSearch.Enabled) {
			fmt.Println("\n=== Stopping OpenSearch Dashboard ===")
			stopDashboardCmd := exec.Command("podman", "rm", "-f", "opensearch-dashboard")
			if err := stopDashboardCmd.Run(); err != nil {
				fmt.Printf("❌ Failed to stop OpenSearch Dashboard: %v\n", err)
			} else {
				fmt.Println("✅ OpenSearch Dashboard stopped")
			}

			fmt.Println("\n=== Stopping OpenSearch ===")

			// Check if OpenSearch container is running
			checkCmd := exec.Command("podman", "ps", "--filter", "name=opensearch-node", "--format", "{{.Names}}")
			output, err := checkCmd.CombinedOutput()
			if err == nil && strings.Contains(string(output), "opensearch-node") {
				// Stop and remove the container
				stopCmd := exec.Command("podman", "rm", "-f", "opensearch-node")
				if err := stopCmd.Run(); err != nil {
					fmt.Printf("❌ Failed to stop OpenSearch container: %v\n", err)
					if !force {
						return
					}
				} else {
					fmt.Println("✅ OpenSearch stopped")
				}
			} else {
				fmt.Println("ℹ️ OpenSearch is not running")
			}
		} else {
			fmt.Println("\nℹ️ Skipping OpenSearch")
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
	stopCmd.Flags().Bool("skip-opensearch", false, "Skip stopping OpenSearch")
	stopCmd.Flags().Bool("force", false, "Force stop all components even if errors occur")
	stopCmd.Flags().Bool("clean-logs", false, "Remove log files when stopping components")
	stopCmd.Flags().StringP("config", "c", "", "Path to environment configuration file (default: localenv.yaml)")
}
