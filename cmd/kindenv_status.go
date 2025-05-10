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

	"github.com/ShieldFC-RD/devhelper-cli/internal/kindenv"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// checkClusterExists checks if a Kind cluster with the given name exists
func checkClusterExists(clusterName string) (bool, error) {
	// Run 'kind get clusters' command
	cmd := exec.Command("kind", "get", "clusters")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("failed to list Kind clusters: %w", err)
	}

	// Check if cluster name exists in the output
	clusters := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, cluster := range clusters {
		if cluster == clusterName {
			return true, nil
		}
	}

	return false, nil
}

// kindenvStatusCmd represents the kindenv status command
var kindenvStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of a Kind-based development environment",
	Long: `Show status of a Kind-based development environment.

This command checks if a Kind cluster exists and shows its status.

By default, the command will use the cluster name from kindenv.yaml.
You can override this with the --name flag.

It uses native Go libraries instead of external CLI tools for improved reliability.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create colored output helpers
		green := color.New(color.FgGreen).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()
		red := color.New(color.FgRed).SprintFunc()

		// Get flags
		verbose, _ := cmd.Flags().GetBool("verbose")
		configPath, _ := cmd.Flags().GetString("config")
		clusterName, _ := cmd.Flags().GetString("name")

		if verbose {
			fmt.Printf("Flags: verbose=%v, config=%s, name=%s\n", verbose, configPath, clusterName)
		}

		fmt.Println(green("Checking Kind-based development environment status..."))

		// Load config
		if verbose {
			fmt.Printf("Attempting to load config from: %s\n", configPath)
		}
		config, err := kindenv.LoadConfig(configPath)
		if err != nil {
			fmt.Printf("%s Error loading config: %v\n", red("❌"), err)
			os.Exit(1)
		}

		if verbose {
			fmt.Println(yellow("Verbose mode enabled"))
			fmt.Printf("Config path: %s\n", configPath)
			fmt.Printf("Config loaded with cluster name: %s\n", config.Cluster.Name)
		}

		// Override cluster name only if explicitly provided with --name flag
		if cmd.Flags().Changed("name") {
			config.Cluster.Name = clusterName
			if verbose {
				fmt.Printf("%s Using specified cluster name: %s\n", yellow("⚙️"), clusterName)
			}
		} else if verbose {
			fmt.Printf("%s Using cluster name from config: %s\n", yellow("📄"), config.Cluster.Name)
		}

		if verbose {
			fmt.Printf("Using cluster name: %s\n", config.Cluster.Name)

			// Display port mappings
			if len(config.Cluster.MapPorts) > 0 {
				fmt.Println(yellow("Port mappings:"))
				for _, portMap := range config.Cluster.MapPorts {
					fmt.Printf("  - containerPort: %v, hostPort: %d, protocol: %s\n",
						portMap.ContainerPort, portMap.HostPort, portMap.Protocol)
				}
			}
		}

		// Check if cluster exists
		exists, err := checkClusterExists(config.Cluster.Name)
		if err != nil {
			fmt.Printf("%s Error checking cluster status: %v\n", red("❌"), err)
			os.Exit(1)
		}

		if exists {
			fmt.Printf("%s Kind cluster '%s' is running\n", green("✅"), config.Cluster.Name)

			// Execute kubectl to verify if the cluster is responsive
			kubectlCmd := exec.Command("kubectl", "cluster-info", "--context", fmt.Sprintf("kind-%s", config.Cluster.Name))
			kubectlOutput, err := kubectlCmd.CombinedOutput()
			if err != nil {
				fmt.Printf("%s Kubernetes cluster exists but may not be fully functional: %v\n", yellow("⚠️"), err)
			} else if verbose {
				fmt.Println(yellow("Kubernetes cluster information:"))
				fmt.Println(string(kubectlOutput))
			}

			// Check component status
			fmt.Println(green("Component status:"))

			// Check Temporal status
			if config.Components.Temporal.Enabled {
				temporalCmd := exec.Command("kubectl", "get", "deployment", "-n", config.Components.Temporal.Namespace, "--no-headers")
				temporalOutput, err := temporalCmd.CombinedOutput()
				if err == nil && string(temporalOutput) != "" {
					fmt.Printf("- %s Temporal is installed and running\n", green("✅"))
				} else {
					fmt.Printf("- %s Temporal is not running or not installed\n", yellow("⚠️"))
				}
			} else {
				fmt.Printf("- %s Temporal is disabled in config\n", yellow("ℹ️"))
			}

			// Check Redis status
			if config.Components.Redis.Enabled {
				redisCmd := exec.Command("kubectl", "get", "pod", "-n", "redis", "--no-headers")
				redisOutput, err := redisCmd.CombinedOutput()
				if err == nil && string(redisOutput) != "" {
					fmt.Printf("- %s Redis is installed and running\n", green("✅"))
				} else {
					fmt.Printf("- %s Redis is not running or not installed\n", yellow("⚠️"))
				}
			} else {
				fmt.Printf("- %s Redis is disabled in config\n", yellow("ℹ️"))
			}

			// Check Dapr status
			if config.Components.Dapr.Enabled {
				daprCmd := exec.Command("kubectl", "get", "deployment", "-n", "dapr-system", "--no-headers")
				daprOutput, err := daprCmd.CombinedOutput()
				if err == nil && string(daprOutput) != "" {
					fmt.Printf("- %s Dapr is installed and running\n", green("✅"))
				} else {
					fmt.Printf("- %s Dapr is not running or not installed\n", yellow("⚠️"))
				}
			} else {
				fmt.Printf("- %s Dapr is disabled in config\n", yellow("ℹ️"))
			}

			// Check OpenSearch status
			if config.Components.OpenSearch.Enabled {
				openSearchCmd := exec.Command("kubectl", "get", "pod", "-n", config.Components.OpenSearch.Namespace, "-l", "app=opensearch", "--no-headers")
				openSearchOutput, err := openSearchCmd.CombinedOutput()
				if err == nil && string(openSearchOutput) != "" {
					fmt.Printf("- %s OpenSearch is installed and running\n", green("✅"))
				} else {
					fmt.Printf("- %s OpenSearch is not running or not installed\n", yellow("⚠️"))
				}
			} else {
				fmt.Printf("- %s OpenSearch is disabled in config\n", yellow("ℹ️"))
			}

			// Check OpenSearch Dashboards status
			if config.Components.OpenSearchDashboards.Enabled {
				dashboardsCmd := exec.Command("kubectl", "get", "pod", "-n", config.Components.OpenSearchDashboards.Namespace, "-l", "app=opensearch-dashboards", "--no-headers")
				dashboardsOutput, err := dashboardsCmd.CombinedOutput()
				if err == nil && string(dashboardsOutput) != "" {
					fmt.Printf("- %s OpenSearch Dashboards is installed and running\n", green("✅"))
				} else {
					fmt.Printf("- %s OpenSearch Dashboards is not running or not installed\n", yellow("⚠️"))
				}
			} else {
				fmt.Printf("- %s OpenSearch Dashboards is disabled in config\n", yellow("ℹ️"))
			}

			// Check Temporal Worker Operator status
			if config.Components.TemporalWorkerOperator.Enabled {
				operatorCmd := exec.Command("kubectl", "get", "deployment", "-n", "shield-system", "--no-headers")
				operatorOutput, err := operatorCmd.CombinedOutput()

				// Check CRDs
				crdCmd := exec.Command("kubectl", "get", "crd", "temporalworkers.orchestration.shieldfc.com", "--no-headers")
				crdOutput, crdErr := crdCmd.CombinedOutput()

				if err == nil && string(operatorOutput) != "" {
					fmt.Printf("- %s Temporal Worker Operator is installed and running\n", green("✅"))
					if crdErr == nil && string(crdOutput) != "" {
						fmt.Printf("  %s CRDs are properly installed\n", green("✓"))
					} else {
						fmt.Printf("  %s CRDs not found, operator may not function correctly\n", yellow("⚠️"))
					}
				} else {
					fmt.Printf("- %s Temporal Worker Operator is not running or not installed\n", yellow("⚠️"))
				}
			} else {
				fmt.Printf("- %s Temporal Worker Operator is disabled in config\n", yellow("ℹ️"))
			}

			// Print service access information
			fmt.Println(green("Access services:"))

			// Find host ports from the port mappings
			var temporalWebPort, temporalFrontendPort, redisPort, openSearchPort, openSearchDashboardsPort int

			for _, portMap := range config.Cluster.MapPorts {
				// Check if containerPort matches any of our known nodePort values
				switch cp := portMap.ContainerPort.(type) {
				case int:
					// Temporal Web UI
					if cp == config.Components.Temporal.NodePorts.Web {
						temporalWebPort = portMap.HostPort
					}
					// Temporal Frontend
					if cp == config.Components.Temporal.NodePorts.Frontend {
						temporalFrontendPort = portMap.HostPort
					}
					// Redis
					if cp == config.Components.Redis.NodePorts.Redis {
						redisPort = portMap.HostPort
					}
					// OpenSearch REST
					if cp == config.Components.OpenSearch.NodePorts.Rest {
						openSearchPort = portMap.HostPort
					}
					// OpenSearch Dashboards
					if cp == config.Components.OpenSearchDashboards.NodePorts.Http {
						openSearchDashboardsPort = portMap.HostPort
					}
				}
			}

			if config.Components.Temporal.Enabled {
				// Check if Temporal is actually deployed
				temporalCmd := exec.Command("kubectl", "get", "namespace", config.Components.Temporal.Namespace)
				_, err := temporalCmd.CombinedOutput()

				if err == nil && temporalWebPort > 0 && temporalFrontendPort > 0 {
					fmt.Printf("- Temporal Web UI: http://localhost:%d\n", temporalWebPort)
					fmt.Printf("- Temporal Frontend: localhost:%d\n", temporalFrontendPort)
				} else if verbose {
					fmt.Printf("%s Temporal namespace not found or port mapping missing. Temporal may not be installed.\n", yellow("⚠️"))
				}
			}

			if config.Components.Redis.Enabled {
				// Check if Redis is actually deployed
				redisCmd := exec.Command("kubectl", "get", "namespace", "redis")
				_, err := redisCmd.CombinedOutput()

				if err == nil && redisPort > 0 {
					fmt.Printf("- Redis: localhost:%d\n", redisPort)
				} else if verbose {
					fmt.Printf("%s Redis namespace not found or port mapping missing. Redis may not be installed.\n", yellow("⚠️"))
				}
			}

			if config.Components.OpenSearch.Enabled {
				// Check if OpenSearch is actually deployed
				openSearchCmd := exec.Command("kubectl", "get", "namespace", config.Components.OpenSearch.Namespace)
				_, err := openSearchCmd.CombinedOutput()

				if err == nil && openSearchPort > 0 {
					fmt.Printf("- OpenSearch: http://localhost:%d\n", openSearchPort)
				} else if verbose {
					fmt.Printf("%s OpenSearch namespace not found or port mapping missing. OpenSearch may not be installed.\n", yellow("⚠️"))
				}
			}

			if config.Components.OpenSearchDashboards.Enabled {
				// Check if OpenSearch Dashboards is actually deployed
				dashboardsCmd := exec.Command("kubectl", "get", "namespace", config.Components.OpenSearchDashboards.Namespace)
				_, err := dashboardsCmd.CombinedOutput()

				if err == nil && openSearchDashboardsPort > 0 {
					fmt.Printf("- OpenSearch Dashboards: http://localhost:%d\n", openSearchDashboardsPort)
				} else if verbose {
					fmt.Printf("%s OpenSearch Dashboards namespace not found or port mapping missing. OpenSearch Dashboards may not be installed.\n", yellow("⚠️"))
				}
			}

			// Additional component status checks for verbose mode
			if verbose {
				// List all namespaces
				nsCmd := exec.Command("kubectl", "get", "namespaces")
				nsOutput, err := nsCmd.CombinedOutput()
				if err == nil {
					fmt.Println(yellow("Available namespaces:"))
					fmt.Println(string(nsOutput))
				}

				// Check for TemporalWorker resources
				if config.Components.TemporalWorkerOperator.Enabled {
					twCmd := exec.Command("kubectl", "get", "temporalworkers", "--all-namespaces")
					twOutput, twErr := twCmd.CombinedOutput()
					if twErr == nil && string(twOutput) != "" {
						fmt.Println(yellow("Installed TemporalWorker resources:"))
						fmt.Println(string(twOutput))
					}

					tnsCmd := exec.Command("kubectl", "get", "temporalnamespaces", "--all-namespaces")
					tnsOutput, tnsErr := tnsCmd.CombinedOutput()
					if tnsErr == nil && string(tnsOutput) != "" {
						fmt.Println(yellow("Installed TemporalNamespace resources:"))
						fmt.Println(string(tnsOutput))
					}
				}
			}
		} else {
			fmt.Printf("%s Kind cluster '%s' is not running\n", yellow("⚠️"), config.Cluster.Name)
		}
	},
}

func init() {
	kindenvCmd.AddCommand(kindenvStatusCmd)

	// Add flags for kindenv status command
	kindenvStatusCmd.Flags().StringP("config", "f", "", "Path to configuration file")
	kindenvStatusCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
	kindenvStatusCmd.Flags().String("name", "", "Cluster name (defaults to current directory name)")
}
