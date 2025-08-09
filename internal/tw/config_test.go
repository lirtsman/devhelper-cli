package tw

import (
	"testing"
)

func TestExtractWorkerTypeFromName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty name",
			input:    "",
			expected: "",
		},
		{
			name:     "single word name",
			input:    "worker",
			expected: "",
		},
		{
			name:     "temporal prefix",
			input:    "temporal-ingestion-parsing",
			expected: "ingestion",
		},
		{
			name:     "temporal without hyphen",
			input:    "temporalingestion",
			expected: "",
		},
		{
			name:     "two-part name without temporal",
			input:    "data-processor",
			expected: "processor",
		},
		{
			name:     "with temporal in middle",
			input:    "my-temporal-ingestion",
			expected: "ingestion",
		},
		{
			name:     "temporal at end",
			input:    "processing-temporal",
			expected: "processing",
		},
		{
			name:     "multi-part name",
			input:    "complex-multi-part-name",
			expected: "multi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractWorkerTypeFromName(tt.input)
			if result != tt.expected {
				t.Errorf("extractWorkerTypeFromName(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCreateDefaultConfig(t *testing.T) {
	tests := []struct {
		name           string
		projectName    string
		workerType     string
		expectedName   string
		expectedType   string
	}{
		{
			name:           "explicit worker type",
			projectName:    "temporal-ingestion-worker",
			workerType:     "parser",
			expectedName:   "temporal-ingestion-worker",
			expectedType:   "parser",
		},
		{
			name:           "extracted worker type",
			projectName:    "temporal-ingestion-worker",
			workerType:     "",
			expectedName:   "temporal-ingestion-worker",
			expectedType:   "ingestion",
		},
		{
			name:           "simple name with no extraction",
			projectName:    "myworker",
			workerType:     "",
			expectedName:   "myworker",
			expectedType:   "myworker",
		},
		{
			name:           "empty name defaults",
			projectName:    "",
			workerType:     "",
			expectedName:   "temporal-worker", // This might be different based on implementation
			expectedType:   "temporal-worker", // Depends on what name is set to by default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The last test (empty name) would use current directory, so we'll skip validation for it
			if tt.projectName == "" {
				return
			}

			config := CreateDefaultConfig(tt.projectName, tt.workerType)
			
			if config.Metadata.Name != tt.expectedName {
				t.Errorf("CreateDefaultConfig(%q, %q): expected name %q, got %q", 
					tt.projectName, tt.workerType, tt.expectedName, config.Metadata.Name)
			}
			
			if config.Spec.WorkerType != tt.expectedType {
				t.Errorf("CreateDefaultConfig(%q, %q): expected worker type %q, got %q", 
					tt.projectName, tt.workerType, tt.expectedType, config.Spec.WorkerType)
			}
		})
	}
}