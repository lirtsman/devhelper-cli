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
)

// version variables - these can be set during build time with ldflags
var (
	Version   = "0.1.0"
	BuildDate = "unknown"
	Commit    = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of devhelper-cli",
	Long:  `All software has versions. This is devhelper-cli's.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("devhelper-cli version: %s\n", Version)
		fmt.Printf("Build Date: %s\n", BuildDate)
		fmt.Printf("Git Commit: %s\n", Commit)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
