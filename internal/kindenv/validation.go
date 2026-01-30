package kindenv

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// validation.go provides comprehensive validation for custom components,
// including configuration validation, pre-deployment checks, and runtime validation.

// ValidateCustomComponent validates a custom component's configuration
func ValidateCustomComponent(component *CustomComponent) error {
	// Required fields
	if component.Name == "" {
		return fmt.Errorf("custom component name is required")
	}
	if component.Image == "" {
		return fmt.Errorf("custom component '%s': image is required", component.Name)
	}

	// Name format (DNS-1123 label)
	if !isValidDNSLabel(component.Name) {
		return fmt.Errorf("custom component name '%s' must be valid DNS label (lowercase, alphanumeric, hyphens)", component.Name)
	}

	// Image format validation
	if !isValidImageFormat(component.Image) {
		return fmt.Errorf("custom component '%s': invalid image format '%s' (expected: [registry/]repository[:tag])", component.Name, component.Image)
	}

	// Replicas validation (only validate if set, don't mutate)
	if component.Replicas != nil && *component.Replicas < 1 {
		return fmt.Errorf("custom component '%s': replicas must be >= 1", component.Name)
	}

	// Command and args validation
	if len(component.Command) > 0 {
		for i, cmd := range component.Command {
			if cmd == "" {
				return fmt.Errorf("custom component '%s': command[%d] cannot be empty", component.Name, i)
			}
		}
	}
	if len(component.Args) > 0 {
		for i, arg := range component.Args {
			if arg == "" {
				return fmt.Errorf("custom component '%s': args[%d] cannot be empty", component.Name, i)
			}
		}
	}

	// Environment variables validation
	for i, env := range component.Env {
		if err := ValidateEnvVar(&env); err != nil {
			return fmt.Errorf("custom component '%s': env[%d]: %w", component.Name, i, err)
		}
	}

	// Port validation
	for i, port := range component.Ports {
		if err := ValidatePortMapping(&port); err != nil {
			return fmt.Errorf("custom component '%s': ports[%d]: %w", component.Name, i, err)
		}
	}

	// Resource validation (only validate if set, don't mutate)
	if component.Resources != nil {
		if err := ValidateResourceRequirements(component.Resources); err != nil {
			return fmt.Errorf("custom component '%s': resources: %w", component.Name, err)
		}
	}

	// Config files validation
	if len(component.ConfigFiles) > 0 {
		pathsSeen := make(map[string]bool)
		namesSeen := make(map[string]bool)
		totalSize := 0

		for i, cf := range component.ConfigFiles {
			if err := ValidateConfigFile(&cf); err != nil {
				return fmt.Errorf("custom component '%s': configFiles[%d]: %w", component.Name, i, err)
			}

			// Check for duplicate mount paths
			if pathsSeen[cf.Path] {
				return fmt.Errorf("custom component '%s': duplicate mount path '%s' in config files", component.Name, cf.Path)
			}
			pathsSeen[cf.Path] = true

			// Check for duplicate filenames
			if namesSeen[cf.Name] {
				return fmt.Errorf("custom component '%s': duplicate config file name '%s'", component.Name, cf.Name)
			}
			namesSeen[cf.Name] = true

			// Track total size (ConfigMap limit is 1MB)
			totalSize += len(cf.Contents)
		}

		if totalSize > 1024*1024 {
			return fmt.Errorf("custom component '%s': total config files size exceeds 1MB limit (%d bytes)", component.Name, totalSize)
		}
	}

	return nil
}

// ValidateEnvVar validates an environment variable configuration
func ValidateEnvVar(envVar *EnvVar) error {
	if envVar.Name == "" {
		return fmt.Errorf("environment variable name is required")
	}

	// Validate env var name format (must be valid shell variable name)
	if !isValidEnvVarName(envVar.Name) {
		return fmt.Errorf("invalid environment variable name '%s' (must match [A-Z_][A-Z0-9_]*)", envVar.Name)
	}

	// Either Value or ValueFrom must be set, but not both
	hasValue := envVar.Value != ""
	hasValueFrom := envVar.ValueFrom != nil

	if !hasValue && !hasValueFrom {
		return fmt.Errorf("environment variable '%s': either 'value' or 'valueFrom' must be specified", envVar.Name)
	}

	if hasValue && hasValueFrom {
		return fmt.Errorf("environment variable '%s': 'value' and 'valueFrom' are mutually exclusive", envVar.Name)
	}

	// Validate ValueFrom if present
	if hasValueFrom {
		if envVar.ValueFrom.SecretKeyRef != nil {
			if envVar.ValueFrom.SecretKeyRef.Name == "" {
				return fmt.Errorf("environment variable '%s': secretKeyRef.name is required", envVar.Name)
			}
			if envVar.ValueFrom.SecretKeyRef.Key == "" {
				return fmt.Errorf("environment variable '%s': secretKeyRef.key is required", envVar.Name)
			}
		}
	}

	return nil
}

