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
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Configuration cache to detect changes between runs
type ConfigCache struct {
	DaprDashboardPort int
	TemporalUIPort    int
	TemporalGRPCPort  int
	TemporalNamespace string
}

// Returns true if current config differs from previous state
func hasConfigChanged(config LocalEnvConfig, currentCache ConfigCache) (bool, ConfigCache, []string) {
	changes := []string{}
	newCache := ConfigCache{
		DaprDashboardPort: config.Dapr.DashboardPort,
		TemporalUIPort:    config.Temporal.UIPort,
		TemporalGRPCPort:  config.Temporal.GRPCPort,
		TemporalNamespace: config.Temporal.Namespace,
	}

	hasChanges := false

	// Check for dashboard port change
	if currentCache.DaprDashboardPort != 0 &&
		currentCache.DaprDashboardPort != config.Dapr.DashboardPort {
		changes = append(changes, fmt.Sprintf("Dapr Dashboard port changed: %d → %d",
			currentCache.DaprDashboardPort, config.Dapr.DashboardPort))
		hasChanges = true
	}

	// Check for Temporal UI port change
	if currentCache.TemporalUIPort != 0 &&
		currentCache.TemporalUIPort != config.Temporal.UIPort {
		changes = append(changes, fmt.Sprintf("Temporal UI port changed: %d → %d",
			currentCache.TemporalUIPort, config.Temporal.UIPort))
		hasChanges = true
	}

	// Check for Temporal GRPC port change
	if currentCache.TemporalGRPCPort != 0 &&
		currentCache.TemporalGRPCPort != config.Temporal.GRPCPort {
		changes = append(changes, fmt.Sprintf("Temporal GRPC port changed: %d → %d",
			currentCache.TemporalGRPCPort, config.Temporal.GRPCPort))
		hasChanges = true
	}

	// Check for Temporal namespace change
	if currentCache.TemporalNamespace != "" &&
		currentCache.TemporalNamespace != config.Temporal.Namespace {
		changes = append(changes, fmt.Sprintf("Temporal namespace changed: %s → %s",
			currentCache.TemporalNamespace, config.Temporal.Namespace))
		hasChanges = true
	}

	return hasChanges, newCache, changes
}

// Reads last used configuration from a cache file
func loadConfigCache() ConfigCache {
	cache := ConfigCache{}
	cacheFile := filepath.Join(os.Getenv("HOME"), ".config", "devhelper-cli", "config-cache.yaml")

	if _, err := os.Stat(cacheFile); err == nil {
		// Cache file exists, try to read it
		data, err := os.ReadFile(cacheFile)
		if err == nil {
			yaml.Unmarshal(data, &cache)
		}
	}

	return cache
}

// Saves current configuration to a cache file for future comparison
func saveConfigCache(cache ConfigCache) {
	cacheDir := filepath.Join(os.Getenv("HOME"), ".config", "devhelper-cli")
	cacheFile := filepath.Join(cacheDir, "config-cache.yaml")

	// Create directory if it doesn't exist
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		os.MkdirAll(cacheDir, 0755)
	}

	data, err := yaml.Marshal(cache)
	if err == nil {
		os.WriteFile(cacheFile, data, 0644)
	}
}

