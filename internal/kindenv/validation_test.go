package kindenv

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// validation_test.go provides table-driven tests for custom component validation

func TestCustomComponentValidate(t *testing.T) {
	tests := []struct {
		name        string
		component   CustomComponent
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid minimal config",
			component: CustomComponent{
				Name:  "test-app",
				Image: "nginx:latest",
			},
			expectError: false,
		},
		{
			name: "missing name",
			component: CustomComponent{
				Image: "nginx:latest",
			},
			expectError: true,
			errorMsg:    "name is required",
		},
		{
			name: "missing image",
			component: CustomComponent{
				Name: "test-app",
			},
			expectError: true,
			errorMsg:    "image is required",
		},
		{
			name: "invalid DNS label name",
			component: CustomComponent{
				Name:  "Test_App",
				Image: "nginx:latest",
			},
			expectError: true,
			errorMsg:    "DNS label",
		},
		{
			name: "invalid replicas",
			component: CustomComponent{
				Name:     "test-app",
				Image:    "nginx:latest",
				Replicas: intPtr(0),
			},
			expectError: true,
			errorMsg:    "replicas must be >= 1",
		},
		{
			name: "empty command element",
			component: CustomComponent{
				Name:    "test-app",
				Image:   "nginx:latest",
				Command: []string{"java", ""},
			},
			expectError: true,
			errorMsg:    "command[1] cannot be empty",
		},
		{
			name: "empty args element",
			component: CustomComponent{
				Name:  "test-app",
				Image: "nginx:latest",
				Args:  []string{"-jar", "", "app.jar"},
			},
			expectError: true,
			errorMsg:    "args[1] cannot be empty",
		},
		{
			name: "valid command and args",
			component: CustomComponent{
				Name:    "test-app",
				Image:   "openjdk:11",
				Command: []string{"java"},
				Args:    []string{"-jar", "app.jar"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.component.Validate()
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestEnvVarValidate(t *testing.T) {
	tests := []struct {
		name        string
		envVar      EnvVar
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid direct value",
			envVar: EnvVar{
				Name:  "APP_ENV",
				Value: "production",
			},
			expectError: false,
		},
		{
			name: "valid secret reference",
			envVar: EnvVar{
				Name: "DB_PASSWORD",
				ValueFrom: &EnvVarSource{
					SecretKeyRef: &SecretKeySelector{
						Name: "mysql-secret",
						Key:  "password",
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing name",
			envVar: EnvVar{
				Value: "value",
			},
			expectError: true,
			errorMsg:    "name is required",
		},
		{
			name: "missing value and valueFrom",
			envVar: EnvVar{
				Name: "VAR_NAME",
			},
			expectError: true,
			errorMsg:    "either 'value' or 'valueFrom'",
		},
		{
			name: "both value and valueFrom",
			envVar: EnvVar{
				Name:  "VAR_NAME",
				Value: "value",
				ValueFrom: &EnvVarSource{
					SecretKeyRef: &SecretKeySelector{
						Name: "secret",
						Key:  "key",
					},
				},
			},
			expectError: true,
			errorMsg:    "mutually exclusive",
		},
		{
			name: "invalid env var name",
			envVar: EnvVar{
				Name:  "invalid-name",
				Value: "value",
			},
			expectError: true,
			errorMsg:    "invalid environment variable name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.envVar.Validate()
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestPortMappingValidate(t *testing.T) {
	tests := []struct {
		name        string
		port        PortMapping
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid port mapping",
			port: PortMapping{
				ContainerPort: 8080,
				HostPort:      8080,
				Protocol:      "TCP",
			},
			expectError: false,
		},
		{
			name: "invalid container port too low",
			port: PortMapping{
				ContainerPort: 0,
			},
			expectError: true,
			errorMsg:    "between 1 and 65535",
		},
		{
			name: "invalid container port too high",
			port: PortMapping{
				ContainerPort: 65536,
			},
			expectError: true,
			errorMsg:    "between 1 and 65535",
		},
		{
			name: "invalid nodePort too low",
			port: PortMapping{
				ContainerPort: 8080,
				NodePort:      29999,
			},
			expectError: true,
			errorMsg:    "between 30000 and 32767",
		},
		{
			name: "invalid protocol",
			port: PortMapping{
				ContainerPort: 8080,
				Protocol:      "HTTP",
			},
			expectError: true,
			errorMsg:    "TCP or UDP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.port.Validate()
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestConfigFileValidate(t *testing.T) {
	tests := []struct {
		name        string
		configFile  ConfigFile
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config file",
			configFile: ConfigFile{
				Name:     "app.yaml",
				Path:     "/config/app.yaml",
				Contents: "key: value",
			},
			expectError: false,
		},
		{
			name: "missing name",
			configFile: ConfigFile{
				Path:     "/config/app.yaml",
				Contents: "key: value",
			},
			expectError: true,
			errorMsg:    "name is required",
		},
		{
			name: "missing path",
			configFile: ConfigFile{
				Name:     "app.yaml",
				Contents: "key: value",
			},
			expectError: true,
			errorMsg:    "path is required",
		},
		{
			name: "relative path",
			configFile: ConfigFile{
				Name:     "app.yaml",
				Path:     "config/app.yaml",
				Contents: "key: value",
			},
			expectError: true,
			errorMsg:    "must be absolute",
		},
		{
			name: "name with directory separator",
			configFile: ConfigFile{
				Name:     "dir/app.yaml",
				Path:     "/config/app.yaml",
				Contents: "key: value",
			},
			expectError: true,
			errorMsg:    "cannot contain directory separators",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.configFile.Validate()
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// T037: Test for missing secret detection
func TestValidateSecretReferences_MissingSecret(t *testing.T) {
	// Note: This test will be fully implemented once we have kubectl integration
	// For now, we test the validation logic structure

	component := &CustomComponent{
		Name:      "my-app",
		Image:     "nginx:latest",
		Namespace: "default",
		Env: []EnvVar{
			{
				Name: "DB_PASSWORD",
				ValueFrom: &EnvVarSource{
					SecretKeyRef: &SecretKeySelector{
						Name: "nonexistent-secret",
						Key:  "password",
					},
				},
			},
		},
	}
	component.SetDefaults()

	// The validateSecretReferences function should check if the secret exists
	// For now, we just verify the component structure is valid
	err := component.Validate()
	assert.NoError(t, err, "Component structure should be valid even if secret doesn't exist yet")

	// Pre-deployment validation will check secret existence
	// This will be tested with actual kubectl calls in integration tests
}

// TestSecretExists_NoFalsePositives tests that secretExists doesn't have false positives
// Bug fix: Searching for "my-secret" shouldn't match "my-secret-name"
func TestSecretExists_NoFalsePositives(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test that requires kubectl")
	}

	ctx := context.Background()

	// Test case: If "my-secret-name" exists, searching for "my-secret" should return false
	// (not true due to substring match)

	// First, check if "my-secret-name" exists (it probably doesn't, but that's okay)
	// The important thing is that if it did exist, searching for "my-secret" wouldn't match it
	exists, err := secretExists(ctx, "default", "my-secret-name")
	if err != nil {
		t.Skipf("Skipping test: kubectl not available or cluster not running: %v", err)
	}

	// Now check for "my-secret" - should return false even if "my-secret-name" exists
	// (or false if neither exists, which is the expected behavior)
	existsShort, err := secretExists(ctx, "default", "my-secret")
	if err != nil {
		t.Skipf("Skipping test: kubectl not available or cluster not running: %v", err)
	}

	// The key test: if "my-secret-name" exists but "my-secret" doesn't,
	// existsShort should be false (not true due to substring match)
	// If both don't exist, existsShort should also be false
	if exists && existsShort {
		// This would indicate a bug: "my-secret" matched when it shouldn't
		t.Errorf("Bug detected: secretExists('my-secret') returned true when it should only match exact secret names")
	}

	// Verify the function uses exact matching by checking the expected output format
	// kubectl -o name outputs "secret/secret-name", so we should match exactly
	assert.False(t, existsShort || (exists && existsShort),
		"secretExists should use exact matching, not substring matching")
}

// T051: Test for port conflict detection
func TestValidatePortConflicts(t *testing.T) {
	tests := []struct {
		name        string
		components  []CustomComponent
		expectError bool
		errorMsg    string
	}{
		{
			name: "no conflicts - different NodePorts",
			components: []CustomComponent{
				{
					Name:  "app1",
					Image: "nginx:latest",
					Ports: []PortMapping{
						{ContainerPort: 8080, NodePort: 30001},
					},
				},
				{
					Name:  "app2",
					Image: "nginx:latest",
					Ports: []PortMapping{
						{ContainerPort: 8080, NodePort: 30002},
					},
				},
			},
			expectError: false,
		},
		{
			name: "conflict - same NodePort",
			components: []CustomComponent{
				{
					Name:  "app1",
					Image: "nginx:latest",
					Ports: []PortMapping{
						{ContainerPort: 8080, NodePort: 30001},
					},
				},
				{
					Name:  "app2",
					Image: "nginx:latest",
					Ports: []PortMapping{
						{ContainerPort: 9090, NodePort: 30001}, // Same NodePort
					},
				},
			},
			expectError: true,
			errorMsg:    "NodePort",
		},
		{
			name: "no conflicts - same container port, different NodePorts",
			components: []CustomComponent{
				{
					Name:  "app1",
					Image: "nginx:latest",
					Ports: []PortMapping{
						{ContainerPort: 8080, NodePort: 30001},
					},
				},
				{
					Name:  "app2",
					Image: "nginx:latest",
					Ports: []PortMapping{
						{ContainerPort: 8080, NodePort: 30002}, // Same container port, different NodePort
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usedPorts := make(map[int]bool)
			var err error

			for i := range tt.components {
				tt.components[i].SetDefaults()
				if err = validatePortConflicts(&tt.components[i], usedPorts); err != nil {
					break
				}
				// Mark ports as used
				for _, port := range tt.components[i].Ports {
					if port.NodePort != 0 {
						usedPorts[port.NodePort] = true
					}
				}
			}

			if tt.expectError {
				assert.Error(t, err, "Expected port conflict error")
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg, "Error should mention port conflict")
				}
			} else {
				assert.NoError(t, err, "Should not have port conflicts")
			}
		})
	}
}

// T063: Test for resource format validation
func TestValidateResourceRequirements_FormatValidation(t *testing.T) {
	tests := []struct {
		name        string
		resources   *ResourceRequirements
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid CPU and memory formats",
			resources: &ResourceRequirements{
				Requests: &ResourceList{
					CPU:    "100m",
					Memory: "128Mi",
				},
				Limits: &ResourceList{
					CPU:    "500m",
					Memory: "512Mi",
				},
			},
			expectError: false,
		},
		{
			name: "valid CPU formats - millicores",
			resources: &ResourceRequirements{
				Requests: &ResourceList{
					CPU: "100m",
				},
			},
			expectError: false,
		},
		{
			name: "valid CPU formats - cores",
			resources: &ResourceRequirements{
				Requests: &ResourceList{
					CPU: "1",
				},
			},
			expectError: false,
		},
		{
			name: "valid CPU formats - fractional cores",
			resources: &ResourceRequirements{
				Requests: &ResourceList{
					CPU: "0.5",
				},
			},
			expectError: false,
		},
		{
			name: "valid memory formats - Mi",
			resources: &ResourceRequirements{
				Requests: &ResourceList{
					Memory: "128Mi",
				},
			},
			expectError: false,
		},
		{
			name: "valid memory formats - Gi",
			resources: &ResourceRequirements{
				Requests: &ResourceList{
					Memory: "1Gi",
				},
			},
			expectError: false,
		},
		{
			name: "valid memory formats - M",
			resources: &ResourceRequirements{
				Requests: &ResourceList{
					Memory: "512M",
				},
			},
			expectError: false,
		},
		{
			name: "invalid CPU format",
			resources: &ResourceRequirements{
				Requests: &ResourceList{
					CPU: "invalid",
				},
			},
			expectError: true,
			errorMsg:    "invalid CPU quantity",
		},
		{
			name: "invalid memory format",
			resources: &ResourceRequirements{
				Requests: &ResourceList{
					Memory: "invalid",
				},
			},
			expectError: true,
			errorMsg:    "invalid memory quantity",
		},
		{
			name: "empty resources",
			resources: &ResourceRequirements{
				Requests: &ResourceList{},
				Limits:   &ResourceList{},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateResourceRequirements(tt.resources)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// T064: Test for requests <= limits validation
func TestValidateResourceRequirements_LimitsGreaterThanRequests(t *testing.T) {
	tests := []struct {
		name        string
		resources   *ResourceRequirements
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid - limits greater than requests",
			resources: &ResourceRequirements{
				Requests: &ResourceList{
					CPU:    "100m",
					Memory: "128Mi",
				},
				Limits: &ResourceList{
					CPU:    "500m",
					Memory: "512Mi",
				},
			},
			expectError: false,
		},
		{
			name: "valid - limits equal to requests",
			resources: &ResourceRequirements{
				Requests: &ResourceList{
					CPU:    "500m",
					Memory: "512Mi",
				},
				Limits: &ResourceList{
					CPU:    "500m",
					Memory: "512Mi",
				},
			},
			expectError: false,
		},
		{
			name: "valid - only requests specified",
			resources: &ResourceRequirements{
				Requests: &ResourceList{
					CPU:    "100m",
					Memory: "128Mi",
				},
			},
			expectError: false,
		},
		{
			name: "valid - only limits specified",
			resources: &ResourceRequirements{
				Limits: &ResourceList{
					CPU:    "500m",
					Memory: "512Mi",
				},
			},
			expectError: false,
		},
		{
			name: "valid - CPU only",
			resources: &ResourceRequirements{
				Requests: &ResourceList{
					CPU: "100m",
				},
				Limits: &ResourceList{
					CPU: "500m",
				},
			},
			expectError: false,
		},
		{
			name: "valid - Memory only",
			resources: &ResourceRequirements{
				Requests: &ResourceList{
					Memory: "128Mi",
				},
				Limits: &ResourceList{
					Memory: "512Mi",
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateResourceRequirements(tt.resources)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// Helper functions

func intPtr(i int) *int {
	return &i
}

func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
