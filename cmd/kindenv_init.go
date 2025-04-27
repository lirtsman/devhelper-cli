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

	"github.com/ShieldFC-RD/devhelper-cli/internal/kindenv"
	"github.com/spf13/cobra"
	yamlv3 "gopkg.in/yaml.v3"
)

// detectToolPath tries to find the path of a tool in the system
func detectToolPath(name string) string {
	path, err := exec.LookPath(name)
	if err != nil {
		return ""
	}
	return path
}

// getCommandVersion tries to get the version of a command
func getCommandVersion(command string, versionArgs ...string) string {
	if command == "" {
		return ""
	}

	args := versionArgs
	if len(args) == 0 {
		args = []string{"--version"}
	}

	cmd := exec.Command(command, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(out))
}

// parseVersion extracts a clean version string from command output
func parseVersion(versionOutput string, command string) string {
	if versionOutput == "" {
		return ""
	}

	// Different commands have different version output formats
	// Here we try to handle common formats
	lines := strings.Split(versionOutput, "\n")
	version := lines[0]

	// Extract just the version number for common formats
	switch command {
	case "kind":
		parts := strings.Split(version, " ")
		if len(parts) >= 2 {
			return parts[1]
		}
	case "kubectl":
		parts := strings.Fields(version)
		if len(parts) >= 2 && strings.HasPrefix(parts[0], "Client") {
			return parts[1]
		}
	case "helm":
		parts := strings.Fields(version)
		if len(parts) >= 3 && parts[0] == "version.BuildInfo" {
			return strings.TrimPrefix(parts[2], "v")
		}
	case "podman", "docker":
		parts := strings.Fields(version)
		for i, part := range parts {
			if part == "version" && i+1 < len(parts) {
				return parts[i+1]
			}
		}
	}

	return version
}

// detectAndPopulateToolVersions detects tool paths and versions and populates them in the config
func detectAndPopulateToolVersions(config *kindenv.KindEnvConfig) {
	// Detect essential tools paths and versions
	kindPath := detectToolPath("kind")
	kubectlPath := detectToolPath("kubectl")
	helmPath := detectToolPath("helm")

	// Detect optional tools
	dockerPath := detectToolPath("docker")
	podmanPath := detectToolPath("podman")
	awsPath := detectToolPath("aws")

	// Get versions
	kindVersion := parseVersion(getCommandVersion(kindPath), "kind")
	kubectlVersion := parseVersion(getCommandVersion(kubectlPath), "kubectl")
	helmVersion := parseVersion(getCommandVersion(helmPath), "helm")
	dockerVersion := parseVersion(getCommandVersion(dockerPath), "docker")
	podmanVersion := parseVersion(getCommandVersion(podmanPath), "podman")
	awsVersion := parseVersion(getCommandVersion(awsPath), "aws")

	// Tools section
	config.Tools.Podman.Path = podmanPath
	config.Tools.Podman.Version = podmanVersion
	config.Tools.Docker.Path = dockerPath
	config.Tools.Docker.Version = dockerVersion
	config.Tools.Kind.Path = kindPath
	config.Tools.Kind.Version = kindVersion
	config.Tools.Kubectl.Path = kubectlPath
	config.Tools.Kubectl.Version = kubectlVersion
	config.Tools.Helm.Path = helmPath
	config.Tools.Helm.Version = helmVersion
	config.Tools.AWS.Path = awsPath
	config.Tools.AWS.Version = awsVersion
}

