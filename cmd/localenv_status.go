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

// statusCmd represents the status command for localenv
var localenvStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check status of local development environment",
	Long: `Check the status of the local development environment components.

This command will check if all required components for the local development
environment are running, including:
- Dapr runtime
- Temporal server
- Related dependencies`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Checking local environment status...")

		verbose, _ := cmd.Flags().GetBool("verbose")
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
			fmt.Println("   Run 'devhelper-cli localenv init' to create a configuration")
		}

		// Define status checks for each component
		components := []struct {
			Name          string
			CheckCommand  string
			CheckArgs     []string
			StatusMessage string
			Available     bool
			Enabled       bool
			WebUIURL      string
			CheckUI       bool
			IsBinary      bool
		}{
			{
				Name:         "Podman",
				CheckCommand: "podman",
				CheckArgs:    []string{"ps", "--format", "{{.Names}} - {{.Status}}"},
				Available:    isCommandAvailable("podman"),
				Enabled:      true, // Always required
				IsBinary:     true,
			},
			{
				Name:         "Kind",
				CheckCommand: "kind",
				CheckArgs:    []string{"get", "clusters"},
				Available:    isCommandAvailable("kind"),
				Enabled:      true, // Always required
				IsBinary:     true,
			},
			{
				Name:         "Dapr",
				CheckCommand: "dapr",
				CheckArgs:    []string{"list"},
				Available:    isCommandAvailable("dapr"),
				Enabled:      getDaprStatusRequirement(configLoaded, config.Components.Dapr.Enabled),
				WebUIURL:     getDaprWebUIURL(configLoaded, config),
				CheckUI:      isDaprDashboardAvailable(),
				IsBinary:     false,
			},
			{
				Name:         "Temporal",
				CheckCommand: "temporal",
				CheckArgs:    getTemporalNamespaceArgs(configLoaded, config),
				Available:    isCommandAvailable("temporal"),
				Enabled:      getTemporalStatusRequirement(configLoaded, config.Components.Temporal.Enabled),
				WebUIURL:     getTemporalUIURL(configLoaded, config),
				CheckUI:      true,
				IsBinary:     false,
			},
		}

		fmt.Println("\n=== Local Environment Status ===")

		// First check required tools
		fmt.Println("\n== Required Tools ==")
		allToolsInstalled := true
		for _, comp := range components {
			if !comp.Available {
				fmt.Printf("❌ %s: Not installed\n", comp.Name)
				allToolsInstalled = false
			} else if comp.IsBinary {
				fmt.Printf("✅ %s: Installed\n", comp.Name)
			}
		}

		if !allToolsInstalled {
			fmt.Println("\n❌ Some required tools are not installed.")
			fmt.Println("Run 'devhelper-cli localenv init' to check required dependencies and create a configuration.")
			return
		}

		// Check Podman functionality
		fmt.Println("\n== Tool Functionality ==")
		podmanWorking := checkToolFunctionality("podman", []string{"ps"}, verbose)
		if podmanWorking {
			fmt.Println("✅ Podman: Can run containers")
		} else {
			fmt.Println("❌ Podman: Not able to run containers")
			fmt.Println("   Make sure Podman is installed correctly and has proper permissions.")
		}

		// Check Kind functionality
		kindWorking := checkToolFunctionality("kind", []string{"get", "clusters"}, verbose)
		if kindWorking {
			// Execute kind get clusters and check if any clusters are listed
			cmd := exec.Command("kind", "get", "clusters")
			output, err := cmd.Output()
			outputStr := strings.TrimSpace(string(output))

			if err == nil && outputStr != "" {
				fmt.Println("✅ Kind: Clusters configured")
			} else {
				fmt.Println("✅ Kind: Tool working, but no clusters configured")
				fmt.Println("   Note: Kubernetes functionality is not required for local development.")
			}
		} else {
			fmt.Println("❌ Kind: No clusters available")
			fmt.Println("   Note: Kubernetes functionality is not required for local development.")
		}

		if !podmanWorking || !kindWorking {
			fmt.Println("\n❌ Required tools are not working properly.")
			fmt.Println("Fix the issues above before continuing.")
			return
		}

		// Check enabled services
		fmt.Println("\n== Enabled Components ==")
		allRunning := true
		anyEnabled := false

		for _, comp := range components {
			if !comp.Enabled || comp.IsBinary {
				continue
			}

			anyEnabled = true

			// Special handling for Dapr
			if comp.Name == "Dapr" {
				// Check if Dapr binaries exist
				_, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".dapr", "bin", "daprd"))
				if err != nil {
					fmt.Printf("❌ %s: Not initialized\n", comp.Name)
					fmt.Println("   Run 'devhelper-cli localenv start' to initialize Dapr")
					allRunning = false
					continue
				}

				// Check if we can run dapr list
				cmd := exec.Command("dapr", "list")
				output, err := cmd.CombinedOutput()
				if err != nil {
					fmt.Printf("❌ %s: Not running properly\n", comp.Name)
					if verbose && len(output) > 0 {
						fmt.Printf("   Details: %s\n", strings.TrimSpace(string(output)))
					}
					allRunning = false
					continue
				}

				fmt.Printf("✅ %s: Running\n", comp.Name)
				if verbose {
					fmt.Println("   No Dapr apps are currently running, but the runtime is available")
				}
				continue
			}

			// Standard check for other components
			cmd := exec.Command(comp.CheckCommand, comp.CheckArgs...)
			output, err := cmd.CombinedOutput()
			outputStr := strings.TrimSpace(string(output))

			if err != nil {
				// Special case for Temporal - sometimes workflow list returns an error even when Temporal is running
				if comp.Name == "Temporal" {
					// Try a simpler check - just use the UI availability as the primary indicator
					client := http.Client{
						Timeout: 2 * time.Second,
					}
					resp, err := client.Get(comp.WebUIURL)
					if err == nil && resp.StatusCode < 400 {
						resp.Body.Close()
						fmt.Printf("✅ %s: Running\n", comp.Name)
						fmt.Printf("   UI: %s (Accessible)\n", comp.WebUIURL)
						continue
					}
				}

				fmt.Printf("❌ %s: Not running\n", comp.Name)
				if verbose && outputStr != "" {
					fmt.Printf("   Details: %s\n", outputStr)
				}
				allRunning = false
				continue
			}

			fmt.Printf("✅ %s: Running\n", comp.Name)

			// For Temporal, check if the UI is accessible
			if comp.CheckUI && comp.WebUIURL != "" {
				uiAccessible := false
				client := http.Client{
					Timeout: 2 * time.Second,
				}
				resp, err := client.Get(comp.WebUIURL)
				if err == nil && resp.StatusCode < 400 {
					uiAccessible = true
					resp.Body.Close()
				}

				if uiAccessible {
					fmt.Printf("   UI: %s (Accessible)\n", comp.WebUIURL)
				} else {
					fmt.Printf("   UI: %s (Not accessible yet, may still be starting up)\n", comp.WebUIURL)
				}

				// Check if the configured namespace exists (if not default)
				if comp.Name == "Temporal" && configLoaded && config.Components.Temporal.Enabled && config.Components.Temporal.Namespace != "" && config.Components.Temporal.Namespace != "default" {
					namespaceCmd := exec.Command("temporal", "operator", "namespace", "describe", config.Components.Temporal.Namespace)
					if err := namespaceCmd.Run(); err != nil {
						fmt.Printf("   ⚠️ Namespace '%s' does not exist. It will be created when starting the environment.\n", config.Components.Temporal.Namespace)
					} else {
						fmt.Printf("   ✅ Temporal namespace '%s' exists\n", config.Components.Temporal.Namespace)
					}
				}
			}

			if verbose && outputStr != "" {
				// Format and print relevant output details
				lines := strings.Split(outputStr, "\n")
				if len(lines) > 5 && !verbose {
					// Truncate output if it's too long and we're not in verbose mode
					for i := 0; i < 3; i++ {
						if i < len(lines) {
							fmt.Printf("   %s\n", lines[i])
						}
					}
					fmt.Printf("   ... %d more lines ...\n", len(lines)-3)
				} else {
					for _, line := range lines {
						fmt.Printf("   %s\n", line)
					}
				}
			}
		}

		if !anyEnabled {
			fmt.Println("ℹ️ No components are enabled in the configuration.")
			fmt.Println("   Edit localenv.yaml to enable components or run 'devhelper-cli localenv init' to create a new config.")
			return
		}

		// Check OpenSearch
		if configLoaded && config.Components.OpenSearch.Enabled {
			// First check if OpenSearch service container is running
			checkCmd := exec.Command("podman", "ps", "--filter", "name=opensearch-node", "--format", "{{.Names}}")
			output, err := checkCmd.CombinedOutput()
			opensearchRunning := err == nil && strings.Contains(string(output), "opensearch-node")

			// Check if security is disabled
			securityDisabled := false
			if opensearchRunning {
				inspectCmd := exec.Command("podman", "inspect", "--format", "{{range .Config.Env}}{{.}}{{println}}{{end}}", "opensearch-node")
				inspectOutput, _ := inspectCmd.CombinedOutput()
				if strings.Contains(string(inspectOutput), "DISABLE_SECURITY_PLUGIN=true") {
					securityDisabled = true
				}
			}

			if opensearchRunning {
				fmt.Println("✅ OpenSearch: Running")

				// Attempt to check the health of the OpenSearch service
				url := fmt.Sprintf("http://localhost:%d", config.Components.OpenSearch.Port)

				client := http.Client{
					Timeout: 2 * time.Second,
				}

				var resp *http.Response
				var httpErr error

				// Send request with or without credentials based on security status
				if securityDisabled {
					// No credentials needed when security is disabled
					resp, httpErr = client.Get(url)
				} else {
					// Use credentials when security is enabled
					req, _ := http.NewRequest("GET", url, nil)
					resp, httpErr = client.Do(req)
				}

				if httpErr == nil {
					defer resp.Body.Close()
					if resp.StatusCode >= 200 && resp.StatusCode < 300 {
						fmt.Printf("   API: http://localhost:%d (available)\n", config.Components.OpenSearch.Port)
					} else {
						fmt.Printf("   API: http://localhost:%d (unhealthy, status code: %d)\n", config.Components.OpenSearch.Port, resp.StatusCode)
					}
				} else {
					fmt.Printf("   API: http://localhost:%d (unavailable, service may still be starting)\n", config.Components.OpenSearch.Port)
					if verbose {
						fmt.Printf("   Error: %v\n", httpErr)
					}
				}

				// Check if Dashboard container is running
				dashCmd := exec.Command("podman", "ps", "--filter", "name=opensearch-dashboard", "--format", "{{.Names}}")
				dashOutput, dashErr := dashCmd.CombinedOutput()
				dashboardRunning := dashErr == nil && strings.Contains(string(dashOutput), "opensearch-dashboard")

				if dashboardRunning {
					fmt.Println("✅ OpenSearch Dashboard: Running")
					fmt.Printf("   Dashboard: http://localhost:%d (available)\n", config.Components.OpenSearch.DashboardPort)
				} else {
					fmt.Println("❌ OpenSearch Dashboard: Not running")
					fmt.Println("   Run 'devhelper-cli localenv start' to start the OpenSearch Dashboard")
				}

				fmt.Println("   Security plugin disabled - no credentials required for API")

			} else {
				fmt.Println("❌ OpenSearch: Not running")
				fmt.Println("   Run 'devhelper-cli localenv start' to start OpenSearch")
			}
		} else if configLoaded {
			fmt.Println("⏹️ OpenSearch: Disabled in configuration")
		} else {
			fmt.Println("⚠️ OpenSearch: No configuration found")
			fmt.Println("   Run 'devhelper-cli localenv init' to create a configuration")
		}

		fmt.Println("\n=== Summary ===")
		if allRunning {
			fmt.Println("✅ All components are running properly.")

			// Show concise component information
			if configLoaded && config.Components.Temporal.Enabled {
				// Extract Temporal configuration values
				uiPort := 8233
				grpcPort := 7233
				namespace := "default"

				if config.Components.Temporal.UIPort != 0 {
					uiPort = config.Components.Temporal.UIPort
				}
				if config.Components.Temporal.GRPCPort != 0 {
					grpcPort = config.Components.Temporal.GRPCPort
				}
				if config.Components.Temporal.Namespace != "" {
					namespace = config.Components.Temporal.Namespace
				}

				fmt.Printf("\nTemporal UI: http://localhost:%d\n", uiPort)
				fmt.Printf("Temporal Server: localhost:%d (namespace: %s)\n", grpcPort, namespace)
			}

			// Show Dapr connection information if enabled
			if configLoaded && config.Components.Dapr.Enabled {
				// Check for Dapr Dashboard
				if isDaprDashboardAvailable() {
					dashboardURL := getDaprDashboardURL(configLoaded, config)
					if isDaprDashboardAccessible(dashboardURL) {
						fmt.Printf("\nDapr Dashboard: %s\n", dashboardURL)
					} else {
						fmt.Printf("\nDapr Dashboard: %s (Not accessible)\n", dashboardURL)
						fmt.Println("   If the dashboard is running but not accessible on this port,")
						fmt.Println("   update the dashboardPort in localenv.yaml and restart the environment.")
					}
				}

				// Show Zipkin URL for tracing
				zipkinURL := getZipkinURL(configLoaded, config)
				fmt.Printf("Zipkin UI (tracing): %s\n", zipkinURL)

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

					// Check if Dapr Dashboard process is running
					dashboardPID := getDaprDashboardPID()
					if dashboardPID != "" {
						fmt.Println("- dapr_dashboard")
					}
				} else {
					// If we couldn't find containers but the dashboard is running, still show it
					dashboardPID := getDaprDashboardPID()
					if dashboardPID != "" {
						fmt.Println("\nDapr Services:")
						fmt.Println("- dapr_dashboard")
					}
				}
			}

			// Show OpenSearch information if enabled and running
			if configLoaded && config.Components.OpenSearch.Enabled {
				// First check if OpenSearch service container is running
				checkCmd := exec.Command("podman", "ps", "--filter", "name=opensearch-node", "--format", "{{.Names}}")
				output, err := checkCmd.CombinedOutput()
				opensearchRunning := err == nil && strings.Contains(string(output), "opensearch-node")

				// Check if security is disabled
				securityDisabled := false
				if opensearchRunning {
					inspectCmd := exec.Command("podman", "inspect", "--format", "{{range .Config.Env}}{{.}}{{println}}{{end}}", "opensearch-node")
					inspectOutput, _ := inspectCmd.CombinedOutput()
					if strings.Contains(string(inspectOutput), "DISABLE_SECURITY_PLUGIN=true") {
						securityDisabled = true
					}
				}

				// Then check if Dashboard container is running
				dashCmd := exec.Command("podman", "ps", "--filter", "name=opensearch-dashboard", "--format", "{{.Names}}")
				dashOutput, dashErr := dashCmd.CombinedOutput()
				dashboardRunning := dashErr == nil && strings.Contains(string(dashOutput), "opensearch-dashboard")

				if opensearchRunning {
					// Try to check if the API is accessible
					client := http.Client{
						Timeout: 2 * time.Second,
					}

					var resp *http.Response
					var httpErr error

					// Send request with or without credentials based on security status
					if securityDisabled {
						// No credentials needed when security is disabled
						resp, httpErr = client.Get(fmt.Sprintf("http://localhost:%d", config.Components.OpenSearch.Port))
					} else {
						// Use credentials when security is enabled
						req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d", config.Components.OpenSearch.Port), nil)
						resp, httpErr = client.Do(req)
					}

					apiAccessible := false
					if httpErr == nil {
						defer resp.Body.Close()
						if resp.StatusCode >= 200 && resp.StatusCode < 300 {
							apiAccessible = true
						}
					}

					// Display OpenSearch information
					fmt.Println("\nOpenSearch:")
					fmt.Printf("- API: http://localhost:%d", config.Components.OpenSearch.Port)
					if apiAccessible {
						fmt.Println(" (accessible)")
					} else {
						fmt.Println(" (not verified)")
					}

					if dashboardRunning {
						fmt.Printf("- Dashboard: http://localhost:%d (running)\n", config.Components.OpenSearch.DashboardPort)
					} else {
						fmt.Printf("- Dashboard: http://localhost:%d (not running)\n", config.Components.OpenSearch.DashboardPort)
					}

					fmt.Println("- Security plugin disabled - no credentials required for API")

				}
			}
		} else {
			fmt.Println("⚠️  Some components are not running.")
			fmt.Println("Run 'devhelper-cli localenv start' to start the environment.")
		}
	},
}

