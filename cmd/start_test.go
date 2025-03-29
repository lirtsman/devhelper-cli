package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConfigCacheFunctions tests the configuration cache handling functions
func TestConfigCacheFunctions(t *testing.T) {
	t.Run("hasConfigChanged should detect no changes", func(t *testing.T) {
		// Given the same configuration
		currentConfig := LocalEnvConfig{
			Dapr: struct {
				DashboardPort int `yaml:"dashboardPort"`
				ZipkinPort    int `yaml:"zipkinPort"`
			}{
				DashboardPort: 8080,
				ZipkinPort:    9411,
			},
			Temporal: struct {
				Namespace string `yaml:"namespace"`
				UIPort    int    `yaml:"uiPort"`
				GRPCPort  int    `yaml:"grpcPort"`
			}{
				Namespace: "default",
				UIPort:    8233,
				GRPCPort:  7233,
			},
		}

		currentCache := ConfigCache{
			DaprDashboardPort: 8080,
			TemporalUIPort:    8233,
			TemporalGRPCPort:  7233,
			TemporalNamespace: "default",
		}

		// When checking for changes
		hasChanges, newCache, changes := hasConfigChanged(currentConfig, currentCache)

		// Then no changes should be detected
		assert.False(t, hasChanges, "Should not detect changes when config matches cache")
		assert.Empty(t, changes, "Changes slice should be empty")
		assert.Equal(t, currentCache, newCache, "New cache should match current cache")
	})

	t.Run("hasConfigChanged should detect Dapr dashboard port change", func(t *testing.T) {
		// Given a configuration with changed Dapr dashboard port
		currentConfig := LocalEnvConfig{
			Dapr: struct {
				DashboardPort int `yaml:"dashboardPort"`
				ZipkinPort    int `yaml:"zipkinPort"`
			}{
				DashboardPort: 8081, // Changed port
				ZipkinPort:    9411,
			},
		}

		currentCache := ConfigCache{
			DaprDashboardPort: 8080, // Original port
			TemporalUIPort:    8233,
			TemporalGRPCPort:  7233,
			TemporalNamespace: "default",
		}

		// When checking for changes
		hasChanges, newCache, changes := hasConfigChanged(currentConfig, currentCache)

		// Then changes should be detected
		assert.True(t, hasChanges, "Should detect changes when Dapr dashboard port changed")
		assert.Contains(t, changes[0], "Dapr Dashboard port changed: 8080 → 8081")
		assert.Equal(t, 8081, newCache.DaprDashboardPort)
	})

	t.Run("hasConfigChanged should detect Temporal UI port change", func(t *testing.T) {
		// Given a configuration with changed Temporal UI port
		currentConfig := LocalEnvConfig{
			Dapr: struct {
				DashboardPort int `yaml:"dashboardPort"`
				ZipkinPort    int `yaml:"zipkinPort"`
			}{
				DashboardPort: 8080, // Match cache value to avoid detecting unrelated changes
				ZipkinPort:    9411,
			},
			Temporal: struct {
				Namespace string `yaml:"namespace"`
				UIPort    int    `yaml:"uiPort"`
				GRPCPort  int    `yaml:"grpcPort"`
			}{
				Namespace: "default",
				UIPort:    8234, // Changed port
				GRPCPort:  7233,
			},
		}

		currentCache := ConfigCache{
			DaprDashboardPort: 8080,
			TemporalUIPort:    8233, // Original port
			TemporalGRPCPort:  7233,
			TemporalNamespace: "default",
		}

		// When checking for changes
		hasChanges, newCache, changes := hasConfigChanged(currentConfig, currentCache)

		// Then changes should be detected
		assert.True(t, hasChanges, "Should detect changes when Temporal UI port changed")
		assert.Contains(t, changes[0], "Temporal UI port changed: 8233 → 8234")
		assert.Equal(t, 8234, newCache.TemporalUIPort)
	})

	t.Run("hasConfigChanged should detect Temporal namespace change", func(t *testing.T) {
		// Given a configuration with changed Temporal namespace
		currentConfig := LocalEnvConfig{
			Dapr: struct {
				DashboardPort int `yaml:"dashboardPort"`
				ZipkinPort    int `yaml:"zipkinPort"`
			}{
				DashboardPort: 8080, // Match cache value to avoid detecting unrelated changes
				ZipkinPort:    9411,
			},
			Temporal: struct {
				Namespace string `yaml:"namespace"`
				UIPort    int    `yaml:"uiPort"`
				GRPCPort  int    `yaml:"grpcPort"`
			}{
				Namespace: "testing", // Changed namespace
				UIPort:    8233,
				GRPCPort:  7233,
			},
		}

		currentCache := ConfigCache{
			DaprDashboardPort: 8080,
			TemporalUIPort:    8233,
			TemporalGRPCPort:  7233,
			TemporalNamespace: "default", // Original namespace
		}

		// When checking for changes
		hasChanges, newCache, changes := hasConfigChanged(currentConfig, currentCache)

		// Then changes should be detected
		assert.True(t, hasChanges, "Should detect changes when Temporal namespace changed")
		assert.Contains(t, changes[0], "Temporal namespace changed: default → testing")
		assert.Equal(t, "testing", newCache.TemporalNamespace)
	})

	t.Run("hasConfigChanged should handle empty cache", func(t *testing.T) {
		// Given a configuration and an empty cache (first run)
		currentConfig := LocalEnvConfig{
			Dapr: struct {
				DashboardPort int `yaml:"dashboardPort"`
				ZipkinPort    int `yaml:"zipkinPort"`
			}{
				DashboardPort: 8080,
				ZipkinPort:    9411,
			},
			Temporal: struct {
				Namespace string `yaml:"namespace"`
				UIPort    int    `yaml:"uiPort"`
				GRPCPort  int    `yaml:"grpcPort"`
			}{
				Namespace: "default",
				UIPort:    8233,
				GRPCPort:  7233,
			},
		}

		currentCache := ConfigCache{} // Empty cache

		// When checking for changes
		hasChanges, newCache, changes := hasConfigChanged(currentConfig, currentCache)

		// Then no changes should be detected (first run)
		assert.False(t, hasChanges, "Should not detect changes on first run with empty cache")
		assert.Empty(t, changes, "Changes slice should be empty on first run")
		assert.Equal(t, 8080, newCache.DaprDashboardPort)
		assert.Equal(t, 8233, newCache.TemporalUIPort)
		assert.Equal(t, 7233, newCache.TemporalGRPCPort)
		assert.Equal(t, "default", newCache.TemporalNamespace)
	})

	t.Run("loadConfigCache and saveConfigCache should work together", func(t *testing.T) {
		// Given a temporary configuration directory
		origConfigDir := filepath.Join(os.Getenv("HOME"), ".config", "devhelper-cli")
		tmpDir := t.TempDir()
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", os.Getenv("HOME")) // Restore original HOME

		// And a test cache
		testCache := ConfigCache{
			DaprDashboardPort: 8888,
			TemporalUIPort:    9999,
			TemporalGRPCPort:  7777,
			TemporalNamespace: "test-namespace",
		}

		// When saving and loading the cache
		saveConfigCache(testCache)

		// Then the loaded cache should match the saved cache
		loadedCache := loadConfigCache()
		assert.Equal(t, testCache, loadedCache, "Loaded cache should match saved cache")

		// Cleanup: restore original HOME
		os.MkdirAll(origConfigDir, 0755) // Ensure directory exists if it was created during test
	})
}