// executeCommandWithOutput executes a command and returns its output
func executeCommandWithOutput(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

var kindenvInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Kind-based environment configuration",
	Long: `Initialize configuration for the Kind-based Kubernetes development environment.

This command creates a default configuration file (kindenv.yaml) that can be
customized for setting up a Kind cluster with Temporal, Dapr, Redis, and other
required components.

It also adds required Helm repositories for components like Dapr and Temporal.`,
	Run: func(cmd *cobra.Command, args []string) {
		configPath, _ := cmd.Flags().GetString("output")
		force, _ := cmd.Flags().GetBool("force")
		skipRepos, _ := cmd.Flags().GetBool("skip-repos")

		// If no config path is provided, use default
		if configPath == "" {
			configPath = "kindenv.yaml"
		}

		// Check if the file already exists and we're not forcing overwrite
		if _, err := os.Stat(configPath); err == nil && !force {
			fmt.Printf("❌ Configuration file already exists at %s. Use --force to overwrite.\n", configPath)
			return
		}

		// Create parent directories if needed
		configDir := filepath.Dir(configPath)
		if configDir != "." && configDir != ".." {
			if err := os.MkdirAll(configDir, 0755); err != nil {
				fmt.Printf("❌ Failed to create directory %s: %v\n", configDir, err)
				os.Exit(1)
			}
		}

		// Create default configuration from internal package
		config := kindenv.CreateDefaultConfig()

		// Detect and populate tool paths and versions
		detectAndPopulateToolVersions(config)

		// Serialize to YAML
		yamlData, err := yamlv3.Marshal(config)
		if err != nil {
			fmt.Printf("❌ Failed to generate YAML: %v\n", err)
			os.Exit(1)
		}

		// Write to file
		err = os.WriteFile(configPath, yamlData, 0644)
		if err != nil {
			fmt.Printf("❌ Failed to write configuration file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Configuration initialized at %s\n", configPath)
		fmt.Println("Edit this file to customize your Kind environment configuration.")

		// Add Helm repositories if not skipped
		if !skipRepos {
			fmt.Println("Adding Helm repositories...")

			// Add Temporal Helm repository
			fmt.Println("Adding Temporal Helm repository...")
			temporalOutput, err := executeCommandWithOutput("helm", "repo", "add", "temporal", "https://go.temporal.io/helm-charts")
			if err != nil {
				if strings.Contains(temporalOutput, "already exists") {
					fmt.Println("✅ Temporal Helm repository already configured")
				} else {
					fmt.Printf("⚠️  Warning: Failed to add Temporal Helm repository: %v\n", err)
					if temporalOutput != "" {
						fmt.Printf("  Output: %s\n", temporalOutput)
					}
				}
			} else {
				fmt.Println("✅ Temporal Helm repository added successfully")
			}

			// Add Dapr Helm repository
			fmt.Println("Adding Dapr Helm repository...")
			daprOutput, err := executeCommandWithOutput("helm", "repo", "add", "dapr", "https://dapr.github.io/helm-charts")
			if err != nil {
				if strings.Contains(daprOutput, "already exists") {
					fmt.Println("✅ Dapr Helm repository already configured")
				} else {
					fmt.Printf("⚠️  Warning: Failed to add Dapr Helm repository: %v\n", err)
					if daprOutput != "" {
						fmt.Printf("  Output: %s\n", daprOutput)
					}
				}
			} else {
				fmt.Println("✅ Dapr Helm repository added successfully")
			}

			// Add Bitnami Helm repository (for Redis)
			fmt.Println("Adding Bitnami (Redis) Helm repository...")
			bitnamiOutput, err := executeCommandWithOutput("helm", "repo", "add", "bitnami", "https://charts.bitnami.com/bitnami")
			if err != nil {
				if strings.Contains(bitnamiOutput, "already exists") {
					fmt.Println("✅ Bitnami Helm repository already configured")
				} else {
					fmt.Printf("⚠️  Warning: Failed to add Bitnami Helm repository: %v\n", err)
					if bitnamiOutput != "" {
						fmt.Printf("  Output: %s\n", bitnamiOutput)
					}
				}
			} else {
				fmt.Println("✅ Bitnami Helm repository added successfully")
			}

			// Update Helm repositories
			fmt.Println("Updating Helm repositories...")
			updateOutput, err := executeCommandWithOutput("helm", "repo", "update")
			if err != nil {
				fmt.Printf("⚠️  Warning: Failed to update Helm repositories: %v\n", err)
				if updateOutput != "" {
					fmt.Printf("  Output: %s\n", updateOutput)
				}
			} else {
				fmt.Println("✅ Helm repositories updated successfully")
			}
		}
	},
}

func init() {
	kindenvCmd.AddCommand(kindenvInitCmd)

	// Add flags for kindenv init command
	kindenvInitCmd.Flags().StringP("output", "o", "", "Output path for configuration file (default: kindenv.yaml)")
	kindenvInitCmd.Flags().BoolP("force", "f", false, "Force overwrite if configuration file already exists")
	kindenvInitCmd.Flags().Bool("skip-repos", false, "Skip adding Helm repositories")
}
