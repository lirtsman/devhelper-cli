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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// twCmd represents the tw command
var twCmd = &cobra.Command{
	Use:   "tw",
	Short: "Manage Temporal Worker projects",
	Long: `The tw command provides tools for managing Temporal Worker projects,
including initialization, configuration, building, running, and deployment.

It supports the complete development lifecycle of Temporal Workers from
local development to production deployment.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Use one of the tw subcommands. Run 'devhelper-cli tw --help' for usage.")
	},
}

func init() {
	rootCmd.AddCommand(twCmd)

	// Add persistent flags for the tw command
	twCmd.PersistentFlags().StringP("config", "c", "tw.yaml", "Path to temporal worker config file")
	twCmd.PersistentFlags().StringP("name", "n", "", "Name of the temporal worker project")
	twCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")

	// Bind flags to viper for configuration
	viper.BindPFlag("tw.config", twCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("tw.name", twCmd.PersistentFlags().Lookup("name"))
	viper.BindPFlag("tw.verbose", twCmd.PersistentFlags().Lookup("verbose"))
}
