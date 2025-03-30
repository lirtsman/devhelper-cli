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
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	yamlv3 "gopkg.in/yaml.v3"
)

// ToolVersion defines version requirements for a CLI tool
type ToolVersion struct {
	Name            string // Name of the tool
	MinVersion      string // Minimum required version
	InstallCommand  string // Command to install the tool
	UpdateCommand   string // Command to update the tool
	InstallURL      string // URL with installation instructions
	VersionRegex    string // Regex to extract version from output
	AutoInstallable bool   // Whether the tool can be auto-installed
}

// Required tool versions
var requiredVersions = map[string]ToolVersion{
	"podman": {
		Name:            "Podman",
		MinVersion:      "5.4.1", // Updated to match user's current version
		InstallCommand:  "brew install podman",
		UpdateCommand:   "brew upgrade podman",
		InstallURL:      "https://podman.io/getting-started/installation",
		VersionRegex:    `version (\d+\.\d+\.\d+)`,
		AutoInstallable: true,
	},
	"kind": {
		Name:            "Kind",
		MinVersion:      "0.14.0", // Keeping existing version as we don't know user's version
		InstallCommand:  "brew install kind",
		UpdateCommand:   "brew upgrade kind",
		InstallURL:      "https://kind.sigs.k8s.io/docs/user/quick-start/#installation",
		VersionRegex:    `v(\d+\.\d+\.\d+)`,
		AutoInstallable: true,
	},
	"dapr": {
		Name:            "Dapr CLI",
		MinVersion:      "1.14.1", // Updated to match user's current version
		InstallCommand:  "curl -fsSL https://raw.githubusercontent.com/dapr/cli/master/install/install.sh | /bin/bash",
		UpdateCommand:   "dapr upgrade",
		InstallURL:      "https://docs.dapr.io/getting-started/install-dapr-cli/",
		VersionRegex:    `CLI version: (\d+\.\d+\.\d+)`,
		AutoInstallable: true,
	},
	"temporal": {
		Name:            "Temporal CLI",
		MinVersion:      "1.2.0", // Updated to match user's current version
		InstallCommand:  "brew install temporalio/tap/temporal",
		UpdateCommand:   "brew upgrade temporal",
		InstallURL:      "https://docs.temporal.io/cli#install",
		VersionRegex:    `temporal version (\d+\.\d+\.\d+)`,
		AutoInstallable: true,
	},
	// Add a test tool for testing version checking
	"test-tool": {
		Name:            "Test Tool",
		MinVersion:      "1.0.0", // Higher than our test script's 0.1.0
		InstallCommand:  "echo 'This is a test install command'",
		UpdateCommand:   "echo 'This is a test update command'",
		InstallURL:      "https://example.com",
		VersionRegex:    `version (\d+\.\d+\.\d+)`,
		AutoInstallable: true,
	},
}

// LocalEnvConfig represents the configuration for the local environment
type LocalEnvConfig struct {
	Tools struct {
		Podman struct {
			Path    string `yaml:"path"`
			Version string `yaml:"version"`
		} `yaml:"podman"`
		Kind struct {
			Path    string `yaml:"path"`
			Version string `yaml:"version"`
		} `yaml:"kind"`
		DaprCli struct {
			Path    string `yaml:"path"`
			Version string `yaml:"version"`
		} `yaml:"daprCli"`
		TemporalCli struct {
			Path    string `yaml:"path"`
			Version string `yaml:"version"`
		} `yaml:"temporalCli"`
	} `yaml:"tools"`
	Components struct {
		Dapr struct {
			Enabled       bool `yaml:"enabled"`
			Dashboard     bool `yaml:"dashboard"`
			DashboardPort int  `yaml:"dashboardPort"`
			ZipkinPort    int  `yaml:"zipkinPort"`
		} `yaml:"dapr"`
		Temporal struct {
			Enabled   bool   `yaml:"enabled"`
			Namespace string `yaml:"namespace"`
			UIPort    int    `yaml:"uiPort"`
			GRPCPort  int    `yaml:"grpcPort"`
		} `yaml:"temporal"`
		OpenSearch struct {
			Enabled       bool   `yaml:"enabled"`
			Version       string `yaml:"version"`
			Port          int    `yaml:"port"`
			DashboardPort int    `yaml:"dashboardPort"`
		} `yaml:"openSearch"`
	} `yaml:"components"`
}

