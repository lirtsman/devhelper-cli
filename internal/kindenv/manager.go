package kindenv

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ShieldFC-RD/devhelper-cli/internal/aws"
	"github.com/ShieldFC-RD/devhelper-cli/internal/container"
	"github.com/ShieldFC-RD/devhelper-cli/internal/helm"
	"github.com/ShieldFC-RD/devhelper-cli/internal/kubernetes"
	"github.com/ShieldFC-RD/devhelper-cli/internal/progress"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Manager handles Kind environment operations
type Manager struct {
	config     *KindEnvConfig
	kindClient container.KindClient
	k8sClient  kubernetes.Client
	helmClient helm.Client
	awsClient  aws.ECRClient
	verbose    bool
}

// NewManager creates a new Kind environment manager
func NewManager(config *KindEnvConfig, verbose bool) (*Manager, error) {
	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create Kind client
	kindClient, err := container.NewKindClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Kind client: %w", err)
	}

	// Initialize ECR client if needed
	var awsClient aws.ECRClient
	if config.Images.UseAwsEcr {
		awsClient, err = aws.NewECRClient(config.Images.AWS.Region)
		if err != nil {
			return nil, fmt.Errorf("failed to create AWS ECR client: %w", err)
		}
	}

	// Kubernetes client will be initialized later when the cluster is ready

	// Create Helm client
	helmClient, err := helm.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Helm client: %w", err)
	}

	return &Manager{
		config:     config,
		kindClient: kindClient,
		helmClient: helmClient,
		awsClient:  awsClient,
		verbose:    verbose,
	}, nil
}

// StartCluster creates and sets up a Kind cluster
func (m *Manager) StartCluster(sequential bool) error {
	tracker := progress.NewTracker("Starting Kind cluster")
	tracker.Start()
	defer tracker.Done()

	// Check if cluster exists
	tracker.Step("Checking if cluster exists")
	exists, err := m.kindClient.ClusterExists(m.config.Cluster.Name)
	if err != nil {
		tracker.Fail(fmt.Sprintf("Failed to check if cluster exists: %v", err))
		return err
	}

	if exists {
		tracker.Step("Cluster already exists")
		if m.verbose {
			fmt.Println("[DEBUG] Found existing Kind cluster:", m.config.Cluster.Name)
		}
	} else {
		if !m.config.Cluster.CreateIfNotExists {
			tracker.Fail(fmt.Sprintf("Cluster %s does not exist and createIfNotExists is false", m.config.Cluster.Name))
			return fmt.Errorf("cluster %s does not exist and createIfNotExists is false", m.config.Cluster.Name)
		}

		// Create cluster configuration
		tracker.Step("Creating cluster configuration")
		if m.verbose {
			fmt.Println("[DEBUG] Creating Kind cluster configuration")
		}

		clusterConfig := createKindClusterConfig(m.config)

		if m.verbose {
			fmt.Println("[DEBUG] Kind cluster configuration:", clusterConfig)
		}

		// Create cluster
		tracker.Step("Creating Kind cluster")
		if m.verbose {
			fmt.Println("[DEBUG] Creating Kind cluster:", m.config.Cluster.Name)
		}

		if err := m.kindClient.CreateCluster(m.config.Cluster.Name, []byte(clusterConfig)); err != nil {
			tracker.Fail(fmt.Sprintf("Failed to create cluster: %v", err))
			return err
		}
		if m.verbose {
			fmt.Println("[DEBUG] Kind cluster created successfully")
		}
	}

	// Initialize Kubernetes client
	tracker.Step("Initializing Kubernetes client")
	if m.verbose {
		fmt.Println("[DEBUG] Initializing Kubernetes client for cluster:", m.config.Cluster.Name)
	}

	k8sClient, err := kubernetes.NewClientForCluster(m.config.Cluster.Name)
	if err != nil {
		tracker.Fail(fmt.Sprintf("Failed to initialize Kubernetes client: %v", err))
		return err
	}
	m.k8sClient = k8sClient

	if m.verbose {
		fmt.Println("[DEBUG] Kubernetes client initialized")
		fmt.Println("[DEBUG] Checking cluster status")
		err := m.k8sClient.PrintClusterInfo()
		if err != nil {
			fmt.Println("[DEBUG] Failed to print cluster info:", err)
		}
	}

	// Install components
	tracker.Step("Installing components")
	if m.verbose {
		fmt.Println("[DEBUG] Starting component installation")
	}

	if err := m.installComponents(tracker, sequential); err != nil {
		if m.verbose {
			fmt.Println("[DEBUG] Failed to install components:", err)
		}
		return err
	}

	if m.verbose {
		fmt.Println("[DEBUG] Component installation completed")
	}

	tracker.Success("Kind cluster started successfully")
	return nil
}

