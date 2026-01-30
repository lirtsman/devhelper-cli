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
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	// Import necessary packages
	"bufio"
	"time"

	"github.com/ShieldFC-RD/devhelper-cli/internal/kindenv"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Define the setupECRCreds function at the file level, outside of any function

// setupECRCreds creates ECR pull secrets in a given namespace
func setupECRCreds(namespace, registry, password string) error {
	fmt.Printf("Creating ECR pull secret in the %s namespace\n", namespace)

	// Ensure namespace exists
	namespaceYaml, err := executeCommand("kubectl", "create", "namespace", namespace, "--dry-run=client", "-o", "yaml")
	if err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	// Apply the namespace using stdin
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(namespaceYaml)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply namespace: %w", err)
	}

	// Create ECR credentials secret
	secretYaml, err := executeCommand("kubectl", "create", "secret", "docker-registry",
		"ecr-credentials",
		fmt.Sprintf("--docker-server=%s", registry),
		"--docker-username=AWS",
		fmt.Sprintf("--docker-password=%s", password),
		fmt.Sprintf("--namespace=%s", namespace),
		"--dry-run=client", "-o", "yaml")
	if err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	// Apply the secret using stdin
	cmd = exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(secretYaml)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply secret: %w", err)
	}

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s ECR pull secret created in the %s namespace\n", green("✅"), namespace)
	return nil
}

// executeCommand runs a shell command and returns the combined output and error
func executeCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("command failed: %v\nOutput: %s\nError: %s", err, stdout.String(), stderr.String())
	}
	return stdout.String(), nil
}

// commandExists checks if a command exists in the system
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// waitForDeployment waits for a deployment to be available
func waitForDeployment(namespace, deployment string, timeout int) error {
	fmt.Printf("Waiting for deployment %s in namespace %s to be ready...\n", deployment, namespace)
	_, err := executeCommand("kubectl", "wait", "--for=condition=Available",
		fmt.Sprintf("--timeout=%dm", timeout),
		fmt.Sprintf("deployment/%s", deployment),
		fmt.Sprintf("-n=%s", namespace))
	return err
}

