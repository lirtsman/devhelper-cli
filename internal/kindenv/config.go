package kindenv

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"gopkg.in/yaml.v3"
)

// KindEnvConfig defines the structure for Kind environment configuration
type KindEnvConfig struct {
	Tools struct {
		Podman struct {
			Path    string `yaml:"path"`
			Version string `yaml:"version"`
		} `yaml:"podman"`
		Docker struct {
			Path    string `yaml:"path"`
			Version string `yaml:"version"`
		} `yaml:"docker"`
		Kind struct {
			Path    string `yaml:"path"`
			Version string `yaml:"version"`
		} `yaml:"kind"`
		Kubectl struct {
			Path    string `yaml:"path"`
			Version string `yaml:"version"`
		} `yaml:"kubectl"`
		Helm struct {
			Path    string `yaml:"path"`
			Version string `yaml:"version"`
		} `yaml:"helm"`
		AWS struct {
			Path    string `yaml:"path"`
			Version string `yaml:"version"`
		} `yaml:"aws"`
	} `yaml:"tools"`
	Cluster struct {
		Name              string `yaml:"name"`
		CreateIfNotExists bool   `yaml:"createIfNotExists"`
		MapPorts          []struct {
			ContainerPort interface{} `yaml:"containerPort"`
			HostPort      int         `yaml:"hostPort"`
			Protocol      string      `yaml:"protocol"`
		} `yaml:"mapPorts"`
	} `yaml:"cluster"`
	Components struct {
		Temporal struct {
			Enabled      bool   `yaml:"enabled"`
			Namespace    string `yaml:"namespace"`
			ChartVersion string `yaml:"chartVersion"`
			NodePorts    struct {
				Web      int `yaml:"web"`
				Frontend int `yaml:"frontend"`
			} `yaml:"nodePorts"`
		} `yaml:"temporal"`
		Redis struct {
			Enabled   bool `yaml:"enabled"`
			NodePorts struct {
				Redis int `yaml:"redis"`
			} `yaml:"nodePorts"`
			ChartVersion string `yaml:"chartVersion"`
			Auth         struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"auth"`
		} `yaml:"redis"`
		Dapr struct {
			Enabled      bool   `yaml:"enabled"`
			ChartVersion string `yaml:"chartVersion"`
			NodePorts    struct {
				Dashboard int `yaml:"dashboard"`
			} `yaml:"nodePorts"`
			LogLevel string `yaml:"logLevel"`
			Mtls     struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"mtls"`
			Ha struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"ha"`
		} `yaml:"dapr"`
	} `yaml:"components"`
	Images struct {
		SkipPull  bool `yaml:"skipPull"`
		DockerHub struct {
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		} `yaml:"dockerHub"`
		UseAwsEcr bool `yaml:"useAwsEcr"`
		AWS       struct {
			Region         string `yaml:"region"`
			EcrRegistry    string `yaml:"ecrRegistry"`
			ServiceAccount string `yaml:"serviceAccount"`
		} `yaml:"aws"`
	} `yaml:"images"`
	Secrets struct {
		MySQL struct {
			Enabled   bool   `yaml:"enabled"`
			Name      string `yaml:"name"`
			Namespace string `yaml:"namespace"`
			Username  string `yaml:"username"`
			Password  string `yaml:"password"`
		} `yaml:"mysql"`
	} `yaml:"secrets"`
}

// LoadConfig loads the configuration from a file
func LoadConfig(configPath string) (*KindEnvConfig, error) {
	// Set default values
	config := &KindEnvConfig{}

	// Set defaults
	config.Cluster.Name = "kindenv"
	config.Cluster.CreateIfNotExists = true

	// Set component defaults
	config.Components.Temporal.Enabled = true
	config.Components.Temporal.Namespace = "temporal"
	config.Components.Temporal.ChartVersion = "0.62.0"
	config.Components.Temporal.NodePorts.Web = 30080
	config.Components.Temporal.NodePorts.Frontend = 30733

	config.Components.Redis.Enabled = true
	config.Components.Redis.NodePorts.Redis = 30679
	config.Components.Redis.ChartVersion = "17.3.7"
	config.Components.Redis.Auth.Enabled = false

	config.Components.Dapr.Enabled = true
	config.Components.Dapr.ChartVersion = "1.15.3"
	config.Components.Dapr.NodePorts.Dashboard = 30479
	config.Components.Dapr.LogLevel = "debug"

	// If configPath is empty, check if kindenv.yaml exists in the current directory
	if configPath == "" {
		if _, err := os.Stat("kindenv.yaml"); err == nil {
			configPath = "kindenv.yaml"
		} else {
			// No config file provided or found, return defaults
			return config, nil
		}
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Process variable substitutions if mapPorts is provided
	if len(config.Cluster.MapPorts) > 0 {
		err = processVariableSubstitutions(config)
		if err != nil {
			return nil, fmt.Errorf("failed to process variable substitutions: %w", err)
		}
	} else {
		// No mapPorts in config file, generate default port mappings based on component settings
		config.Cluster.MapPorts = generateDefaultPortMappings(config)
	}

	return config, nil
}

// generateDefaultPortMappings creates default port mappings if none are provided
func generateDefaultPortMappings(config *KindEnvConfig) []struct {
	ContainerPort interface{} `yaml:"containerPort"`
	HostPort      int         `yaml:"hostPort"`
	Protocol      string      `yaml:"protocol"`
} {
	var mappings []struct {
		ContainerPort interface{} `yaml:"containerPort"`
		HostPort      int         `yaml:"hostPort"`
		Protocol      string      `yaml:"protocol"`
	}

	// // Add standard ports
	// mappings = append(mappings, struct {
	// 	ContainerPort interface{} `yaml:"containerPort"`
	// 	HostPort      int         `yaml:"hostPort"`
	// 	Protocol      string      `yaml:"protocol"`
	// }{
	// 	ContainerPort: 80,
	// 	HostPort:      80,
	// 	Protocol:      "TCP",
	// })
	// mappings = append(mappings, struct {
	// 	ContainerPort interface{} `yaml:"containerPort"`
	// 	HostPort      int         `yaml:"hostPort"`
	// 	Protocol      string      `yaml:"protocol"`
	// }{
	// 	ContainerPort: 443,
	// 	HostPort:      443,
	// 	Protocol:      "TCP",
	// })

	// Add component-specific ports if components are enabled
	if config.Components.Temporal.Enabled {
		// Temporal web UI
		mappings = append(mappings, struct {
			ContainerPort interface{} `yaml:"containerPort"`
			HostPort      int         `yaml:"hostPort"`
			Protocol      string      `yaml:"protocol"`
		}{
			ContainerPort: "${{ components.temporal.nodePorts.web }}",
			HostPort:      8080,
			Protocol:      "TCP",
		})

		// Temporal frontend
		mappings = append(mappings, struct {
			ContainerPort interface{} `yaml:"containerPort"`
			HostPort      int         `yaml:"hostPort"`
			Protocol      string      `yaml:"protocol"`
		}{
			ContainerPort: "${{ components.temporal.nodePorts.frontend }}",
			HostPort:      7233,
			Protocol:      "TCP",
		})
	}

	// Add Redis if enabled
	if config.Components.Redis.Enabled {
		mappings = append(mappings, struct {
			ContainerPort interface{} `yaml:"containerPort"`
			HostPort      int         `yaml:"hostPort"`
			Protocol      string      `yaml:"protocol"`
		}{
			ContainerPort: "${{ components.redis.nodePorts.redis }}",
			HostPort:      6379,
			Protocol:      "TCP",
		})
	}

	return mappings
}

// processVariableSubstitutions handles variable substitution in the config
func processVariableSubstitutions(config *KindEnvConfig) error {
	// Process containerPort references in mapPorts
	for i, portMap := range config.Cluster.MapPorts {
		// Check if containerPort is a string (potential variable reference)
		if strValue, ok := portMap.ContainerPort.(string); ok {
			// Pattern for ${{ components.x.nodePorts.y }}
			pattern := regexp.MustCompile(`\$\{\{\s*components\.([a-zA-Z0-9]+)\.nodePorts\.([a-zA-Z0-9]+)\s*\}\}`)
			matches := pattern.FindStringSubmatch(strValue)

			if len(matches) == 3 {
				componentName := matches[1]
				portName := matches[2]

				// Resolve the variable based on component and property
				var value int
				switch componentName {
				case "temporal":
					switch portName {
					case "web":
						value = config.Components.Temporal.NodePorts.Web
					case "frontend":
						value = config.Components.Temporal.NodePorts.Frontend
					default:
						return fmt.Errorf("unknown temporal port: %s", portName)
					}
				case "redis":
					switch portName {
					case "redis":
						value = config.Components.Redis.NodePorts.Redis
					default:
						return fmt.Errorf("unknown redis port: %s", portName)
					}
				case "dapr":
					switch portName {
					case "dashboard":
						value = config.Components.Dapr.NodePorts.Dashboard
					default:
						return fmt.Errorf("unknown dapr port: %s", portName)
					}
				default:
					return fmt.Errorf("unknown component: %s", componentName)
				}

				// Replace the variable with the resolved value
				config.Cluster.MapPorts[i].ContainerPort = value
			} else if strValue != "" {
				// If it's not a variable reference but a string, try to convert to int
				intValue, err := strconv.Atoi(strValue)
				if err != nil {
					return fmt.Errorf("invalid containerPort value: %s", strValue)
				}
				config.Cluster.MapPorts[i].ContainerPort = intValue
			}
		}
	}

	return nil
}

// Validate validates the configuration
func (c *KindEnvConfig) Validate() error {
	if c.Cluster.Name == "" {
		return errors.New("cluster name cannot be empty")
	}

	// Validate Temporal configuration
	if c.Components.Temporal.Enabled {
		if c.Components.Temporal.Namespace == "" {
			c.Components.Temporal.Namespace = "temporal"
		}
		if c.Components.Temporal.NodePorts.Web <= 0 {
			return errors.New("temporal web node port must be greater than 0")
		}
		if c.Components.Temporal.NodePorts.Frontend <= 0 {
			return errors.New("temporal frontend node port must be greater than 0")
		}
	}

	// Validate Redis configuration
	if c.Components.Redis.Enabled {
		if c.Components.Redis.NodePorts.Redis <= 0 {
			return errors.New("redis node port must be greater than 0")
		}
	}

	// Validate AWS ECR configuration
	if c.Images.UseAwsEcr {
		if c.Images.AWS.Region == "" {
			return errors.New("aws region must be specified when using ECR")
		}
		if c.Images.AWS.EcrRegistry == "" {
			return errors.New("aws ECR registry must be specified when using ECR")
		}
	}

	// Validate MySQL secret configuration
	if c.Secrets.MySQL.Enabled {
		if c.Secrets.MySQL.Name == "" {
			return errors.New("mysql secret name must be specified when enabled")
		}
		if c.Secrets.MySQL.Namespace == "" {
			return errors.New("mysql secret namespace must be specified when enabled")
		}
		if c.Secrets.MySQL.Username == "" {
			return errors.New("mysql username must be specified when enabled")
		}
		if c.Secrets.MySQL.Password == "" {
			return errors.New("mysql password must be specified when enabled")
		}
	}

	return nil
}

// CreateDefaultConfig creates a default configuration for Kind environments
func CreateDefaultConfig() *KindEnvConfig {
	config := &KindEnvConfig{}

	// Cluster section
	config.Cluster.Name = "kindenv"
	config.Cluster.CreateIfNotExists = true

	// Components section
	config.Components.Temporal.Enabled = true
	config.Components.Temporal.Namespace = "temporal"
	config.Components.Temporal.ChartVersion = "0.62.0"
	config.Components.Temporal.NodePorts.Web = 30080
	config.Components.Temporal.NodePorts.Frontend = 30733

	config.Components.Redis.Enabled = true
	config.Components.Redis.NodePorts.Redis = 30679
	config.Components.Redis.ChartVersion = "17.3.7"
	config.Components.Redis.Auth.Enabled = false

	config.Components.Dapr.Enabled = true
	config.Components.Dapr.ChartVersion = "1.15.3"
	config.Components.Dapr.NodePorts.Dashboard = 30479
	config.Components.Dapr.LogLevel = "debug"
	config.Components.Dapr.Mtls.Enabled = false
	config.Components.Dapr.Ha.Enabled = false

	// Images section
	config.Images.SkipPull = false
	config.Images.DockerHub.Username = ""
	config.Images.DockerHub.Password = ""
	config.Images.UseAwsEcr = true
	config.Images.AWS.Region = "eu-west-1"
	config.Images.AWS.EcrRegistry = ""
	config.Images.AWS.ServiceAccount = "ecr-pull-service-account"

	// Secrets section
	config.Secrets.MySQL.Enabled = true
	config.Secrets.MySQL.Name = "mysql-credentials"
	config.Secrets.MySQL.Namespace = "default"
	config.Secrets.MySQL.Username = "root"
	config.Secrets.MySQL.Password = "password"

	// Generate default port mappings based on enabled components
	config.Cluster.MapPorts = generateDefaultPortMappings(config)

	return config
}
