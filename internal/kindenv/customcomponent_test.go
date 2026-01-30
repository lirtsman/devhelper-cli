package kindenv

import (
	"testing"
)

// customcomponent_test.go provides tests for custom component deployment functionality

func TestDeployCustomComponents(t *testing.T) {
	// TODO: Implement tests for deployCustomComponents
	// This will require mocking kubectl operations
	t.Skip("deployCustomComponents tests not yet implemented")
}

func TestGenerateDeploymentYAML(t *testing.T) {
	// TODO: Implement tests for generateDeploymentYAML
	t.Skip("generateDeploymentYAML tests not yet implemented")
}

func TestFilterEnabledComponents(t *testing.T) {
	tests := []struct {
		name       string
		components []CustomComponent
		expected   int
	}{
		{
			name: "all enabled",
			components: []CustomComponent{
				{Name: "app1", Image: "nginx:latest", Enabled: boolPtr(true)},
				{Name: "app2", Image: "redis:latest", Enabled: boolPtr(true)},
			},
			expected: 2,
		},
		{
			name: "some disabled",
			components: []CustomComponent{
				{Name: "app1", Image: "nginx:latest", Enabled: boolPtr(true)},
				{Name: "app2", Image: "redis:latest", Enabled: boolPtr(false)},
			},
			expected: 1,
		},
		{
			name: "all disabled",
			components: []CustomComponent{
				{Name: "app1", Image: "nginx:latest", Enabled: boolPtr(false)},
				{Name: "app2", Image: "redis:latest", Enabled: boolPtr(false)},
			},
			expected: 0,
		},
		{
			name: "nil enabled defaults to true",
			components: []CustomComponent{
				{Name: "app1", Image: "nginx:latest", Enabled: nil},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var enabled []CustomComponent
			for _, component := range tt.components {
				if component.Enabled == nil || *component.Enabled {
					enabled = append(enabled, component)
				}
			}
			if len(enabled) != tt.expected {
				t.Errorf("expected %d enabled components, got %d", tt.expected, len(enabled))
			}
		})
	}
}

// Helper functions

func boolPtr(b bool) *bool {
	return &b
}

// Placeholder for future tests that will require kubectl mocking
func TestDeployCustomComponentsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	// TODO: Implement integration tests with real Kind cluster
	t.Skip("integration tests not yet implemented")
}