// kindenvStartCmd represents the kindenv start command
var kindenvStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a Kind-based development environment",
	Long: `Start a Kind-based development environment for Shield applications.

This command creates a Kind cluster if it doesn't exist and installs required components:
- Temporal
- Redis
- Dapr
- OpenSearch
- OpenSearch Dashboards
- Temporal Worker Operator

By default, the command will use the cluster name from kindenv.yaml.
You can override this with the --name flag.

When AWS ECR is enabled, you can specify an AWS profile in the config file (images.aws.profile)
or override it with the --aws-profile flag.

It also verifies and switches to the correct Kubernetes context if needed, and creates
required namespaces for component installation.

Note: When starting the environment, the command ensures you're using the correct Kubernetes
context for the kind cluster. If you're using a different context, it will prompt you to switch.
Use --force-context to automatically switch without prompting.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create colored output helpers
		green := color.New(color.FgGreen).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()
		red := color.New(color.FgRed).SprintFunc()

		// Get flags
		verbose, _ := cmd.Flags().GetBool("verbose")
		configPath, _ := cmd.Flags().GetString("config")
		clusterName, _ := cmd.Flags().GetString("name")
		useAwsEcr, _ := cmd.Flags().GetBool("use-aws-ecr")
		skipTemporal, _ := cmd.Flags().GetBool("skip-temporal")
		skipDapr, _ := cmd.Flags().GetBool("skip-dapr")
		skipRedis, _ := cmd.Flags().GetBool("skip-redis")
		skipOpenSearch, _ := cmd.Flags().GetBool("skip-opensearch")
		skipOpenSearchDashboards, _ := cmd.Flags().GetBool("skip-opensearch-dashboards")
		skipOpenSearchIndexManagement, _ := cmd.Flags().GetBool("skip-opensearch-index-management")
		skipTemporalWorkerOperator, _ := cmd.Flags().GetBool("skip-temporal-worker-operator")
		skipIndicesOperator, _ := cmd.Flags().GetBool("skip-indices-operator")
		skipMetricsServer, _ := cmd.Flags().GetBool("skip-metrics-server")
		awsProfile, _ := cmd.Flags().GetString("aws-profile")
		forceContext, _ := cmd.Flags().GetBool("force-context")

		// Load config file
		fmt.Println(green("Setting up Kind-based development environment..."))

		// Load config
		config, err := kindenv.LoadConfig(configPath)
		if err != nil {
			fmt.Printf("%s Error loading config: %v\n", red("❌"), err)
			os.Exit(1)
		}

		// Override cluster name only if explicitly provided with --name flag
		if cmd.Flags().Changed("name") {
			config.Cluster.Name = clusterName
			fmt.Printf("%s Using specified cluster name: %s\n", yellow("⚙️"), clusterName)
		} else {
			fmt.Printf("%s Using cluster name from config: %s\n", yellow("📄"), config.Cluster.Name)
		}

		// Override config with command line args
		if useAwsEcr {
			config.Images.UseAwsEcr = true
		}

		// Set component enabled/disabled based on skip flags
		// Override config values with flags
		if skipTemporal {
			config.Components.Temporal.Enabled = false
		}
		if skipDapr {
			config.Components.Dapr.Enabled = false
		}
		if skipRedis {
			config.Components.Redis.Enabled = false
		}
		if skipOpenSearch {
			config.Components.OpenSearch.Enabled = false
		}
		if skipOpenSearchDashboards {
			config.Components.OpenSearchDashboards.Enabled = false
		}
		if skipOpenSearchIndexManagement {
			config.Components.OpenSearch.IndexManagement.Enabled = false
		}
		if skipTemporalWorkerOperator {
			config.Components.TemporalWorkerOperator.Enabled = false
		}
		if skipIndicesOperator {
			config.Components.IndicesOperator.Enabled = false
		}
		if skipMetricsServer {
			config.Components.MetricsServer.Enabled = false
		}

		// Show warning if deprecated operator-namespace flag is used
		if cmd.Flags().Changed("operator-namespace") {
			fmt.Printf("%s The --operator-namespace flag is deprecated. The operator will be installed in the default namespace.\n", yellow("⚠️"))
		}

		// Print component configuration
		fmt.Println(green("Component configuration:"))
		fmt.Printf("- Temporal: %v\n", config.Components.Temporal.Enabled)
		fmt.Printf("- Redis: %v\n", config.Components.Redis.Enabled)
		fmt.Printf("- Dapr: %v\n", config.Components.Dapr.Enabled)
		fmt.Printf("- OpenSearch: %v\n", config.Components.OpenSearch.Enabled)
		fmt.Printf("  - Index Management: %v\n", config.Components.OpenSearch.IndexManagement.Enabled)
		fmt.Printf("- OpenSearch Dashboards: %v\n", config.Components.OpenSearchDashboards.Enabled)
		fmt.Printf("- Temporal Worker Operator: %v\n", config.Components.TemporalWorkerOperator.Enabled)
		if config.Components.TemporalWorkerOperator.Enabled {
			fmt.Printf("  - Temporal namespace: '%s'\n", config.Components.TemporalWorkerOperator.TemporalNamespace)
			fmt.Printf("    (To modify namespace, edit temporalNamespace in kindenv.yaml)\n")
		}
		fmt.Printf("- Indices Operator: %v\n", config.Components.IndicesOperator.Enabled)
		fmt.Printf("- Metrics Server: %v\n", config.Components.MetricsServer.Enabled)
		fmt.Printf("- AWS ECR: %v\n", config.Images.UseAwsEcr)

		if verbose {
			fmt.Println(yellow("Verbose mode enabled"))
		}

		// Check for required tools
		requiredTools := []string{"kind", "kubectl", "helm"}
		for _, tool := range requiredTools {
			if !commandExists(tool) {
				fmt.Printf("%s Error: %s is not installed. Please install it first.\n", red("❌"), tool)
				os.Exit(1)
			}
		}

		// Detect container engine
		containerEngine := ""
		if commandExists("podman") {
			containerEngine = "podman"
			fmt.Println(green("Using podman as container engine"))
		} else if commandExists("docker") {
			containerEngine = "docker"
			fmt.Println(green("Using docker as container engine"))
		} else {
			fmt.Printf("%s Error: Neither docker nor podman found. Please install one of them.\n", red("❌"))
			os.Exit(1)
		}

		// Create a Kind cluster if it doesn't exist
		clusterExists := false
		clusterOutput, err := executeCommand("kind", "get", "clusters")
		if err == nil && strings.Contains(clusterOutput, config.Cluster.Name) {
			clusterExists = true
			fmt.Printf("%s Kind cluster %s already exists, reusing it\n", yellow("🔄"), config.Cluster.Name)
		}

		if !clusterExists && config.Cluster.CreateIfNotExists {
			fmt.Printf("%s Creating Kind cluster: %s\n", yellow("⚙️"), config.Cluster.Name)

			// Create Kind cluster configuration
			kindConfig := "kind: Cluster\napiVersion: kind.x-k8s.io/v1alpha4\nnodes:\n- role: control-plane\n  kubeadmConfigPatches:\n  - |\n    kind: InitConfiguration\n    nodeRegistration:\n      kubeletExtraArgs:\n        node-labels: \"ingress-ready=true\"\n  extraPortMappings:\n"

			// Track mapped ports to avoid duplicates
			mappedPorts := make(map[string]bool)

			// Helper function to add port mapping and track it to avoid duplicates
			addPortMapping := func(containerPort, hostPort int, protocol string) {
				portKey := fmt.Sprintf("%d/%s", containerPort, protocol)
				if _, exists := mappedPorts[portKey]; !exists {
					kindConfig += fmt.Sprintf("  - containerPort: %d\n    hostPort: %d\n    protocol: %s\n",
						containerPort, hostPort, protocol)
					mappedPorts[portKey] = true
					if verbose {
						fmt.Printf("%s Added port mapping: container %d -> host %d (%s)\n",
							yellow("📡"), containerPort, hostPort, protocol)
					}
				} else if verbose {
					fmt.Printf("%s Skipping duplicate port mapping: container port %d/%s\n",
						yellow("ℹ️"), containerPort, protocol)
				}
			}

			// Process port mappings from config.Cluster.MapPorts
			if verbose {
				fmt.Println(yellow("Processing port mappings from configuration"))
			}

			for _, portMap := range config.Cluster.MapPorts {
				// Convert containerPort to int (it could be an interface{})
				var containerPort int
				switch cp := portMap.ContainerPort.(type) {
				case int:
					containerPort = cp
				case float64:
					containerPort = int(cp)
				case string:
					// If it's a string, it might be a variable reference or a numeric string
					// Try to parse as a number first
					if val, err := strconv.Atoi(cp); err == nil {
						containerPort = val
					} else {
						// Skip this port mapping if containerPort is not a valid number
						fmt.Printf("%s Skipping invalid containerPort: %v (cannot be converted to a number)\n", yellow("⚠️"), cp)
						continue
					}
				default:
					// Skip this port mapping if containerPort is not a valid number or convertible type
					fmt.Printf("%s Skipping invalid containerPort: %v (unsupported type)\n", yellow("⚠️"), portMap.ContainerPort)
					continue
				}

				// Add the port mapping
				addPortMapping(containerPort, portMap.HostPort, portMap.Protocol)
			}

			// Make sure we don't miss any essential component port mappings
			if verbose {
				fmt.Println(yellow("Verifying essential component port mappings"))
			}

			// Ensure Temporal ports are mapped if enabled
			if config.Components.Temporal.Enabled {
				// Check if Temporal Web UI port is already mapped
				webPortKey := fmt.Sprintf("%d/TCP", config.Components.Temporal.NodePorts.Web)
				if !mappedPorts[webPortKey] {
					fmt.Printf("%s Adding missing Temporal Web UI port mapping\n", yellow("➕"))
					addPortMapping(config.Components.Temporal.NodePorts.Web, 8080, "TCP")
				}

				// Check if Temporal Frontend port is already mapped
				frontendPortKey := fmt.Sprintf("%d/TCP", config.Components.Temporal.NodePorts.Frontend)
				if !mappedPorts[frontendPortKey] {
					fmt.Printf("%s Adding missing Temporal Frontend port mapping\n", yellow("➕"))
					addPortMapping(config.Components.Temporal.NodePorts.Frontend, 7233, "TCP")
				}
			}

			// Ensure Redis port is mapped if enabled
			if config.Components.Redis.Enabled {
				redisPortKey := fmt.Sprintf("%d/TCP", config.Components.Redis.NodePorts.Redis)
				if !mappedPorts[redisPortKey] {
					fmt.Printf("%s Adding missing Redis port mapping\n", yellow("➕"))
					addPortMapping(config.Components.Redis.NodePorts.Redis, 6379, "TCP")
				}
			}

			// Ensure MySQL port is mapped if enabled
			if config.Components.MySQL.Enabled {
				mysqlPortKey := fmt.Sprintf("%d/TCP", config.Components.MySQL.NodePorts.MySQL)
				if !mappedPorts[mysqlPortKey] {
					fmt.Printf("%s Adding missing MySQL port mapping\n", yellow("➕"))
					addPortMapping(config.Components.MySQL.NodePorts.MySQL, 3306, "TCP")
				}
			}

			if verbose {
				fmt.Println(yellow("Port mapping configuration complete"))
			}

			// Create the cluster using the configuration
			kindConfigFile, err := os.CreateTemp("", "kind-config-*.yaml")
			if err != nil {
				fmt.Printf("%s Error creating temporary file for Kind config: %v\n", red("❌"), err)
				os.Exit(1)
			}
			defer os.Remove(kindConfigFile.Name())

			_, err = kindConfigFile.WriteString(kindConfig)
			if err != nil {
				fmt.Printf("%s Error writing Kind config: %v\n", red("❌"), err)
				os.Exit(1)
			}
			kindConfigFile.Close()

			// Handle kind cluster creation differently based on container engine
			if containerEngine == "podman" {
				fmt.Println(yellow("Using podman provider for Kind..."))

				// Create the command with proper environment variables
				cmd := exec.Command("kind", "create", "cluster", "--name", config.Cluster.Name, "--config", kindConfigFile.Name())

				// Set the environment properly for podman
				env := os.Environ()
				cmd.Env = append(env, "KIND_EXPERIMENTAL_PROVIDER=podman")

				// Capture output
				var stdout, stderr bytes.Buffer
				cmd.Stdout = &stdout
				cmd.Stderr = &stderr

				// Run the command
				fmt.Println(yellow("Running kind create cluster with podman provider..."))
				err := cmd.Run()
				if err != nil {
					fmt.Printf("%s Error creating Kind cluster: %v\n", red("❌"), err)
					fmt.Println("Output:", stdout.String())
					fmt.Println("Error:", stderr.String())
					os.Exit(1)
				}

				fmt.Println(stdout.String())
			} else {
				// For docker, use the standard executeCommand function
				_, err = executeCommand("kind", "create", "cluster", "--name", config.Cluster.Name, "--config", kindConfigFile.Name())
				if err != nil {
					fmt.Printf("%s Error creating Kind cluster: %v\n", red("❌"), err)
					os.Exit(1)
				}
			}

			fmt.Printf("%s Kind cluster created successfully\n", green("✅"))
		} else if !clusterExists && !config.Cluster.CreateIfNotExists {
			fmt.Printf("%s Error: Cluster %s does not exist and createIfNotExists is false\n", red("❌"), config.Cluster.Name)
			os.Exit(1)
		}

		// Check current kubectl context before proceeding
		expectedContext := fmt.Sprintf("kind-%s", config.Cluster.Name)
		currentContextCmd := exec.Command("kubectl", "config", "current-context")
		currentContextOutput, err := currentContextCmd.CombinedOutput()
		currentContext := strings.TrimSpace(string(currentContextOutput))

		if err == nil && currentContext != expectedContext {
			fmt.Printf("%s Current kubectl context is: %s\n", yellow("⚠️"), currentContext)
			fmt.Printf("%s Expected kubectl context for this environment is: %s\n", yellow("ℹ️"), expectedContext)

			shouldSwitch := forceContext
			if !forceContext {
				fmt.Print("Switch to the correct context? (y/n): ")
				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				response = strings.TrimSpace(response)

				shouldSwitch = strings.ToLower(response) == "y" || strings.ToLower(response) == "yes"
				if !shouldSwitch {
					fmt.Println(yellow("Context switch declined. This may cause issues with deployment."))
					fmt.Println(yellow("You can manually switch context with:"))
					fmt.Printf("  kubectl config use-context %s\n", expectedContext)
					fmt.Println(yellow("Or run this command with --force-context to switch automatically."))
					os.Exit(1)
				}
			} else {
				fmt.Printf("%s Automatically switching kubectl context from %s to %s\n",
					yellow("⚙️"), currentContext, expectedContext)
			}

			// Explicitly switch context
			switchCmd := exec.Command("kubectl", "config", "use-context", expectedContext)
			switchOutput, err := switchCmd.CombinedOutput()
			if err != nil {
				fmt.Printf("%s Error switching kubectl context: %v\n", red("❌"), err)
				if len(switchOutput) > 0 {
					fmt.Println(string(switchOutput))
				}
				os.Exit(1)
			}
		}

		// Verify we're using the right context
		_, err = executeCommand("kubectl", "cluster-info", "--context", expectedContext)
		if err != nil {
			fmt.Printf("%s Error connecting to cluster with context %s: %v\n", red("❌"), expectedContext, err)
			os.Exit(1)
		}
		fmt.Printf("%s Using kubectl context: %s\n", green("✅"), expectedContext)

		// Setup AWS ECR if needed
		var ecrPassword string
		var ecrRegistry string
		if config.Images.UseAwsEcr {
			fmt.Println(yellow("Setting up AWS ECR credentials"))

			if !commandExists("aws") {
				fmt.Printf("%s Error: AWS CLI is not installed. Please install it first.\n", red("❌"))
				os.Exit(1)
			}

			// Determine AWS profile to use - command line flag takes precedence over config file
			awsProfileToUse := config.Images.AWS.Profile
			if cmd.Flags().Changed("aws-profile") {
				awsProfileToUse = awsProfile
				fmt.Printf("%s Overriding AWS profile from command line: %s\n", yellow("🔑"), awsProfile)
			} else if awsProfileToUse != "" {
				fmt.Printf("%s Using AWS profile from config: %s\n", yellow("🔑"), awsProfileToUse)
			} else {
				// No profile provided in config or command line, prompt the user
				fmt.Printf("%s AWS ECR is enabled but no AWS profile specified\n", yellow("⚠️"))
				fmt.Println(yellow("You can:"))
				fmt.Println("  1. Add a profile field to the images.aws section in kindenv.yaml:")
				fmt.Println("     images:")
				fmt.Println("       aws:")
				fmt.Println("         profile: \"your-profile-name\"")
				fmt.Println("  2. Run this command with --aws-profile flag:")
				fmt.Println("     kindenv start --aws-profile your-profile-name")
				fmt.Println("  3. Ensure your default AWS credentials are configured and continue")
				fmt.Println("")

				// Ask if user wants to continue with default credentials
				fmt.Print("Do you want to continue with default AWS credentials? (y/n): ")
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
					fmt.Println(yellow("Exiting as requested."))
					os.Exit(0)
				}
				fmt.Println(yellow("Continuing with default AWS credentials..."))
			}

			// Use AWS profile if specified
			awsArgs := []string{"sts", "get-caller-identity", "--query", "Account", "--output", "text"}
			if awsProfileToUse != "" {
				awsArgs = append([]string{"--profile", awsProfileToUse}, awsArgs...)
			}

			// Get AWS account ID and ECR credentials
			accountOutput, err := executeCommand("aws", awsArgs...)
			if err != nil {
				fmt.Printf("%s Error getting AWS account ID: %v\n", red("❌"), err)
				fmt.Println(yellow("Authentication failed with AWS CLI."))

				if awsProfileToUse != "" {
					fmt.Printf(yellow("The profile '%s' may not be correctly configured or has expired tokens.\n"), awsProfileToUse)
				} else {
					fmt.Println(yellow("No AWS profile was specified, and default credentials failed."))
				}

				fmt.Println(yellow("You can resolve this by:"))
				fmt.Println("  1. Run 'aws configure' to set up your credentials")
				fmt.Println("  2. Run 'aws sso login' if using AWS SSO")
				fmt.Println("  3. Set up environment variables AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY")
				fmt.Println("  4. Use --aws-profile flag to specify a valid profile")
				fmt.Println("  5. Add a valid profile in your kindenv.yaml configuration:")
				fmt.Println("     images:")
				fmt.Println("       aws:")
				fmt.Println("         profile: \"your-profile-name\"")

				// List available profiles to help the user
				fmt.Println("")
				fmt.Println(yellow("Available AWS profiles:"))
				profileOutput, profileErr := executeCommand("aws", "configure", "list-profiles")
				if profileErr == nil && profileOutput != "" {
					profiles := strings.Split(strings.TrimSpace(profileOutput), "\n")
					for _, profile := range profiles {
						fmt.Printf("  - %s\n", profile)
					}
				} else {
					fmt.Println("  No profiles found or unable to list profiles")
				}

				os.Exit(1)
			}
			accountID := strings.TrimSpace(accountOutput)

			// Use provided region or default
			awsRegion := config.Images.AWS.Region
			if awsRegion == "" {
				awsRegion = "eu-west-1"
			}

			// Set or validate ECR registry
			if config.Images.AWS.EcrRegistry == "" {
				config.Images.AWS.EcrRegistry = fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", accountID, awsRegion)
			}

			// Set ecrRegistry to the config value
			ecrRegistry = config.Images.AWS.EcrRegistry

			// Get ecrPassword from AWS with profile if specified
			ecrPasswordArgs := []string{"ecr", "get-login-password", "--region", awsRegion}
			if awsProfileToUse != "" {
				ecrPasswordArgs = append([]string{"--profile", awsProfileToUse}, ecrPasswordArgs...)
			}

			ecrPasswordOutput, err := executeCommand("aws", ecrPasswordArgs...)
			if err != nil {
				fmt.Printf("%s Error getting ECR credentials: %v\n", red("❌"), err)
				os.Exit(1)
			}
			ecrPassword = strings.TrimSpace(ecrPasswordOutput)

			// Create ECR credentials in default namespace
			err = setupECRCreds("default", ecrRegistry, ecrPassword)
			if err != nil {
				fmt.Printf("%s Error setting up ECR credentials: %v\n", red("❌"), err)
				os.Exit(1)
			}

			fmt.Printf("%s AWS ECR credentials configured\n", green("✅"))
		}

		// Install Metrics Server
		if config.Components.MetricsServer.Enabled {
			fmt.Println(yellow("Installing Metrics Server"))

			// Define Helm arguments
			helmArgs := []string{
				"upgrade",
				"--install",
				"metrics-server", "metrics-server/metrics-server",
				"--namespace", "kube-system",
				"--version", config.Components.MetricsServer.ChartVersion,
				"--set", "args={--kubelet-insecure-tls}",
			}

			// Execute Helm command
			if verbose {
				fmt.Printf("Command: helm %s\n", strings.Join(helmArgs, " "))
			}

			helmOutput, err := executeCommand("helm", helmArgs...)
			if err != nil {
				fmt.Printf("%s Error installing Metrics Server: %v\n", red("❌"), err)
				if helmOutput != "" {
					fmt.Println("Output:")
					fmt.Println(helmOutput)
				}
				fmt.Println(yellow("Continuing despite Metrics Server installation failure..."))
			} else {
				fmt.Printf("%s Metrics Server installed successfully\n", green("✅"))

				// Wait for Metrics Server to be ready
				fmt.Println(yellow("Waiting for Metrics Server to be ready..."))

				// Wait a moment for resources to be created
				time.Sleep(5 * time.Second)

				err = waitForDeployment("kube-system", "metrics-server", 2)
				if err != nil {
					fmt.Printf("%s Error waiting for Metrics Server: %v\n", red("❌"), err)
					fmt.Println(yellow("Continuing despite Metrics Server not being ready..."))
				} else {
					fmt.Printf("%s Metrics Server is ready\n", green("✅"))
					fmt.Println(yellow("You can now use these commands to view resource usage:"))
					fmt.Println("  kubectl top nodes    - Shows CPU and memory usage for each node")
					fmt.Println("  kubectl top pods -A  - Shows CPU and memory usage for all pods")
					fmt.Println(yellow("Note: It may take a few minutes for metrics to be available after installation"))
				}
			}
		}

		// Create MySQL secret if MySQL secrets are enabled
		if config.Secrets.MySQL.Enabled {
			fmt.Println(yellow("Creating MySQL credentials secret"))

			mysqlSecretYaml := fmt.Sprintf(`
