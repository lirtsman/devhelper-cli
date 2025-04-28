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

// twConfigCmd represents the tw config command
var twConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Temporal Worker configuration",
	Long: `Manage Temporal Worker configuration.

This command allows you to view and modify the temporal worker configuration.
You can view the entire configuration, get specific values, or set new values.

Examples:
  devhelper-cli tw config view
  devhelper-cli tw config get spec.temporal.namespace
  devhelper-cli tw config set spec.temporal.namespace=my-namespace
  devhelper-cli tw config --profile dev
  devhelper-cli tw config write --redis localhost:6379`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default action is to view the configuration
		viewConfig(cmd)
	},
}

// viewConfig displays the current configuration
func viewConfig(cmd *cobra.Command) {
	// Create colored output helpers
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	// Get the config file path from flags
	configPath, _ := cmd.Flags().GetString("config")
	profile, _ := cmd.Flags().GetString("profile")

	// Add profile suffix if profile is specified
	if profile != "" {
		// Check if configPath has an extension
		ext := ""
		if idx := strings.LastIndex(configPath, "."); idx != -1 {
			ext = configPath[idx:]
			configPath = configPath[:idx]
		}
		configPath = fmt.Sprintf("%s-%s%s", configPath, profile, ext)
	}

	// Load the configuration
	config, err := tw.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("%s Error loading configuration: %v\n", red("❌"), err)
		os.Exit(1)
	}

	// Convert config to YAML for display
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		fmt.Printf("%s Error marshaling configuration to YAML: %v\n", red("❌"), err)
		os.Exit(1)
	}

	// Print the configuration
	fmt.Println(green("Current Temporal Worker configuration:"))
	fmt.Println(string(yamlData))
}

// getConfigValue gets a specific value from the configuration
func getConfigValue(cmd *cobra.Command, key string) {
	// Create colored output helpers
	red := color.New(color.FgRed).SprintFunc()

	// Get the config file path from flags
	configPath, _ := cmd.Flags().GetString("config")
	profile, _ := cmd.Flags().GetString("profile")

	// Add profile suffix if profile is specified
	if profile != "" {
		// Check if configPath has an extension
		ext := ""
		if idx := strings.LastIndex(configPath, "."); idx != -1 {
			ext = configPath[idx:]
			configPath = configPath[:idx]
		}
		configPath = fmt.Sprintf("%s-%s%s", configPath, profile, ext)
	}

	// Load the configuration
	config, err := tw.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("%s Error loading configuration: %v\n", red("❌"), err)
		os.Exit(1)
	}

	// TODO: Implement traversing the config object to get the specific value
	// This will require parsing the key path (e.g., "spec.temporal.namespace")
	// and accessing the nested values dynamically

	fmt.Printf("Getting value for key %s is not implemented yet\n", key)
	_ = config // Temporary fix to avoid unused variable error
}

// setConfigValue sets a specific value in the configuration
func setConfigValue(cmd *cobra.Command, keyValue string) {
	// Create colored output helpers
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	// Split the key=value string
	parts := strings.SplitN(keyValue, "=", 2)
	if len(parts) != 2 {
		fmt.Printf("%s Invalid format. Use key=value format (e.g., spec.temporal.namespace=my-namespace)\n", red("❌"))
		os.Exit(1)
	}

	key := parts[0]
	value := parts[1]

	// Get the config file path from flags
	configPath, _ := cmd.Flags().GetString("config")
	profile, _ := cmd.Flags().GetString("profile")

	// Add profile suffix if profile is specified
	if profile != "" {
		// Check if configPath has an extension
		ext := ""
		if idx := strings.LastIndex(configPath, "."); idx != -1 {
			ext = configPath[idx:]
			configPath = configPath[:idx]
		}
		configPath = fmt.Sprintf("%s-%s%s", configPath, profile, ext)
	}

	// Load the configuration
	config, err := tw.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("%s Error loading configuration: %v\n", red("❌"), err)
		os.Exit(1)
	}

	// TODO: Implement traversing the config object to set the specific value
	// This will require parsing the key path (e.g., "spec.temporal.namespace")
	// and modifying the nested values dynamically

	fmt.Printf("Setting value %s=%s is not implemented yet\n", key, value)

	// Save the configuration
	err = tw.SaveConfig(config, configPath)
	if err != nil {
		fmt.Printf("%s Error saving configuration: %v\n", red("❌"), err)
		os.Exit(1)
	}

	fmt.Printf("%s Configuration updated successfully\n", green("✅"))
}