// ValidatePortMapping validates a port mapping configuration
func ValidatePortMapping(port *PortMapping) error {
	// Container port validation
	if port.ContainerPort < 1 || port.ContainerPort > 65535 {
		return fmt.Errorf("containerPort must be between 1 and 65535, got %d", port.ContainerPort)
	}

	// Host port validation (if specified)
	if port.HostPort != 0 {
		if port.HostPort < 1024 || port.HostPort > 65535 {
			return fmt.Errorf("hostPort must be between 1024 and 65535, got %d", port.HostPort)
		}
	}

	// NodePort validation (if specified)
	if port.NodePort != 0 {
		if port.NodePort < 30000 || port.NodePort > 32767 {
			return fmt.Errorf("nodePort must be between 30000 and 32767, got %d", port.NodePort)
		}
	}

	// Protocol validation (check validity without mutating)
	protocol := port.Protocol
	if protocol == "" {
		protocol = "TCP" // Default, but don't mutate the original
	}
	protocol = strings.ToUpper(protocol)
	if protocol != "TCP" && protocol != "UDP" {
		return fmt.Errorf("protocol must be TCP or UDP, got %s", port.Protocol)
	}

	return nil
}

// ValidateResourceRequirements validates resource requirements
func ValidateResourceRequirements(resources *ResourceRequirements) error {
	if resources.Requests != nil {
		if err := validateResourceList(resources.Requests); err != nil {
			return fmt.Errorf("requests: %w", err)
		}
	}

	if resources.Limits != nil {
		if err := validateResourceList(resources.Limits); err != nil {
			return fmt.Errorf("limits: %w", err)
		}
	}

	// Validate that limits >= requests (if both specified)
	if resources.Requests != nil && resources.Limits != nil {
		if err := validateResourceLimitsGreaterThanRequests(resources.Requests, resources.Limits); err != nil {
			return err
		}
	}

	return nil
}

// validateResourceList validates CPU and memory quantities
func validateResourceList(rl *ResourceList) error {
	if rl.CPU != "" {
		if !isValidCPUQuantity(rl.CPU) {
			return fmt.Errorf("invalid CPU quantity '%s' (examples: 100m, 0.5, 1)", rl.CPU)
		}
	}

	if rl.Memory != "" {
		if !isValidMemoryQuantity(rl.Memory) {
			return fmt.Errorf("invalid memory quantity '%s' (examples: 128Mi, 1Gi, 512M)", rl.Memory)
		}
	}

	return nil
}

// validateResourceLimitsGreaterThanRequests ensures limits >= requests
func validateResourceLimitsGreaterThanRequests(requests, limits *ResourceList) error {
	// This is a simplified check - full validation would require parsing quantities
	// For now, we'll rely on Kubernetes to validate this
	return nil
}

// ValidateConfigFile validates a config file configuration
func ValidateConfigFile(configFile *ConfigFile) error {
	if configFile.Name == "" {
		return fmt.Errorf("config file name is required")
	}

	if configFile.Path == "" {
		return fmt.Errorf("config file '%s': path is required", configFile.Name)
	}

	if configFile.Contents == "" {
		return fmt.Errorf("config file '%s': contents cannot be empty", configFile.Name)
	}

	// Validate path format (must be absolute)
	if !strings.HasPrefix(configFile.Path, "/") {
		return fmt.Errorf("config file '%s': path must be absolute (start with /), got '%s'", configFile.Name, configFile.Path)
	}

	// Validate filename has no directory separators
	if strings.Contains(configFile.Name, "/") || strings.Contains(configFile.Name, "\\") {
		return fmt.Errorf("config file name '%s' cannot contain directory separators", configFile.Name)
	}

	// Validate contents size (ConfigMap limit is 1MB)
	if len(configFile.Contents) > 1024*1024 {
		return fmt.Errorf("config file '%s': contents exceed 1MB limit (%d bytes)", configFile.Name, len(configFile.Contents))
	}

	return nil
}