apiVersion: v1
kind: Secret
metadata:
  name: %s
  namespace: %s
type: Opaque
stringData:
  mysql-root-password: "%s"
  mysql-password: "%s"
`, config.Secrets.MySQL.Name, config.Secrets.MySQL.Namespace,
				config.Secrets.MySQL.Password, config.Secrets.MySQL.Password)

			cmd := exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = strings.NewReader(mysqlSecretYaml)
			err = cmd.Run()
			if err != nil {
				fmt.Printf("%s Error creating MySQL secret: %v\n", red("❌"), err)
				os.Exit(1)
			}

			fmt.Printf("%s MySQL credentials secret created\n", green("✅"))
		}

		// Install Redis
		if config.Components.Redis.Enabled {
			fmt.Println(yellow("Installing Redis"))

			// Create namespace
			namespaceYaml, err := executeCommand("kubectl", "create", "namespace", "redis", "--dry-run=client", "-o", "yaml")
			if err != nil {
				fmt.Printf("%s Error creating Redis namespace: %v\n", red("❌"), err)
				os.Exit(1)
			}

			cmd := exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = strings.NewReader(namespaceYaml)
			if err := cmd.Run(); err != nil {
				fmt.Printf("%s Error applying Redis namespace: %v\n", red("❌"), err)
				os.Exit(1)
			}

			// Set up ECR credentials if needed
			if config.Images.UseAwsEcr {
				err = setupECRCreds("redis", ecrRegistry, ecrPassword)
				if err != nil {
					fmt.Printf("%s Error setting up ECR credentials for Redis: %v\n", red("❌"), err)
					os.Exit(1)
				}
			}

			// Install Redis with Helm
			_, err = executeCommand("helm", "upgrade", "--install",
				"redis", "bitnami/redis",
				"--namespace", "redis",
				"--version", config.Components.Redis.ChartVersion,
				"--set", "master.service.type=NodePort",
				"--set", fmt.Sprintf("master.service.nodePorts.redis=%d", config.Components.Redis.NodePorts.Redis),
				"--set", fmt.Sprintf("auth.enabled=%t", config.Components.Redis.Auth.Enabled),
				"--set", "replica.replicaCount=0",
				"--set", "image.repository=bitnamilegacy/redis")
			if err != nil {
				fmt.Printf("%s Error installing Redis: %v\n", red("❌"), err)
				os.Exit(1)
			}

			// Wait for Redis to be ready using label selectors
			fmt.Println(yellow("Waiting for Redis to be ready..."))

			// Give resources a moment to be created
			fmt.Println(yellow("Pausing briefly to allow resources to be created..."))
			time.Sleep(10 * time.Second) // Increased wait time to ensure pod is created

			// Wait specifically for the redis-master-0 pod
			fmt.Println(yellow("Waiting for redis-master-0 pod to be created..."))
			// First check if the pod exists
			podCheckCmd := exec.Command("kubectl", "get", "pod", "redis-master-0", "-n", "redis", "--no-headers")

			// Retry a few times for the pod to appear
			var podExists bool
			for i := 0; i < 6; i++ { // Try for up to 30 seconds (6 * 5s)
				podOutput, err := podCheckCmd.CombinedOutput()
				if err == nil && len(podOutput) > 0 {
					podExists = true
					if verbose {
						fmt.Println(string(podOutput))
					}
					break
				}
				if i < 5 { // Don't sleep on the last iteration
					fmt.Printf("Waiting for redis-master-0 pod to appear (attempt %d/6)...\n", i+1)
					time.Sleep(5 * time.Second)
				}
			}

			if podExists {
				fmt.Println(yellow("Found redis-master-0 pod, waiting for it to be ready..."))
				_, err = executeCommand("kubectl", "wait", "--for=condition=Ready", "pod/redis-master-0", "-n", "redis", "--timeout=2m")
				if err != nil {
					fmt.Printf("%s Warning: Redis master pod is not ready: %v\n", yellow("⚠️"), err)
					fmt.Println(yellow("Continuing despite Redis not being fully ready..."))
				} else {
					fmt.Printf("%s Redis is ready\n", green("✅"))
				}
			} else {
				fmt.Printf("%s Redis master pod (redis-master-0) not found\n", yellow("⚠️"))
				fmt.Println(yellow("Continuing despite Redis pod not being detected..."))
			}

			// Create kvv2-redis secret with redis address
			redisPassword := ""
			// If Redis auth is enabled, use default password
			if config.Components.Redis.Auth.Enabled {
				redisPassword = "redis"
			}

			kvv2RedisSecretYaml := fmt.Sprintf(`
