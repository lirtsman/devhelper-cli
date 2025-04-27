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

// kindenvCmd represents the kindenv command
var kindenvCmd = &cobra.Command{
	Use:   "kindenv",
	Short: "Manage Kind-based development environment",
	Long: `Manage a Kind-based Kubernetes development environment for Shield applications.

The kindenv command provisions and manages Kind clusters with all required components:
- Kind Kubernetes cluster
- Temporal server
- Redis
- Dapr
- Cert-manager
- Additional required services

This allows developers to run and test Shield applications in a local Kubernetes environment.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Use one of the kindenv subcommands. Run 'devhelper-cli kindenv --help' for usage.")
	},
}

func init() {
	rootCmd.AddCommand(kindenvCmd)

	// Add persistent flags that are available to all subcommands
	kindenvCmd.PersistentFlags().StringP("cluster-name", "n", "kindenv", "Name of the Kind cluster")
	kindenvCmd.PersistentFlags().StringP("config", "c", "", "Path to kindenv config file")
	kindenvCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")

	// Bind to viper
	viper.BindPFlag("kindenv.clusterName", kindenvCmd.PersistentFlags().Lookup("cluster-name"))
	viper.BindPFlag("kindenv.config", kindenvCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("kindenv.verbose", kindenvCmd.PersistentFlags().Lookup("verbose"))
}
