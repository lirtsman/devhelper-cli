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
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

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
		wait, _ := cmd.Flags().GetBool("wait")
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
			checkCmd := exec.Command("temporal", "workflow", "list")
			if err := checkCmd.Run(); err != nil {
				if verbose {
					fmt.Printf("Temporal server check failed: %v\n", err)
				}
				return false
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
				CheckArgs:       []string{"operator", "list", "--namespace", "default"},
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
			fmt.Println("Run 'shielddev-cli localenv init' to check required dependencies and create a configuration.")
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
							fmt.Println("   Create a cluster with: 'kind create cluster --name my-cluster'")
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
				// For Dapr Dashboard, we need special handling to make sure it stays running
				fmt.Println("Starting DaprDashboard in background mode...")

				// Try the configured port first
				dashboardPort := config.Dapr.DashboardPort
				dashboardStarted := tryStartDashboard(comp.Command, dashboardPort, nil)

				if dashboardStarted {
					components[i].IsRunning = true
					dashboardURL := fmt.Sprintf("http://%s:%d", config.Dapr.DashboardIP, dashboardPort)
					fmt.Printf("✅ Dapr Dashboard started at %s\n", dashboardURL)
				} else {
					fmt.Printf("❌ Failed to start Dapr Dashboard on port %d\n", dashboardPort)
					fmt.Println("   This could be because the port is already in use.")
					fmt.Println("   Update the dashboardPort in localenv.yaml to a different value and try again.")
					fmt.Println("   For example: dapr.dashboardPort: 8081")
				}
				continue
			}

			// For Temporal and other components that need to be started
			cmd := exec.Command(comp.Command, comp.Args...)

			// Start the process in the background
			if err := cmd.Start(); err != nil {
				fmt.Printf("❌ Failed to start %s: %v\n", comp.Name, err)
				continue
			}

			components[i].IsRunning = true

			// Wait a bit for the component to start
			time.Sleep(comp.StartupDelay)

			// If there's a verification function, use it to make sure the component is accessible
			if comp.VerifyAvailable != nil && wait {
				maxRetries := 5
				for retry := 0; retry < maxRetries; retry++ {
					if comp.VerifyAvailable() {
						break
					}
					if retry == maxRetries-1 {
						fmt.Printf("⚠️ %s may not be fully running yet. It might need more time to initialize.\n", comp.Name)
					} else {
						if verbose {
							fmt.Printf("Waiting for %s to become available (attempt %d/%d)...\n", comp.Name, retry+1, maxRetries)
						}
						time.Sleep(2 * time.Second)
					}
				}
			}

			if verbose {
				fmt.Printf("✅ Started %s (PID: %d)\n", comp.Name, cmd.Process.Pid)
			}
		}

		// Check if all required components are running
		allRunning := true
		for _, comp := range components {
			if comp.IsRequired && !comp.IsRunning {
				allRunning = false
				break
			}
		}

		if !allRunning {
			fmt.Println("\n⚠️ Not all components are running. Please check the errors above.")
			os.Exit(1)
		}

		fmt.Println("\n✅ Local development environment is running!")

		// Show connection information
		if configLoaded && config.Components.Temporal {
			// Default values
			frontendIP := "localhost"
			uiPort := 8233
			grpcPort := 7233
			namespace := "default"

			// Override with config if available
			if config.Temporal.FrontendIP != "" {
				frontendIP = config.Temporal.FrontendIP
			}
			if config.Temporal.UIPort != 0 {
				uiPort = config.Temporal.UIPort
			}
			if config.Temporal.GRPCPort != 0 {
				grpcPort = config.Temporal.GRPCPort
			}
			if config.Temporal.Namespace != "" {
				namespace = config.Temporal.Namespace
			}

			fmt.Printf("Temporal UI: http://%s:%d\n", frontendIP, uiPort)
			fmt.Printf("Temporal Server: %s:%d (namespace: %s)\n", frontendIP, grpcPort, namespace)
		} else {
			fmt.Printf("Temporal UI is available at http://localhost:8233\n")
		}

		// Show Dapr container information if enabled
		if configLoaded && config.Components.Dapr && !skipDaprDashboard {
			// Check for Dapr Dashboard
			dashboardCmd := exec.Command("dapr", "dashboard", "--help")
			if dashboardCmd.Run() == nil {
				// Dashboard command is available
				fmt.Printf("Dapr Dashboard: http://%s:%d\n", config.Dapr.DashboardIP, config.Dapr.DashboardPort)
			}

			// Show Zipkin URL for tracing
			zipkinIP := "localhost"
			zipkinPort := 9411

			if config.Dapr.ZipkinIP != "" {
				zipkinIP = config.Dapr.ZipkinIP
			}
			if config.Dapr.ZipkinPort != 0 {
				zipkinPort = config.Dapr.ZipkinPort
			}

			fmt.Printf("Zipkin UI (tracing): http://%s:%d\n", zipkinIP, zipkinPort)

			// Check if Dapr containers are running using podman
			cmd := exec.Command("podman", "ps", "--format", "{{.Names}}", "--filter", "name=dapr_")
			output, err := cmd.CombinedOutput()
			if err == nil && len(output) > 0 {
				fmt.Println("\nDapr Services:")
				containers := strings.Split(strings.TrimSpace(string(output)), "\n")
				for _, container := range containers {
					if container == "" {
						continue
					}
					fmt.Printf("- %s\n", container)
				}

				fmt.Println("\nYou can check running containers with: podman ps")
			}
		}

		fmt.Println("Use 'shielddev-cli localenv stop' to stop the environment.")
	},
}

func init() {
	localenvCmd.AddCommand(startCmd)

	// Add flags specific to the start command
	startCmd.Flags().Bool("skip-dapr", false, "Skip starting Dapr runtime")
	startCmd.Flags().Bool("skip-temporal", false, "Skip starting Temporal server")
	startCmd.Flags().Bool("skip-dapr-dashboard", false, "Skip starting Dapr Dashboard")
	startCmd.Flags().StringP("config", "c", "", "Path to environment configuration file (default: localenv.yaml)")
	startCmd.Flags().Bool("wait", true, "Wait for all components to be ready before exiting")
}

// Helper functions for determining if components are required
func getDaprRequirement(configLoaded bool, configValue bool, skipFlag bool) bool {
	if configLoaded {
		return configValue
	}
	return !skipFlag
}

func getTemporalRequirement(configLoaded bool, configValue bool, skipFlag bool) bool {
	if configLoaded {
		return configValue
	}
	return !skipFlag
}

func getDaprDashboardRequirement(configLoaded bool, configValue bool, skipFlag bool) bool {
	if configLoaded {
		return configValue
	}
	return !skipFlag
}

func tryStartDashboard(command string, port int, logFile *os.File) bool {
	dashboardCmd := exec.Command(command, "dashboard", "-p", strconv.Itoa(port), "--address", "0.0.0.0")

	// Redirect output to null device
	devNull, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	dashboardCmd.Stdout = devNull
	dashboardCmd.Stderr = devNull

	// Start the dashboard in a goroutine
	resultChan := make(chan error, 1)
	go func() {
		resultChan <- dashboardCmd.Run()
	}()

	// Wait a bit for the dashboard to start
	time.Sleep(3 * time.Second)

	// Check if the process exited quickly (indicating failure)
	select {
	case err := <-resultChan:
		// If we get here, the process exited before our timeout
		if err != nil {
			fmt.Printf("Dashboard failed to start: %v\n", err)
		}
		return false
	default:
		// No error yet, process is still running
		return true
	}
}
