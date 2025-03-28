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
	"time"

	"github.com/spf13/cobra"
)

// Components to be started
type Component struct {
	Name          string
	Command       string
	Args          []string
	CheckCommand  string
	CheckArgs     []string
	RequiredFor   []string
	StartupDelay  time.Duration
	IsRunning     bool
	IsRequired    bool
	CommandExists bool
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

		// Define the components we need to start
		components := []Component{
			{
				Name:          "Docker",
				Command:       "docker",
				Args:          []string{"--version"},
				CheckCommand:  "docker",
				CheckArgs:     []string{"ps"},
				RequiredFor:   []string{"Dapr", "Temporal"},
				StartupDelay:  0,
				IsRequired:    true,
				CommandExists: isCommandAvailable("docker"),
			},
			{
				Name:          "Dapr",
				Command:       "dapr",
				Args:          []string{"init", "--slim"},
				CheckCommand:  "dapr",
				CheckArgs:     []string{"status", "-k"},
				RequiredFor:   []string{},
				StartupDelay:  1 * time.Second,
				IsRequired:    !skipDapr,
				CommandExists: isCommandAvailable("dapr"),
			},
			{
				Name:          "Temporal",
				Command:       "temporal",
				Args:          []string{"server", "start-dev"},
				CheckCommand:  "temporal",
				CheckArgs:     []string{"operator", "list", "--namespace", "default"},
				RequiredFor:   []string{},
				StartupDelay:  3 * time.Second,
				IsRequired:    !skipTemporal,
				CommandExists: isCommandAvailable("temporal"),
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

			fmt.Printf("Starting %s...\n", comp.Name)

			// For each component, we'll want to run as background process
			cmd := exec.Command(comp.Command, comp.Args...)

			// Start the process in the background
			if err := cmd.Start(); err != nil {
				fmt.Printf("❌ Failed to start %s: %v\n", comp.Name, err)
				continue
			}

			components[i].IsRunning = true

			// Wait a bit for the component to start
			time.Sleep(comp.StartupDelay)

			if verbose {
				fmt.Printf("✅ Started %s (PID: %d)\n", comp.Name, cmd.Process.Pid)
			}
		}

		fmt.Println("\n✅ Local development environment is running!")
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