apiVersion: v1
kind: Secret
metadata:
  name: kvv2-redis
  namespace: redis
type: Opaque
stringData:
  address: "redis-master.redis.svc.cluster.local:6379"
  redis-password: "%s"
`, redisPassword)

			cmd = exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = strings.NewReader(kvv2RedisSecretYaml)
			err = cmd.Run()
			if err != nil {
				fmt.Printf("%s Error creating Redis secret in redis namespace: %v\n", red("❌"), err)
			} else {
				fmt.Printf("%s Redis secret created successfully in redis namespace\n", green("✅"))
			}

			// Also create kvv2-redis secret in default namespace for other components
			defaultRedisSecretYaml := fmt.Sprintf(`
apiVersion: v1
kind: Secret
metadata:
  name: kvv2-redis
  namespace: default
type: Opaque
stringData:
  address: "redis-master.redis.svc.cluster.local:6379"
  redis-password: "%s"
`, redisPassword)
			cmd = exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = strings.NewReader(defaultRedisSecretYaml)
			err = cmd.Run()
			if err != nil {
				fmt.Printf("%s Error creating Redis secret in default namespace: %v\n", red("❌"), err)
			} else {
				fmt.Printf("%s Redis secret created successfully in default namespace\n", green("✅"))
			}
		}

		// Install MySQL
		if config.Components.MySQL.Enabled {
			fmt.Println(yellow("Installing MySQL"))

			// Create namespace
			namespaceYaml, err := executeCommand("kubectl", "create", "namespace", config.Components.MySQL.Namespace, "--dry-run=client", "-o", "yaml")
			if err != nil {
				fmt.Printf("%s Error creating MySQL namespace: %v\n", red("❌"), err)
				os.Exit(1)
			}

			cmd := exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = strings.NewReader(namespaceYaml)
			if err := cmd.Run(); err != nil {
				fmt.Printf("%s Error applying MySQL namespace: %v\n", red("❌"), err)
				os.Exit(1)
			}

			// Set up ECR credentials if needed
			if config.Images.UseAwsEcr {
				err = setupECRCreds(config.Components.MySQL.Namespace, ecrRegistry, ecrPassword)
				if err != nil {
					fmt.Printf("%s Error setting up ECR credentials for MySQL: %v\n", red("❌"), err)
					os.Exit(1)
				}
			}

			// Build Helm arguments for MySQL installation
			helmArgs := []string{
				"upgrade", "--install",
				"mysql", "bitnami/mysql",
				"--namespace", config.Components.MySQL.Namespace,
				"--version", config.Components.MySQL.ChartVersion,
				"--set", "primary.service.type=NodePort",
				"--set", fmt.Sprintf("primary.service.nodePorts.mysql=%d", config.Components.MySQL.NodePorts.MySQL),
				"--set", fmt.Sprintf("auth.database=%s", config.Components.MySQL.Database),
				"--set", fmt.Sprintf("primary.persistence.enabled=%t", config.Components.MySQL.Persistence.Enabled),
				"--set", fmt.Sprintf("primary.resources.requests.cpu=%s", config.Components.MySQL.Resources.CPU),
				"--set", fmt.Sprintf("primary.resources.requests.memory=%s", config.Components.MySQL.Resources.Memory),
				"--set", "secondary.replicaCount=0",
			}

			// Add secret configuration if MySQL secrets are enabled
			if config.Secrets.MySQL.Enabled {
				helmArgs = append(helmArgs, "--set", fmt.Sprintf("auth.existingSecret=%s", config.Secrets.MySQL.Name))
			} else {
				// Use default credentials
				helmArgs = append(helmArgs,
					"--set", "auth.rootPassword=password",
					"--set", "auth.username=mysql",
					"--set", "auth.password=password")
			}

			// ECR-specific image configuration
			if config.Images.UseAwsEcr {
				helmArgs = append(helmArgs,
					"--set", fmt.Sprintf("global.imageRegistry=%s", ecrRegistry),
					"--set", "image.repository=bitnamilegacy/mysql")
			}

			// Add persistence size if enabled
			if config.Components.MySQL.Persistence.Enabled {
				helmArgs = append(helmArgs, "--set", fmt.Sprintf("primary.persistence.size=%s", config.Components.MySQL.Persistence.Size))
			}

			_, err = executeCommand("helm", helmArgs...)
			if err != nil {
				fmt.Printf("%s Error installing MySQL: %v\n", red("❌"), err)
				os.Exit(1)
			}

			// Wait for MySQL to be ready
			fmt.Println(yellow("Waiting for MySQL to be ready..."))
			time.Sleep(10 * time.Second)

			// Wait for MySQL pod
			fmt.Println(yellow("Waiting for mysql-primary-0 pod to be created..."))
			podCheckCmd := exec.Command("kubectl", "get", "pod", "mysql-primary-0", "-n", config.Components.MySQL.Namespace, "--no-headers")

			var podExists bool
			for i := 0; i < 10; i++ { // Try for up to 5 minutes (10 * 30s)
				podOutput, err := podCheckCmd.CombinedOutput()
				if err == nil && len(podOutput) > 0 {
					podExists = true
					if verbose {
						fmt.Println(string(podOutput))
					}
					break
				}
				if i < 9 {
					fmt.Printf("Waiting for mysql-primary-0 pod to appear (attempt %d/10)...\n", i+1)
					time.Sleep(30 * time.Second)
				}
			}

			if podExists {
				fmt.Println(yellow("Found mysql-primary-0 pod, waiting for it to be ready..."))
				_, err = executeCommand("kubectl", "wait", "--for=condition=Ready", "pod/mysql-primary-0", "-n", config.Components.MySQL.Namespace, "--timeout=5m")
				if err != nil {
					fmt.Printf("%s Warning: MySQL pod is not ready: %v\n", yellow("⚠️"), err)
					fmt.Println(yellow("Continuing despite MySQL not being fully ready..."))
				} else {
					fmt.Printf("%s MySQL installed successfully\n", green("✅"))
				}
			} else {
				fmt.Printf("%s MySQL master pod (mysql-primary-0) not found\n", yellow("⚠️"))
				fmt.Println(yellow("Continuing despite MySQL pod not being detected..."))
			}
		}

		// Install Dapr
		if config.Components.Dapr.Enabled {
			fmt.Println(yellow("Installing Dapr"))

			// Create namespace
			namespaceYaml, err := executeCommand("kubectl", "create", "namespace", "dapr-system", "--dry-run=client", "-o", "yaml")
			if err != nil {
				fmt.Printf("%s Error creating Dapr namespace: %v\n", red("❌"), err)
				os.Exit(1)
			}

			cmd := exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = strings.NewReader(namespaceYaml)
			if err := cmd.Run(); err != nil {
				fmt.Printf("%s Error applying Dapr namespace: %v\n", red("❌"), err)
				os.Exit(1)
			}

			// Set up ECR credentials if needed
			if config.Images.UseAwsEcr {
				err = setupECRCreds("dapr-system", ecrRegistry, ecrPassword)
				if err != nil {
					fmt.Printf("%s Error setting up ECR credentials for Dapr: %v\n", red("❌"), err)
					os.Exit(1)
				}
			}

			// Install Dapr with Helm
			_, err = executeCommand("helm", "upgrade", "--install",
				"dapr", "dapr/dapr",
				"--namespace", "dapr-system",
				"--version", config.Components.Dapr.ChartVersion,
				"--set", fmt.Sprintf("global.logLevel=%s", config.Components.Dapr.LogLevel),
				"--set", fmt.Sprintf("global.ha.enabled=%t", config.Components.Dapr.Ha.Enabled),
				"--set", fmt.Sprintf("global.mtls.enabled=%t", config.Components.Dapr.Mtls.Enabled),
				"--wait", "--timeout", "5m")
			if err != nil {
				fmt.Printf("%s Error installing Dapr: %v\n", red("❌"), err)
				os.Exit(1)
			}

			// Wait for Dapr components to be ready
			err = waitForDeployment("dapr-system", "dapr-operator", 2)
			if err != nil {
				fmt.Printf("%s Error waiting for Dapr operator: %v\n", red("❌"), err)
				fmt.Println(yellow("Continuing despite Dapr operator not being ready..."))
			}

			err = waitForDeployment("dapr-system", "dapr-sentry", 2)
			if err != nil {
				fmt.Printf("%s Error waiting for Dapr sentry: %v\n", red("❌"), err)
				fmt.Println(yellow("Continuing despite Dapr sentry not being ready..."))
			}

			err = waitForDeployment("dapr-system", "dapr-sidecar-injector", 2)
			if err != nil {
				fmt.Printf("%s Error waiting for Dapr sidecar injector: %v\n", red("❌"), err)
				fmt.Println(yellow("Continuing despite Dapr sidecar injector not being ready..."))
			} else {
				fmt.Printf("%s Dapr installed successfully\n", green("✅"))
			}
		}

		// Install OpenSearch
		if config.Components.OpenSearch.Enabled {
			fmt.Println(yellow("Installing OpenSearch"))

			// Create namespace
			namespaceYaml, err := executeCommand("kubectl", "create", "namespace", config.Components.OpenSearch.Namespace, "--dry-run=client", "-o", "yaml")
			if err != nil {
				fmt.Printf("%s Error creating OpenSearch namespace: %v\n", red("❌"), err)
				os.Exit(1)
			}

			cmd := exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = strings.NewReader(namespaceYaml)
			if err := cmd.Run(); err != nil {
				fmt.Printf("%s Error applying OpenSearch namespace: %v\n", red("❌"), err)
				os.Exit(1)
			}

			// Set up ECR credentials if needed
			if config.Images.UseAwsEcr {
				err = setupECRCreds(config.Components.OpenSearch.Namespace, ecrRegistry, ecrPassword)
				if err != nil {
					fmt.Printf("%s Error setting up ECR credentials for OpenSearch: %v\n", red("❌"), err)
					os.Exit(1)
				}
			}

			// Deploy OpenSearch using direct Kubernetes manifests
			opensearchYaml := fmt.Sprintf(`
