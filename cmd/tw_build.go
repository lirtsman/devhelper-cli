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

Examples:
  devhelper-cli tw build
  devhelper-cli tw build --tag v1.0.0
  devhelper-cli tw build --arg KEY=VALUE
  devhelper-cli tw build --no-cache`,
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

		if verbose {
			fmt.Printf("Flags: config=%s, tag=%s, no-cache=%v, args=%v\n", configPath, tag, noCache, buildArgs)
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
		registry := config.Spec.Image.Registry
		if registry != "" && !strings.HasSuffix(registry, "/") {
			registry += "/"
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
		kindCluster, _ := cmd.Flags().GetString("kind-load")
		if kindCluster != "" {
			fmt.Printf("%s Loading image into Kind cluster: %s\n", yellow("⚙️"), kindCluster)

			// Execute kind load command
			loadCmd := exec.Command("kind", "load", "docker-image", fullImageName, "--name", kindCluster)
			loadCmd.Stdout = os.Stdout
			loadCmd.Stderr = os.Stderr

			if err := loadCmd.Run(); err != nil {
				fmt.Printf("%s Failed to load image into Kind cluster: %v\n", red("❌"), err)
				os.Exit(1)
			}

			fmt.Printf("%s Image loaded into Kind cluster successfully\n", green("✅"))
		}
	},
}

func init() {
	twCmd.AddCommand(twBuildCmd)

	// Add flags for tw build command
	twBuildCmd.Flags().String("tag", "", "Image tag (default: 'latest' or from config)")
	twBuildCmd.Flags().Bool("no-cache", false, "Do not use cache when building the image")
	twBuildCmd.Flags().StringArray("arg", []string{}, "Build arguments for Docker (KEY=VALUE)")
	twBuildCmd.Flags().String("kind-load", "", "Load the image into a Kind cluster after building")
}
