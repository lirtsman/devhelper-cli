package kindenv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// volume_test.go provides tests for volume and volumeMount generation

// T075: Test for volume mount generation
func TestGenerateVolumeMounts_SingleFile(t *testing.T) {
	component := &CustomComponent{
		Name:      "test-app",
		Image:     "nginx:latest",
		Namespace: "default",
		ConfigFiles: []ConfigFile{
			{
				Name:     "app.yaml",
				Path:     "/config/app.yaml",
				Contents: "server:\n  port: 8080",
			},
		},
	}
	component.SetDefaults()

	volumeMounts, err := generateVolumeMounts(component)
	require.NoError(t, err)
	require.Len(t, volumeMounts, 1)

	vm := volumeMounts[0]
	assert.Equal(t, "test-app-config-volume", vm.Name)
	assert.Equal(t, "/config/app.yaml", vm.MountPath)
	assert.Equal(t, "app.yaml", vm.SubPath)
	assert.True(t, vm.ReadOnly)
}

// Test for multiple volume mounts
func TestGenerateVolumeMounts_MultipleFiles(t *testing.T) {
	component := &CustomComponent{
		Name:      "test-app",
		Image:     "nginx:latest",
		Namespace: "default",
		ConfigFiles: []ConfigFile{
			{
				Name:     "app.yaml",
				Path:     "/config/app.yaml",
				Contents: "server:\n  port: 8080",
			},
			{
				Name:     "logback.xml",
				Path:     "/config/logback.xml",
				Contents: "<configuration></configuration>",
			},
			{
				Name:     "database.properties",
				Path:     "/etc/database.properties",
				Contents: "host=localhost",
			},
		},
	}
	component.SetDefaults()

	volumeMounts, err := generateVolumeMounts(component)
	require.NoError(t, err)
	require.Len(t, volumeMounts, 3)

	// Verify all mounts use the same volume name
	volumeName := "test-app-config-volume"
	for _, vm := range volumeMounts {
		assert.Equal(t, volumeName, vm.Name)
		assert.True(t, vm.ReadOnly)
	}

	// Verify mount paths and subPaths
	mountPaths := make(map[string]string)
	for _, vm := range volumeMounts {
		mountPaths[vm.MountPath] = vm.SubPath
	}

	assert.Equal(t, "app.yaml", mountPaths["/config/app.yaml"])
	assert.Equal(t, "logback.xml", mountPaths["/config/logback.xml"])
	assert.Equal(t, "database.properties", mountPaths["/etc/database.properties"])
}

// Test for volume generation
func TestGenerateVolumes_SingleFile(t *testing.T) {
	component := &CustomComponent{
		Name:      "test-app",
		Image:     "nginx:latest",
		Namespace: "default",
		ConfigFiles: []ConfigFile{
			{
				Name:     "app.yaml",
				Path:     "/config/app.yaml",
				Contents: "server:\n  port: 8080",
			},
		},
	}
	component.SetDefaults()

	volumes, err := generateVolumes(component)
	require.NoError(t, err)
	require.Len(t, volumes, 1)

	vol := volumes[0]
	assert.Equal(t, "test-app-config-volume", vol.Name)
	require.NotNil(t, vol.ConfigMap)
	assert.Equal(t, "test-app-config", vol.ConfigMap.Name)
	assert.Equal(t, 0644, vol.ConfigMap.DefaultMode)
}

// Test for multiple files (should still generate single volume)
func TestGenerateVolumes_MultipleFiles(t *testing.T) {
	component := &CustomComponent{
		Name:      "test-app",
		Image:     "nginx:latest",
		Namespace: "default",
		ConfigFiles: []ConfigFile{
			{
				Name:     "app.yaml",
				Path:     "/config/app.yaml",
				Contents: "server:\n  port: 8080",
			},
			{
				Name:     "logback.xml",
				Path:     "/config/logback.xml",
				Contents: "<configuration></configuration>",
			},
		},
	}
	component.SetDefaults()

	volumes, err := generateVolumes(component)
	require.NoError(t, err)
	require.Len(t, volumes, 1) // Single volume for all files

	vol := volumes[0]
	assert.Equal(t, "test-app-config-volume", vol.Name)
	require.NotNil(t, vol.ConfigMap)
	assert.Equal(t, "test-app-config", vol.ConfigMap.Name)
}

// Test for no config files (should return nil)
func TestGenerateVolumes_NoConfigFiles(t *testing.T) {
	component := &CustomComponent{
		Name:        "test-app",
		Image:       "nginx:latest",
		Namespace:   "default",
		ConfigFiles: []ConfigFile{},
	}
	component.SetDefaults()

	volumes, err := generateVolumes(component)
	require.NoError(t, err)
	assert.Nil(t, volumes)
}

// Test for no config files volumeMounts (should return nil)
func TestGenerateVolumeMounts_NoConfigFiles(t *testing.T) {
	component := &CustomComponent{
		Name:        "test-app",
		Image:       "nginx:latest",
		Namespace:   "default",
		ConfigFiles: []ConfigFile{},
	}
	component.SetDefaults()

	volumeMounts, err := generateVolumeMounts(component)
	require.NoError(t, err)
	assert.Nil(t, volumeMounts)
}
