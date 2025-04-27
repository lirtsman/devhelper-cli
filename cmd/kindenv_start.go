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
	"strings"
	"time"

	"github.com/ShieldFC-RD/devhelper-cli/internal/kindenv"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Define the setupECRCreds function at the file level, outside of any function
// Add before the kindenvStartCmd definition

// setupECRCreds creates ECR pull secrets in a given namespace
func setupECRCreds(namespace, registry, password string) error {
	fmt.Printf("Creating ECR pull secret in the %s namespace\n", namespace)

	// Ensure namespace exists
	namespaceYaml, err := executeCommand("kubectl", "create", "namespace", namespace, "--dry-run=client", "-o", "yaml")
	if err != nil {
		return fmt.Errorf("failed to create namespace: %v", err)
	}

	// Apply the namespace using stdin
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(namespaceYaml)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply namespace: %v", err)
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
		return fmt.Errorf("failed to create secret: %v", err)
	}

	// Apply the secret using stdin
	cmd = exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(secretYaml)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply secret: %v", err)
	}

	// Patch default service account to use ECR credentials
	patchCmd := exec.Command("kubectl", "patch", "serviceaccount", "default",
		"-p", "{\"imagePullSecrets\": [{\"name\": \"ecr-credentials\"}]}",
		fmt.Sprintf("-n=%s", namespace))
	err = patchCmd.Run()
	if err != nil {
		return fmt.Errorf("failed to patch service account: %v", err)
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

It also sets up the Kubernetes context and creates required namespaces.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create colored output helpers
		green := color.New(color.FgGreen).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()
		red := color.New(color.FgRed).SprintFunc()

		// Get flags
		verbose, _ := cmd.Flags().GetBool("verbose")
		configPath, _ := cmd.Flags().GetString("config")
		useAwsEcr, _ := cmd.Flags().GetBool("use-aws-ecr")
		skipTemporal, _ := cmd.Flags().GetBool("skip-temporal")
		skipDapr, _ := cmd.Flags().GetBool("skip-dapr")
		skipRedis, _ := cmd.Flags().GetBool("skip-redis")

		// Load config file
		fmt.Println(green("Setting up Kind-based development environment..."))

		// Load config
		config, err := kindenv.LoadConfig(configPath)
		if err != nil {
			fmt.Printf("%s Error loading config: %v\n", red("❌"), err)
			os.Exit(1)
		}

		// Override config with command line args
		if useAwsEcr {
			config.Images.UseAwsEcr = true
		}

		// Set component enabled/disabled based on skip flags
		if skipTemporal {
			config.Components.Temporal.Enabled = false
		}

		if skipDapr {
			config.Components.Dapr.Enabled = false
		}

		if skipRedis {
			config.Components.Redis.Enabled = false
		}

		// Print component configuration
		fmt.Println(green("Component configuration:"))
		fmt.Printf("- Temporal: %v\n", config.Components.Temporal.Enabled)
		fmt.Printf("- Redis: %v\n", config.Components.Redis.Enabled)
		fmt.Printf("- Dapr: %v\n", config.Components.Dapr.Enabled)
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

		// Setup AWS ECR if needed
		var ecrPassword string
		var ecrRegistry string
		if config.Images.UseAwsEcr {
			fmt.Println(yellow("Setting up AWS ECR credentials"))

			if !commandExists("aws") {
				fmt.Printf("%s Error: AWS CLI is not installed. Please install it first.\n", red("❌"))
				os.Exit(1)
			}

			// Get AWS account ID and ECR credentials
			accountOutput, err := executeCommand("aws", "sts", "get-caller-identity", "--query", "Account", "--output", "text")
			if err != nil {
				fmt.Printf("%s Error getting AWS account ID: %v\n", red("❌"), err)
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

			// Get ecrPassword from AWS
			ecrPasswordOutput, err := executeCommand("aws", "ecr", "get-login-password", "--region", awsRegion)
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

			// Create a service account that uses the secret
			serviceAccountYaml := `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ecr-pull-service-account
  namespace: default
imagePullSecrets:
- name: ecr-credentials
`

			cmd := exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = strings.NewReader(serviceAccountYaml)
			err = cmd.Run()
			if err != nil {
				fmt.Printf("%s Error creating service account: %v\n", red("❌"), err)
				os.Exit(1)
			}

			// Wait for resources to be created
			time.Sleep(2 * time.Second)

			fmt.Printf("%s AWS ECR credentials configured\n", green("✅"))
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

			// Add port mappings from config
			for _, portMap := range config.Cluster.MapPorts {
				kindConfig += fmt.Sprintf("  - containerPort: %d\n    hostPort: %d\n    protocol: %s\n",
					portMap.ContainerPort, portMap.HostPort, portMap.Protocol)
			}

			// Add temporal and redis port mappings if enabled
			if config.Components.Temporal.Enabled {
				kindConfig += fmt.Sprintf("  - containerPort: %d\n    hostPort: %d\n    protocol: TCP\n",
					config.Components.Temporal.FrontendNodePort, config.Components.Temporal.FrontendPort)

				kindConfig += fmt.Sprintf("  - containerPort: %d\n    hostPort: %d\n    protocol: TCP\n",
					config.Components.Temporal.WebNodePort, config.Components.Temporal.WebPort)
			}

			if config.Components.Redis.Enabled {
				kindConfig += fmt.Sprintf("  - containerPort: %d\n    hostPort: %d\n    protocol: TCP\n",
					config.Components.Redis.NodePort, config.Components.Redis.Port)
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

			// Use selected container engine for kind operations if needed (e.g. KIND_EXPERIMENTAL_PROVIDER=podman)
			if containerEngine == "podman" && verbose {
				fmt.Println(yellow("Setting KIND_EXPERIMENTAL_PROVIDER=podman"))
				os.Setenv("KIND_EXPERIMENTAL_PROVIDER", "podman")
			}

			_, err = executeCommand("kind", "create", "cluster", "--name", config.Cluster.Name, "--config", kindConfigFile.Name())
			if err != nil {
				fmt.Printf("%s Error creating Kind cluster: %v\n", red("❌"), err)
				os.Exit(1)
			}

			fmt.Printf("%s Kind cluster created successfully\n", green("✅"))
		} else if !clusterExists && !config.Cluster.CreateIfNotExists {
			fmt.Printf("%s Error: Cluster %s does not exist and createIfNotExists is false\n", red("❌"), config.Cluster.Name)
			os.Exit(1)
		}

		// Switch kubectl context to the kind cluster
		_, err = executeCommand("kubectl", "cluster-info", "--context", fmt.Sprintf("kind-%s", config.Cluster.Name))
		if err != nil {
			fmt.Printf("%s Error switching kubectl context: %v\n", red("❌"), err)
			os.Exit(1)
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
  username: "%s"
  password: "%s"
`, config.Secrets.MySQL.Name, config.Secrets.MySQL.Namespace,
				config.Secrets.MySQL.Username, config.Secrets.MySQL.Password)

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
				"--set", fmt.Sprintf("master.service.nodePorts.redis=%d", config.Components.Redis.NodePort),
				"--set", fmt.Sprintf("auth.enabled=%t", config.Components.Redis.Auth.Enabled),
				"--set", "replica.replicaCount=0")
			if err != nil {
				fmt.Printf("%s Error installing Redis: %v\n", red("❌"), err)
				os.Exit(1)
			}

			// Wait for Redis to be ready using label selectors
			fmt.Println(yellow("Waiting for Redis to be ready..."))

			// Give resources a moment to be created
			fmt.Println(yellow("Pausing briefly to allow resources to be created..."))
			time.Sleep(5 * time.Second)

			// First, find out what resources are available to identify the correct name
			redisResourcesOutput, err := executeCommand("kubectl", "get", "all", "-n", "redis")
			if verbose {
				fmt.Println(yellow("Redis resources detected:"))
				fmt.Println(redisResourcesOutput)
			}

			// Try multiple approaches to wait for Redis, with retries
			var redisReady bool
			for attempt := 1; attempt <= 3 && !redisReady; attempt++ {
				if attempt > 1 {
					fmt.Printf("Retry attempt %d of 3 for finding Redis resources...\n", attempt)
					time.Sleep(5 * time.Second)
				}

				// Try to find pods directly with common Redis labels
				podOutput, err := executeCommand("kubectl", "get", "pods", "-n", "redis", "-l", "app.kubernetes.io/name=redis", "-o", "name")
				if err == nil && podOutput != "" {
					fmt.Println(yellow("Redis pods detected with app.kubernetes.io/name=redis label"))
					redisReady = true
					continue
				}

				podOutput, err = executeCommand("kubectl", "get", "pods", "-n", "redis", "-l", "app=redis", "-o", "name")
				if err == nil && podOutput != "" {
					fmt.Println(yellow("Redis pods detected with app=redis label"))
					redisReady = true
					continue
				}

				// Try specific Redis pod patterns
				podOutput, err = executeCommand("kubectl", "get", "pods", "-n", "redis", "--no-headers")
				if err == nil && podOutput != "" {
					// Just check if any pods exist in the namespace
					lines := strings.Split(strings.TrimSpace(podOutput), "\n")
					if len(lines) > 0 {
						fmt.Println(yellow("Found Redis pods:"))
						for _, line := range lines {
							fmt.Println(line)
						}
						redisReady = true
						continue
					}
				}

				// Try statefulset
				statefulsetOutput, err := executeCommand("kubectl", "get", "statefulset", "-n", "redis", "-o", "name")
				if err == nil && statefulsetOutput != "" {
					fmt.Println(yellow("Redis statefulset detected, waiting for rollout..."))
					redisReady = true
					continue
				}

				// Try deployments
				deploymentOutput, err := executeCommand("kubectl", "get", "deployment", "-n", "redis", "-o", "name")
				if err == nil && deploymentOutput != "" {
					fmt.Println(yellow("Redis deployment detected"))
					redisReady = true
					continue
				}
			}

			if redisReady {
				fmt.Printf("%s Redis resources detected\n", green("✅"))
			} else {
				fmt.Printf("%s Could not detect Redis resources, but continuing anyway\n", yellow("⚠️"))
			}

			// Now that we confirmed resources exist, wait for them to be ready
			fmt.Println(yellow("Waiting for Redis pods to become ready..."))
			_, err = executeCommand("kubectl", "wait", "--for=condition=Ready", "pods", "--all", "-n", "redis", "--timeout=2m")
			if err != nil {
				fmt.Printf("%s Error waiting for Redis pods: %v\n", yellow("⚠️"), err)
				fmt.Println(yellow("Continuing despite error waiting for Redis pods"))
			} else {
				fmt.Printf("%s Redis is ready\n", green("✅"))
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
				"--version", config.Components.Dapr.Version,
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
				"--set", "server.replicaCount=1",
				"--set", "cassandra.config.cluster_size=1",
				"--set", "elasticsearch.replicas=1",
				"--set", "prometheus.enabled=false",
				"--set", "grafana.enabled=false",
				"--set", "server.frontend.service.type=NodePort",
				"--set", fmt.Sprintf("server.frontend.service.nodePort=%d", config.Components.Temporal.FrontendNodePort),
				"--set", "web.enabled=true",
				"--set", "web.service.type=NodePort",
				"--set", fmt.Sprintf("web.service.nodePort=%d", config.Components.Temporal.WebNodePort),
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

		fmt.Println(green("Kind-based development environment setup complete!"))
		fmt.Println(green("Access services:"))
		if config.Components.Temporal.Enabled {
			fmt.Printf("- Temporal Web UI: http://localhost:%d\n", config.Components.Temporal.WebPort)
			fmt.Printf("- Temporal Frontend: localhost:%d\n", config.Components.Temporal.FrontendPort)
		}
		if config.Components.Redis.Enabled {
			fmt.Printf("- Redis: localhost:%d\n", config.Components.Redis.Port)
		}
	},
}

func init() {
	kindenvCmd.AddCommand(kindenvStartCmd)

	// Add flags for kindenv start command
	kindenvStartCmd.Flags().Bool("skip-temporal", false, "Skip installing Temporal")
	kindenvStartCmd.Flags().Bool("skip-dapr", false, "Skip installing Dapr")
	kindenvStartCmd.Flags().Bool("skip-redis", false, "Skip installing Redis")
	kindenvStartCmd.Flags().String("operator-namespace", "temporal-worker-operator-system",
		"Namespace for Temporal worker operator")
	kindenvStartCmd.Flags().Bool("use-aws-ecr", false, "Use AWS ECR for pulling images")
}
