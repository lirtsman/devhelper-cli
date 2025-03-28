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
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// Components to be started
type Component struct {
	Name            string
	Command         string
	Args            []string
	CheckCommand    string
	CheckArgs       []string
	RequiredFor     []string
	StartupDelay    time.Duration
	IsRunning       bool
	IsRequired      bool
	CommandExists   bool
	VerifyAvailable func() bool // Function to verify the component is accessible
	RequiresStartup bool        // Whether the component needs to be started or just verified
	IsBinary        bool        // Whether the component is a binary command (like Podman, Kind) rather than a service
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start local development environment",
	Long: `Start a local development environment with all necessary components
for ShieldDev application development including:

- Dapr runtime
- Temporal server
- Required dependencies

This command will check for necessary dependencies and start them
in the correct order.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting local development environment...")

		verbose, _ := cmd.Flags().GetBool("verbose")
		skipDapr, _ := cmd.Flags().GetBool("skip-dapr")
		skipTemporal, _ := cmd.Flags().GetBool("skip-temporal")
		wait, _ := cmd.Flags().GetBool("wait")

		// Function to check if Temporal server is accessible
		checkTemporalServerRunning := func() bool {
			checkCmd := exec.Command("temporal", "workflow", "list")
			if err := checkCmd.Run(); err != nil {
				if verbose {
					fmt.Printf("Temporal server check failed: %v\n", err)
				}
				return false
			}
			return true
		}

		// Function to check if Dapr is accessible
		checkDaprRunning := func() bool {
			// Use 'dapr list' instead of 'dapr status' which requires -k flag
			checkCmd := exec.Command("dapr", "list")
			output, err := checkCmd.CombinedOutput()
			if err != nil {
				if verbose {
					fmt.Printf("Dapr check failed: %v\n", err)
					fmt.Printf("Output: %s\n", string(output))
				}
				return false
			}

			// Check if the output indicates Dapr is running
			// Even if no apps are running, the command will succeed if Dapr is available
			return true
		}

		// Function to check if Podman is running
		checkPodmanRunning := func() bool {
			checkCmd := exec.Command("podman", "ps")
			if err := checkCmd.Run(); err != nil {
				if verbose {
					fmt.Printf("Podman check failed: %v\n", err)
				}
				return false
			}
			return true
		}

		// Function to check if Kind has clusters
		checkKindRunning := func() bool {
			checkCmd := exec.Command("kind", "get", "clusters")
			output, err := checkCmd.CombinedOutput()
			if err != nil {
				if verbose {
					fmt.Printf("Kind check failed: %v\n", err)
				}
				return false
			}
			// Check if there's at least one cluster
			return len(strings.TrimSpace(string(output))) > 0
		}

		// Define the components we need to start
		components := []Component{
			{
				Name:            "Podman",
				Command:         "podman",
				Args:            []string{"--version"},
				CheckCommand:    "podman",
				CheckArgs:       []string{"ps"},
				RequiredFor:     []string{"Kind", "Dapr", "Temporal"},
				StartupDelay:    0,
				IsRequired:      true,
				CommandExists:   isCommandAvailable("podman"),
				VerifyAvailable: checkPodmanRunning,
				RequiresStartup: false, // We don't start Podman, just verify it's running
				IsBinary:        true,
			},
			{
				Name:            "Kind",
				Command:         "kind",
				Args:            []string{"--version"},
				CheckCommand:    "kind",
				CheckArgs:       []string{"get", "clusters"},
				RequiredFor:     []string{"Dapr", "Temporal"},
				StartupDelay:    0,
				IsRequired:      true,
				CommandExists:   isCommandAvailable("kind"),
				VerifyAvailable: checkKindRunning,
				RequiresStartup: false, // We don't start Kind, just verify it's configured
				IsBinary:        true,
			},
			{
				Name:            "Dapr",
				Command:         "dapr",
				Args:            []string{"init", "--slim"},
				CheckCommand:    "dapr",
				CheckArgs:       []string{"status"},
				RequiredFor:     []string{},
				StartupDelay:    2 * time.Second,
				IsRequired:      !skipDapr,
				CommandExists:   isCommandAvailable("dapr"),
				VerifyAvailable: checkDaprRunning,
				RequiresStartup: true, // Dapr needs to be started
				IsBinary:        false,
			},
			{
				Name:            "Temporal",
				Command:         "temporal",
				Args:            []string{"server", "start-dev"},
				CheckCommand:    "temporal",
				CheckArgs:       []string{"operator", "list", "--namespace", "default"},
				RequiredFor:     []string{},
				StartupDelay:    5 * time.Second, // Increased delay for Temporal to start fully
				IsRequired:      !skipTemporal,
				CommandExists:   isCommandAvailable("temporal"),
				VerifyAvailable: checkTemporalServerRunning,
				RequiresStartup: true, // Temporal needs to be started
				IsBinary:        false,
			},
		}

		// First, check if required components are installed
		allInstalled := true
		for _, comp := range components {
			if comp.IsRequired && !comp.CommandExists {
				fmt.Printf("❌ Required component '%s' is not installed or not in PATH.\n", comp.Name)
				allInstalled = false
			} else if verbose {
				fmt.Printf("✅ Component '%s' is installed.\n", comp.Name)
			}
		}

		if !allInstalled {
			fmt.Println("\nSome required components are missing. Please install them and try again.")
			fmt.Println("See README.md for installation instructions.")
			os.Exit(1)
		}

		// Next, check dependencies between components
		for i, comp := range components {
			if !comp.IsRequired {
				if verbose {
					fmt.Printf("⏭️  Skipping '%s' as it's not required.\n", comp.Name)
				}
				continue
			}

			// Check if any required components are missing
			missingDeps := false
			for _, dep := range comp.RequiredFor {
				for _, depComp := range components {
					if depComp.Name == dep && !depComp.CommandExists {
						fmt.Printf("❌ '%s' requires '%s', but it's not installed.\n", comp.Name, dep)
						missingDeps = true
					}
				}
			}

			if missingDeps {
				continue
			}

			// Handle components differently based on RequiresStartup
			if !comp.RequiresStartup {
				// For components like Podman and Kind, just check if they're running
				if comp.IsBinary {
					fmt.Printf("Checking if %s is available...\n", comp.Name)
					if comp.VerifyAvailable != nil && comp.VerifyAvailable() {
						if comp.Name == "Podman" {
							fmt.Printf("✅ %s is available and can run containers.\n", comp.Name)
						} else if comp.Name == "Kind" {
							fmt.Printf("✅ %s is available and has clusters configured.\n", comp.Name)
						} else {
							fmt.Printf("✅ %s is available.\n", comp.Name)
						}
						components[i].IsRunning = true
					} else {
						if comp.Name == "Podman" {
							fmt.Printf("❌ %s is not working properly.\n", comp.Name)
							fmt.Println("   Make sure Podman is installed correctly and has proper permissions.")
						} else if comp.Name == "Kind" {
							fmt.Printf("❌ %s does not have any clusters configured.\n", comp.Name)
							fmt.Println("   Create a cluster with: 'kind create cluster --name my-cluster'")
						} else {
							fmt.Printf("❌ %s is not available.\n", comp.Name)
						}
						allInstalled = false
					}
				} else {
					// For non-binary components that still don't require startup
					fmt.Printf("Verifying %s is ready...\n", comp.Name)
					if comp.VerifyAvailable != nil && comp.VerifyAvailable() {
						fmt.Printf("✅ %s is running and ready.\n", comp.Name)
						components[i].IsRunning = true
					} else {
						fmt.Printf("❌ %s is not running or not ready.\n", comp.Name)
						allInstalled = false
					}
				}
				continue
			}

			// For components that need to be started (Dapr, Temporal)
			fmt.Printf("Starting %s...\n", comp.Name)

			// Special handling for Dapr - check if it's already initialized
			if comp.Name == "Dapr" {
				// First check if Dapr is already running
				if checkDaprRunning() {
					fmt.Println("✅ Dapr is already running, skipping initialization.")
					components[i].IsRunning = true
					continue
				}

				// Run dapr init with --slim
				initCmd := exec.Command("dapr", "init", "--slim")
				initOutput, err := initCmd.CombinedOutput()
				if err != nil {
					fmt.Printf("❌ Failed to initialize Dapr: %v\n", err)
					if verbose {
						fmt.Printf("Output: %s\n", string(initOutput))
					}
					continue
				}

				if verbose {
					fmt.Printf("Dapr initialization output: %s\n", string(initOutput))
				}

				// Wait a moment for Dapr to start
				time.Sleep(2 * time.Second)

				// Verify Dapr is running
				if checkDaprRunning() {
					fmt.Println("✅ Dapr started successfully.")
					components[i].IsRunning = true
				} else {
					fmt.Println("⚠️ Dapr initialization completed, but the runtime may not be fully ready.")
					components[i].IsRunning = true // Consider it running anyway
				}

				continue
			}

			// For Temporal and other components that need to be started
			cmd := exec.Command(comp.Command, comp.Args...)

			// Start the process in the background
			if err := cmd.Start(); err != nil {
				fmt.Printf("❌ Failed to start %s: %v\n", comp.Name, err)
				continue
			}

			components[i].IsRunning = true

			// Wait a bit for the component to start
			time.Sleep(comp.StartupDelay)

			// If there's a verification function, use it to make sure the component is accessible
			if comp.VerifyAvailable != nil && wait {
				maxRetries := 5
				for retry := 0; retry < maxRetries; retry++ {
					if comp.VerifyAvailable() {
						break
					}
					if retry == maxRetries-1 {
						fmt.Printf("⚠️ %s may not be fully running yet. It might need more time to initialize.\n", comp.Name)
					} else {
						if verbose {
							fmt.Printf("Waiting for %s to become available (attempt %d/%d)...\n", comp.Name, retry+1, maxRetries)
						}
						time.Sleep(2 * time.Second)
					}
				}
			}

			if verbose {
				fmt.Printf("✅ Started %s (PID: %d)\n", comp.Name, cmd.Process.Pid)
			}
		}

		// Check if all required components are running
		allRunning := true
		for _, comp := range components {
			if comp.IsRequired && !comp.IsRunning {
				allRunning = false
				break
			}
		}

		if !allRunning {
			fmt.Println("\n⚠️ Not all components are running. Please check the errors above.")
			os.Exit(1)
		}

		fmt.Println("\n✅ Local development environment is running!")
		fmt.Printf("Temporal UI is available at http://localhost:8233\n")
		// Dapr Dashboard is not available in slim installation
		fmt.Println("Use 'shielddev-cli localenv stop' to stop the environment.")
	},
}

func init() {
	localenvCmd.AddCommand(startCmd)

	// Add flags specific to the start command
	startCmd.Flags().Bool("skip-dapr", false, "Skip starting Dapr runtime")
	startCmd.Flags().Bool("skip-temporal", false, "Skip starting Temporal server")
	startCmd.Flags().StringP("config", "c", "", "Path to environment configuration file")
	startCmd.Flags().Bool("wait", true, "Wait for all components to be ready before exiting")
}