// createKindClusterConfig creates a Kind cluster configuration
func createKindClusterConfig(config *KindEnvConfig) string {
	clusterConfig := "kind: Cluster\napiVersion: kind.x-k8s.io/v1alpha4\nnodes:\n- role: control-plane\n  kubeadmConfigPatches:\n  - |\n    kind: InitConfiguration\n    nodeRegistration:\n      kubeletExtraArgs:\n        node-labels: \"ingress-ready=true\"\n  extraPortMappings:\n"

	// Add port mappings
	for _, portMap := range config.Cluster.MapPorts {
		clusterConfig += fmt.Sprintf("  - containerPort: %d\n    hostPort: %d\n    protocol: %s\n",
			portMap.ContainerPort, portMap.HostPort, portMap.Protocol)
	}

	// Add temporal and redis port mappings
	if config.Components.Temporal.Enabled {
		clusterConfig += fmt.Sprintf("  - containerPort: %d\n    hostPort: %d\n    protocol: TCP\n",
			config.Components.Temporal.FrontendNodePort, config.Components.Temporal.FrontendPort)

		clusterConfig += fmt.Sprintf("  - containerPort: %d\n    hostPort: %d\n    protocol: TCP\n",
			config.Components.Temporal.WebNodePort, config.Components.Temporal.WebPort)
	}

	if config.Components.Redis.Enabled {
		clusterConfig += fmt.Sprintf("  - containerPort: %d\n    hostPort: %d\n    protocol: TCP\n",
			config.Components.Redis.NodePort, config.Components.Redis.Port)
	}

	return clusterConfig
}

