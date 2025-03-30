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

	"github.com/spf13/cobra"
)

var testVersionCmd = &cobra.Command{
	Use:   "test-version",
	Short: "Test the version checking functionality",
	Long:  `A test command to verify the version checking functionality works correctly.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Testing version checking functionality...")

		// Get the test script path
		testScriptPath, _ := filepath.Abs("test")
		testToolScript := filepath.Join(testScriptPath, "test-tool.sh")

		// Make sure the script is executable
		chmod := exec.Command("chmod", "+x", testToolScript)
		chmod.Run()

		fmt.Printf("Using test script: %s\n", testToolScript)

		// Create a symbolic link to make it accessible as "test-tool"
		testToolLink := filepath.Join(testScriptPath, "test-tool")
		_ = os.Remove(testToolLink) // Remove any existing link
		_ = os.Symlink(testToolScript, testToolLink)

		// Add test directory to PATH for testing
		os.Setenv("PATH", testScriptPath+":"+os.Getenv("PATH"))

		// Execute the script directly to verify it works
		output, err := exec.Command(testToolScript).CombinedOutput()
		if err != nil {
			fmt.Printf("Error running test script: %v\n", err)
		} else {
			fmt.Printf("Test script output: %s\n", output)
		}

		// Now check the test tool which should trigger a version warning
		fmt.Println("\nRunning validateTool on test-tool...")
		testToolPath, testToolErr := validateTool("test-tool", "", true)
		if testToolErr != nil {
			fmt.Printf("Test tool validation error: %v\n", testToolErr)
		} else {
			fmt.Printf("Test tool found at: %s\n", testToolPath)
		}
	},
}

func init() {
	rootCmd.AddCommand(testVersionCmd)
}
