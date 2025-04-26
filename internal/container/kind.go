package container

import (
	"fmt"
	"os"
	"time"

	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cmd"
)

// KindClient defines the interface for Kind operations
type KindClient interface {
	// Cluster operations
	ClusterExists(name string) (bool, error)
	CreateCluster(name string, configContent []byte) error
	DeleteCluster(name string) error
	GetClusters() ([]string, error)
}

type kindClientImpl struct {
	provider *cluster.Provider
}

// NewKindClient creates a new Kind client
func NewKindClient() (KindClient, error) {
	provider := cluster.NewProvider(
		cluster.ProviderWithLogger(cmd.NewLogger()),
	)
	return &kindClientImpl{
		provider: provider,
	}, nil
}

// ClusterExists checks if a Kind cluster exists
func (k *kindClientImpl) ClusterExists(name string) (bool, error) {
	clusters, err := k.GetClusters()
	if err != nil {
		return false, err
	}

	for _, cluster := range clusters {
		if cluster == name {
			return true, nil
		}
	}

	return false, nil
}

// CreateCluster creates a new Kind cluster with the given configuration
func (k *kindClientImpl) CreateCluster(name string, configContent []byte) error {
	exists, err := k.ClusterExists(name)
	if err != nil {
		return fmt.Errorf("failed to check if cluster exists: %w", err)
	}

	if exists {
		return nil // Cluster already exists
	}

	// Write config to a temporary file
	tmpfile, err := os.CreateTemp("", "kind-config-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temporary file for Kind config: %w", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(configContent); err != nil {
		return fmt.Errorf("failed to write Kind config to temp file: %w", err)
	}

	if err := tmpfile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Create the cluster
	err = k.provider.Create(
		name,
		cluster.CreateWithConfigFile(tmpfile.Name()),
		cluster.CreateWithWaitForReady(time.Minute*5),
		cluster.CreateWithDisplayUsage(true),
		cluster.CreateWithDisplaySalutation(true),
	)
	if err != nil {
		return fmt.Errorf("failed to create Kind cluster: %w", err)
	}

	return nil
}

// DeleteCluster deletes a Kind cluster
func (k *kindClientImpl) DeleteCluster(name string) error {
	exists, err := k.ClusterExists(name)
	if err != nil {
		return fmt.Errorf("failed to check if cluster exists: %w", err)
	}

	if !exists {
		return nil // Cluster doesn't exist
	}

	if err := k.provider.Delete(name, ""); err != nil {
		return fmt.Errorf("failed to delete Kind cluster: %w", err)
	}

	return nil
}

// GetClusters returns a list of all Kind clusters
func (k *kindClientImpl) GetClusters() ([]string, error) {
	clusters, err := k.provider.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list Kind clusters: %w", err)
	}

	return clusters, nil
}
