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
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
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

		// Define status checks for each component
		components := []struct {
			Name          string
			CheckCommand  string
			CheckArgs     []string
			StatusMessage string
			Available     bool
		}{
			{
				Name:         "Docker",
				CheckCommand: "docker",
				CheckArgs:    []string{"ps", "--format", "{{.Names}} - {{.Status}}"},
				Available:    isCommandAvailable("docker"),
			},
			{
				Name:         "Dapr",
				CheckCommand: "dapr",
				CheckArgs:    []string{"status", "-k"},
				Available:    isCommandAvailable("dapr"),
			},
			{
				Name:         "Temporal",
				CheckCommand: "temporal",
				CheckArgs:    []string{"server", "list"},
				Available:    isCommandAvailable("temporal"),
			},
		}

		fmt.Println("\n=== Local Environment Status ===")

		allRunning := true
		anyAvailable := false

		for _, comp := range components {
			if !comp.Available {
				fmt.Printf("❌ %s: Not installed\n", comp.Name)
				allRunning = false
				continue
			}

			anyAvailable = true

			// Run the check command
			cmd := exec.Command(comp.CheckCommand, comp.CheckArgs...)
			output, err := cmd.CombinedOutput()
			outputStr := strings.TrimSpace(string(output))

			if err != nil {
				fmt.Printf("❌ %s: Not running\n", comp.Name)
				if verbose && outputStr != "" {
					fmt.Printf("   Details: %s\n", outputStr)
				}
				allRunning = false
				continue
			}

			fmt.Printf("✅ %s: Running\n", comp.Name)
			if verbose && outputStr != "" {
				// Format and print relevant output details
				lines := strings.Split(outputStr, "\n")
				if len(lines) > 5 && !verbose {
					// Truncate output if it's too long and we're not in verbose mode
					for i := 0; i < 3; i++ {
						fmt.Printf("   %s\n", lines[i])
					}
					fmt.Printf("   ... %d more lines ...\n", len(lines)-3)
				} else {
					for _, line := range lines {
						fmt.Printf("   %s\n", line)
					}
				}
			}
		}

		if !anyAvailable {
			fmt.Println("\n❌ No components are installed. Please install the required components.")
			fmt.Println("See README.md for installation instructions.")
			return
		}

		fmt.Println("\n=== Summary ===")
		if allRunning {
			fmt.Println("✅ All components are running properly.")
		} else {
			fmt.Println("⚠️  Some components are not running.")
			fmt.Println("Run 'shielddev-cli localenv start' to start the environment.")
		}
	},
}

func init() {
	localenvCmd.AddCommand(localenvStatusCmd)
}