// Helper function to check if a tool is working properly
func checkToolFunctionality(command string, args []string, verbose bool) bool {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if verbose {
			fmt.Printf("   Details: %s\n", strings.TrimSpace(string(output)))
		}
		return false
	}
	return true
}

// Helper functions for status command
func getDaprStatusRequirement(configLoaded bool, enabled bool) bool {
	if !configLoaded {
		return true // Default to enabled if no config
	}
	return enabled
}

func getTemporalStatusRequirement(configLoaded bool, enabled bool) bool {
	if !configLoaded {
		return true // Default to enabled if no config
	}
	return enabled
}

// Helper function to get Temporal namespace args
func getTemporalNamespaceArgs(configLoaded bool, config LocalEnvConfig) []string {
	namespace := "default"
	if configLoaded && config.Components.Temporal.Namespace != "" {
		namespace = config.Components.Temporal.Namespace
	}
	return []string{"operator", "namespace", "describe", namespace}
}

// Helper function to get Temporal UI URL
func getTemporalUIURL(configLoaded bool, config LocalEnvConfig) string {
	uiPort := 8233

	if configLoaded && config.Components.Temporal.UIPort != 0 {
		uiPort = config.Components.Temporal.UIPort
	}

	return fmt.Sprintf("http://localhost:%d", uiPort)
}