apiVersion: v1
kind: Service
metadata:
  name: opensearch
  namespace: %s
  labels:
    app: opensearch
spec:
  type: NodePort
  ports:
  - port: 9200
    targetPort: 9200
    nodePort: %d
    name: rest
  - port: 9300
    name: inter-node
  selector:
    app: opensearch
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: opensearch
  namespace: %s
spec:
  serviceName: opensearch
  replicas: 1
  selector:
    matchLabels:
      app: opensearch
  template:
    metadata:
      labels:
        app: opensearch
    spec:
      containers:
      - name: opensearch
        image: opensearchproject/opensearch:%s
        command:
        - bash
        - -c
        - |
          echo "Starting OpenSearch..."
          until [ -f /usr/share/opensearch/config/opensearch.yml ]; do sleep 1; done

          # Create snapshots directory with proper permissions
          mkdir -p /tmp/opensearch-snapshots
          chmod 777 /tmp/opensearch-snapshots

          # Configure repository paths and plugin settings
          echo "path.repo: [\"/tmp/opensearch-snapshots\"]" >> /usr/share/opensearch/config/opensearch.yml
          echo "plugins.security.disabled: %t" >> /usr/share/opensearch/config/opensearch.yml
          echo "action.auto_create_index: true" >> /usr/share/opensearch/config/opensearch.yml

          # Enable index management plugin (includes snapshot management functionality)
          echo "plugins.index_state_management.enabled: %t" >> /usr/share/opensearch/config/opensearch.yml

          # Start OpenSearch in foreground
          /usr/share/opensearch/opensearch-docker-entrypoint.sh
        env:
        - name: discovery.type
          value: single-node
        - name: bootstrap.memory_lock
          value: "true"
        - name: OPENSEARCH_JAVA_OPTS
          value: "-Xms512m -Xmx512m"
        - name: _JAVA_OPTIONS
          value: "-XX:UseSVE=0"
        - name: DISABLE_SECURITY_PLUGIN
          value: "%t"
        ports:
        - containerPort: 9200
          name: rest
        - containerPort: 9300
          name: inter-node
        resources:
          limits:
            cpu: 1000m
            memory: 1Gi
          requests:
            cpu: 100m
            memory: 1Gi
        readinessProbe:
          httpGet:
            path: /_cluster/health?local=true
            port: 9200
          initialDelaySeconds: 90
          periodSeconds: 15
          failureThreshold: 15
          timeoutSeconds: 10
`, config.Components.OpenSearch.Namespace, config.Components.OpenSearch.NodePorts.Rest,
				config.Components.OpenSearch.Namespace, config.Components.OpenSearch.Version,
				config.Components.OpenSearch.Security.Disabled,
				config.Components.OpenSearch.Security.Disabled,
				config.Components.OpenSearch.IndexManagement.Enabled)

			cmd = exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = strings.NewReader(opensearchYaml)
			if err := cmd.Run(); err != nil {
				fmt.Printf("%s Error deploying OpenSearch: %v\n", red("❌"), err)
				os.Exit(1)
			}

			// Wait for OpenSearch pod to be created
			fmt.Println(yellow("Waiting for OpenSearch pod to be created..."))
			maxPodCreateRetries := 20
			podCreateRetryCount := 0
			var opensearchPodExists bool

			for podCreateRetryCount < maxPodCreateRetries {
				podCheckCmd := exec.Command("kubectl", "get", "pod", "-l", "app=opensearch",
					"-n", config.Components.OpenSearch.Namespace, "--no-headers")
				podOutput, err := podCheckCmd.CombinedOutput()
				if err == nil && len(podOutput) > 0 && !strings.Contains(string(podOutput), "No resources found") {
					opensearchPodExists = true
					if verbose {
						fmt.Println(string(podOutput))
					}
					fmt.Printf("%s OpenSearch pod is created\n", green("✅"))
					break
				}
				podCreateRetryCount++
				fmt.Print(".")
				time.Sleep(3 * time.Second)
			}

			if !opensearchPodExists {
				fmt.Printf("%s Timed out waiting for OpenSearch pod to be created\n", yellow("⚠️"))
				fmt.Println(yellow("Continuing despite OpenSearch pod not being detected..."))
			} else {
				// Wait for OpenSearch pod to be ready
				fmt.Println(yellow("Waiting for OpenSearch pod to become ready..."))
				_, err = executeCommand("kubectl", "wait", "--for=condition=Ready", "pod", "-l", "app=opensearch",
					"-n", config.Components.OpenSearch.Namespace, "--timeout=5m")
				if err != nil {
					fmt.Printf("%s Warning: OpenSearch pod is not ready: %v\n", yellow("⚠️"), err)
					fmt.Println(yellow("Continuing despite OpenSearch not being fully ready..."))
				} else {
					fmt.Printf("%s OpenSearch is ready\n", green("✅"))

					// Find the proper host port for OpenSearch
					var openSearchHostPort int
					for _, portMap := range config.Cluster.MapPorts {
						// Handle different types of containerPort values
						switch cp := portMap.ContainerPort.(type) {
						case int:
							if cp == config.Components.OpenSearch.NodePorts.Rest {
								openSearchHostPort = portMap.HostPort
								break
							}
						}
					}

					// Use the mapped host port or default to 9200 if not found
					if openSearchHostPort == 0 {
						openSearchHostPort = 9200
					}

					fmt.Printf("%s OpenSearch is accessible at http://localhost:%d\n",
						green("✅"), openSearchHostPort)
				}
			}
		}

		// Install OpenSearch Dashboards
		if config.Components.OpenSearchDashboards.Enabled && config.Components.OpenSearch.Enabled {
			fmt.Println(yellow("Installing OpenSearch Dashboards"))

			// OpenSearch Dashboards uses the same namespace as OpenSearch
			// Deploy OpenSearch Dashboards using direct Kubernetes manifests
			dashboardsYaml := fmt.Sprintf(`