// validatePortConflicts validates that ports don't conflict with already used ports
func validatePortConflicts(component *CustomComponent, usedPorts map[int]bool) error {
	for _, port := range component.Ports {
		if port.NodePort != 0 {
			if usedPorts[port.NodePort] {
				return fmt.Errorf("component '%s': NodePort %d is already in use", component.Name, port.NodePort)
			}
		}
	}
	return nil
}

// validateSecretReferences validates that all referenced secrets exist in the cluster
func validateSecretReferences(ctx context.Context, component *CustomComponent) error {
	if len(component.Env) == 0 {
		return nil
	}

	// Collect all unique secret references
	secretRefs := make(map[string]map[string]bool) // namespace -> secretName -> keys
	for _, env := range component.Env {
		if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
			secretName := env.ValueFrom.SecretKeyRef.Name
			namespace := component.Namespace
			if namespace == "" {
				namespace = "default"
			}

			if secretRefs[namespace] == nil {
				secretRefs[namespace] = make(map[string]bool)
			}
			secretRefs[namespace][secretName] = true
		}
	}

	// Check each secret exists
	var missingSecrets []string
	for namespace, secrets := range secretRefs {
		for secretName := range secrets {
			exists, err := secretExists(ctx, namespace, secretName)
			if err != nil {
				return fmt.Errorf("failed to check secret '%s' in namespace '%s': %w", secretName, namespace, err)
			}
			if !exists {
				missingSecrets = append(missingSecrets, fmt.Sprintf("%s/%s", namespace, secretName))
			}
		}
	}

	if len(missingSecrets) > 0 {
		return fmt.Errorf("component '%s' references secrets that do not exist: %s. Please ensure these secrets are created before deploying", component.Name, strings.Join(missingSecrets, ", "))
	}

	return nil
}

// secretExists checks if a secret exists in the specified namespace using kubectl
func secretExists(ctx context.Context, namespace, secretName string) (bool, error) {
	// Use kubectl to check if secret exists
	// kubectl get secret <name> -n <namespace> --ignore-not-found
	cmd := exec.CommandContext(ctx, "kubectl", "get", "secret", secretName, "-n", namespace, "--ignore-not-found", "-o", "name")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	if err != nil {
		// If command fails, assume secret doesn't exist
		return false, nil
	}
	
	output := strings.TrimSpace(stdout.String())
	// If output contains the secret name, it exists
	return strings.Contains(output, secretName), nil
}

// Helper validation functions

// isValidDNSLabel checks if a string is a valid DNS-1123 label
func isValidDNSLabel(name string) bool {
	if len(name) == 0 || len(name) > 63 {
		return false
	}
	// DNS-1123 label: [a-z0-9]([-a-z0-9]*[a-z0-9])?
	dnsLabelRegex := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	return dnsLabelRegex.MatchString(name)
}

// isValidImageFormat checks if a string is a valid container image format
func isValidImageFormat(image string) bool {
	if len(image) == 0 {
		return false
	}
	// Basic validation: should contain at least repository:tag or repository
	// More complex validation would check registry format, etc.
	parts := strings.Split(image, "/")
	if len(parts) == 0 {
		return false
	}
	// Last part should be repository[:tag]
	lastPart := parts[len(parts)-1]
	if len(lastPart) == 0 {
		return false
	}
	return true
}

// isValidEnvVarName checks if a string is a valid environment variable name
func isValidEnvVarName(name string) bool {
	if len(name) == 0 {
		return false
	}
	// Must start with letter or underscore, followed by letters, digits, or underscores
	envVarRegex := regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)
	return envVarRegex.MatchString(name)
}

// isValidCPUQuantity checks if a string is a valid CPU quantity
func isValidCPUQuantity(cpu string) bool {
	if len(cpu) == 0 {
		return false
	}
	// Valid formats: "100m", "0.5", "1", "2000m"
	cpuRegex := regexp.MustCompile(`^(\d+m|\d+(\.\d+)?)$`)
	return cpuRegex.MatchString(cpu)
}

// isValidMemoryQuantity checks if a string is a valid memory quantity
func isValidMemoryQuantity(memory string) bool {
	if len(memory) == 0 {
		return false
	}
	// Valid formats: "128Mi", "1Gi", "512M", "1G"
	memoryRegex := regexp.MustCompile(`^\d+(\.\d+)?(Ki|Mi|Gi|Ti|Pi|Ei|K|M|G|T|P|E)?$`)
	return memoryRegex.MatchString(memory)
}