// writeConfigToRedis writes the worker settings to Redis
func writeConfigToRedis(cmd *cobra.Command, args []string) {
	// Create colored output helpers
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	// Get flags
	configPath, _ := cmd.Flags().GetString("config")
	redisAddr, _ := cmd.Flags().GetString("redis")
	profile, _ := cmd.Flags().GetString("profile")
	verbose, _ := cmd.Flags().GetBool("verbose")

	if verbose {
		fmt.Printf("Writing configuration from %s to Redis at %s\n", configPath, redisAddr)
	}

	// Add profile suffix if profile is specified
	if profile != "" {
		// Check if configPath has an extension
		ext := ""
		if idx := strings.LastIndex(configPath, "."); idx != -1 {
			ext = configPath[idx:]
			configPath = configPath[:idx]
		}
		configPath = fmt.Sprintf("%s-%s%s", configPath, profile, ext)
	}

	// Load the configuration
	config, err := tw.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("%s Error loading configuration: %v\n", red("❌"), err)
		os.Exit(1)
	}

	// Skip if worker settings is empty
	if config.Spec.WorkerSettings == "" {
		fmt.Printf("%s No worker settings defined in the configuration\n", yellow("⚠️"))
		return
	}

	// Create redis key based on worker name
	redisKey := fmt.Sprintf("configuration||%s", config.Metadata.Name)

	// Parse the YAML workerSettings to a map
	var settingsMap map[string]interface{}
	if err := yaml.Unmarshal([]byte(config.Spec.WorkerSettings), &settingsMap); err != nil {
		fmt.Printf("%s Failed to parse workerSettings YAML: %v\n", red("❌"), err)
		os.Exit(1)
	}

	// Check if redis-cli is available
	if _, err := exec.LookPath("redis-cli"); err != nil {
		fmt.Printf("%s redis-cli not found. Please install Redis CLI tools\n", red("❌"))
		os.Exit(1)
	}

	// Convert the settings map to JSON
	settingsJSON, err := json.Marshal(settingsMap)
	if err != nil {
		fmt.Printf("%s Failed to serialize settings to JSON: %v\n", red("❌"), err)
		os.Exit(1)
	}

	// Check if Redis is available at the specified address
	redisCliArgs := []string{"ping"}
	if redisAddr != "" {
		host, port, err := parseRedisAddress(redisAddr)
		if err != nil {
			fmt.Printf("%s Invalid Redis address: %v\n", red("❌"), err)
			os.Exit(1)
		}
		redisCliArgs = append(redisCliArgs, "-h", host, "-p", port)
	}

	pingCmd := exec.Command("redis-cli", redisCliArgs...)
	pingOutput, err := pingCmd.CombinedOutput()
	if err != nil || !strings.Contains(string(pingOutput), "PONG") {
		fmt.Printf("%s Redis is not available at %s: %v\n", red("❌"), redisAddr, err)
		os.Exit(1)
	}

	// Prepare redis-cli command for setting the key
	redisCliSetArgs := []string{"set", redisKey}
	if redisAddr != "" {
		host, port, _ := parseRedisAddress(redisAddr)
		redisCliSetArgs = []string{"-h", host, "-p", port, "set", redisKey}
	}

	// Add the JSON string as the value
	redisCliSetArgs = append(redisCliSetArgs, string(settingsJSON))

	if verbose {
		fmt.Printf("Redis key: %s\n", redisKey)
		fmt.Printf("Settings JSON: %s\n", string(settingsJSON))
	}

	// Run redis-cli to set the key
	setCmd := exec.Command("redis-cli", redisCliSetArgs...)
	setOutput, err := setCmd.CombinedOutput()
	if err != nil || !strings.Contains(string(setOutput), "OK") {
		fmt.Printf("%s Failed to write settings to Redis: %v\n", red("❌"), err)
		if len(setOutput) > 0 {
			fmt.Printf("Output: %s\n", string(setOutput))
		}
		os.Exit(1)
	}

	fmt.Printf("%s Successfully wrote worker settings to Redis with key: %s\n", green("✅"), redisKey)
}

// parseRedisAddress parses a Redis address in the format host:port
// and returns the host and port as separate strings
func parseRedisAddress(addr string) (string, string, error) {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid Redis address format, expected host:port")
	}

	host := parts[0]
	port := parts[1]

	// Use default Redis port if not specified
	if port == "" {
		port = "6379"
	}

	return host, port, nil
}

// twConfigViewCmd represents the config view subcommand
var twConfigViewCmd = &cobra.Command{
	Use:   "view",
	Short: "View the current Temporal Worker configuration",
	Long:  `Display the current Temporal Worker configuration in YAML format.`,
	Run: func(cmd *cobra.Command, args []string) {
		viewConfig(cmd)
	},
}

// twConfigGetCmd represents the config get subcommand
var twConfigGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a specific value from the configuration",
	Long:  `Get a specific value from the Temporal Worker configuration by key path.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		getConfigValue(cmd, args[0])
	},
}

// twConfigSetCmd represents the config set subcommand
var twConfigSetCmd = &cobra.Command{
	Use:   "set [key=value]",
	Short: "Set a specific value in the configuration",
	Long:  `Set a specific value in the Temporal Worker configuration by key path.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setConfigValue(cmd, args[0])
	},
}

// twConfigWriteCmd represents the config write subcommand
var twConfigWriteCmd = &cobra.Command{
	Use:   "write",
	Short: "Write worker settings to Redis",
	Long: `Write worker settings from the configuration file to Redis.

This command parses the workerSettings section from the configuration file,
converts it to JSON, and writes it to Redis with the key 'configuration||{worker-name}'.
This simulates how the Temporal Worker Operator would store configuration.

Examples:
  devhelper-cli tw config write
  devhelper-cli tw config write --redis localhost:6379
  devhelper-cli tw config write --profile dev`,
	Run: writeConfigToRedis,
}

func init() {
	twCmd.AddCommand(twConfigCmd)

	// Add subcommands to twConfigCmd
	twConfigCmd.AddCommand(twConfigViewCmd)
	twConfigCmd.AddCommand(twConfigGetCmd)
	twConfigCmd.AddCommand(twConfigSetCmd)
	twConfigCmd.AddCommand(twConfigWriteCmd)

	// Add flags for tw config command
	twConfigCmd.PersistentFlags().String("profile", "", "Configuration profile (dev, staging, prod)")

	// Add flags for tw config write command
	twConfigWriteCmd.Flags().String("redis", "", "Redis address in the format host:port (default: localhost:6379)")
}
