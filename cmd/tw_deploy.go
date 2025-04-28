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
	"strings"

	"github.com/ShieldFC-RD/devhelper-cli/internal/tw"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// twDeployCmd represents the tw deploy command
var twDeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a Temporal Worker",
	Long: `Deploy a Temporal Worker to a Kubernetes environment.

This command deploys a Temporal Worker to a Kubernetes environment using
the configuration from tw.yaml. It can deploy to a local Kind cluster
or to a remote environment.

Examples:
  devhelper-cli tw deploy --kind
  devhelper-cli tw deploy --remote dev
  devhelper-cli tw deploy --build`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create colored output helpers
		green := color.New(color.FgGreen).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()
		red := color.New(color.FgRed).SprintFunc()

		// Temporary fix to avoid unused variable error if not used in this function
		_ = green

		// Get flags
		configPath, _ := cmd.Flags().GetString("config")
		kind, _ := cmd.Flags().GetBool("kind")
		remote, _ := cmd.Flags().GetString("remote")
		buildFirst, _ := cmd.Flags().GetBool("build")
		kindCluster, _ := cmd.Flags().GetString("kind-cluster")
		namespace, _ := cmd.Flags().GetString("namespace")
		verbose, _ := cmd.Flags().GetBool("verbose")

		if verbose {
			fmt.Printf("Flags: config=%s, kind=%v, remote=%s, build=%v, cluster=%s, namespace=%s\n",
				configPath, kind, remote, buildFirst, kindCluster, namespace)
		}

		// Check that exactly one deployment target is specified
		if (kind && remote != "") || (!kind && remote == "") {
			fmt.Printf("%s Please specify exactly one deployment target: --kind or --remote\n", red("❌"))
			os.Exit(1)
		}

		// Load configuration
		config, err := tw.LoadConfig(configPath)
		if err != nil {
			fmt.Printf("%s Error loading configuration: %v\n", red("❌"), err)
			os.Exit(1)
		}

		// Validate configuration
		if err := config.Validate(); err != nil {
			fmt.Printf("%s Configuration validation failed: %v\n", red("❌"), err)
			os.Exit(1)
		}

		// Use default namespace if not provided
		if namespace == "" {
			namespace = "default"
		}

		// Build image first if requested
		if buildFirst {
			fmt.Println(yellow("Building Docker image first..."))

			// Create build command with appropriate flags
			buildArgs := []string{"tw", "build"}
			if kind {
				if kindCluster == "" {
					kindCluster = "kind" // Default kind cluster name
				}
				buildArgs = append(buildArgs, "--kind-load", kindCluster)
			}

			// Add verbose flag if needed
			if verbose {
				buildArgs = append(buildArgs, "--verbose")
			}

			// Execute the build command
			buildCmd := exec.Command("devhelper-cli", buildArgs...)
			buildCmd.Stdout = os.Stdout
			buildCmd.Stderr = os.Stderr

			if err := buildCmd.Run(); err != nil {
				fmt.Printf("%s Failed to build Docker image: %v\n", red("❌"), err)
				os.Exit(1)
			}
		}

		// Handle different deployment targets
		if kind {
			deployToKind(config, kindCluster, namespace, verbose)
		} else if remote != "" {
			deployToRemote(config, remote, namespace, verbose)
		}
	},
}

// deployToKind deploys a Temporal Worker to a Kind cluster
func deployToKind(config *tw.TemporalWorkerConfig, cluster, namespace string, verbose bool) {
	// Create colored output helpers
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	fmt.Println(green("Deploying Temporal Worker to Kind cluster..."))

	// Use default cluster name if not provided
	if cluster == "" {
		cluster = "kind"
		fmt.Println(yellow("Using default Kind cluster name: kind"))
	}

	// Verify that the Kind cluster exists
	clusterCheckCmd := exec.Command("kind", "get", "clusters")
	clusterOutput, err := clusterCheckCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s Failed to list Kind clusters: %v\n", red("❌"), err)
		os.Exit(1)
	}

	clusterFound := false
	clusters := strings.Split(strings.TrimSpace(string(clusterOutput)), "\n")
	for _, c := range clusters {
		if c == cluster {
			clusterFound = true
			break
		}
	}

	if !clusterFound {
		fmt.Printf("%s Kind cluster '%s' not found. Available clusters: %s\n",
			red("❌"), cluster, strings.Join(clusters, ", "))
		os.Exit(1)
	}

	// Create YAML for the TemporalWorker resource
	fmt.Println(yellow("Creating TemporalWorker resource..."))

	// Convert config to YAML
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		fmt.Printf("%s Failed to marshal configuration to YAML: %v\n", red("❌"), err)
		os.Exit(1)
	}

	// Create a temporary file for the YAML
	tempFile, err := os.CreateTemp("", "temporal-worker-*.yaml")
	if err != nil {
		fmt.Printf("%s Failed to create temporary file: %v\n", red("❌"), err)
		os.Exit(1)
	}
	defer os.Remove(tempFile.Name())

	// Write the YAML to the temporary file
	if _, err := tempFile.Write(yamlData); err != nil {
		fmt.Printf("%s Failed to write YAML to temporary file: %v\n", red("❌"), err)
		os.Exit(1)
	}
	tempFile.Close()

	// Apply the YAML using kubectl
	fmt.Printf("Applying TemporalWorker resource to namespace %s...\n", namespace)

	kubectlArgs := []string{
		"apply",
		"-f", tempFile.Name(),
		"-n", namespace,
	}

	kubectlCmd := exec.Command("kubectl", kubectlArgs...)
	kubectlCmd.Stdout = os.Stdout
	kubectlCmd.Stderr = os.Stderr

	if verbose {
		fmt.Printf("Running command: kubectl %s\n", strings.Join(kubectlArgs, " "))
	}

	if err := kubectlCmd.Run(); err != nil {
		fmt.Printf("%s Failed to apply TemporalWorker resource: %v\n", red("❌"), err)
		os.Exit(1)
	}

	fmt.Printf("%s Temporal Worker deployed successfully to Kind cluster\n", green("✅"))

	// Print next steps
	fmt.Println(green("\nNext steps:"))
	fmt.Printf("1. Check the status of your worker:\n   kubectl get temporalworkers -n %s\n", namespace)
	fmt.Printf("2. View worker logs:\n   kubectl logs -l worker-type=%s -n %s\n", config.Spec.WorkerType, namespace)
}

// deployToRemote deploys a Temporal Worker to a remote environment
func deployToRemote(config *tw.TemporalWorkerConfig, env, namespace string, verbose bool) {
	// Create colored output helpers
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Println(yellow("Deploying to remote environments is not implemented yet"))
	fmt.Printf("Would deploy to environment: %s, namespace: %s\n", env, namespace)

	// TODO: Implement remote deployment with environment-specific authentication
}

func init() {
	twCmd.AddCommand(twDeployCmd)

	// Add flags for tw deploy command
	twDeployCmd.Flags().Bool("kind", false, "Deploy to a Kind cluster")
	twDeployCmd.Flags().String("remote", "", "Deploy to a remote environment (dev, staging, prod)")
	twDeployCmd.Flags().Bool("build", false, "Build the Docker image before deploying")
	twDeployCmd.Flags().String("kind-cluster", "", "Kind cluster name (default: kind)")
	twDeployCmd.Flags().String("namespace", "", "Kubernetes namespace (default: default)")
}
