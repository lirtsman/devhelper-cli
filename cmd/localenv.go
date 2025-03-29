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
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// Helper function that can be used by subcommands to check if a command is installed
// Make it a variable holding a function so it can be mocked in tests
var isCommandAvailable = func(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// localenvCmd represents the localenv command
var localenvCmd = &cobra.Command{
	Use:   "localenv",
	Short: "Manage local development environment",
	Long: `Manage the local development environment for ShieldDev applications.

The localenv command provides functionality to start, stop, and manage 
local development components including:
- Dapr runtime
- Temporal server
- Required dependencies and infrastructure

This allows developers to run and test ShieldDev applications locally.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Use one of the localenv subcommands. Run 'devhelper-cli localenv --help' for usage.")
	},
}

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs [component]",
	Short: "View logs for a local development component",
	Long: `View logs for a locally running component like Temporal server.
Optionally follow the logs in real-time.

Examples:
  devhelper-cli localenv logs temporal      # View Temporal server logs
  devhelper-cli localenv logs temporal -f   # Follow Temporal server logs
`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Determine which component logs to view
		component := "temporal" // Default to temporal
		if len(args) > 0 {
			component = args[0]
		}

		follow, _ := cmd.Flags().GetBool("follow")
		verbose, _ := cmd.Flags().GetBool("verbose")
		lines, _ := cmd.Flags().GetInt("lines")

		// Base logs directory
		logsDir := filepath.Join(os.Getenv("HOME"), ".logs", "devhelper-cli")
		if _, err := os.Stat(logsDir); os.IsNotExist(err) {
			fmt.Printf("❌ Logs directory not found: %s\n", logsDir)
			fmt.Println("   No logs have been generated yet. Try starting the component first.")
			os.Exit(1)
		}

		var logPath string

		// Determine the path to the log file based on component
		switch component {
		case "temporal", "temporal-server":
			logPath = filepath.Join(logsDir, "temporal-server.log")
		default:
			fmt.Printf("❌ Unknown component: %s\n", component)
			fmt.Println("   Supported components: temporal")
			os.Exit(1)
		}

		// Check if log file exists
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			fmt.Printf("❌ Log file not found for component '%s': %s\n", component, logPath)
			fmt.Printf("   Try starting %s first using 'devhelper-cli localenv start'\n", component)
			os.Exit(1)
		}

		// If follow is false, just display the last N lines of the file
		if !follow {
			// Use tail to display the last N lines
			tailCmd := exec.Command("tail", "-n", fmt.Sprintf("%d", lines), logPath)
			tailCmd.Stdout = os.Stdout
			tailCmd.Stderr = os.Stderr
			if err := tailCmd.Run(); err != nil {
				if verbose {
					fmt.Printf("Error running tail: %v\n", err)
				}
				// Fallback to Go implementation if tail fails
				displayLastNLines(logPath, lines)
			}
			return
		}

		// For follow mode, use tail -f to stream the logs in real-time
		fmt.Printf("Following logs for %s. Press Ctrl+C to stop...\n", component)
		tailCmd := exec.Command("tail", "-f", "-n", fmt.Sprintf("%d", lines), logPath)
		tailCmd.Stdout = os.Stdout
		tailCmd.Stderr = os.Stderr
		if err := tailCmd.Run(); err != nil {
			fmt.Printf("Error following logs: %v\n", err)
		}
	},
}

// displayLastNLines reads the last N lines from a file
// Used as a fallback if the tail command is not available
func displayLastNLines(filePath string, n int) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := make([]string, 0, n)

	// Scan all lines, keeping only the last n
	for scanner.Scan() {
		if len(lines) >= n {
			lines = append(lines[1:], scanner.Text())
		} else {
			lines = append(lines, scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading log file: %v\n", err)
		return
	}

	for _, line := range lines {
		fmt.Println(line)
	}
}

func init() {
	rootCmd.AddCommand(localenvCmd)

	// Add logs command
	localenvCmd.AddCommand(logsCmd)

	// Add flags for logs command
	logsCmd.Flags().BoolP("follow", "f", false, "Follow the logs (like tail -f)")
	logsCmd.Flags().IntP("lines", "n", 50, "Number of lines to display")

	// Add persistent flags that are available to all subcommands
	localenvCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// localenvCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// localenvCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
