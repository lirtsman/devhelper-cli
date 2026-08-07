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
	"path/filepath"

	"github.com/ShieldFC-RD/devhelper-cli/internal/tw"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// twInitCmd represents the tw init command
var twInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a Temporal Worker project",
	Long: `Initialize a Temporal Worker project with a configuration file.

This command creates a default tw.yaml configuration file with the necessary
structure for defining a Temporal Worker. It can also generate template code
for different languages.

By default, it will use the current directory name as the worker name.
You can override this with the --name flag.

The worker type will be parsed from the name (e.g., temporal-ingestion-parsing -> ingestion).
You can override this with the --worker-type flag.

Example:
  devhelper-cli tw init
  devhelper-cli tw init --name my-worker
  devhelper-cli tw init --name temporal-ingestion-worker --worker-type ingestion
  devhelper-cli tw init --template typescript
  devhelper-cli tw init --force`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create colored output helpers
		green := color.New(color.FgGreen).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()
		red := color.New(color.FgRed).SprintFunc()

		// Get flags
		configPath, _ := cmd.Flags().GetString("output")
		force, _ := cmd.Flags().GetBool("force")
		template, _ := cmd.Flags().GetString("template")
		workerName, _ := cmd.Flags().GetString("name")
		workerType, _ := cmd.Flags().GetString("worker-type")
		verbose, _ := cmd.Flags().GetBool("verbose")

		if verbose {
			fmt.Printf("Flags: output=%s, force=%v, template=%s, name=%s, worker-type=%s\n",
				configPath, force, template, workerName, workerType)
		}

		// If no config path is provided, use default
		if configPath == "" {
			configPath = "tw.yaml"
		}

		// Check if the file already exists and we're not forcing overwrite
		if _, err := os.Stat(configPath); err == nil && !force {
			fmt.Printf("%s Configuration file already exists at %s. Use --force to overwrite.\n", red("❌"), configPath)
			return
		}

		// Create parent directories if needed
		configDir := filepath.Dir(configPath)
		if configDir != "." && configDir != ".." {
			if err := os.MkdirAll(configDir, 0755); err != nil {
				fmt.Printf("%s Failed to create directory %s: %v\n", red("❌"), configDir, err)
				os.Exit(1)
			}
		}

		fmt.Println(green("Initializing Temporal Worker project..."))

		// Create default configuration
		config := tw.CreateDefaultConfig(workerName, workerType)

		// Save configuration to file
		err := tw.SaveConfig(config, configPath)
		if err != nil {
			fmt.Printf("%s Failed to write configuration file: %v\n", red("❌"), err)
			os.Exit(1)
		}

		fmt.Printf("%s Temporal Worker configuration initialized at %s\n", green("✅"), configPath)
		fmt.Println("Edit this file to customize your Temporal Worker configuration.")

		// Create template code if requested
		if template != "" {
			fmt.Printf("%s Template support for %s is not implemented yet\n", yellow("⚠️"), template)
			// TODO: Implement template generation
		}

		// Print next steps
		fmt.Println(green("\nNext steps:"))
		fmt.Println("1. Edit the configuration file to customize your worker settings")
		fmt.Println("2. Build and run your worker with 'devhelper-cli tw build' and 'devhelper-cli tw run'")
		fmt.Println("3. Deploy your worker to a Kind cluster with 'devhelper-cli tw deploy --kind'")
	},
}

func init() {
	twCmd.AddCommand(twInitCmd)

	// Add flags for tw init command
	twInitCmd.Flags().StringP("output", "o", "", "Output path for configuration file (default: tw.yaml)")
	twInitCmd.Flags().BoolP("force", "f", false, "Force overwrite if configuration file already exists")
	twInitCmd.Flags().String("template", "", "Template to use for generating code (golang, typescript)")
	twInitCmd.Flags().String("worker-type", "", "Type of worker (e.g., ingestion, processing). Default: parsed from name")
}