apiVersion: v1
kind: Service
metadata:
  name: opensearch-dashboards
  namespace: %s
  labels:
    app: opensearch-dashboards
spec:
  type: NodePort
  ports:
  - port: 5601
    targetPort: 5601
    nodePort: %d
    name: http
  selector:
    app: opensearch-dashboards
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: opensearch-dashboards
  namespace: %s
spec:
  replicas: 1
  selector:
    matchLabels:
      app: opensearch-dashboards
  template:
    metadata:
      labels:
        app: opensearch-dashboards
    spec:
      containers:
      - name: opensearch-dashboards
        image: opensearchproject/opensearch-dashboards:%s
        env:
        - name: OPENSEARCH_HOSTS
          value: '["http://opensearch:9200"]'
        - name: DISABLE_SECURITY_DASHBOARDS_PLUGIN
          value: "%t"
        ports:
        - containerPort: 5601
          name: http
        resources:
          limits:
            cpu: 1000m
            memory: 1Gi
          requests:
            cpu: 100m
            memory: 512Mi
        readinessProbe:
          httpGet:
            path: /api/status
            port: 5601
          initialDelaySeconds: 60
          periodSeconds: 15
          failureThreshold: 10
          timeoutSeconds: 10
`, config.Components.OpenSearchDashboards.Namespace, config.Components.OpenSearchDashboards.NodePorts.Http,
				config.Components.OpenSearchDashboards.Namespace, config.Components.OpenSearchDashboards.Version,
				config.Components.OpenSearch.Security.Disabled)

			cmd := exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = strings.NewReader(dashboardsYaml)
			if err := cmd.Run(); err != nil {
				fmt.Printf("%s Error deploying OpenSearch Dashboards: %v\n", red("❌"), err)
				os.Exit(1)
			}

			// Wait for OpenSearch Dashboards pod to be created
			fmt.Println(yellow("Waiting for OpenSearch Dashboards pod to be created..."))
			maxPodCreateRetries := 20
			podCreateRetryCount := 0
			var dashboardsPodExists bool

			for podCreateRetryCount < maxPodCreateRetries {
				podCheckCmd := exec.Command("kubectl", "get", "pod", "-l", "app=opensearch-dashboards",
					"-n", config.Components.OpenSearchDashboards.Namespace, "--no-headers")
				podOutput, err := podCheckCmd.CombinedOutput()
				if err == nil && len(podOutput) > 0 && !strings.Contains(string(podOutput), "No resources found") {
					dashboardsPodExists = true
					if verbose {
						fmt.Println(string(podOutput))
					}
					fmt.Printf("%s OpenSearch Dashboards pod is created\n", green("✅"))
					break
				}
				podCreateRetryCount++
				fmt.Print(".")
				time.Sleep(3 * time.Second)
			}

			if !dashboardsPodExists {
				fmt.Printf("%s Timed out waiting for OpenSearch Dashboards pod to be created\n", yellow("⚠️"))
				fmt.Println(yellow("Continuing despite OpenSearch Dashboards pod not being detected..."))
			} else {
				// Wait for OpenSearch Dashboards pod to be ready
				fmt.Println(yellow("Waiting for OpenSearch Dashboards pod to become ready..."))
				_, err := executeCommand("kubectl", "wait", "--for=condition=Ready", "pod", "-l", "app=opensearch-dashboards",
					"-n", config.Components.OpenSearchDashboards.Namespace, "--timeout=5m")
				if err != nil {
					fmt.Printf("%s Warning: OpenSearch Dashboards pod is not ready: %v\n", yellow("⚠️"), err)
					fmt.Println(yellow("Continuing despite OpenSearch Dashboards not being fully ready..."))
				} else {
					fmt.Printf("%s OpenSearch Dashboards is ready\n", green("✅"))

					// Find the proper host port for OpenSearch Dashboards
					var dashboardsHostPort int
					for _, portMap := range config.Cluster.MapPorts {
						// Handle different types of containerPort values
						switch cp := portMap.ContainerPort.(type) {
						case int:
							if cp == config.Components.OpenSearchDashboards.NodePorts.Http {
								dashboardsHostPort = portMap.HostPort
								break
							}
						}
					}

					// Use the mapped host port or default to 5601 if not found
					if dashboardsHostPort == 0 {
						dashboardsHostPort = 5601
					}

					fmt.Printf("%s OpenSearch Dashboards is accessible at http://localhost:%d\n",
						green("✅"), dashboardsHostPort)
				}
			}
		}

		// Install Temporal
		if config.Components.Temporal.Enabled {
			fmt.Println(yellow("Installing Temporal"))

			// Create namespace
			namespaceYaml, err := executeCommand("kubectl", "create", "namespace", config.Components.Temporal.Namespace, "--dry-run=client", "-o", "yaml")
			if err != nil {
				fmt.Printf("%s Error creating Temporal namespace: %v\n", red("❌"), err)
				os.Exit(1)
			}

			cmd := exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = strings.NewReader(namespaceYaml)
			if err := cmd.Run(); err != nil {
				fmt.Printf("%s Error applying Temporal namespace: %v\n", red("❌"), err)
				os.Exit(1)
			}

			// Set up ECR credentials if needed
			if config.Images.UseAwsEcr {
				err = setupECRCreds(config.Components.Temporal.Namespace, ecrRegistry, ecrPassword)
				if err != nil {
					fmt.Printf("%s Error setting up ECR credentials for Temporal: %v\n", red("❌"), err)
					os.Exit(1)
				}
			}

			// Install Temporal with Helm
			_, err = executeCommand("helm", "upgrade", "--install",
				"temporal", "temporalio/temporal",
				"--namespace", config.Components.Temporal.Namespace,
				"--version", config.Components.Temporal.ChartVersion,
				"--set", "server.replicaCount=1",
				"--set", "cassandra.config.cluster_size=1",
				"--set", "elasticsearch.replicas=1",
				"--set", "prometheus.enabled=false",
				"--set", "grafana.enabled=false",
				"--set", "server.frontend.service.type=NodePort",
				"--set", fmt.Sprintf("server.frontend.service.nodePort=%d", config.Components.Temporal.NodePorts.Frontend),
				"--set", "web.enabled=true",
				"--set", "web.service.type=NodePort",
				"--set", fmt.Sprintf("web.service.nodePort=%d", config.Components.Temporal.NodePorts.Web),
				"--timeout", "15m")
			if err != nil {
				fmt.Printf("%s Error installing Temporal: %v\n", red("❌"), err)
				os.Exit(1)
			}

			// Wait for Temporal to be ready
			err = waitForDeployment(config.Components.Temporal.Namespace, "temporal-frontend", 5)
			if err != nil {
				fmt.Printf("%s Error waiting for Temporal frontend: %v\n", red("❌"), err)
				fmt.Println(yellow("Continuing despite Temporal frontend not being ready..."))
			}

			err = waitForDeployment(config.Components.Temporal.Namespace, "temporal-web", 5)
			if err != nil {
				fmt.Printf("%s Error waiting for Temporal web: %v\n", red("❌"), err)
				fmt.Println(yellow("Continuing despite Temporal web not being ready..."))
			} else {
				fmt.Printf("%s Temporal installed successfully\n", green("✅"))
			}
		}

		// Install Temporal Worker Operator
		if config.Components.TemporalWorkerOperator.Enabled {
			fmt.Println(yellow("Installing Temporal Worker Operator"))

			// Set up ECR credentials if needed (default namespace already has them)
			if config.Images.UseAwsEcr {
				fmt.Println(yellow("Setting up ECR credentials in shield-system namespace"))
				err = setupECRCreds("shield-system", ecrRegistry, ecrPassword)
				if err != nil {
					fmt.Printf("%s Error setting up ECR credentials for shield-system: %v\n", red("❌"), err)
					os.Exit(1)
				}
			}

			// Install Temporal Worker Operator with CRDs first
			fmt.Println(yellow("Installing Temporal Worker Operator (with CRDs)..."))

			var helmArgs []string
			if config.Components.Redis.Enabled {
				// Create kvv2-redis secret with redis address
				redisPassword := ""
				// If Redis auth is enabled, use default password
				if config.Components.Redis.Auth.Enabled {
					redisPassword = "redis"
				}

				kvv2RedisSecretYaml := fmt.Sprintf(`
