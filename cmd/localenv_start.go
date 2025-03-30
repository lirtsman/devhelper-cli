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
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	yamlv3 "gopkg.in/yaml.v3"
)

// Configuration cache to detect changes between runs
type ConfigCache struct {
	DaprDashboardPort  int
	TemporalUIPort     int
	TemporalGRPCPort   int
	TemporalNamespace  string
	OpenSearchPort     int
	OpenSearchDashPort int
}

// Returns true if current config differs from previous state
func hasConfigChanged(config LocalEnvConfig, currentCache ConfigCache) (bool, ConfigCache, []string) {
	changes := []string{}
	newCache := ConfigCache{
		DaprDashboardPort:  config.Dapr.DashboardPort,
		TemporalUIPort:     config.Temporal.UIPort,
		TemporalGRPCPort:   config.Temporal.GRPCPort,
		TemporalNamespace:  config.Temporal.Namespace,
		OpenSearchPort:     config.OpenSearch.Port,
		OpenSearchDashPort: config.OpenSearch.DashboardPort,
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

	// Check for OpenSearch port change
	if currentCache.OpenSearchPort != 0 &&
		currentCache.OpenSearchPort != config.OpenSearch.Port {
		changes = append(changes, fmt.Sprintf("OpenSearch port changed: %d → %d",
			currentCache.OpenSearchPort, config.OpenSearch.Port))
		hasChanges = true
	}

	// Check for OpenSearch Dashboard port change
	if currentCache.OpenSearchDashPort != 0 &&
		currentCache.OpenSearchDashPort != config.OpenSearch.DashboardPort {
		changes = append(changes, fmt.Sprintf("OpenSearch Dashboard port changed: %d → %d",
			currentCache.OpenSearchDashPort, config.OpenSearch.DashboardPort))
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
			yamlv3.Unmarshal(data, &cache)
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

	data, err := yamlv3.Marshal(cache)
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

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start local development environment",
	Long: `Start a local development environment with all necessary components
for Shield application development including:

- Dapr runtime
- Temporal server
- OpenSearch (for search and analytics)
- Required dependencies

This command will check for necessary dependencies and start them
in the correct order.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting local development environment...")

		verbose, _ := cmd.Flags().GetBool("verbose")
		skipDapr, _ := cmd.Flags().GetBool("skip-dapr")
		skipTemporal, _ := cmd.Flags().GetBool("skip-temporal")
		skipDaprDashboard, _ := cmd.Flags().GetBool("skip-dapr-dashboard")
		skipOpenSearch, _ := cmd.Flags().GetBool("skip-opensearch")
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
				err = yamlv3.Unmarshal(configData, &config)
				if err == nil {
					configLoaded = true
					fmt.Printf("✅ Loaded configuration from %s\n", configPath)

					// Check if OpenSearch config is missing and add it if needed
					openSearchMissing := config.OpenSearch.Port == 0 &&
						config.OpenSearch.DashboardPort == 0

					// Check if Podman is available (required for OpenSearch)
					podmanAvailable := isCommandAvailable("podman")

					if openSearchMissing && podmanAvailable {
						fmt.Println("ℹ️ Adding default OpenSearch configuration to localenv.yaml")

						// Enable OpenSearch component
						config.Components.OpenSearch = true

						// Set default OpenSearch configuration
						config.OpenSearch.Port = 9200
						config.OpenSearch.DashboardPort = 5601

						// Find Podman path
						podmanPath, err := exec.LookPath("podman")
						if err == nil {
							config.Paths.Podman = podmanPath
						}

						// Update the config file
						updatedConfigData, err := yamlv3.Marshal(config)
						if err == nil {
							err = os.WriteFile(configPath, updatedConfigData, 0644)
							if err == nil {
								fmt.Println("✅ Updated localenv.yaml with OpenSearch configuration")
							} else if verbose {
								fmt.Printf("⚠️ Failed to update configuration file: %v\n", err)
							}
						} else if verbose {
							fmt.Printf("⚠️ Failed to marshal updated configuration: %v\n", err)
						}
					}
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
		if skipOpenSearch {
			config.Components.OpenSearch = false
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

		// Function to check if OpenSearch is running
		checkOpenSearchRunning := func() bool {
			// Check if container is running
			checkCmd := exec.Command("podman", "ps", "--filter", "name=opensearch-node", "--format", "{{.Names}}")
			output, err := checkCmd.CombinedOutput()
			if err != nil || !strings.Contains(string(output), "opensearch-node") {
				if verbose {
					fmt.Printf("OpenSearch container check failed: %v\n", err)
					if len(output) > 0 {
						fmt.Printf("Output: %s\n", string(output))
					}

					// Try to get logs from the container to help with debugging
					logsCmd := exec.Command("podman", "logs", "opensearch-node")
					logsOutput, logsErr := logsCmd.CombinedOutput()
					if logsErr == nil && len(logsOutput) > 0 {
						fmt.Println("\nOpenSearch container logs:")
						fmt.Println(string(logsOutput))
					} else {
						fmt.Println("\nUnable to retrieve OpenSearch container logs")
					}
				}
				return false
			}

			// Add retry logic for OpenSearch connection
			maxRetries := 5
			retryDelay := 5 * time.Second

			// Use HTTP for dev environment
			url := fmt.Sprintf("http://localhost:%d/_cluster/health", config.OpenSearch.Port)

			for i := 0; i < maxRetries; i++ {
				if i > 0 {
					fmt.Printf("Retrying OpenSearch connection (%d/%d)...\n", i+1, maxRetries)
					time.Sleep(retryDelay)
				}

				client := http.Client{
					Timeout: 10 * time.Second, // Increased timeout for OpenSearch to respond
				}

				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					if verbose {
						fmt.Printf("OpenSearch request creation failed: %v\n", err)
					}
					continue
				}

				resp, err := client.Do(req)
				if err != nil {
					if verbose {
						fmt.Printf("OpenSearch health check failed: %v\n", err)
					}
					continue
				}

				// Read and log response for debugging
				if verbose {
					body, _ := io.ReadAll(resp.Body)
					fmt.Printf("OpenSearch response (status %d): %s\n", resp.StatusCode, string(body))
				}

				resp.Body.Close()

				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					fmt.Println("✅ Successfully connected to OpenSearch")
					return true
				}
			}

			fmt.Println("❌ Failed to connect to OpenSearch after multiple attempts")
			return false
		}

		// Function to check if OpenSearch Dashboard is running
		checkOpenSearchDashboardRunning := func() bool {
			// Check if container is running
			checkCmd := exec.Command("podman", "ps", "--filter", "name=opensearch-dashboard", "--format", "{{.Names}}")
			output, err := checkCmd.CombinedOutput()
			if err != nil || !strings.Contains(string(output), "opensearch-dashboard") {
				return false
			}

			// Dashboard startup can take longer, so let it start
			time.Sleep(10 * time.Second) // Increased from 5 to 10 seconds
			fmt.Println("⏳ Waiting for OpenSearch Dashboard to initialize...")

			// Add retry logic for Dashboard connection - may take some time to initialize
			maxRetries := 15 // Increased retries from 10 to 15
			retryDelay := 5 * time.Second

			// Check dashboard availability
			url := fmt.Sprintf("http://localhost:%d", config.OpenSearch.DashboardPort)

			for i := 0; i < maxRetries; i++ {
				if i > 0 {
					fmt.Printf("Checking OpenSearch Dashboard connection (%d/%d)...\n", i+1, maxRetries)
				}

				client := http.Client{
					Timeout: 10 * time.Second, // Increased timeout
				}

				// Create a request with basic auth
				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					if verbose {
						fmt.Printf("Failed to create request for Dashboard: %v\n", err)
					}
					time.Sleep(retryDelay)
					continue
				}

				// Send the request
				resp, err := client.Do(req)
				if err != nil {
					if verbose {
						fmt.Printf("OpenSearch Dashboard check failed: %v\n", err)
					}
					time.Sleep(retryDelay)
					continue
				}

				// Just checking if we get any response is enough
				resp.Body.Close()

				// Accept any status code as long as we get a response
				fmt.Println("✅ OpenSearch Dashboard is accessible")
				return true
			}

			fmt.Println("❌ OpenSearch Dashboard is not yet responding after multiple attempts")
			fmt.Println("   The Dashboard container is running but may still be initializing.")
			fmt.Println("   This is normal, especially on first startup. It may take up to 2-3 minutes.")
			fmt.Println("   You can check the logs with: podman logs opensearch-dashboard")

			if verbose {
				// Show logs to help diagnose issues
				logsCmd := exec.Command("podman", "logs", "opensearch-dashboard")
				logsOutput, _ := logsCmd.CombinedOutput()
				if len(logsOutput) > 0 {
					fmt.Println("\nOpenSearch Dashboard container logs:")
					fmt.Println(string(logsOutput))
				}
			}

			// Return true anyway to avoid blocking the startup flow, as the container is running
			// The dashboard is likely just slow to initialize
			return true
		}

		// Define the components we need to start
		components := []Component{
			{
				Name:            "Podman",
				Command:         "podman",
				Args:            []string{"--version"},
				CheckCommand:    "podman",
				CheckArgs:       []string{"ps"},
				RequiredFor:     []string{"Kind", "Dapr", "Temporal", "OpenSearch"},
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
			{
				Name:    "OpenSearch",
				Command: "podman",
				Args: []string{
					"run",
					"-d",
					"--name", "opensearch-node",
					"-p", fmt.Sprintf("%d:9200", config.OpenSearch.Port),
					"-e", "cluster.name=devhelper-cluster",
					"-e", "node.name=opensearch-node",
					"-e", "discovery.type=single-node",
					"-e", "DISABLE_SECURITY_PLUGIN=true",
					"-e", "DISABLE_INSTALL_DEMO_CONFIG=true",
					"--health-cmd", fmt.Sprintf("curl -u %s:%s -f http://localhost:9200/_cluster/health || exit 1", "admin", "admin"),
					"--health-interval", "30s",
					"--health-timeout", "10s",
					"--health-retries", "5",
					"--network", "opensearch-network",
					"--restart", "unless-stopped",
					"opensearchproject/opensearch:2.17.1",
				},
				CheckCommand:    "podman",
				CheckArgs:       []string{"ps", "--filter", "name=opensearch-node", "--format", "{{.Names}}"},
				RequiredFor:     []string{"OpenSearchDashboard"},
				StartupDelay:    30 * time.Second, // OpenSearch needs more time to initialize
				IsRequired:      getOpenSearchRequirement(configLoaded, config.Components.OpenSearch, skipOpenSearch),
				CommandExists:   isCommandAvailable("podman"),
				VerifyAvailable: checkOpenSearchRunning,
				RequiresStartup: true,
				IsBinary:        false,
			},
			{
				Name:    "OpenSearchDashboard",
				Command: "podman",
				Args: []string{
					"run",
					"-d",
					"--name", "opensearch-dashboard",
					"-p", fmt.Sprintf("%d:5601", config.OpenSearch.DashboardPort),
					"-e", "OPENSEARCH_HOSTS=[\"http://opensearch-node:9200\"]",
					"-e", "DISABLE_SECURITY_DASHBOARDS_PLUGIN=true",
					"--network", "opensearch-network",
					"--restart", "unless-stopped",
					"opensearchproject/opensearch-dashboards:2.17.1",
				},
				CheckCommand:    "podman",
				CheckArgs:       []string{"ps", "--filter", "name=opensearch-dashboard", "--format", "{{.Names}}"},
				RequiredFor:     []string{},
				StartupDelay:    15 * time.Second,
				IsRequired:      getOpenSearchRequirement(configLoaded, config.Components.OpenSearch, skipOpenSearch),
				CommandExists:   isCommandAvailable("podman"),
				VerifyAvailable: checkOpenSearchDashboardRunning,
				RequiresStartup: true,
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
							fmt.Printf("✅ %s is available and can create clusters.\n", comp.Name)
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
				temporalNamespaceToCreate := ""
				if configLoaded && config.Components.Temporal && temporalNamespace != "" && temporalNamespace != "default" {
					fmt.Printf("Configuring Temporal with namespace: %s\n", temporalNamespace)
					// Store the namespace name for creation after server starts
					temporalNamespaceToCreate = temporalNamespace
				}

				// Start the server normally
				temporalCmd = exec.Command("temporal", "server", "start-dev")

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

						// Create the custom namespace if needed, now that the server is running
						if temporalNamespaceToCreate != "" {
							// First check if the namespace exists
							namespaceCheckCmd := exec.Command("temporal", "operator", "namespace", "describe", "--namespace", temporalNamespaceToCreate)
							if err := namespaceCheckCmd.Run(); err != nil {
								// Namespace doesn't exist, create it now
								fmt.Printf("Creating Temporal namespace '%s'...\n", temporalNamespaceToCreate)
								createCmd := exec.Command("temporal", "operator", "namespace", "create", "--namespace", temporalNamespaceToCreate)
								if output, err := createCmd.CombinedOutput(); err != nil {
									fmt.Printf("❌ Failed to create namespace: %v\n", err)
									if verbose {
										fmt.Printf("Output: %s\n", string(output))
									}
								} else {
									fmt.Printf("✅ Created Temporal namespace '%s'\n", temporalNamespaceToCreate)
								}
							} else {
								fmt.Printf("✅ Temporal namespace '%s' already exists\n", temporalNamespaceToCreate)
							}
						}

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

			// Special handling for OpenSearch
			if comp.Name == "OpenSearch" {
				fmt.Println("Starting OpenSearch...")

				// Ensure the network exists
				networkCmd := exec.Command("podman", "network", "create", "opensearch-network")
				networkCmd.Run() // Ignore errors, network may already exist

				// Check if a container with the same name already exists and remove it
				checkExistingCmd := exec.Command("podman", "ps", "-a", "--filter", "name=opensearch-node", "--format", "{{.Names}}")
				existingOutput, _ := checkExistingCmd.CombinedOutput()
				if strings.Contains(string(existingOutput), "opensearch-node") {
					fmt.Println("Found existing OpenSearch container, removing it...")
					removeCmd := exec.Command("podman", "rm", "-f", "opensearch-node")
					removeOutput, removeErr := removeCmd.CombinedOutput()
					if removeErr != nil {
						fmt.Printf("❌ Failed to remove existing OpenSearch container: %v\n", removeErr)
						if verbose {
							fmt.Printf("Output: %s\n", string(removeOutput))
						}
						continue
					}
				}

				// Run the podman command to start OpenSearch
				if verbose {
					fmt.Println("Executing command: podman", strings.Join(comp.Args, " "))
				}

				startCmd := exec.Command(comp.Command, comp.Args...)
				startOutput, startErr := startCmd.CombinedOutput()

				if startErr != nil {
					fmt.Printf("❌ Failed to start OpenSearch: %v\n", startErr)
					if verbose || strings.Contains(string(startOutput), "Error:") {
						fmt.Printf("Output: %s\n", string(startOutput))
					}
					continue
				}

				// Wait for the container to start
				fmt.Println("⏳ Waiting for OpenSearch container to start...")
				time.Sleep(5 * time.Second)

				// Check if container is running
				containerRunning := false
				for i := 0; i < 5; i++ {
					checkCmd := exec.Command("podman", "ps", "--filter", "name=opensearch-node", "--format", "{{.Names}}")
					output, err := checkCmd.CombinedOutput()
					if err == nil && strings.Contains(string(output), "opensearch-node") {
						containerRunning = true
						break
					}

					if i < 4 {
						fmt.Println("Waiting for container to start...")
						time.Sleep(2 * time.Second)
					}
				}

				if !containerRunning {
					fmt.Println("❌ OpenSearch container failed to start")
					if verbose {
						logsCmd := exec.Command("podman", "logs", "opensearch-node")
						logsOutput, _ := logsCmd.CombinedOutput()
						if len(logsOutput) > 0 {
							fmt.Println("\nOpenSearch container logs:")
							fmt.Println(string(logsOutput))
						}
					}
					continue
				}

				// Now wait for OpenSearch service to be ready
				fmt.Println("⏳ Waiting for OpenSearch service to be ready...")
				serviceReady := false

				// Add retry logic for OpenSearch connection
				maxRetries := 10 // Increased retries
				retryDelay := 5 * time.Second

				// Use HTTP for dev environment
				url := fmt.Sprintf("http://localhost:%d/_cluster/health", config.OpenSearch.Port)

				for i := 0; i < maxRetries; i++ {
					if i > 0 {
						fmt.Printf("Checking OpenSearch connection (%d/%d)...\n", i+1, maxRetries)
					}

					client := http.Client{
						Timeout: 10 * time.Second, // Increased timeout for OpenSearch to respond
					}

					req, err := http.NewRequest("GET", url, nil)
					if err != nil {
						if verbose {
							fmt.Printf("OpenSearch request creation failed: %v\n", err)
						}
						time.Sleep(retryDelay)
						continue
					}

					// Use credentials from config instead of hardcoded values
					resp, err := client.Do(req)
					if err != nil {
						if verbose {
							fmt.Printf("OpenSearch health check failed: %v\n", err)
						}
						time.Sleep(retryDelay)
						continue
					}

					// Read and log response for debugging
					if verbose {
						body, _ := io.ReadAll(resp.Body)
						fmt.Printf("OpenSearch response (status %d): %s\n", resp.StatusCode, string(body))
					}

					resp.Body.Close()

					if resp.StatusCode >= 200 && resp.StatusCode < 300 {
						fmt.Println("✅ OpenSearch is running and ready")
						serviceReady = true
						break
					}

					time.Sleep(retryDelay)
				}

				if serviceReady {
					components[i].IsRunning = true
				} else {
					fmt.Println("❌ OpenSearch is not running. Please check its logs for errors.")
					if verbose {
						logsCmd := exec.Command("podman", "logs", "opensearch-node")
						logsOutput, _ := logsCmd.CombinedOutput()
						if len(logsOutput) > 0 {
							fmt.Println("\nOpenSearch container logs:")
							fmt.Println(string(logsOutput))
						}
					}
				}

				continue
			}

			// Special handling for OpenSearch Dashboard
			if comp.Name == "OpenSearchDashboard" {
				fmt.Println("Starting OpenSearchDashboard...")

				// First check if OpenSearch is running as the Dashboard depends on it
				checkOpenSearchCmd := exec.Command("podman", "ps", "--filter", "name=opensearch-node", "--format", "{{.Names}}")
				osOutput, osErr := checkOpenSearchCmd.CombinedOutput()
				if osErr != nil || !strings.Contains(string(osOutput), "opensearch-node") {
					fmt.Println("❌ OpenSearch is not running. Dashboard cannot start without OpenSearch.")
					continue
				}

				// Check if a container with the same name already exists and remove it
				checkExistingCmd := exec.Command("podman", "ps", "-a", "--filter", "name=opensearch-dashboard", "--format", "{{.Names}}")
				existingOutput, _ := checkExistingCmd.CombinedOutput()
				if strings.Contains(string(existingOutput), "opensearch-dashboard") {
					fmt.Println("Found existing OpenSearch Dashboard container, removing it...")
					removeCmd := exec.Command("podman", "rm", "-f", "opensearch-dashboard")
					removeOutput, removeErr := removeCmd.CombinedOutput()
					if removeErr != nil {
						fmt.Printf("❌ Failed to remove existing OpenSearch Dashboard container: %v\n", removeErr)
						if verbose {
							fmt.Printf("Output: %s\n", string(removeOutput))
						}
						continue
					}
				}

				// Run the podman command to start OpenSearch Dashboard
				if verbose {
					fmt.Println("Executing command: podman", strings.Join(comp.Args, " "))
				}

				startCmd := exec.Command(comp.Command, comp.Args...)
				startOutput, startErr := startCmd.CombinedOutput()

				if startErr != nil {
					fmt.Printf("❌ Failed to start OpenSearch Dashboard: %v\n", startErr)
					if verbose || strings.Contains(string(startOutput), "Error:") {
						fmt.Printf("Output: %s\n", string(startOutput))
					}
					continue
				}

				// Wait for the container to start
				fmt.Println("⏳ Waiting for OpenSearch Dashboard container to start...")
				time.Sleep(5 * time.Second)

				// Check if container is running
				containerRunning := false
				for i := 0; i < 5; i++ {
					checkCmd := exec.Command("podman", "ps", "--filter", "name=opensearch-dashboard", "--format", "{{.Names}}")
					output, err := checkCmd.CombinedOutput()
					if err == nil && strings.Contains(string(output), "opensearch-dashboard") {
						containerRunning = true
						break
					}

					if i < 4 {
						fmt.Println("Waiting for Dashboard container to start...")
						time.Sleep(2 * time.Second)
					}
				}

				if !containerRunning {
					fmt.Println("❌ OpenSearch Dashboard container failed to start")
					if verbose {
						logsCmd := exec.Command("podman", "logs", "opensearch-dashboard")
						logsOutput, _ := logsCmd.CombinedOutput()
						if len(logsOutput) > 0 {
							fmt.Println("\nOpenSearch Dashboard container logs:")
							fmt.Println(string(logsOutput))
						}
					}
					continue
				}

				// Give the Dashboard more time to initialize
				fmt.Println("⏳ Waiting for OpenSearch Dashboard to initialize...")
				// Dashboard needs more time to initialize than just the container start
				time.Sleep(10 * time.Second)

				// Now check if the Dashboard is accessible
				dashboardReady := false
				maxRetries := 12 // More retries for dashboard
				retryDelay := 5 * time.Second

				// Check dashboard URL
				url := fmt.Sprintf("http://localhost:%d", config.OpenSearch.DashboardPort)

				for i := 0; i < maxRetries; i++ {
					if i > 0 {
						fmt.Printf("Checking OpenSearch Dashboard connection (%d/%d)...\n", i+1, maxRetries)
					}

					client := http.Client{
						Timeout: 10 * time.Second,
					}

					// Create a request with basic auth
					req, err := http.NewRequest("GET", url, nil)
					if err != nil {
						if verbose {
							fmt.Printf("Failed to create request for Dashboard: %v\n", err)
						}
						time.Sleep(retryDelay)
						continue
					}

					// Send the request
					resp, err := client.Do(req)
					if err != nil {
						if verbose {
							fmt.Printf("OpenSearch Dashboard check failed: %v\n", err)
						}
						time.Sleep(retryDelay)
						continue
					}

					// Just need to close the body
					resp.Body.Close()

					// Any response (even 404) is ok as it means the server is up
					fmt.Println("✅ OpenSearch Dashboard is accessible")
					dashboardReady = true
					break
				}

				if dashboardReady {
					components[i].IsRunning = true
				} else {
					fmt.Println("❌ OpenSearch Dashboard is not responding. It may still be initializing.")
					fmt.Println("   OpenSearch Dashboard can take longer to start up than OpenSearch itself.")
					fmt.Println("   The container is running but may need more time to fully initialize.")

					// Show logs to help diagnose issues
					if verbose {
						logsCmd := exec.Command("podman", "logs", "opensearch-dashboard")
						logsOutput, _ := logsCmd.CombinedOutput()
						if len(logsOutput) > 0 {
							fmt.Println("\nOpenSearch Dashboard container logs:")
							fmt.Println(string(logsOutput))
						}
					}

					// Mark as running anyway as the container is up
					// This prevents the entire localenv from failing when just the Dashboard UI is slow to start
					components[i].IsRunning = true
				}

				continue
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

		// Show summary of available components and their URLs
		if configLoaded {
			fmt.Println("\n=== Component URLs ===")

			// Temporal URLs
			if config.Components.Temporal && !skipTemporal {
				// Extract Temporal configuration values
				uiPort := 8233
				grpcPort := 7233
				namespace := "default"

				if config.Temporal.UIPort != 0 {
					uiPort = config.Temporal.UIPort
				}
				if config.Temporal.GRPCPort != 0 {
					grpcPort = config.Temporal.GRPCPort
				}
				if config.Temporal.Namespace != "" {
					namespace = config.Temporal.Namespace
				}

				fmt.Printf("Temporal UI: http://localhost:%d\n", uiPort)
				fmt.Printf("Temporal Server: localhost:%d (namespace: %s)\n", grpcPort, namespace)
				fmt.Println()
			}

			// Dapr URLs
			if config.Components.Dapr && !skipDapr {
				// Show Dapr Dashboard URL if enabled
				if config.Components.DaprDashboard && !skipDaprDashboard {
					dashboardPort := config.Dapr.DashboardPort
					fmt.Printf("Dapr Dashboard: http://localhost:%d\n", dashboardPort)
				}

				// Show Zipkin URL for tracing
				zipkinPort := 9411
				if config.Dapr.ZipkinPort != 0 {
					zipkinPort = config.Dapr.ZipkinPort
				}
				fmt.Printf("Zipkin UI (tracing): http://localhost:%d\n", zipkinPort)
				fmt.Println()
			}

			// OpenSearch URLs
			if config.Components.OpenSearch && !skipOpenSearch {
				fmt.Printf("OpenSearch API: http://localhost:%d\n", config.OpenSearch.Port)
				fmt.Printf("OpenSearch Dashboard: http://localhost:%d\n", config.OpenSearch.DashboardPort)
				fmt.Println()
			}
		}
	},
}

func init() {
	localenvCmd.AddCommand(startCmd)
	startCmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	startCmd.Flags().Bool("skip-dapr", false, "Skip starting Dapr")
	startCmd.Flags().Bool("skip-temporal", false, "Skip starting Temporal")
	startCmd.Flags().Bool("skip-dapr-dashboard", false, "Skip starting Dapr Dashboard")
	startCmd.Flags().Bool("skip-opensearch", false, "Skip starting OpenSearch")
	startCmd.Flags().Bool("force-restart", false, "Force restart of components even if already running")
	startCmd.Flags().StringP("config", "c", "", "Path to localenv configuration file")
	startCmd.Flags().Bool("stream-logs", false, "Stream Temporal server logs to terminal")
}
