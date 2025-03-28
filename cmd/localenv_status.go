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
	"net/http"
	"os/exec"
	"strings"
	"time"

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
			WebUIURL      string
			CheckUI       bool
		}{
			{
				Name:         "Podman",
				CheckCommand: "podman",
				CheckArgs:    []string{"ps", "--format", "{{.Names}} - {{.Status}}"},
				Available:    isCommandAvailable("podman"),
			},
			{
				Name:         "Kind",
				CheckCommand: "kind",
				CheckArgs:    []string{"get", "clusters"},
				Available:    isCommandAvailable("kind"),
			},
			{
				Name:         "Dapr",
				CheckCommand: "dapr",
				CheckArgs:    []string{"list"},
				Available:    isCommandAvailable("dapr"),
				WebUIURL:     "", // Dashboard not available in slim installation
				CheckUI:      false,
			},
			{
				Name:         "Temporal",
				CheckCommand: "temporal",
				CheckArgs:    []string{"workflow", "list"},
				Available:    isCommandAvailable("temporal"),
				WebUIURL:     "http://localhost:8233",
				CheckUI:      true,
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
				if comp.Name == "Podman" {
					fmt.Printf("❌ %s: Not able to run containers\n", comp.Name)
				} else if comp.Name == "Kind" {
					fmt.Printf("❌ %s: No clusters available\n", comp.Name)
				} else {
					fmt.Printf("❌ %s: Not running\n", comp.Name)
				}

				if verbose && outputStr != "" {
					fmt.Printf("   Details: %s\n", outputStr)
				}
				allRunning = false
				continue
			}

			if comp.Name == "Podman" {
				fmt.Printf("✅ %s: Available (can run containers)\n", comp.Name)
			} else if comp.Name == "Kind" {
				fmt.Printf("✅ %s: Available (has clusters)\n", comp.Name)
			} else {
				fmt.Printf("✅ %s: Running\n", comp.Name)
			}

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

		if !anyAvailable {
			fmt.Println("\n❌ No components are installed. Please install the required components.")
			fmt.Println("See README.md for installation instructions.")
			return
		}

		fmt.Println("\n=== Summary ===")
		if allRunning {
			fmt.Println("✅ All components are running properly.")
			fmt.Println("Temporal UI: http://localhost:8233")
			// Dapr Dashboard not available in slim installation
		} else {
			fmt.Println("⚠️  Some components are not running.")
			fmt.Println("Run 'shielddev-cli localenv start' to start the environment.")
		}
	},
}

func init() {
	localenvCmd.AddCommand(localenvStatusCmd)
}