// initCmd represents the init command
var localenvInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize local development environment configuration",
	Long: `Initialize the local development environment by:
1. Validating required tools (Podman, Kind, Dapr CLI, Temporal CLI)
2. Creating a configuration file (localenv.yaml) in the current directory
3. Allowing you to customize which components to enable

This command should be run once before using other localenv commands.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Initializing local development environment...")

		force, _ := cmd.Flags().GetBool("force")
		verbose, _ := cmd.Flags().GetBool("verbose")

		// Check if localenv.yaml already exists
		configPath := "localenv.yaml"
		if _, err := os.Stat(configPath); err == nil && !force {
			fmt.Println("❌ Configuration file already exists. Use --force to overwrite.")
			return
		}

		// Initialize default configuration
		config := LocalEnvConfig{}

		// Initialize components
		config.Components.Dapr.Enabled = true
		config.Components.Temporal.Enabled = true
		config.Components.OpenSearch.Enabled = true

		// Set default Temporal configuration
		config.Components.Temporal.Namespace = "default"
		config.Components.Temporal.UIPort = 8233
		config.Components.Temporal.GRPCPort = 7233

		// Set default Dapr configuration
		config.Components.Dapr.DashboardPort = 8080
		config.Components.Dapr.ZipkinPort = 9411

		// Set default OpenSearch configuration
		config.Components.OpenSearch.Port = 9200
		config.Components.OpenSearch.DashboardPort = 5601
		config.Components.OpenSearch.Version = "2.17.1"

		// Check for required tools and record their paths
		fmt.Println("\n=== Validating Required Tools ===")

		// Check Podman
		podmanPath, podmanErr, podmanVersion := validateToolWithVersionDetection("podman", "--version", verbose)
		if podmanErr != nil {
			fmt.Printf("❌ Podman: %v\n", podmanErr)
			fmt.Println("   Please install Podman from: https://podman.io/getting-started/installation")
		} else {
			fmt.Printf("✅ Podman: Found at %s (Version: %s)\n", podmanPath, podmanVersion)
			config.Tools.Podman.Path = podmanPath
			config.Tools.Podman.Version = podmanVersion

			// Check if Podman can run containers
			cmd := exec.Command(podmanPath, "ps")
			if err := cmd.Run(); err != nil {
				fmt.Println("⚠️  Podman is installed but may not be configured correctly to run containers")
				fmt.Println("   Make sure you have proper permissions and Podman is configured correctly")
			}
		}

		// Check Kind
		kindPath, kindErr, kindVersion := validateToolWithVersionDetection("kind", "--version", verbose)
		if kindErr != nil {
			fmt.Printf("❌ Kind: %v\n", kindErr)
			fmt.Println("   Please install Kind from: https://kind.sigs.k8s.io/docs/user/quick-start/#installation")
		} else {
			fmt.Printf("✅ Kind: Found at %s (Version: %s)\n", kindPath, kindVersion)
			config.Tools.Kind.Path = kindPath
			config.Tools.Kind.Version = kindVersion

			// Check if Kind has any clusters configured
			clusterCmd := exec.Command(kindPath, "get", "clusters")
			output, err := clusterCmd.CombinedOutput()
			if err != nil || strings.TrimSpace(string(output)) == "" {
				fmt.Println("⚠️  Kind is installed but no clusters are configured")
				fmt.Println("   Note: Kubernetes functionality is not required for local development")
			} else {
				fmt.Printf("   Clusters found: %s\n", strings.ReplaceAll(string(output), "\n", ", "))
			}
		}

		// Check Dapr CLI
		daprPath, daprErr, daprVersion := validateToolWithVersionDetection("dapr", "--version", verbose)
		if daprErr != nil {
			fmt.Printf("❌ Dapr CLI: %v\n", daprErr)
			fmt.Println("   Please install Dapr CLI from: https://docs.dapr.io/getting-started/install-dapr-cli/")
			config.Components.Dapr.Enabled = false
		} else {
			fmt.Printf("✅ Dapr CLI: Found at %s (Version: %s)\n", daprPath, daprVersion)
			config.Tools.DaprCli.Path = daprPath
			config.Tools.DaprCli.Version = daprVersion

			// Check if Docker is available (for original Dapr support)
			dockerAvailable := isCommandAvailable("docker")

			// If Podman is available but Docker is not, add a note about using --container-runtime
			if podmanErr == nil && !dockerAvailable {
				fmt.Println("   Note: Dapr will be initialized with '--container-runtime podman' since Docker is not available")
				fmt.Println("   This will start Redis, Zipkin, placement, and scheduler containers using Podman")
			}

			// Check if Dapr is initialized
			listCmd := exec.Command(daprPath, "list")
			if err := listCmd.Run(); err != nil {
				fmt.Println("⚠️  Dapr CLI is installed but Dapr may not be initialized")
				fmt.Println("   Dapr will be initialized when you run 'devhelper-cli localenv start'")
			} else {
				fmt.Println("   Dapr is initialized and ready to use")
			}

			// Check if Dapr Dashboard is available
			dashboardCmd := exec.Command(daprPath, "dashboard", "--help")
			if err := dashboardCmd.Run(); err == nil {
				fmt.Println("✅ Dapr Dashboard: Available")
				fmt.Println("   Run with: 'dapr dashboard'")
				fmt.Println("   Or use 'devhelper-cli localenv start' with the Dapr Dashboard enabled")

				// Enable Dapr Dashboard by default if available
				config.Components.Dapr.Dashboard = true
			} else {
				fmt.Println("ℹ️ Dapr Dashboard: Not available")
				fmt.Println("   Install with: 'dapr dashboard install'")
				config.Components.Dapr.Dashboard = false
			}
		}

		// Check Temporal CLI
		temporalPath, temporalErr, temporalVersion := validateToolWithVersionDetection("temporal", "--version", verbose)
		if temporalErr != nil {
			fmt.Printf("❌ Temporal CLI: %v\n", temporalErr)
			fmt.Println("   Please install Temporal CLI from: https://docs.temporal.io/cli#install")
			config.Components.Temporal.Enabled = false
		} else {
			fmt.Printf("✅ Temporal CLI: Found at %s (Version: %s)\n", temporalPath, temporalVersion)
			config.Tools.TemporalCli.Path = temporalPath
			config.Tools.TemporalCli.Version = temporalVersion
		}

		// Check Podman (for OpenSearch)
		podmanPath, podmanErr, _ = validateToolWithVersionDetection("podman", "--version", verbose)
		if podmanErr != nil {
			fmt.Printf("❌ Podman: %v\n", podmanErr)
			fmt.Println("   Please install Podman from: https://podman.io/getting-started/installation")
			fmt.Println("   Note: Podman is required for running OpenSearch")
			config.Components.OpenSearch.Enabled = false
		} else {
			fmt.Printf("✅ Podman: Found at %s\n", podmanPath)
			config.Tools.Podman.Path = podmanPath

			// Check if Podman is running
			cmd := exec.Command(podmanPath, "ps")
			if err := cmd.Run(); err != nil {
				fmt.Println("⚠️  Podman is installed but may not be running")
				fmt.Println("   Start Podman and try again")
				config.Components.OpenSearch.Enabled = false
			}
		}

		fmt.Println("=== Components ===")
		fmt.Println("The following components will be enabled:")
		if config.Components.Dapr.Enabled {
			fmt.Println("✅ Dapr")
		}
		if config.Components.Dapr.Dashboard {
			fmt.Println("✅ Dapr Dashboard")
		}
		if config.Components.Temporal.Enabled {
			fmt.Println("✅ Temporal")
		}
		if config.Components.OpenSearch.Enabled {
			fmt.Println("✅ OpenSearch")
		}

		// Write configuration to file
		var buf bytes.Buffer
		encoder := yamlv3.NewEncoder(&buf)
		encoder.SetIndent(2)

		if err := encoder.Encode(config); err != nil {
			fmt.Printf("❌ Failed to generate configuration: %v\n", err)
			return
		}

		if err := os.WriteFile(configPath, buf.Bytes(), 0644); err != nil {
			fmt.Printf("❌ Failed to write configuration file: %v\n", err)
			return
		}

		fmt.Printf("\n✅ Configuration written to %s\n", configPath)
		fmt.Println("\nYou can now use 'devhelper-cli localenv start' to start the local environment")
		fmt.Println("or edit the configuration file to customize enabled components.")
	},
}

// validateToolWithVersionDetection checks if a tool is installed, verifies its version, and offers to install/update
func validateToolWithVersionDetection(name, versionFlag string, verbose bool) (string, error, string) {
	// Check if the tool is in PATH
	path, err := exec.LookPath(name)
	if err != nil {
		// Tool not found, offer to install
		fmt.Printf("❌ %s not found. Would you like to install it? (y/n): ", requiredVersions[name].Name)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response == "y" || response == "yes" {
			if err := installTool(name); err != nil {
				return "", fmt.Errorf("installation failed: %v", err), ""
			}
			// Recheck after installation
			path, err = exec.LookPath(name)
			if err != nil {
				return "", fmt.Errorf("still not found after installation attempt"), ""
			}
			fmt.Printf("✅ %s successfully installed\n", requiredVersions[name].Name)
		} else {
			return "", fmt.Errorf("not found in PATH"), ""
		}
	}

	// Check if the tool works by running version command
	cmd := exec.Command(path, versionFlag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return path, fmt.Errorf("found but failed to run: %v", err), ""
	}

	outputStr := string(output)
	if verbose {
		fmt.Printf("   %s version output: %s\n", name, strings.TrimSpace(outputStr))
	}

	// Extract version
	var currentVersion string
	versionInfo, ok := requiredVersions[name]
	if ok {
		currentVersion = extractVersion(outputStr, versionInfo.VersionRegex)

		if currentVersion != "" {
			if compareVersions(currentVersion, versionInfo.MinVersion) < 0 {
				fmt.Printf("⚠️ %s version %s is below recommended minimum %s\n", versionInfo.Name, currentVersion, versionInfo.MinVersion)
				fmt.Printf("Would you like to update %s? (y/n): ", versionInfo.Name)

				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				response = strings.TrimSpace(strings.ToLower(response))

				if response == "y" || response == "yes" {
					if err := updateTool(name); err != nil {
						fmt.Printf("Update failed: %v. Please update manually.\n", err)
					} else {
						fmt.Printf("✅ %s successfully updated\n", versionInfo.Name)

						// Get new version after update
						cmd = exec.Command(path, versionFlag)
						newOutput, err := cmd.CombinedOutput()
						if err == nil {
							newVersion := extractVersion(string(newOutput), versionInfo.VersionRegex)
							if newVersion != "" {
								currentVersion = newVersion
								fmt.Printf("   New version: %s\n", currentVersion)
							}
						}
					}
				}
			} else if verbose {
				fmt.Printf("   %s version %s meets minimum requirement of %s\n", versionInfo.Name, currentVersion, versionInfo.MinVersion)
			}
		}
	}

	return path, nil, currentVersion
}

// Modify the existing validateTool to use the new function
func validateTool(name, versionFlag string, verbose bool) (string, error) {
	path, err, _ := validateToolWithVersionDetection(name, versionFlag, verbose)
	return path, err
}

// extractVersion extracts a semantic version using the provided regex
func extractVersion(output, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(output)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// compareVersions compares two semantic versions
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func compareVersions(v1, v2 string) int {
	v1Parts := strings.Split(v1, ".")
	v2Parts := strings.Split(v2, ".")

	// Normalize length
	maxLen := len(v1Parts)
	if len(v2Parts) > maxLen {
		maxLen = len(v2Parts)
	}

	for i := 0; i < maxLen; i++ {
		var v1Part, v2Part int

		if i < len(v1Parts) {
			fmt.Sscanf(v1Parts[i], "%d", &v1Part)
		}

		if i < len(v2Parts) {
			fmt.Sscanf(v2Parts[i], "%d", &v2Part)
		}

		if v1Part < v2Part {
			return -1
		} else if v1Part > v2Part {
			return 1
		}
	}

	return 0
}

// installTool installs a tool
func installTool(name string) error {
	toolInfo, ok := requiredVersions[name]
	if !ok || !toolInfo.AutoInstallable {
		return fmt.Errorf("auto-installation not supported for %s", name)
	}

	fmt.Printf("Installing %s...\n", toolInfo.Name)

	// Split install command to handle shell commands with arguments
	cmdParts := strings.Split(toolInfo.InstallCommand, " ")
	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// updateTool updates a tool
func updateTool(name string) error {
	toolInfo, ok := requiredVersions[name]
	if !ok || !toolInfo.AutoInstallable {
		return fmt.Errorf("auto-update not supported for %s", name)
	}

	fmt.Printf("Updating %s...\n", toolInfo.Name)

	// Split update command to handle shell commands with arguments
	cmdParts := strings.Split(toolInfo.UpdateCommand, " ")
	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func init() {
	localenvCmd.AddCommand(localenvInitCmd)

	// Add flags
	localenvInitCmd.Flags().BoolP("force", "f", false, "Force overwrite of existing configuration")
	localenvInitCmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
}
