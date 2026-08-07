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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ShieldFC-RD/devhelper-cli/internal/tw"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// twRunCmd represents the tw run command
var twRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a Temporal Worker locally",
	Long: `Run a Temporal Worker locally for development.

This command runs a Temporal Worker using the configuration from tw.yaml.
It stores the worker settings in Redis to match how the operator would
do it in a production environment.

Examples:
  devhelper-cli tw run
  devhelper-cli tw run --local
  devhelper-cli tw run --env dev`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create colored output helpers
		green := color.New(color.FgGreen).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()
		red := color.New(color.FgRed).SprintFunc()

		// Get flags
		configPath, _ := cmd.Flags().GetString("config")
		local, _ := cmd.Flags().GetBool("local")
		env, _ := cmd.Flags().GetString("env")
		verbose, _ := cmd.Flags().GetBool("verbose")

		if verbose {
			fmt.Printf("Flags: config=%s, local=%v, env=%s\n", configPath, local, env)
		}

		// Add environment suffix to config path if env is specified
		if env != "" {
			// Check if configPath has an extension
			ext := ""
			if idx := strings.LastIndex(configPath, "."); idx != -1 {
				ext = configPath[idx:]
				configPath = configPath[:idx]
			}
			configPath = fmt.Sprintf("%s-%s%s", configPath, env, ext)

			if verbose {
				fmt.Printf("Using environment-specific config: %s\n", configPath)
			}
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

		fmt.Println(green("Running Temporal Worker locally..."))

		// Check if local option is specified
		if local {
			fmt.Println(yellow("Running in local mode (without Redis configuration)"))
			// TODO: Implement local run without Redis
			fmt.Println(yellow("Local mode is not fully implemented yet"))
		} else {
			// Write configuration to Redis to simulate operator behavior
			fmt.Println(yellow("Writing worker settings to Redis..."))

			// Check if Redis is available
			if !isRedisAvailable() {
				fmt.Printf("%s Redis is not available. Make sure Redis is running or use --local flag.\n", red("❌"))
				os.Exit(1)
			}

			// Create redis key based on worker name
			redisKey := fmt.Sprintf("configuration||%s", config.Metadata.Name)

			// Parse the YAML workerSettings to a map
			var settingsMap map[string]interface{}
			if err := yaml.Unmarshal([]byte(config.Spec.WorkerSettings), &settingsMap); err != nil {
				fmt.Printf("%s Failed to parse workerSettings YAML: %v\n", red("❌"), err)
				os.Exit(1)
			}

			// Convert the settings map to JSON
			settingsJSON, err := json.Marshal(settingsMap)
			if err != nil {
				fmt.Printf("%s Failed to serialize settings to JSON: %v\n", red("❌"), err)
				os.Exit(1)
			}

			// Run redis-cli to set the key
			setCmd := exec.Command("redis-cli", "set", redisKey, string(settingsJSON))
			setOutput, err := setCmd.CombinedOutput()
			if err != nil || !strings.Contains(string(setOutput), "OK") {
				fmt.Printf("%s Failed to write settings to Redis: %v\n", red("❌"), err)
				if len(setOutput) > 0 {
					fmt.Printf("Output: %s\n", string(setOutput))
				}
				os.Exit(1)
			}

			fmt.Printf("%s Worker settings written to Redis with key: %s\n", green("✅"), redisKey)
		}

		// Get the worker project directory
		workingDir, err := os.Getwd()
		if err != nil {
			fmt.Printf("%s Failed to get current directory: %v\n", red("❌"), err)
			os.Exit(1)
		}

		// TODO: Based on the project, determine how to run the worker
		// For now, check if there's a main binary, npm start, etc.

		// Check if Dockerfile exists and suggest building first
		if _, err := os.Stat("Dockerfile"); err == nil {
			fmt.Println(yellow("Dockerfile detected. You may need to build the image first:"))
			fmt.Println("  devhelper-cli tw build")
		}

		// Check if package.json exists (Node.js project)
		if _, err := os.Stat("package.json"); err == nil {
			fmt.Println(yellow("Detected Node.js project, attempting to run..."))

			// Check if node_modules exists, if not suggest npm install
			if _, err := os.Stat("node_modules"); os.IsNotExist(err) {
				fmt.Println(yellow("node_modules not found, running npm install first..."))

				cmd := exec.Command("npm", "install")
				cmd.Dir = workingDir
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				if err := cmd.Run(); err != nil {
					fmt.Printf("%s Failed to run npm install: %v\n", red("❌"), err)
					os.Exit(1)
				}
			}

			// Run with npm start
			npmCmd := exec.Command("npm", "start")
			npmCmd.Dir = workingDir
			npmCmd.Stdout = os.Stdout
			npmCmd.Stderr = os.Stderr

			fmt.Printf("%s Starting worker with npm start\n", green("✅"))
			if err := npmCmd.Run(); err != nil {
				fmt.Printf("%s Failed to run worker: %v\n", red("❌"), err)
				os.Exit(1)
			}

			return
		}

		// Check if go.mod exists (Go project)
		if _, err := os.Stat("go.mod"); err == nil {
			fmt.Println(yellow("Detected Go project, attempting to run..."))

			// Run with go run
			goCmd := exec.Command("go", "run", ".")
			goCmd.Dir = workingDir
			goCmd.Stdout = os.Stdout
			goCmd.Stderr = os.Stderr

			fmt.Printf("%s Starting worker with go run\n", green("✅"))
			if err := goCmd.Run(); err != nil {
				fmt.Printf("%s Failed to run worker: %v\n", red("❌"), err)
				os.Exit(1)
			}

			return
		}

		// If we get here, we couldn't determine how to run the worker
		fmt.Printf("%s Unable to determine how to run the worker.\n", red("❌"))
		fmt.Println("Please build and run the worker manually, or use:")
		fmt.Println("  devhelper-cli tw build")
		fmt.Println("  devhelper-cli tw deploy --kind")
	},
}

// isRedisAvailable checks if Redis is available
func isRedisAvailable() bool {
	// Try to ping Redis using redis-cli
	cmd := exec.Command("redis-cli", "ping")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}

	// Check if the output contains PONG
	return strings.Contains(string(output), "PONG")
}

func init() {
	twCmd.AddCommand(twRunCmd)

	// Add flags for tw run command
	twRunCmd.Flags().Bool("local", false, "Run locally without Redis configuration")
	twRunCmd.Flags().String("env", "", "Environment to use (dev, staging, prod)")
}