// Helper function to check if Dapr Dashboard is available
func isDaprDashboardAvailable() bool {
	// Check if dapr dashboard command is available
	cmd := exec.Command("dapr", "dashboard", "--help")
	err := cmd.Run()
	return err == nil
}

// Helper function to get Dapr Dashboard URL
func getDaprDashboardURL(configLoaded bool, config LocalEnvConfig) string {
	dashboardPort := 8080

	if configLoaded && config.Components.Dapr.DashboardPort != 0 {
		dashboardPort = config.Components.Dapr.DashboardPort
	}

	// Simply return the configured URL
	return fmt.Sprintf("http://localhost:%d", dashboardPort)
}

// Helper function to check if Dapr Dashboard is running and get its PID
func getDaprDashboardPID() string {
	// First try to find the process using ps command
	cmd := exec.Command("ps", "-ef")
	output, err := cmd.CombinedOutput()
	if err == nil {
		outputLines := strings.Split(string(output), "\n")
		for _, line := range outputLines {
			if strings.Contains(line, "dapr dashboard") && !strings.Contains(line, "grep") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					return fields[1] // Return PID
				}
			}
		}
	}

	// If we can't find it with ps, try a different approach using lsof
	// Look for the 'dashboard' process on common Dapr Dashboard ports
	commonPorts := []string{"8080", "8081", "8082", "8083", "8084", "8085"}

	for _, port := range commonPorts {
		cmd = exec.Command("lsof", "-i", ":"+port, "-sTCP:LISTEN")
		output, err = cmd.CombinedOutput()
		if err == nil && len(output) > 0 {
			outputLines := strings.Split(string(output), "\n")
			for _, line := range outputLines {
				if strings.HasPrefix(line, "dashboard") {
					fields := strings.Fields(line)
					if len(fields) >= 2 {
						return fields[1] // Return PID
					}
				}
			}
		}
	}

	return ""
}