apiVersion: v1
kind: Secret
metadata:
  name: kvv2-redis
  namespace: shield-system
type: Opaque
stringData:
  address: "redis-master.redis.svc.cluster.local:6379"
  redis-password: "%s"
`, redisPassword)

				cmd := exec.Command("kubectl", "apply", "-f", "-")
				cmd.Stdin = strings.NewReader(kvv2RedisSecretYaml)
				err = cmd.Run()
				if err != nil {
					fmt.Printf("%s Error creating Redis secret: %v\n", red("❌"), err)
				}
				helmArgs = []string{
					"upgrade",
					"--install",
					"temporal-worker-operator", "shield/temporal-worker-operator",
					"--namespace", "shield-system",
					"--version", config.Components.TemporalWorkerOperator.ChartVersion,
					"--create-namespace",
					"--set", "imagePullSecrets[0].name=ecr-credentials",
					"--set", "redis.deployChart=false",
					"--set", "redis.external.secretName=kvv2-redis",
					"--set", "redis.auth.enabled=false",
					"--set", fmt.Sprintf("temporal.namespaces.items[0].name=%s", config.Components.TemporalWorkerOperator.TemporalNamespace),
					"--set", "temporal.namespaces.items[0].description=Default namespace for general workloads",
					"--set", "temporal.namespaces.items[0].retentionPeriod=7d",
					"--set", "temporal.namespaces.enabled=true",
				}
			} else {
				helmArgs = []string{
					"upgrade",
					"--install",
					"temporal-worker-operator", "shield/temporal-worker-operator",
					"--namespace", "shield-system",
					"--version", config.Components.TemporalWorkerOperator.ChartVersion,
					"--create-namespace",
					"--set", "imagePullSecrets[0].name=ecr-credentials",
					"--set", "redis.deployChart=true",
					"--set", "redis.global.imageRegistry=docker.io",
					"--set", fmt.Sprintf("temporal.namespaces.items[0].name=%s", config.Components.TemporalWorkerOperator.TemporalNamespace),
					"--set", "temporal.namespaces.items[0].description=Default namespace for general workloads",
					"--set", "temporal.namespaces.items[0].retentionPeriod=7d",
					"--set", "temporal.namespaces.enabled=true",
				}
			}

			// Execute Helm command
			if verbose {
				fmt.Printf("Command: helm %s\n", strings.Join(helmArgs, " "))
			}

			helmOutput, err := executeCommand("helm", helmArgs...)
			if err != nil {
				fmt.Printf("%s Error installing Temporal Worker Operator: %v\n", red("❌"), err)
				if helmOutput != "" {
					fmt.Println("Helm output:")
					fmt.Println(helmOutput)
				}

				// Check if the error is due to chart not found
				if strings.Contains(err.Error(), "chart not found") ||
					strings.Contains(helmOutput, "chart not found") {
					fmt.Println(yellow("The chart was not found. Make sure:"))
					fmt.Println("1. The Shield Helm repository is properly added: helm repo add shield https://harbor.shieldfis.com/chartrepo/stable")
					fmt.Println("2. The Helm repositories are updated: helm repo update")
					fmt.Println("3. The chart exists: helm search repo shield/temporal-worker-operator")
				}

				// Exit since we can't continue without the operator
				fmt.Println(red("Unable to continue without the Temporal Worker Operator."))
				os.Exit(1)
			}

			fmt.Printf("%s Temporal Worker Operator installed successfully\n", green("✅"))
			fmt.Printf("%s Temporal namespace '%s' will be created by the operator\n", green("✅"), config.Components.TemporalWorkerOperator.TemporalNamespace)

			// Wait for Temporal Worker Operator to be ready
			fmt.Println(yellow("Waiting for Temporal Worker Operator to be ready..."))

			// Wait a moment for CRDs to be established and resources to be created
			fmt.Println(yellow("Waiting for resources to be established..."))
			time.Sleep(10 * time.Second)

			// First check if there's a deployment for the operator
			deploymentOutput, _ := executeCommand("kubectl", "get", "deployment",
				"-n", "default", "--no-headers")

			if deploymentOutput != "" {
				// Try to find the operator deployment name
				operatorDeployments := strings.Split(strings.TrimSpace(deploymentOutput), "\n")
				if len(operatorDeployments) > 0 {
					for _, deploymentLine := range operatorDeployments {
						parts := strings.Fields(deploymentLine)
						if len(parts) > 0 {
							deploymentName := parts[0]
							if strings.Contains(deploymentName, "temporal-worker-operator") {
								err = waitForDeployment("default", deploymentName, 5)
								if err != nil {
									fmt.Printf("%s Error waiting for Temporal Worker Operator: %v\n", red("❌"), err)
									fmt.Println(yellow("Continuing despite Temporal Worker Operator not being ready..."))
								}
							}
						}
					}
				}
			}

			fmt.Printf("%s Temporal Worker Operator installation completed\n", green("✅"))
		}

		// Install Indices Operator
		if config.Components.IndicesOperator.Enabled && config.Components.OpenSearch.Enabled {
			fmt.Println(yellow("Installing Indices Operator"))

			// Create kvv2-opensearch secret with opensearch connection information
			// This will be used by the indices operator
			kvv2OpenSearchSecretYaml := fmt.Sprintf(`
apiVersion: v1
kind: Secret
metadata:
  name: kvv2-opensearch
  namespace: default
type: Opaque
stringData:
  address: "http://opensearch.%s.svc.cluster.local:9200"
  username: "admin"
  password: "admin"