// installComponents installs components in the cluster
func (m *Manager) installComponents(tracker progress.Tracker, sequential bool) error {
	// Print verbose info about enabled components
	if m.verbose {
		fmt.Println("[DEBUG] Beginning component installation with these settings:")
		fmt.Printf("[DEBUG] - AWS ECR: %v\n", m.config.Images.UseAwsEcr)
		fmt.Printf("[DEBUG] - Temporal: %v\n", m.config.Components.Temporal.Enabled)
		fmt.Printf("[DEBUG] - Redis: %v\n", m.config.Components.Redis.Enabled)
		fmt.Printf("[DEBUG] - Dapr: %v\n", m.config.Components.Dapr.Enabled)
		fmt.Printf("[DEBUG] - Sequential installation: %v\n", sequential)
	}

	// Setup ECR credentials if needed
	if m.config.Images.UseAwsEcr {
		tracker.Step("Setting up AWS ECR credentials")
		if m.verbose {
			fmt.Println("[DEBUG] Setting up AWS ECR credentials in default namespace")
		}
		if err := m.setupAWSECRCredentials("default"); err != nil {
			tracker.Fail(fmt.Sprintf("Failed to setup AWS ECR credentials: %v", err))
			return err
		}
	}

	// If we're in sequential mode, install one component at a time
	if sequential {
		// Install Redis first as it's the simplest
		if m.config.Components.Redis.Enabled {
			tracker.Step("Installing Redis")
			if m.verbose {
				fmt.Println("[DEBUG] Starting Redis installation")
			}
			if err := m.installRedis(); err != nil {
				tracker.Fail(fmt.Sprintf("Failed to install Redis: %v", err))
				return err
			}
			if m.verbose {
				fmt.Println("[DEBUG] Redis installation completed")
			}
		}

		// Install Dapr
		if m.config.Components.Dapr.Enabled {
			tracker.Step("Installing Dapr")
			if m.verbose {
				fmt.Println("[DEBUG] Starting Dapr installation")
			}
			if err := m.installDapr(); err != nil {
				tracker.Fail(fmt.Sprintf("Failed to install Dapr: %v", err))
				return err
			}
			if m.verbose {
				fmt.Println("[DEBUG] Dapr installation completed")
			}
		}

		// Install Temporal last as it's the most complex
		if m.config.Components.Temporal.Enabled {
			tracker.Step("Installing Temporal")
			if m.verbose {
				fmt.Println("[DEBUG] Starting Temporal installation")
			}
			if err := m.installTemporal(); err != nil {
				tracker.Fail(fmt.Sprintf("Failed to install Temporal: %v", err))
				return err
			}
			if m.verbose {
				fmt.Println("[DEBUG] Temporal installation completed")
			}
		}

		// Setup MySQL secret if needed
		if m.config.Secrets.MySQL.Enabled {
			tracker.Step("Creating MySQL secret")
			if m.verbose {
				fmt.Println("[DEBUG] Creating MySQL secret")
			}
			if err := m.createMySQLSecret(); err != nil {
				tracker.Fail(fmt.Sprintf("Failed to create MySQL secret: %v", err))
				return err
			}
			if m.verbose {
				fmt.Println("[DEBUG] MySQL secret created")
			}
		}

		return nil
	}

	// Otherwise, install components in the original order
	// Install components
	if m.config.Components.Temporal.Enabled {
		tracker.Step("Installing Temporal")
		if m.verbose {
			fmt.Println("[DEBUG] Starting Temporal installation")
		}
		if err := m.installTemporal(); err != nil {
			tracker.Fail(fmt.Sprintf("Failed to install Temporal: %v", err))
			return err
		}
		if m.verbose {
			fmt.Println("[DEBUG] Temporal installation completed")
		}
	}

	if m.config.Components.Redis.Enabled {
		tracker.Step("Installing Redis")
		if m.verbose {
			fmt.Println("[DEBUG] Starting Redis installation")
		}
		if err := m.installRedis(); err != nil {
			tracker.Fail(fmt.Sprintf("Failed to install Redis: %v", err))
			return err
		}
		if m.verbose {
			fmt.Println("[DEBUG] Redis installation completed")
		}
	}

	if m.config.Components.Dapr.Enabled {
		tracker.Step("Installing Dapr")
		if m.verbose {
			fmt.Println("[DEBUG] Starting Dapr installation")
		}
		if err := m.installDapr(); err != nil {
			tracker.Fail(fmt.Sprintf("Failed to install Dapr: %v", err))
			return err
		}
		if m.verbose {
			fmt.Println("[DEBUG] Dapr installation completed")
		}
	}

	// Setup MySQL secret if needed
	if m.config.Secrets.MySQL.Enabled {
		tracker.Step("Creating MySQL secret")
		if m.verbose {
			fmt.Println("[DEBUG] Creating MySQL secret")
		}
		if err := m.createMySQLSecret(); err != nil {
			tracker.Fail(fmt.Sprintf("Failed to create MySQL secret: %v", err))
			return err
		}
		if m.verbose {
			fmt.Println("[DEBUG] MySQL secret created")
		}
	}

	return nil
}

