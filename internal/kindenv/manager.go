package kindenv

import (
	"encoding/json"
	"fmt"
	"time"

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
func (m *Manager) StartCluster() error {
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
	} else {
		if !m.config.Cluster.CreateIfNotExists {
			tracker.Fail(fmt.Sprintf("Cluster %s does not exist and createIfNotExists is false", m.config.Cluster.Name))
			return fmt.Errorf("cluster %s does not exist and createIfNotExists is false", m.config.Cluster.Name)
		}

		// Create cluster configuration
		tracker.Step("Creating cluster configuration")
		clusterConfig := createKindClusterConfig(m.config)

		// Create cluster
		tracker.Step("Creating Kind cluster")
		if err := m.kindClient.CreateCluster(m.config.Cluster.Name, []byte(clusterConfig)); err != nil {
			tracker.Fail(fmt.Sprintf("Failed to create cluster: %v", err))
			return err
		}
	}

	// Initialize Kubernetes client
	tracker.Step("Initializing Kubernetes client")
	k8sClient, err := kubernetes.NewClientForCluster(m.config.Cluster.Name)
	if err != nil {
		tracker.Fail(fmt.Sprintf("Failed to initialize Kubernetes client: %v", err))
		return err
	}
	m.k8sClient = k8sClient

	// Install components
	if err := m.installComponents(tracker); err != nil {
		return err
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
func (m *Manager) installComponents(tracker progress.Tracker) error {
	// Install components based on configuration

	// Setup ECR credentials if needed
	if m.config.Images.UseAwsEcr {
		tracker.Step("Setting up AWS ECR credentials")
		if err := m.setupAWSECRCredentials("default"); err != nil {
			tracker.Fail(fmt.Sprintf("Failed to setup AWS ECR credentials: %v", err))
			return err
		}
	}

	// Install components
	if m.config.Components.Temporal.Enabled {
		tracker.Step("Installing Temporal")
		if err := m.installTemporal(); err != nil {
			tracker.Fail(fmt.Sprintf("Failed to install Temporal: %v", err))
			return err
		}
	}

	if m.config.Components.Redis.Enabled {
		tracker.Step("Installing Redis")
		if err := m.installRedis(); err != nil {
			tracker.Fail(fmt.Sprintf("Failed to install Redis: %v", err))
			return err
		}
	}

	if m.config.Components.CertManager.Enabled {
		tracker.Step("Installing cert-manager")
		if err := m.installCertManager(); err != nil {
			tracker.Fail(fmt.Sprintf("Failed to install cert-manager: %v", err))
			return err
		}
	}

	if m.config.Components.Dapr.Enabled {
		tracker.Step("Installing Dapr")
		if err := m.installDapr(); err != nil {
			tracker.Fail(fmt.Sprintf("Failed to install Dapr: %v", err))
			return err
		}
	}

	// Setup MySQL secret if needed
	if m.config.Secrets.MySQL.Enabled {
		tracker.Step("Creating MySQL secret")
		if err := m.createMySQLSecret(); err != nil {
			tracker.Fail(fmt.Sprintf("Failed to create MySQL secret: %v", err))
			return err
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

	// Add Helm repository
	if err := m.helmClient.AddRepository("temporal", "https://temporal.github.io/helm-charts"); err != nil {
		return err
	}

	// Update Helm repositories
	if err := m.helmClient.UpdateRepositories(); err != nil {
		return err
	}

	// Prepare values for Temporal helm chart
	values := map[string]interface{}{
		"server": map[string]interface{}{
			"replicaCount": 1,
		},
		"web": map[string]interface{}{
			"replicaCount": 1,
			"service": map[string]interface{}{
				"type":     "NodePort",
				"nodePort": m.config.Components.Temporal.WebNodePort,
			},
		},
		"cassandra": map[string]interface{}{
			"enabled": true,
			"config": map[string]interface{}{
				"cluster_size": 1,
			},
		},
		"prometheus": map[string]interface{}{
			"enabled": false,
		},
		"elasticsearch": map[string]interface{}{
			"enabled": false,
		},
		"grafana": map[string]interface{}{
			"enabled": false,
		},
	}

	// Convert values to YAML
	valuesBytes, err := json.Marshal(values)
	if err != nil {
		return fmt.Errorf("failed to marshal Temporal values: %w", err)
	}

	// Install Temporal
	timeout := 5 * time.Minute
	if err := m.helmClient.Install("temporal", "temporal/temporal", namespace, string(valuesBytes), timeout); err != nil {
		return err
	}

	return nil
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

	// Add Helm repository
	if err := m.helmClient.AddRepository("bitnami", "https://charts.bitnami.com/bitnami"); err != nil {
		return err
	}

	// Update Helm repositories
	if err := m.helmClient.UpdateRepositories(); err != nil {
		return err
	}

	// Prepare values for Redis helm chart
	values := map[string]interface{}{
		"master": map[string]interface{}{
			"persistence": map[string]interface{}{
				"enabled": false,
			},
			"service": map[string]interface{}{
				"type": "NodePort",
				"nodePorts": map[string]interface{}{
					"redis": m.config.Components.Redis.NodePort,
				},
			},
		},
		"replica": map[string]interface{}{
			"replicaCount": 0,
		},
		"auth": map[string]interface{}{
			"enabled": m.config.Components.Redis.Auth.Enabled,
		},
	}

	// Convert values to YAML
	valuesBytes, err := json.Marshal(values)
	if err != nil {
		return fmt.Errorf("failed to marshal Redis values: %w", err)
	}

	// Install Redis
	timeout := 5 * time.Minute
	if err := m.helmClient.Install("redis", "bitnami/redis", namespace, string(valuesBytes), timeout); err != nil {
		return err
	}

	return nil
}

// installCertManager installs cert-manager in the cluster
func (m *Manager) installCertManager() error {
	// Set default version if not specified
	version := m.config.Components.CertManager.Version
	if version == "" {
		version = "v1.11.0"
	}

	// Create namespace if it doesn't exist
	namespace := "cert-manager"
	if err := m.k8sClient.CreateNamespace(namespace); err != nil {
		return err
	}

	// Set up ECR credentials in the namespace if needed
	if m.config.Images.UseAwsEcr {
		if err := m.setupAWSECRCredentials(namespace); err != nil {
			return err
		}
	}

	// Install cert-manager using manifest
	// Note: Ideally we would use the Helm chart, but this matches the original implementation
	manifestURL := fmt.Sprintf("https://github.com/cert-manager/cert-manager/releases/download/%s/cert-manager.yaml", version)

	// Implement fetching and applying the manifest
	// This would require HTTP fetching which we'd need to implement
	// For now, just return an error indicating a limitation
	return fmt.Errorf("direct manifest application not implemented for cert-manager; use Helm chart instead")
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

	// Add Helm repository
	if err := m.helmClient.AddRepository("dapr", "https://dapr.github.io/helm-charts"); err != nil {
		return err
	}

	// Update Helm repositories
	if err := m.helmClient.UpdateRepositories(); err != nil {
		return err
	}

	// Prepare values for Dapr helm chart
	values := map[string]interface{}{
		"global": map[string]interface{}{
			"ha": map[string]interface{}{
				"enabled": m.config.Components.Dapr.Ha.Enabled,
			},
			"mtls": map[string]interface{}{
				"enabled": m.config.Components.Dapr.Mtls.Enabled,
			},
			"logLevel": m.config.Components.Dapr.LogLevel,
		},
	}

	// Convert values to YAML
	valuesBytes, err := json.Marshal(values)
	if err != nil {
		return fmt.Errorf("failed to marshal Dapr values: %w", err)
	}

	// Install Dapr
	timeout := 5 * time.Minute
	if err := m.helmClient.Install("dapr", "dapr/dapr", namespace, string(valuesBytes), timeout); err != nil {
		return err
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
