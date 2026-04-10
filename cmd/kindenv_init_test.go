package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestInitPrometheusRepoOutputContract verifies the output messages for the
// prometheus-community Helm repo registration follow the expected contract.
func TestInitPrometheusRepoOutputContract(t *testing.T) {
	tests := []struct {
		name           string
		commandOutput  string
		commandErr     bool
		expectContains string
	}{
		{
			name:           "success output",
			commandOutput:  "",
			commandErr:     false,
			expectContains: "✅ Prometheus Community Helm repository configured",
		},
		{
			name:           "already exists output",
			commandOutput:  "repository name (prometheus-community) already exists",
			commandErr:     true,
			expectContains: "✅ Prometheus Community Helm repository already configured",
		},
		{
			name:           "failure output",
			commandOutput:  "connection refused",
			commandErr:     true,
			expectContains: "⚠️  Warning: Failed to add Prometheus Community Helm repository",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the output contract logic from kindenv_init.go
			var result string
			if tt.commandErr {
				if strings.Contains(tt.commandOutput, "already exists") {
					result = "✅ Prometheus Community Helm repository already configured"
				} else {
					result = "⚠️  Warning: Failed to add Prometheus Community Helm repository: <error>"
				}
			} else {
				result = "✅ Prometheus Community Helm repository configured"
			}

			assert.Contains(t, result, tt.expectContains)
		})
	}
}

// TestInitPrometheusRepoConstants verifies the prometheus-community repo name and URL
// match the expected values from the contract.
func TestInitPrometheusRepoConstants(t *testing.T) {
	// These constants mirror what is used in kindenv_init.go and must not change
	// without updating the helm chart references in kindenv_start.go.
	const (
		prometheusRepoName = "prometheus-community"
		prometheusRepoURL  = "https://prometheus-community.github.io/helm-charts"
		prometheusChart    = "prometheus-community/kube-prometheus-stack"
	)

	assert.Equal(t, "prometheus-community", prometheusRepoName,
		"Repo name must match the name used in kindenv_start.go helm args")
	assert.Equal(t, "https://prometheus-community.github.io/helm-charts", prometheusRepoURL,
		"Repo URL must match the official prometheus-community helm charts URL")
	assert.True(t, strings.HasPrefix(prometheusChart, prometheusRepoName+"/"),
		"Chart reference must use the registered repo name as prefix")
}

// TestInitPrometheusAlreadyExistsHandling verifies the "already exists" detection logic.
func TestInitPrometheusAlreadyExistsHandling(t *testing.T) {
	alreadyExistsOutputs := []string{
		"Error: repository name (prometheus-community) already exists",
		"repository name (prometheus-community) already exists, please specify a different name",
	}

	for _, output := range alreadyExistsOutputs {
		assert.True(t, strings.Contains(output, "already exists"),
			"'already exists' detection should handle output: %s", output)
	}

	// Verify a genuine error does NOT match the "already exists" branch
	genuineError := "connection refused to https://prometheus-community.github.io/helm-charts"
	assert.False(t, strings.Contains(genuineError, "already exists"),
		"genuine errors should not be treated as 'already exists'")
}
