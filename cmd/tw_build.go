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
	"github.com/ShieldFC-RD/devhelper-cli/internal/tw"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// twBuildCmd represents the tw build command
var twBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a Temporal Worker Docker image",
	Long: `Build a Docker image for a Temporal Worker.

This command builds a Docker image for a Temporal Worker using the configuration
from tw.yaml and a Dockerfile in the current directory.

When using --kind-load, the command will:
1. Build the image using the configured registry
2. Load the built image into the Kind cluster defined in kindenv.yaml
This requires a kindenv.yaml file in the current directory and an existing Kind
cluster with the name specified in that configuration.

The command automatically detects if you're using Docker or Podman as your container
engine and uses the appropriate method to load the image into the Kind cluster.

Examples:
  devhelper-cli tw build
  devhelper-cli tw build --tag v1.0.0
  devhelper-cli tw build --arg KEY=VALUE
  devhelper-cli tw build --no-cache
  devhelper-cli tw build --kind-load`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create colored output helpers
		green := color.New(color.FgGreen).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()
		red := color.New(color.FgRed).SprintFunc()

		// Get flags
		configPath, _ := cmd.Flags().GetString("config")
		tag, _ := cmd.Flags().GetString("tag")
		noCache, _ := cmd.Flags().GetBool("no-cache")
		buildArgs, _ := cmd.Flags().GetStringArray("arg")
		verbose, _ := cmd.Flags().GetBool("verbose")
		kindLoad, _ := cmd.Flags().GetBool("kind-load")

		if verbose {
			fmt.Printf("Flags: config=%s, tag=%s, no-cache=%v, args=%v, kind-load=%v\n", 
				configPath, tag, noCache, buildArgs, kindLoad)
		}

		// Load configuration
		config, err := tw.LoadConfig(configPath)
		if err != nil {
			fmt.Printf("%s Error loading configuration: %v\n", red("❌"), err)
			os.Exit(1)
		}

		// Validate configuration
		if err := config.Validate(); err != nil {
			fmt.Printf("%s Configuration validation failed: %v\n", red("❌"), err)
			os.Exit(1)
		}

		// Check if Dockerfile exists
		if _, err := os.Stat("Dockerfile"); os.IsNotExist(err) {
			fmt.Printf("%s Dockerfile not found in the current directory\n", red("❌"))
			os.Exit(1)
		}

		// Get current directory name to use as the image name if not specified
		imageName := config.Metadata.Name
		if imageName == "" {
			currentDir, err := os.Getwd()
			if err != nil {
				fmt.Printf("%s Error getting current directory: %v\n", red("❌"), err)
				os.Exit(1)
			}
			imageName = filepath.Base(currentDir)
		}

		// Use image.registry from config if available
		registry := ""
		if config.Spec.Image.Registry != "" {
			registry = config.Spec.Image.Registry
			if !strings.HasSuffix(registry, "/") {
				registry += "/"
			}
		}

		// Use tag from config if not provided as a flag
		imageTag := tag
		if imageTag == "" {
			imageTag = config.Spec.Image.Tag
			if imageTag == "" {
				imageTag = "latest"
			}
		}

		fullImageName := fmt.Sprintf("%s%s:%s", registry, imageName, imageTag)
		fmt.Printf("%s Building Docker image: %s\n", green("🔨"), fullImageName)

		// Prepare podman build command
		buildCmd := []string{"podman", "build", "-t", fullImageName}

		// Add --no-cache if specified
		if noCache {
			buildCmd = append(buildCmd, "--no-cache")
		}

		// Add build args if specified
		for _, arg := range buildArgs {
			buildCmd = append(buildCmd, "--build-arg", arg)
		}

		// Add current directory as build context
		buildCmd = append(buildCmd, ".")

		// Execute podman build command
		execCmd := exec.Command(buildCmd[0], buildCmd[1:]...)
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr

		if verbose {
			fmt.Printf("Running command: %s\n", strings.Join(buildCmd, " "))
		}

		if err := execCmd.Run(); err != nil {
			fmt.Printf("%s Build failed: %v\n", red("❌"), err)
			os.Exit(1)
		}

		fmt.Printf("%s Docker image built successfully: %s\n", green("✅"), fullImageName)

		// Check if running in a Kind environment
		if kindLoad {
			// Figure out cluster name to use
			var kindClusterName string
			
			// First check for kindenv.yaml configuration
			kindEnvConfig, err := kindenv.LoadConfig("")
			if err == nil && kindEnvConfig.Cluster.Name != "" {
				kindClusterName = kindEnvConfig.Cluster.Name
				fmt.Printf("%s Using Kind cluster name from kindenv.yaml: %s\n", green("✅"), kindClusterName)
			} else {
				// Fall back to using the project name as the cluster name
				kindClusterName = imageName
				fmt.Printf("%s Using image name as Kind cluster name: %s\n", yellow("⚙️"), kindClusterName)
			}
		
			// Check if the kind command exists
			_, err = exec.LookPath("kind")
			if err != nil {
				fmt.Printf("%s Kind command not found: %v\n", red("❌"), err)
				os.Exit(1)
			}
		
			// Get list of clusters to check if ours exists
			kindCmd := exec.Command("kind", "get", "clusters")
			kindOutput, err := kindCmd.Output()
			if err != nil {
				fmt.Printf("%s Failed to get Kind clusters: %v\n", red("❌"), err)
				os.Exit(1)
			}
		
			// Check if our cluster exists in the list
			kindClusters := strings.Split(string(kindOutput), "\n")
			clusterExists := false
			for _, cluster := range kindClusters {
				if strings.TrimSpace(cluster) == kindClusterName {
					clusterExists = true
					break
				}
			}

			if !clusterExists {
				fmt.Printf("%s No Kind cluster found with name: %s\n", red("❌"), kindClusterName)
				os.Exit(1)
			}
		
			fmt.Printf("%s Loading image %s into Kind cluster: %s\n", yellow("⚙️"), fullImageName, kindClusterName)
		
			// Determine if we're using podman or docker
			containerRuntime := "docker"
			
			// First check for podman
			_, podmanErr := exec.LookPath("podman")
			if podmanErr == nil {
				containerRuntime = "podman"
			}
			
			// Then check for docker, prefer docker if both are available
			_, dockerErr := exec.LookPath("docker")
			if dockerErr == nil {
				containerRuntime = "docker"
				
				// Use podman if this is running in a podman environment
				// This detects if we're using kind with its experimental podman provider
				kindInfoCmd := exec.Command("kind", "version")
				kindInfoOutput, _ := kindInfoCmd.CombinedOutput()
				if strings.Contains(string(kindInfoOutput), "podman") {
					containerRuntime = "podman"
				}
			}
			
			if verbose {
				fmt.Printf("Using container runtime: %s\n", containerRuntime)
			}
			
			var loadCmd *exec.Cmd
			
			if containerRuntime == "podman" {
				fmt.Printf("%s Using podman to load image into Kind cluster...\n", yellow("⚙️"))
				// For podman, we need to save the image to a tarball first
				tempDir, err := os.MkdirTemp("", "podman-image-*")
				if err != nil {
					fmt.Printf("%s Failed to create temporary directory: %v\n", red("❌"), err)
					os.Exit(1)
				}
				defer os.RemoveAll(tempDir)
				
				tarballPath := filepath.Join(tempDir, "image.tar")
				fmt.Printf("%s Saving image to tarball...\n", yellow("⚙️"))
				
				// Save the image to a tarball
				saveCmd := exec.Command("podman", "save", "-o", tarballPath, fullImageName)
				saveCmd.Stdout = os.Stdout
				saveCmd.Stderr = os.Stderr
				if err := saveCmd.Run(); err != nil {
					fmt.Printf("%s Failed to save image to tarball: %v\n", red("❌"), err)
					os.Exit(1)
				}
				
				// Load the tarball into kind
				fmt.Printf("%s Loading tarball into Kind cluster...\n", yellow("⚙️"))
				loadCmd = exec.Command("kind", "load", "image-archive", tarballPath, "--name", kindClusterName)
			} else {
				// Using Docker - can load directly
				loadCmd = exec.Command("kind", "load", "docker-image", fullImageName, "--name", kindClusterName)
			}
			
			loadCmd.Stdout = os.Stdout
			loadCmd.Stderr = os.Stderr

			if err := loadCmd.Run(); err != nil {
				fmt.Printf("%s Failed to load image %s into Kind cluster: %v\n", red("❌"), fullImageName, err)
				os.Exit(1)
			} 
			
			fmt.Printf("%s Image %s loaded into Kind cluster %s successfully\n", green("✅"), fullImageName, kindClusterName)
		}
	},
}

func init() {
	twCmd.AddCommand(twBuildCmd)

	// Add flags for tw build command
	twBuildCmd.Flags().String("tag", "", "Image tag (default: 'latest' or from config)")
	twBuildCmd.Flags().Bool("no-cache", false, "Do not use cache when building the image")
	twBuildCmd.Flags().StringArray("arg", []string{}, "Build arguments for Docker (KEY=VALUE)")
	twBuildCmd.Flags().Bool("kind-load", false, "Load the built image into the Kind cluster defined in kindenv.yaml")
}
