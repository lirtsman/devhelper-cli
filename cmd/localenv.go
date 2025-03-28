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
		fmt.Println("Use one of the localenv subcommands. Run 'shielddev-cli localenv --help' for usage.")
	},
}

func init() {
	rootCmd.AddCommand(localenvCmd)

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
