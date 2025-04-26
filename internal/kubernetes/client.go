package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client defines the interface for Kubernetes operations
type Client interface {
	// Namespace operations
	CreateNamespace(name string) error
	NamespaceExists(name string) (bool, error)

	// Secret operations
	CreateSecret(namespace string, secret *corev1.Secret) error
	UpdateSecret(namespace string, secret *corev1.Secret) error

	// ServiceAccount operations
	PatchServiceAccount(namespace, name string, patchData []byte) error

	// General resource operations
	ApplyManifest(manifest []byte) error

	// Debug operations
	PrintClusterInfo() error
}

type clientImpl struct {
	clientset     kubernetes.Interface
	dynamicClient dynamic.Interface
	config        *rest.Config
}

// NewClient creates a new Kubernetes client
func NewClient(kubeconfig string) (Client, error) {
	config, err := getConfig(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get Kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return &clientImpl{
		clientset:     clientset,
		dynamicClient: dynamicClient,
		config:        config,
	}, nil
}

// NewClientForCluster creates a new Kubernetes client for a specific Kind cluster
func NewClientForCluster(clusterName string) (Client, error) {
	// Get kubeconfig path - usually in ~/.kube/kind-config-<cluster-name>
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	kubeConfigPath := filepath.Join(home, ".kube", fmt.Sprintf("kind-config-%s", clusterName))

	// Check if custom kubeconfig exists, otherwise use default
	if _, err := os.Stat(kubeConfigPath); os.IsNotExist(err) {
		kubeConfigPath = "" // Use default from KUBECONFIG env var or ~/.kube/config
	}

	return NewClient(kubeConfigPath)
}

// getConfig returns a Kubernetes client config
func getConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig == "" {
		// Use in-cluster config if running inside Kubernetes
		config, err := rest.InClusterConfig()
		if err == nil {
			return config, nil
		}

		// Otherwise, try the default locations
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, nil).ClientConfig()
	}

	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

// CreateNamespace creates a Kubernetes namespace if it doesn't exist
func (c *clientImpl) CreateNamespace(name string) error {
	exists, err := c.NamespaceExists(name)
	if err != nil {
		return err
	}

	if exists {
		return nil // Namespace already exists
	}

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	_, err = c.clientset.CoreV1().Namespaces().Create(context.Background(), namespace, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create namespace '%s': %w", name, err)
	}

	return nil
}

// NamespaceExists checks if a namespace exists
func (c *clientImpl) NamespaceExists(name string) (bool, error) {
	_, err := c.clientset.CoreV1().Namespaces().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if namespace '%s' exists: %w", name, err)
	}

	return true, nil
}

// CreateSecret creates a new secret in the specified namespace
func (c *clientImpl) CreateSecret(namespace string, secret *corev1.Secret) error {
	_, err := c.clientset.CoreV1().Secrets(namespace).Create(
		context.Background(),
		secret,
		metav1.CreateOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to create secret '%s' in namespace '%s': %w", secret.Name, namespace, err)
	}

	return nil
}

// UpdateSecret updates an existing secret in the specified namespace
func (c *clientImpl) UpdateSecret(namespace string, secret *corev1.Secret) error {
	_, err := c.clientset.CoreV1().Secrets(namespace).Update(
		context.Background(),
		secret,
		metav1.UpdateOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to update secret '%s' in namespace '%s': %w", secret.Name, namespace, err)
	}

	return nil
}

// PatchServiceAccount patches a service account in the specified namespace
func (c *clientImpl) PatchServiceAccount(namespace, name string, patchData []byte) error {
	_, err := c.clientset.CoreV1().ServiceAccounts(namespace).Patch(
		context.Background(),
		name,
		types.StrategicMergePatchType,
		patchData,
		metav1.PatchOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to patch service account '%s' in namespace '%s': %w", name, namespace, err)
	}

	return nil
}

// ApplyManifest applies a Kubernetes manifest
func (c *clientImpl) ApplyManifest(manifest []byte) error {
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(manifest), 4096)

	for {
		var rawObj runtime.RawExtension
		if err := decoder.Decode(&rawObj); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode manifest: %w", err)
		}

		if len(rawObj.Raw) == 0 {
			continue
		}

		obj := &unstructured.Unstructured{}
		if _, _, err := unstructured.UnstructuredJSONScheme.Decode(rawObj.Raw, nil, obj); err != nil {
			return fmt.Errorf("failed to convert manifest: %w", err)
		}

		gvk := obj.GetObjectKind().GroupVersionKind()
		gvr := schema.GroupVersionResource{
			Group:    gvk.Group,
			Version:  gvk.Version,
			Resource: fmt.Sprintf("%ss", strings.ToLower(gvk.Kind)), // Simple pluralization, might need more logic for complex cases
		}

		namespace := obj.GetNamespace()
		if namespace == "" {
			namespace = "default"
		}

		_, err := c.dynamicClient.Resource(gvr).Namespace(namespace).Create(
			context.Background(),
			obj,
			metav1.CreateOptions{},
		)
		if err != nil {
			return fmt.Errorf("failed to apply manifest: %w", err)
		}
	}

	return nil
}

// PrintClusterInfo prints information about the cluster for debugging
func (c *clientImpl) PrintClusterInfo() error {
	namespaces, err := c.clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list namespaces: %w", err)
	}

	fmt.Println("--- Cluster Namespaces ---")
	for _, ns := range namespaces.Items {
		fmt.Printf("Namespace: %s\n", ns.Name)
	}

	return nil
}