// Helper function to check if a specific port is in use
func isPortInUse(port int) bool {
	cmd := exec.Command("lsof", "-i", ":"+strconv.Itoa(port), "-sTCP:LISTEN")
	output, err := cmd.CombinedOutput()
	return err == nil && len(output) > 0
}

// Helper function to check if Dapr Dashboard is accessible
func isDaprDashboardAccessible(url string) bool {
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := client.Get(url)
	if err == nil {
		resp.Body.Close()
		return true
	}
	return false
}

// Helper function to get Dapr Dashboard URL with availability check
func getDaprWebUIURL(configLoaded bool, config LocalEnvConfig) string {
	if !isDaprDashboardAvailable() {
		return ""
	}

	return getDaprDashboardURL(configLoaded, config)
}

// Helper function to get Zipkin URL
func getZipkinURL(configLoaded bool, config LocalEnvConfig) string {
	zipkinPort := 9411

	if configLoaded && config.Components.Dapr.ZipkinPort != 0 {
		zipkinPort = config.Components.Dapr.ZipkinPort
	}

	return fmt.Sprintf("http://localhost:%d", zipkinPort)
}

func init() {
	localenvCmd.AddCommand(localenvStatusCmd)
	localenvStatusCmd.Flags().StringP("config", "c", "", "Path to environment configuration file (default: localenv.yaml)")
}