`, config.Components.OpenSearch.Namespace)

			cmd := exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = strings.NewReader(kvv2OpenSearchSecretYaml)
			err = cmd.Run()
			if err != nil {
				fmt.Printf("%s Error creating OpenSearch secret: %v\n", red("❌"), err)
			} else {
				fmt.Printf("%s OpenSearch secret created successfully\n", green("✅"))
			}

			// Install Indices Operator with Helm
			fmt.Println(yellow("Installing Indices Operator..."))

			// Define Helm arguments
			helmArgs := []string{
				"upgrade",
				"--install",
				"indices-operator", "shield/indices-operator",
				"--namespace", "shield-system",
				"--version", config.Components.IndicesOperator.ChartVersion,
				"--set", "opensearch.skipTlsVerify=true",
				"--set", "opensearch.secretName=kvv2-opensearch",
				"--set", "global.dapr.enabled=false", // Disable Dapr for this operator
			}

			// Add ECR credentials if needed
			if config.Images.UseAwsEcr {
				helmArgs = append(helmArgs, "--set", "imagePullSecrets[0].name=ecr-credentials")
			}

			// Execute Helm command
			if verbose {
				fmt.Printf("Command: helm %s\n", strings.Join(helmArgs, " "))
			}

			helmOutput, err := executeCommand("helm", helmArgs...)
			if err != nil {
				fmt.Printf("%s Error installing Indices Operator: %v\n", red("❌"), err)
				if helmOutput != "" {
					fmt.Println("Helm output:")
					fmt.Println(helmOutput)
				}

				// Check if the error is due to chart not found
				if strings.Contains(err.Error(), "chart not found") ||
					strings.Contains(helmOutput, "chart not found") {
					fmt.Println(yellow("The chart was not found. Make sure:"))
					fmt.Println("1. The Shield Helm repository is properly added: helm repo add shield https://harbor.shieldfis.com/chartrepo/stable")
					fmt.Println("2. The Helm repositories are updated: helm repo update")
					fmt.Println("3. The chart exists: helm search repo shield/indices-operator")
				}

				// Continue despite errors, as the operator is optional
				fmt.Println(yellow("Continuing despite Indices Operator installation failure..."))
			} else {
				fmt.Printf("%s Indices Operator installed successfully\n", green("✅"))

				// Wait for Indices Operator to be ready
				fmt.Println(yellow("Waiting for Indices Operator to be ready..."))

				// Wait a moment for CRDs to be established and resources to be created
				fmt.Println(yellow("Waiting for resources to be established..."))
				time.Sleep(10 * time.Second)

				// Check if there's a deployment for the operator
				deploymentOutput, _ := executeCommand("kubectl", "get", "deployment",
					"-n", "default", "--no-headers")

				if deploymentOutput != "" {
					// Try to find the operator deployment name
					operatorDeployments := strings.Split(strings.TrimSpace(deploymentOutput), "\n")
					if len(operatorDeployments) > 0 {
						for _, deploymentLine := range operatorDeployments {
							parts := strings.Fields(deploymentLine)
							if len(parts) > 0 {
								deploymentName := parts[0]
								if strings.Contains(deploymentName, "indices-operator") {
									err = waitForDeployment("default", deploymentName, 5)
									if err != nil {
										fmt.Printf("%s Error waiting for Indices Operator: %v\n", red("❌"), err)
										fmt.Println(yellow("Continuing despite Indices Operator not being ready..."))
									}
								}
							}
						}
					}
				}

				fmt.Printf("%s Indices Operator installation completed\n", green("✅"))
			}
		}

		// Check if metrics are available now that all components are installed
		if config.Components.MetricsServer.Enabled {
			fmt.Println(yellow("Checking if metrics are available..."))

			// Try to run kubectl top nodes to verify it's working
			topCmd := exec.Command("kubectl", "top", "nodes")
			topOutput, topErr := topCmd.CombinedOutput()
			if topErr == nil {
				fmt.Printf("%s Resource metrics are now available\n", green("✅"))
				if verbose {
					fmt.Println(string(topOutput))
				}
			} else {
				fmt.Printf("%s Resource metrics are not available yet (may need a few more minutes)\n", yellow("⚠️"))
				fmt.Println(yellow("You can check again later with:"))
				fmt.Println("  kubectl top nodes    - Shows CPU and memory usage for each node")
				fmt.Println("  kubectl top pods -A  - Shows CPU and memory usage for all pods")
			}
		}

		fmt.Println(green("Kind-based development environment setup complete!"))

		// Find host ports from the port mappings and component configuration
		var temporalWebPort, temporalFrontendPort, redisPort, mysqlPort int

		if verbose {
			fmt.Println(yellow("Port mapping details for accessing services:"))
			for _, portMap := range config.Cluster.MapPorts {
				fmt.Printf("  - Container: %v → Host: %d (%s)\n",
					portMap.ContainerPort, portMap.HostPort, portMap.Protocol)
			}
		}

		// Helper function to find a host port for a given container port
		findHostPort := func(containerPort int) int {
			for _, portMap := range config.Cluster.MapPorts {
				// Handle different types of containerPort values
				switch cp := portMap.ContainerPort.(type) {
				case int:
					if cp == containerPort {
						return portMap.HostPort
					}
				case float64:
					if int(cp) == containerPort {
						return portMap.HostPort
					}
				case string:
					// If it's a string, it might be a variable reference
					// Check if it directly evaluates to our container port
					if val, err := strconv.Atoi(cp); err == nil && val == containerPort {
						return portMap.HostPort
					}
					// If it's a variable reference like ${{ components.temporal.nodePorts.web }},
					// check for known patterns
					if strings.Contains(cp, "temporal.nodePorts.web") && containerPort == config.Components.Temporal.NodePorts.Web {
						return portMap.HostPort
					}
					if strings.Contains(cp, "temporal.nodePorts.frontend") && containerPort == config.Components.Temporal.NodePorts.Frontend {
						return portMap.HostPort
					}
					if strings.Contains(cp, "redis.nodePorts.redis") && containerPort == config.Components.Redis.NodePorts.Redis {
						return portMap.HostPort
					}
					if strings.Contains(cp, "mysql.nodePorts.mysql") && containerPort == config.Components.MySQL.NodePorts.MySQL {
						return portMap.HostPort
					}
				}
			}
			return 0
		}

		// Find host ports for the services
		if config.Components.Temporal.Enabled {
			temporalWebPort = findHostPort(config.Components.Temporal.NodePorts.Web)
			temporalFrontendPort = findHostPort(config.Components.Temporal.NodePorts.Frontend)

			// If ports are not found in the mappings, use defaults based on NodePorts
			if temporalWebPort == 0 {
				temporalWebPort = 8080 // Default host port for Temporal Web
				if verbose {
					fmt.Printf("%s Using default host port for Temporal Web: %d\n", yellow("ℹ️"), temporalWebPort)
				}
			}
			if temporalFrontendPort == 0 {
				temporalFrontendPort = 7233 // Default host port for Temporal Frontend
				if verbose {
					fmt.Printf("%s Using default host port for Temporal Frontend: %d\n", yellow("ℹ️"), temporalFrontendPort)
				}
			}
		}

		if config.Components.Redis.Enabled {
			redisPort = findHostPort(config.Components.Redis.NodePorts.Redis)

			// If port is not found in the mappings, use default
			if redisPort == 0 {
				redisPort = 6379 // Default host port for Redis
				if verbose {
					fmt.Printf("%s Using default host port for Redis: %d\n", yellow("ℹ️"), redisPort)
				}
			}
		}

		// Find OpenSearch host ports
		var openSearchPort, openSearchDashboardsPort int
		if config.Components.OpenSearch.Enabled {
			openSearchPort = findHostPort(config.Components.OpenSearch.NodePorts.Rest)

			// If port is not found in the mappings, use default
			if openSearchPort == 0 {
				openSearchPort = 9200 // Default host port for OpenSearch
				if verbose {
					fmt.Printf("%s Using default host port for OpenSearch: %d\n", yellow("ℹ️"), openSearchPort)
				}
			}
		}

		if config.Components.OpenSearchDashboards.Enabled {
			openSearchDashboardsPort = findHostPort(config.Components.OpenSearchDashboards.NodePorts.Http)

			// If port is not found in the mappings, use default
			if openSearchDashboardsPort == 0 {
				openSearchDashboardsPort = 5601 // Default host port for OpenSearch Dashboards
				if verbose {
					fmt.Printf("%s Using default host port for OpenSearch Dashboards: %d\n", yellow("ℹ️"), openSearchDashboardsPort)
				}
			}
		}

		// Display service access information
		fmt.Println("\nKind-based development environment setup complete!")
		if config.Components.Temporal.Enabled {
			if temporalWebPort > 0 {
				fmt.Printf("- Temporal Web UI: http://localhost:%d\n", temporalWebPort)
			}
			if temporalFrontendPort > 0 {
				fmt.Printf("- Temporal Frontend: localhost:%d\n", temporalFrontendPort)
			}
		}
		if config.Components.Redis.Enabled && redisPort > 0 {
			fmt.Printf("- Redis: localhost:%d\n", redisPort)
		}
		if config.Components.MySQL.Enabled {
			mysqlPort = findHostPort(config.Components.MySQL.NodePorts.MySQL)
			if mysqlPort == 0 {
				mysqlPort = 3306 // Default host port for MySQL
				if verbose {
					fmt.Printf("%s Using default host port for MySQL: %d\n", yellow("ℹ️"), mysqlPort)
				}
			}
			if mysqlPort > 0 {
				fmt.Printf("- MySQL: localhost:%d\n", mysqlPort)
				fmt.Printf("  Database: %s\n", config.Components.MySQL.Database)
				if config.Secrets.MySQL.Enabled {
					fmt.Printf("  Username: %s\n", config.Secrets.MySQL.Username)
				} else {
					fmt.Printf("  Username: root\n")
				}
			}
		}
		if config.Components.OpenSearch.Enabled && openSearchPort > 0 {
			fmt.Printf("- OpenSearch: http://localhost:%d\n", openSearchPort)
		}
		if config.Components.OpenSearchDashboards.Enabled && openSearchDashboardsPort > 0 {
			fmt.Printf("- OpenSearch Dashboards: http://localhost:%d\n", openSearchDashboardsPort)
		}
	},
}

func init() {
	kindenvCmd.AddCommand(kindenvStartCmd)

	// Add flags for kindenv start command
	kindenvStartCmd.Flags().Bool("skip-temporal", false, "Skip deploying Temporal")
	kindenvStartCmd.Flags().Bool("skip-dapr", false, "Skip deploying Dapr")
	kindenvStartCmd.Flags().Bool("skip-redis", false, "Skip deploying Redis")
	kindenvStartCmd.Flags().Bool("skip-opensearch", false, "Skip deploying OpenSearch")
	kindenvStartCmd.Flags().Bool("skip-opensearch-dashboards", false, "Skip deploying OpenSearch Dashboards")
	kindenvStartCmd.Flags().Bool("skip-opensearch-index-management", false, "Skip enabling OpenSearch Index Management plugin")
	kindenvStartCmd.Flags().Bool("skip-temporal-worker-operator", false, "Skip deploying Temporal Worker Operator")
	kindenvStartCmd.Flags().Bool("skip-indices-operator", false, "Skip deploying Indices Operator")
	kindenvStartCmd.Flags().Bool("skip-metrics-server", false, "Skip deploying Metrics Server")
	kindenvStartCmd.Flags().Bool("force-context", false, "Automatically switch to the correct Kubernetes context without prompting")

	// Deprecated flag, kept for backward compatibility
	kindenvStartCmd.Flags().String("operator-namespace", "default",
		"Deprecated: Namespace for Temporal worker operator (now always uses default namespace)")
	kindenvStartCmd.Flags().MarkDeprecated("operator-namespace", "The Temporal Worker Operator is now always installed in the default namespace")

	kindenvStartCmd.Flags().Bool("use-aws-ecr", false, "Use AWS ECR for pulling images")
	kindenvStartCmd.Flags().String("aws-profile", "", "AWS profile to use for ECR access")
	kindenvStartCmd.Flags().String("name", "", "Cluster name (defaults to current directory name)")
	kindenvStartCmd.Flags().StringP("config", "f", "", "Path to configuration file")
	kindenvStartCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
}