// Components to be started
type Component struct {
	Name            string
	Command         string
	Args            []string
	CheckCommand    string
	CheckArgs       []string
	RequiredFor     []string
	StartupDelay    time.Duration
	IsRunning       bool
	IsRequired      bool
	CommandExists   bool
	VerifyAvailable func() bool // Function to verify the component is accessible
	RequiresStartup bool        // Whether the component needs to be started or just verified
	IsBinary        bool        // Whether the component is a binary command (like Podman, Kind) rather than a service
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start local development environment",
	Long: `Start a local development environment with all necessary components
for ShieldDev application development including:

- Dapr runtime
- Temporal server
- Required dependencies

This command will check for necessary dependencies and start them
in the correct order.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting local development environment...")

		verbose, _ := cmd.Flags().GetBool("verbose")
		skipDapr, _ := cmd.Flags().GetBool("skip-dapr")
		skipTemporal, _ := cmd.Flags().GetBool("skip-temporal")
		skipDaprDashboard, _ := cmd.Flags().GetBool("skip-dapr-dashboard")
		configPath, _ := cmd.Flags().GetString("config")
		forceRestart, _ := cmd.Flags().GetBool("force-restart")
		streamLogs, _ := cmd.Flags().GetBool("stream-logs")

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
			fmt.Println("   Run 'devhelper-cli localenv init' to create a configuration")
		}

		// Load previous configuration cache
		configCache := loadConfigCache()

		// Check if configuration has changed
		configChanged, newCache, changes := hasConfigChanged(config, configCache)

		// If this is a first run (no previous cache), initialize the config cache
		if configCache.DaprDashboardPort == 0 && configCache.TemporalUIPort == 0 {
			saveConfigCache(newCache)
		}

		// Override config with command line flags
		if skipDapr {
			config.Components.Dapr = false
		}
		if skipTemporal {
			config.Components.Temporal = false
		}
		if skipDaprDashboard {
			config.Components.DaprDashboard = false
		}

		// Function to check if Temporal server is accessible
		checkTemporalServerRunning := func() bool {
			checkCmd := exec.Command("temporal", "operator", "namespace", "list")
			if err := checkCmd.Run(); err != nil {
				if verbose {
					fmt.Printf("Temporal server check failed: %v\n", err)
				}

				// Try with explicit server address as fallback
				checkCmdWithAddress := exec.Command("temporal", "operator", "--address", "localhost:7233", "namespace", "list")
				if err := checkCmdWithAddress.Run(); err != nil {
					if verbose {
						fmt.Printf("Temporal server check with explicit address failed: %v\n", err)
					}
					return false
				}
				return true
			}
			return true
		}

		// Function to check if Dapr is accessible
		checkDaprRunning := func() bool {
			// For self-hosted mode, we just check if `dapr list` works
			// The `dapr status` command requires -k which is for Kubernetes
			listCmd := exec.Command("dapr", "list")
			_, err := listCmd.CombinedOutput()
			if err != nil {
				if verbose {
					fmt.Printf("Dapr check failed: %v\n", err)
				}
				return false
			}

			// Also verify the Dapr binaries are installed
			_, err = os.Stat(filepath.Join(os.Getenv("HOME"), ".dapr", "bin", "daprd"))
			if err != nil {
				if verbose {
					fmt.Printf("Dapr binaries not found: %v\n", err)
				}
				return false
			}

			// If the command succeeds and binaries exist, we consider Dapr initialized
			return true
		}

		// Function to check if Podman is running
		checkPodmanRunning := func() bool {
			checkCmd := exec.Command("podman", "ps")
			if err := checkCmd.Run(); err != nil {
				if verbose {
					fmt.Printf("Podman check failed: %v\n", err)
				}
				return false
			}
			return true
		}

		// Function to check if Kind has clusters
		checkKindRunning := func() bool {
			checkCmd := exec.Command("kind", "get", "clusters")
			output, err := checkCmd.CombinedOutput()
			if err != nil {
				if verbose {
					fmt.Printf("Kind check failed: %v\n", err)
				}
				return false
			}
			// Check if there's at least one cluster
			return len(strings.TrimSpace(string(output))) > 0
		}

		// Define the components we need to start
		components := []Component{
			{
				Name:            "Podman",
				Command:         "podman",
				Args:            []string{"--version"},
				CheckCommand:    "podman",
				CheckArgs:       []string{"ps"},
				RequiredFor:     []string{"Kind", "Dapr", "Temporal"},
				StartupDelay:    0,
				IsRequired:      true,
				CommandExists:   isCommandAvailable("podman"),
				VerifyAvailable: checkPodmanRunning,
				RequiresStartup: false, // We don't start Podman, just verify it's running
				IsBinary:        true,
			},
			{
				Name:            "Kind",
				Command:         "kind",
				Args:            []string{"--version"},
				CheckCommand:    "kind",
				CheckArgs:       []string{"get", "clusters"},
				RequiredFor:     []string{"Dapr", "Temporal"},
				StartupDelay:    0,
				IsRequired:      true,
				CommandExists:   isCommandAvailable("kind"),
				VerifyAvailable: checkKindRunning,
				RequiresStartup: false, // We don't start Kind, just verify it's configured
				IsBinary:        true,
			},
			{
				Name:            "Dapr",
				Command:         "dapr",
				Args:            []string{"init", "--container-runtime", "podman"},
				CheckCommand:    "dapr",
				CheckArgs:       []string{"status"},
				RequiredFor:     []string{},
				StartupDelay:    2 * time.Second,
				IsRequired:      getDaprRequirement(configLoaded, config.Components.Dapr, skipDapr),
				CommandExists:   isCommandAvailable("dapr"),
				VerifyAvailable: checkDaprRunning,
				RequiresStartup: true, // Dapr needs to be started
				IsBinary:        false,
			},
			{
				Name:         "DaprDashboard",
				Command:      "dapr",
				Args:         []string{"dashboard", "-p", strconv.Itoa(config.Dapr.DashboardPort), "--address", "0.0.0.0"},
				CheckCommand: "dapr",
				CheckArgs:    []string{"dashboard", "--help"},
				RequiredFor:  []string{},
				StartupDelay: 1 * time.Second,
				IsRequired:   getDaprDashboardRequirement(configLoaded, config.Components.DaprDashboard, skipDaprDashboard),
				CommandExists: func() bool {
					// Check if dapr dashboard command is available
					cmd := exec.Command("dapr", "dashboard", "--help")
					err := cmd.Run()
					return err == nil
				}(),
				VerifyAvailable: func() bool {
					// Dashboard is available if the command exists
					return true
				},
				RequiresStartup: true, // We want the component system to handle it
				IsBinary:        false,
			},
			{
				Name:            "Temporal",
				Command:         "temporal",
				Args:            []string{"server", "start-dev"},
				CheckCommand:    "temporal",
				CheckArgs:       []string{"workflow", "list"},
				RequiredFor:     []string{},
				StartupDelay:    5 * time.Second, // Increased delay for Temporal to start fully
				IsRequired:      getTemporalRequirement(configLoaded, config.Components.Temporal, skipTemporal),
				CommandExists:   isCommandAvailable("temporal"),
				VerifyAvailable: checkTemporalServerRunning,
				RequiresStartup: true, // Temporal needs to be started
				IsBinary:        false,
			},
		}

		// First, check if required components are installed
		allInstalled := true
		for _, comp := range components {
			if comp.IsRequired && !comp.CommandExists {
				fmt.Printf("❌ Required component '%s' is not installed or not in PATH.\n", comp.Name)
				allInstalled = false
			} else if verbose {
				fmt.Printf("✅ Component '%s' is installed.\n", comp.Name)
			}
		}

		if !allInstalled {
			fmt.Println("\nSome required components are missing. Please install them and try again.")
			fmt.Println("Run 'devhelper-cli localenv init' to check required dependencies and create a configuration.")
			os.Exit(1)
		}

		// Next, check dependencies between components
		for i, comp := range components {
			if !comp.IsRequired {
				if verbose {
					fmt.Printf("⏭️  Skipping '%s' as it's not required.\n", comp.Name)
				}
				continue
			}

			// Check if any required components are missing
			missingDeps := false
			for _, dep := range comp.RequiredFor {
				for _, depComp := range components {
					if depComp.Name == dep && !depComp.CommandExists {
						fmt.Printf("❌ '%s' requires '%s', but it's not installed.\n", comp.Name, dep)
						missingDeps = true
					}
				}
			}

			if missingDeps {
				continue
			}

			// Handle components differently based on RequiresStartup
			if !comp.RequiresStartup {
				// For components like Podman and Kind, just check if they're running
				if comp.IsBinary {
					fmt.Printf("Checking if %s is available...\n", comp.Name)
					if comp.VerifyAvailable != nil && comp.VerifyAvailable() {
						if comp.Name == "Podman" {
							fmt.Printf("✅ %s is available and can run containers.\n", comp.Name)
						} else if comp.Name == "Kind" {
							fmt.Printf("✅ %s is available and has clusters configured.\n", comp.Name)
						} else {
							fmt.Printf("✅ %s is available.\n", comp.Name)
						}
						components[i].IsRunning = true
					} else {
						if comp.Name == "Podman" {
							fmt.Printf("❌ %s is not working properly.\n", comp.Name)
							fmt.Println("   Make sure Podman is installed correctly and has proper permissions.")
						} else if comp.Name == "Kind" {
							fmt.Printf("❌ %s does not have any clusters configured.\n", comp.Name)
							fmt.Println("   Note: Kubernetes functionality is not required for local development.")
						} else {
							fmt.Printf("❌ %s is not available.\n", comp.Name)
						}
						allInstalled = false
					}
				} else {
					// For non-binary components that still don't require startup
					fmt.Printf("Verifying %s is ready...\n", comp.Name)
					if comp.VerifyAvailable != nil && comp.VerifyAvailable() {
						fmt.Printf("✅ %s is running and ready.\n", comp.Name)
						components[i].IsRunning = true
					} else {
						fmt.Printf("❌ %s is not running or not ready.\n", comp.Name)
						allInstalled = false
					}
				}
				continue
			}

			// For components that need to be started (Dapr, Temporal)
			fmt.Printf("Starting %s...\n", comp.Name)

			// Special handling for Dapr - check if it's already initialized
			if comp.Name == "Dapr" {
				// First check if Dapr is already running
				if checkDaprRunning() {
					fmt.Println("✅ Dapr is already running, skipping initialization.")
					components[i].IsRunning = true
					continue
				}

				// Run dapr init with Podman container runtime
				initCmd := exec.Command("dapr", "init", "--container-runtime", "podman")
				initOutput, err := initCmd.CombinedOutput()
				if err != nil {
					fmt.Printf("❌ Failed to initialize Dapr: %v\n", err)
					if verbose {
						fmt.Printf("Output: %s\n", string(initOutput))
					}
					continue
				}

				if verbose {
					fmt.Printf("Dapr initialization output: %s\n", string(initOutput))
				}

				// Wait a moment for Dapr to start
				time.Sleep(2 * time.Second)

				// Verify Dapr is running
				if checkDaprRunning() {
					fmt.Println("✅ Dapr started successfully.")
					components[i].IsRunning = true
				} else {
					fmt.Println("⚠️ Dapr initialization completed, but the runtime may not be fully ready.")
					components[i].IsRunning = true // Consider it running anyway
				}

				continue
			}

			// Special handling for Dapr Dashboard
			if comp.Name == "DaprDashboard" {
				// First check if the desired port is already in use
				dashboardPort := config.Dapr.DashboardPort

				// Get dashboard PID if it's running
				dashboardPID := getDaprDashboardPID()

				// Determine if a restart is required
				restartRequired := forceRestart ||
					(configChanged && dashboardPID != "" &&
						(configCache.DaprDashboardPort != dashboardPort))

				// More robust process termination and port cleanup
				if dashboardPID != "" {
					if restartRequired {
						fmt.Printf("Detected configuration change: Dashboard port changed (%d → %d)\n",
							configCache.DaprDashboardPort, dashboardPort)
						fmt.Println("Stopping existing Dapr Dashboard...")

						// First try graceful termination with SIGTERM
						killCmd := exec.Command("kill", dashboardPID)
						if err := killCmd.Run(); err != nil {
							fmt.Printf("Warning: Failed to stop Dapr Dashboard gracefully: %v\n", err)

							// If graceful termination fails, try force kill (SIGKILL)
							forceKillCmd := exec.Command("kill", "-9", dashboardPID)
							if err := forceKillCmd.Run(); err != nil {
								fmt.Printf("Error: Failed to force kill Dapr Dashboard: %v\n", err)
							}
						}

						// Give more time for the process to fully terminate
						time.Sleep(2 * time.Second)

						// Check if port is still in use by anything
						stillInUse := isPortInUse(dashboardPort)
						if stillInUse {
							// Try to find any process using this port and kill it
							cmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", dashboardPort), "-t")
							output, err := cmd.Output()
							if err == nil && len(output) > 0 {
								pids := strings.Split(strings.TrimSpace(string(output)), "\n")
								for _, pid := range pids {
									fmt.Printf("Forcefully terminating process %s that is still using port %d\n", pid, dashboardPort)
									exec.Command("kill", "-9", pid).Run()
								}
								time.Sleep(1 * time.Second)
							}
						}

						// Final verification
						stillInUse = isPortInUse(dashboardPort)
						if stillInUse {
							fmt.Printf("❌ Port %d is still in use after attempts to free it\n", dashboardPort)
							fmt.Printf("   Try a different port or manually kill the process: lsof -i :%d -t | xargs kill -9\n", dashboardPort)
							fmt.Println("   Updating localenv.yaml with a new port is recommended.")
							fmt.Printf("   For example: dapr.dashboardPort: %d\n", dashboardPort+1)
							continue
						}
					} else {
						// Dashboard is already running with current configuration
						dashboardURL := fmt.Sprintf("http://localhost:%d", dashboardPort)
						fmt.Printf("✅ Dapr Dashboard already running at %s\n", dashboardURL)
						components[i].IsRunning = true
						continue
					}
				}

				// Check if the port is in use by something else
				if isPortInUse(dashboardPort) {
					fmt.Printf("❌ Port %d is already in use by another process\n", dashboardPort)
					fmt.Printf("   Run 'lsof -i :%d' to see which process is using it\n", dashboardPort)
					fmt.Println("   Update the dashboardPort in localenv.yaml to a different value and try again.")
					fmt.Printf("   For example: dapr.dashboardPort: %d\n", dashboardPort+1)
					continue
				}

				// For Dapr Dashboard, we need special handling to make sure it stays running
				fmt.Println("Starting DaprDashboard in background mode...")

				// Start the dashboard
				dashboardStarted := tryStartDashboard(comp.Command, dashboardPort, nil)

				if dashboardStarted {
					components[i].IsRunning = true
					dashboardURL := fmt.Sprintf("http://localhost:%d", dashboardPort)
					fmt.Printf("✅ Dapr Dashboard started at %s\n", dashboardURL)
				} else {
					fmt.Printf("❌ Failed to start Dapr Dashboard on port %d\n", dashboardPort)
					fmt.Println("   This could be because the port is already in use.")
					fmt.Printf("   You can check which process is using the port with: lsof -i :%d\n", dashboardPort)
					fmt.Println("   Update the dashboardPort in localenv.yaml to a different value and try again.")
					fmt.Printf("   For example: dapr.dashboardPort: %d\n", dashboardPort+1)
				}
				continue
			}

			// Special handling for Temporal server
			if comp.Name == "Temporal" {
				// Get Temporal configuration
				temporalUIPort := config.Temporal.UIPort
				temporalGRPCPort := config.Temporal.GRPCPort
				temporalNamespace := config.Temporal.Namespace

				// Check if Temporal is already running and if there are config changes
				temporalRunning := checkTemporalServerRunning()

				if temporalRunning {
					// Temporal is already running
					restartRequired := forceRestart || (configChanged &&
						(configCache.TemporalUIPort != temporalUIPort ||
							configCache.TemporalGRPCPort != temporalGRPCPort ||
							configCache.TemporalNamespace != temporalNamespace))

					if !restartRequired {
						fmt.Println("✅ Temporal is already running with current configuration, skipping startup.")
						components[i].IsRunning = true
						continue
					}

					// If we need to restart, kill any existing Temporal server process
					fmt.Println("Detected configuration changes in Temporal settings:")
					for _, change := range changes {
						if strings.Contains(change, "Temporal") {
							fmt.Printf("- %s\n", change)
						}
					}

					fmt.Println("Stopping existing Temporal server...")
					// Find and kill the Temporal process
					found := false
					// First look for the main temporal server process
					cmd := exec.Command("ps", "-ef")
					output, err := cmd.CombinedOutput()
					if err == nil {
						outputLines := strings.Split(string(output), "\n")
						for _, line := range outputLines {
							if strings.Contains(line, "temporal server start-dev") && !strings.Contains(line, "grep") {
								fields := strings.Fields(line)
								if len(fields) >= 2 {
									pid := fields[1]
									fmt.Printf("Stopping Temporal server process (PID: %s)...\n", pid)
									killCmd := exec.Command("kill", pid)
									killCmd.Run()
									found = true
								}
							}
						}
					}

					// Also check for any processes on the Temporal ports
					if !found || isPortInUse(temporalUIPort) || isPortInUse(temporalGRPCPort) {
						fmt.Println("Looking for processes using Temporal ports...")

						// Check UI port
						uiPortCmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", temporalUIPort), "-t")
						uiPortOutput, _ := uiPortCmd.Output()
						if len(uiPortOutput) > 0 {
							pids := strings.Split(strings.TrimSpace(string(uiPortOutput)), "\n")
							for _, pid := range pids {
								fmt.Printf("Forcefully terminating process %s using Temporal UI port %d\n", pid, temporalUIPort)
								exec.Command("kill", "-9", pid).Run()
							}
						}

						// Check GRPC port
						grpcPortCmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", temporalGRPCPort), "-t")
						grpcPortOutput, _ := grpcPortCmd.Output()
						if len(grpcPortOutput) > 0 {
							pids := strings.Split(strings.TrimSpace(string(grpcPortOutput)), "\n")
							for _, pid := range pids {
								fmt.Printf("Forcefully terminating process %s using Temporal GRPC port %d\n", pid, temporalGRPCPort)
								exec.Command("kill", "-9", pid).Run()
							}
						}
					}

					// Give more time for processes to fully terminate
					time.Sleep(3 * time.Second)

					// Final verification
					if isPortInUse(temporalUIPort) {
						fmt.Printf("❌ Temporal UI port %d is still in use after attempts to free it\n", temporalUIPort)
						fmt.Printf("   Try manually killing the process: lsof -i :%d -t | xargs kill -9\n", temporalUIPort)
						continue
					}

					if isPortInUse(temporalGRPCPort) {
						fmt.Printf("❌ Temporal GRPC port %d is still in use after attempts to free it\n", temporalGRPCPort)
						fmt.Printf("   Try manually killing the process: lsof -i :%d -t | xargs kill -9\n", temporalGRPCPort)
						continue
					}
				} else {
					// Temporal is not running, check if ports are available
					if isPortInUse(temporalUIPort) {
						fmt.Printf("❌ Temporal UI port %d is already in use by another process\n", temporalUIPort)
						fmt.Printf("   Run 'lsof -i :%d' to see which process is using it\n", temporalUIPort)
						fmt.Println("   Update the UIPort in localenv.yaml to a different value and try again.")
						continue
					}

					if isPortInUse(temporalGRPCPort) {
						fmt.Printf("❌ Temporal GRPC port %d is already in use by another process\n", temporalGRPCPort)
						fmt.Printf("   Run 'lsof -i :%d' to see which process is using it\n", temporalGRPCPort)
						fmt.Println("   Update the GRPCPort in localenv.yaml to a different value and try again.")
						continue
					}
				}

				// Start Temporal server in background
				fmt.Println("Starting Temporal server in background mode...")

				// Prepare command with namespace flag if configured
				var temporalCmd *exec.Cmd
				if configLoaded && config.Components.Temporal && temporalNamespace != "" && temporalNamespace != "default" {
					fmt.Printf("Configuring Temporal with namespace: %s\n", temporalNamespace)
					// Use a more efficient approach - create the namespace first if needed, then start server
					// This avoids potential issues with the namespace not being created properly during startup

					// First check if the namespace exists
					namespaceCheckCmd := exec.Command("temporal", "operator", "namespace", "describe", temporalNamespace)
					if err := namespaceCheckCmd.Run(); err != nil {
						// Namespace doesn't exist, create it first
						fmt.Printf("Creating Temporal namespace '%s'...\n", temporalNamespace)
						createCmd := exec.Command("temporal", "operator", "namespace", "create", temporalNamespace)
						if output, err := createCmd.CombinedOutput(); err != nil {
							fmt.Printf("❌ Failed to create namespace: %v\n", err)
							if verbose {
								fmt.Printf("Output: %s\n", string(output))
							}
						} else {
							fmt.Printf("✅ Created Temporal namespace '%s'\n", temporalNamespace)
						}
					} else {
						fmt.Printf("✅ Temporal namespace '%s' already exists\n", temporalNamespace)
					}

					// Start the server normally
					temporalCmd = exec.Command("temporal", "server", "start-dev")
				} else {
					temporalCmd = exec.Command("temporal", "server", "start-dev")
				}

				// Create logs directory if it doesn't exist
				logsDir := filepath.Join(os.Getenv("HOME"), ".logs", "devhelper-cli")
				if _, err := os.Stat(logsDir); os.IsNotExist(err) {
					os.MkdirAll(logsDir, 0755)
				}

				logFilePath := filepath.Join(logsDir, "temporal-server.log")

				// Configure logs based on stream-logs flag
				if streamLogs {
					// In streaming mode, we'll use a MultiWriter to write to both terminal and file
					logFile, err := os.OpenFile(
						logFilePath,
						os.O_CREATE|os.O_WRONLY|os.O_APPEND,
						0644,
					)

					if err == nil {
						defer logFile.Close()

						// Create a MultiWriter that sends output to both the terminal and log file
						multiWriter := io.MultiWriter(os.Stdout, logFile)
						temporalCmd.Stdout = multiWriter
						temporalCmd.Stderr = multiWriter

						fmt.Println("📃 Streaming Temporal server logs to terminal and writing to log file...")
						fmt.Printf("📂 Log file: %s\n", logFilePath)
					} else {
						// Fallback to just terminal if can't create log file
						fmt.Printf("⚠️ Warning: Could not create log file: %v\n", err)
						fmt.Println("📃 Streaming Temporal server logs to terminal only...")
						temporalCmd.Stdout = os.Stdout
						temporalCmd.Stderr = os.Stderr
					}
				} else {
					// Standard non-streaming mode, just write to log file
					logFile, err := os.OpenFile(
						logFilePath,
						os.O_CREATE|os.O_WRONLY|os.O_APPEND,
						0644,
					)

					if err == nil {
						defer logFile.Close()
						temporalCmd.Stdout = logFile
						temporalCmd.Stderr = logFile
						fmt.Printf("📂 Temporal server logs will be written to %s\n", logFilePath)
						fmt.Println("💡 Use --stream-logs flag to see logs in terminal")
					} else {
						// Fallback to null device if can't create log file
						fmt.Printf("⚠️ Warning: Could not create log file: %v\n", err)
						devNull, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
						temporalCmd.Stdout = devNull
						temporalCmd.Stderr = devNull
					}
				}

				// Start Temporal server in background
				if err := temporalCmd.Start(); err != nil {
					fmt.Printf("❌ Failed to start Temporal server: %v\n", err)
					continue
				}

				// Wait for Temporal to start up
				fmt.Println("⏳ Waiting for Temporal server to start...")
				time.Sleep(5 * time.Second)

				// Verify Temporal is running with increased retries and timeout
				retries := 5                  // Increased from 3
				retryDelay := 3 * time.Second // Increased from 2
				temporalStarted := false

				for retry := 0; retry < retries; retry++ {
					if checkTemporalServerRunning() {
						fmt.Println("✅ Temporal server started successfully.")
						components[i].IsRunning = true
						temporalStarted = true
						break
					}

					if retry < retries-1 {
						fmt.Println("Waiting for Temporal server to become available...")
						time.Sleep(retryDelay)
					} else {
						fmt.Println("❌ Temporal server did not start properly.")
						fmt.Println("   Check the logs at " + filepath.Join(logsDir, "temporal-server.log") + " for details.")
					}
				}

				if !temporalStarted {
					continue
				}
			}
		}

		// Check if all components are running
		for _, comp := range components {
			if comp.IsRequired && !comp.IsRunning {
				fmt.Printf("❌ %s is not running. Please check its logs for errors.\n", comp.Name)
				allInstalled = false
			}
		}

		if !allInstalled {
			fmt.Println("\nSome components failed to start. Please check the logs for errors.")
			os.Exit(1)
		}

		// Save current configuration for future comparison
		saveConfigCache(newCache)

		fmt.Println("\nAll required components are running successfully!")
	},
}

func init() {
	localenvCmd.AddCommand(startCmd)
	startCmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	startCmd.Flags().Bool("skip-dapr", false, "Skip starting Dapr")
	startCmd.Flags().Bool("skip-temporal", false, "Skip starting Temporal")
	startCmd.Flags().Bool("skip-dapr-dashboard", false, "Skip starting Dapr Dashboard")
	startCmd.Flags().Bool("force-restart", false, "Force restart of components even if already running")
	startCmd.Flags().StringP("config", "c", "", "Path to localenv configuration file")
	startCmd.Flags().Bool("stream-logs", false, "Stream Temporal server logs to terminal")
}