// setupAWSECRCredentials sets up AWS ECR credentials in a namespace
func (m *Manager) setupAWSECRCredentials(namespace string) error {
	// Create namespace if it doesn't exist
	if err := m.k8sClient.CreateNamespace(namespace); err != nil {
		return err
	}

	// Get ECR authorization token
	registry, username, password, err := m.awsClient.GetAuthorizationToken()
	if err != nil {
		return err
	}

	// Create Docker config JSON
	dockerConfig, err := m.awsClient.CreateDockerConfig(registry, username, password)
	if err != nil {
		return err
	}

	// Create secret for Docker registry credentials
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ecr-docker-registry",
			Namespace: namespace,
		},
		Type: corev1.SecretTypeDockerConfigJson,
		StringData: map[string]string{
			".dockerconfigjson": dockerConfig,
		},
	}

	// Apply the secret
	if err := m.k8sClient.CreateSecret(namespace, secret); err != nil {
		return err
	}

	// Add the secret to default service account
	if m.config.Images.AWS.ServiceAccount == "" {
		// Use default service account
		patchData := []byte(fmt.Sprintf(`{"imagePullSecrets":[{"name":"ecr-docker-registry"}]}`))
		if err := m.k8sClient.PatchServiceAccount(namespace, "default", patchData); err != nil {
			return err
		}
	} else {
		// Use specified service account
		patchData := []byte(fmt.Sprintf(`{"imagePullSecrets":[{"name":"ecr-docker-registry"}]}`))
		if err := m.k8sClient.PatchServiceAccount(namespace, m.config.Images.AWS.ServiceAccount, patchData); err != nil {
			return err
		}
	}

	return nil
}

// installTemporal installs Temporal in the cluster
func (m *Manager) installTemporal() error {
	// Create namespace if it doesn't exist
	namespace := m.config.Components.Temporal.Namespace
	if namespace == "" {
		namespace = "temporal"
	}

	if err := m.k8sClient.CreateNamespace(namespace); err != nil {
		return err
	}

	// Set up ECR credentials in the namespace if needed
	if m.config.Images.UseAwsEcr {
		if err := m.setupAWSECRCredentials(namespace); err != nil {
			return err
		}
	}

	// Note: Helm repository is added during initialization, so we don't need to add it here
	if m.verbose {
		fmt.Println("[DEBUG] Using existing Temporal Helm repository")
	}

	// First check if Temporal is already installed
	if m.verbose {
		fmt.Println("[DEBUG] Checking if Temporal is already installed")
	}
	checkCmd := []string{"helm", "list", "--namespace", namespace, "--filter", "temporal", "--output", "json"}
	checkOutput, err := executeCommandWithOutput(checkCmd...)
	if err != nil {
		if m.verbose {
			fmt.Println("[DEBUG] Error checking Temporal installation:", err)
		}
	} else if m.verbose {
		fmt.Println("[DEBUG] Helm list output:", checkOutput)
	}

	// Build the Helm install command with minimal configuration
	installArgs := []string{
		"helm", "upgrade", "--install",
		"--namespace", namespace,
		"--create-namespace",
		// Ultra minimal server configuration
		"--set", "server.replicaCount=1",
		// Minimal database settings
		"--set", "cassandra.config.cluster_size=1",
		"--set", "elasticsearch.replicas=1",
		// Disable optional components
		"--set", "prometheus.enabled=false",
		"--set", "grafana.enabled=false",
		// Configure services
		"--set", "server.frontend.service.type=NodePort",
		"--set", fmt.Sprintf("server.frontend.service.nodePort=%d", m.config.Components.Temporal.FrontendNodePort),
		"--set", "web.enabled=true",
		"--set", "web.service.type=NodePort",
		"--set", fmt.Sprintf("web.service.nodePort=%d", m.config.Components.Temporal.WebNodePort),
		// Reduce resource requirements significantly
		"--set", "cassandra.resources.requests.cpu=100m",
		"--set", "cassandra.resources.requests.memory=256Mi",
		"--set", "cassandra.resources.limits.cpu=200m",
		"--set", "cassandra.resources.limits.memory=512Mi",
		"--set", "elasticsearch.resources.requests.cpu=100m",
		"--set", "elasticsearch.resources.requests.memory=256Mi",
		"--set", "elasticsearch.resources.limits.cpu=200m",
		"--set", "elasticsearch.resources.limits.memory=512Mi",
		"--set", "server.resources.requests.cpu=100m",
		"--set", "server.resources.requests.memory=128Mi",
		"--set", "web.resources.requests.cpu=100m",
		"--set", "web.resources.requests.memory=128Mi",
		// Use single-instance mode for all components
		"--set", "server.frontend.replicaCount=1",
		"--set", "server.history.replicaCount=1",
		"--set", "server.matching.replicaCount=1",
		"--set", "server.worker.replicaCount=1",
		"--timeout", "15m",
		"--debug", // Add debug flag to see more details
	}

	// Add AWS ECR settings if needed
	if m.config.Images.UseAwsEcr {
		installArgs = append(installArgs,
			"--set", "global.imagePullSecrets[0].name=ecr-docker-registry",
		)
	}

	// Add chart reference and release name
	installArgs = append(installArgs, "temporal", "temporal/temporal")

	if m.verbose {
		fmt.Println("[DEBUG] Installing Temporal with command:", strings.Join(installArgs, " "))
	}

	// Check available resources in the cluster
	if m.verbose {
		fmt.Println("[DEBUG] Checking available resources in the cluster")
		nodeResourcesCmd := []string{"kubectl", "describe", "nodes"}
		if err := executeCommand(nodeResourcesCmd...); err != nil {
			fmt.Println("[DEBUG] Error checking node resources:", err)
		}
	}

	// Execute the command with output capturing for debugging
	if output, err := executeCommandWithOutput(installArgs...); err != nil {
		if m.verbose {
			fmt.Println("[DEBUG] Helm installation output:", output)
		}
		return fmt.Errorf("failed to install temporal: %w", err)
	} else if m.verbose {
		fmt.Println("[DEBUG] Helm installation output:", output)
	}

	return nil
}

