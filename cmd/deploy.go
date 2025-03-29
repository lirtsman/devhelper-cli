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

	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy ShieldDev resources",
	Long: `Deploy ShieldDev resources to the target environment.
	
This command allows you to deploy various ShieldDev resources such as
applications, services, or infrastructure components to your target environment.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Use one of the deploy subcommands. Run 'devhelper-cli deploy --help' for usage.")
	},
}

// appCmd represents the deploy app subcommand
var deployAppCmd = &cobra.Command{
	Use:   "app [name]",
	Short: "Deploy an application",
	Long: `Deploy a ShieldDev application to the target environment.
	
This command deploys a specified application with its configuration 
to the target environment.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]
		env, _ := cmd.Flags().GetString("env")
		fmt.Printf("Deploying application '%s' to '%s' environment...\n", appName, env)
		// Here would be the actual deployment logic
		fmt.Println("Deployment completed successfully!")
	},
}

// serviceCmd represents the deploy service subcommand
var deployServiceCmd = &cobra.Command{
	Use:   "service [name]",
	Short: "Deploy a service",
	Long: `Deploy a ShieldDev service to the target environment.
	
This command deploys a specified service with its configuration 
to the target environment.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceName := args[0]
		env, _ := cmd.Flags().GetString("env")
		fmt.Printf("Deploying service '%s' to '%s' environment...\n", serviceName, env)
		// Here would be the actual deployment logic
		fmt.Println("Service deployment completed successfully!")
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)

	// Add subcommands to deploy command
	deployCmd.AddCommand(deployAppCmd)
	deployCmd.AddCommand(deployServiceCmd)

	// Add flags to the deploy app command
	deployAppCmd.Flags().StringP("env", "e", "dev", "Target environment (dev, staging, prod)")
	deployAppCmd.Flags().BoolP("force", "f", false, "Force deployment even if validation fails")
	deployAppCmd.Flags().StringP("version", "v", "latest", "Version to deploy")

	// Add flags to the deploy service command
	deployServiceCmd.Flags().StringP("env", "e", "dev", "Target environment (dev, staging, prod)")
	deployServiceCmd.Flags().BoolP("force", "f", false, "Force deployment even if validation fails")
	deployServiceCmd.Flags().StringP("version", "v", "latest", "Version to deploy")
}