// executeCommandWithOutput executes a command and returns its output
func executeCommandWithOutput(args ...string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// executeCommand executes a command with the given arguments
func executeCommand(args ...string) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// installRedis installs Redis in the cluster
func (m *Manager) installRedis() error {
	// Create namespace if it doesn't exist
	namespace := "redis"
	if err := m.k8sClient.CreateNamespace(namespace); err != nil {
		return err
	}

	// Set up ECR credentials in the namespace if needed
	if m.config.Images.UseAwsEcr {
		if err := m.setupAWSECRCredentials(namespace); err != nil {
			return err
		}
	}

	// Note: Helm repository is added during initialization, so we don't need to add it here
	if m.verbose {
		fmt.Println("[DEBUG] Using existing Bitnami Helm repository")
	}

	// Get chart version, use default if not specified
	chartVersion := m.config.Components.Redis.ChartVersion
	if chartVersion == "" {
		chartVersion = "17.3.7" // Default to this version if not specified
	}

	// Build the Helm install command
	installArgs := []string{
		"helm", "upgrade", "--install",
		"--namespace", namespace,
		"--create-namespace",
		"--version", chartVersion,
		"--set", "master.persistence.enabled=false",
		"--set", "master.service.type=NodePort",
		"--set", fmt.Sprintf("master.service.nodePorts.redis=%s", fmt.Sprintf("%d", m.config.Components.Redis.NodePort)),
		"--set", "replica.replicaCount=0",
		"--set", fmt.Sprintf("auth.enabled=%t", m.config.Components.Redis.Auth.Enabled),
		"--timeout", "5m",
	}

	// Add AWS ECR settings if needed
	if m.config.Images.UseAwsEcr {
		installArgs = append(installArgs,
			"--set", "global.imagePullSecrets[0].name=ecr-docker-registry",
		)
	}

	// Add chart reference and release name
	installArgs = append(installArgs, "redis", "bitnami/redis")

	if m.verbose {
		fmt.Println("[DEBUG] Installing Redis with command:", strings.Join(installArgs, " "))
	}

	if err := executeCommand(installArgs...); err != nil {
		return fmt.Errorf("failed to install redis: %w", err)
	}

	return nil
}

// installDapr installs Dapr in the cluster
func (m *Manager) installDapr() error {
	// Create namespace if it doesn't exist
	namespace := "dapr-system"
	if err := m.k8sClient.CreateNamespace(namespace); err != nil {
		return err
	}

	// Set up ECR credentials in the namespace if needed
	if m.config.Images.UseAwsEcr {
		if err := m.setupAWSECRCredentials(namespace); err != nil {
			return err
		}
	}

	// Note: Helm repository is added during initialization, so we don't need to add it here
	if m.verbose {
		fmt.Println("[DEBUG] Using existing Dapr Helm repository")
	}

	// Build the Helm install command
	installArgs := []string{
		"helm", "upgrade", "--install",
		"--namespace", namespace,
		"--create-namespace",
		"--set", fmt.Sprintf("global.logLevel=%s", m.config.Components.Dapr.LogLevel),
		"--set", fmt.Sprintf("global.ha.enabled=%t", m.config.Components.Dapr.Ha.Enabled),
		"--set", fmt.Sprintf("global.mtls.enabled=%t", m.config.Components.Dapr.Mtls.Enabled),
		// Minimal resource settings
		"--set", "dapr_operator.resources.requests.cpu=50m",
		"--set", "dapr_operator.resources.requests.memory=64Mi",
		"--set", "dapr_placement.resources.requests.cpu=50m",
		"--set", "dapr_placement.resources.requests.memory=64Mi",
		"--set", "dapr_sidecar_injector.resources.requests.cpu=50m",
		"--set", "dapr_sidecar_injector.resources.requests.memory=64Mi",
		"--set", "dapr_sentry.resources.requests.cpu=50m",
		"--set", "dapr_sentry.resources.requests.memory=64Mi",
		"--timeout", "5m",
	}

	// Add AWS ECR settings if needed
	if m.config.Images.UseAwsEcr {
		installArgs = append(installArgs,
			"--set", "global.imagePullSecrets[0].name=ecr-docker-registry",
		)
	}

	// Add chart reference and release name
	installArgs = append(installArgs, "dapr", "dapr/dapr")

	if m.verbose {
		fmt.Println("[DEBUG] Installing Dapr with command:", strings.Join(installArgs, " "))
	}

	if err := executeCommand(installArgs...); err != nil {
		return fmt.Errorf("failed to install dapr: %w", err)
	}

	return nil
}

// createMySQLSecret creates a MySQL secret in the cluster
func (m *Manager) createMySQLSecret() error {
	namespace := m.config.Secrets.MySQL.Namespace
	if err := m.k8sClient.CreateNamespace(namespace); err != nil {
		return err
	}

	// Create secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.config.Secrets.MySQL.Name,
			Namespace: namespace,
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"username": m.config.Secrets.MySQL.Username,
			"password": m.config.Secrets.MySQL.Password,
		},
	}

	// Apply the secret
	if err := m.k8sClient.CreateSecret(namespace, secret); err != nil {
		return err
	}

	return nil
}

// StopCluster stops and deletes the Kind cluster
func (m *Manager) StopCluster() error {
	tracker := progress.NewTracker("Stopping Kind cluster")
	tracker.Start()
	defer tracker.Done()

	// Check if cluster exists
	tracker.Step("Checking if cluster exists")
	exists, err := m.kindClient.ClusterExists(m.config.Cluster.Name)
	if err != nil {
		tracker.Fail(fmt.Sprintf("Failed to check if cluster exists: %v", err))
		return err
	}

	if !exists {
		tracker.Success("Cluster does not exist")
		return nil
	}

	// Delete cluster
	tracker.Step("Deleting Kind cluster")
	if err := m.kindClient.DeleteCluster(m.config.Cluster.Name); err != nil {
		tracker.Fail(fmt.Sprintf("Failed to delete cluster: %v", err))
		return err
	}

	tracker.Success("Kind cluster stopped successfully")
	return nil
}

// GetClusterStatus gets the status of the Kind cluster
func (m *Manager) GetClusterStatus() (bool, error) {
	// Check if cluster exists
	exists, err := m.kindClient.ClusterExists(m.config.Cluster.Name)
	if err != nil {
		return false, err
	}

	return exists, nil
}
