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
		OpenSearch struct {
			Enabled   bool   `yaml:"enabled"`
			Namespace string `yaml:"namespace"`
			Version   string `yaml:"version"`
			NodePorts struct {
				Rest int `yaml:"rest"`
			} `yaml:"nodePorts"`
			Security struct {
				Disabled bool `yaml:"disabled"`
			} `yaml:"security"`
			IndexManagement struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"indexManagement"`
		} `yaml:"openSearch"`
		OpenSearchDashboards struct {
			Enabled   bool   `yaml:"enabled"`
			Namespace string `yaml:"namespace"`
			Version   string `yaml:"version"`
			NodePorts struct {
				Http int `yaml:"http"`
			} `yaml:"nodePorts"`
		} `yaml:"openSearchDashboards"`
		TemporalWorkerOperator struct {
			Enabled           bool   `yaml:"enabled"`
			ChartVersion      string `yaml:"chartVersion"`
			TemporalNamespace string `yaml:"temporalNamespace"`
		} `yaml:"temporalWorkerOperator"`
		IndicesOperator struct {
			Enabled      bool   `yaml:"enabled"`
			ChartVersion string `yaml:"chartVersion"`
		} `yaml:"indicesOperator"`
		MetricsServer struct {
			Enabled      bool   `yaml:"enabled"`
			ChartVersion string `yaml:"chartVersion"`
		} `yaml:"metricsServer"`
		MySQL struct {
			Enabled      bool   `yaml:"enabled"`
			Namespace    string `yaml:"namespace"`
			ChartVersion string `yaml:"chartVersion"`
			Database     string `yaml:"database"`
			NodePorts    struct {
				MySQL int `yaml:"mysql"`
			} `yaml:"nodePorts"`
			Resources struct {
				CPU    string `yaml:"cpu"`
				Memory string `yaml:"memory"`
			} `yaml:"resources"`
			Persistence struct {
				Enabled bool   `yaml:"enabled"`
				Size    string `yaml:"size"`
			} `yaml:"persistence"`
		} `yaml:"mysql"`
	} `yaml:"components"`
	Images struct {
		SkipPull  bool `yaml:"skipPull"`
		DockerHub struct {
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		} `yaml:"dockerHub"`
		UseAwsEcr bool `yaml:"useAwsEcr"`
		AWS       struct {
			Region      string `yaml:"region"`
			EcrRegistry string `yaml:"ecrRegistry"`
			Profile     string `yaml:"profile"`
		} `yaml:"aws"`
		UseHarbor      bool   `yaml:"useHarbor"`
		HarborRegistry string `yaml:"harborRegistry"`
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
	config.Cluster.Name = "kindenv-default" // Will be overridden by project name if available
	config.Cluster.CreateIfNotExists = true

	// Set component defaults
	config.Components.Temporal.Enabled = false
	config.Components.Temporal.Namespace = "temporal"
	config.Components.Temporal.ChartVersion = "0.62.0"
	config.Components.Temporal.NodePorts.Web = 30080
	config.Components.Temporal.NodePorts.Frontend = 30733

	config.Components.Redis.Enabled = false
	config.Components.Redis.NodePorts.Redis = 30679
	config.Components.Redis.ChartVersion = "17.3.7"
	config.Components.Redis.Auth.Enabled = false

	config.Components.Dapr.Enabled = false
	config.Components.Dapr.ChartVersion = "1.15.3"
	config.Components.Dapr.NodePorts.Dashboard = 30479
	config.Components.Dapr.LogLevel = "debug"

	config.Components.OpenSearch.Enabled = true
	config.Components.OpenSearch.Namespace = "opensearch"
	config.Components.OpenSearch.Version = "2.17.1"
	config.Components.OpenSearch.NodePorts.Rest = 30920
	config.Components.OpenSearch.Security.Disabled = true
	config.Components.OpenSearch.IndexManagement.Enabled = true

	config.Components.OpenSearchDashboards.Enabled = true
	config.Components.OpenSearchDashboards.Namespace = "opensearch"
	config.Components.OpenSearchDashboards.Version = "2.17.1"
	config.Components.OpenSearchDashboards.NodePorts.Http = 30601

	config.Components.MySQL.Enabled = true
	config.Components.MySQL.Namespace = "mysql"
	config.Components.MySQL.ChartVersion = "9.4.6"
	config.Components.MySQL.Database = "mysql"
	config.Components.MySQL.NodePorts.MySQL = 30306
	config.Components.MySQL.Resources.CPU = "500m"
	config.Components.MySQL.Resources.Memory = "1Gi"
	config.Components.MySQL.Persistence.Enabled = false
	config.Components.MySQL.Persistence.Size = "8Gi"

	// Set MySQL secret defaults
	config.Secrets.MySQL.Enabled = true
	config.Secrets.MySQL.Name = "mysql-credentials"
	config.Secrets.MySQL.Namespace = "mysql"
	config.Secrets.MySQL.Username = "root"
	config.Secrets.MySQL.Password = "password"

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
	// Add Dapr if enabled
	if config.Components.Dapr.Enabled {
		mappings = append(mappings, struct {
			ContainerPort interface{} `yaml:"containerPort"`
			HostPort      int         `yaml:"hostPort"`
			Protocol      string      `yaml:"protocol"`
		}{
			ContainerPort: "${{ components.dapr.nodePorts.dashboard }}",
			HostPort:      3500,
			Protocol:      "TCP",
		})
	}

	// Add OpenSearch if enabled
	if config.Components.OpenSearch.Enabled {
		mappings = append(mappings, struct {
			ContainerPort interface{} `yaml:"containerPort"`
			HostPort      int         `yaml:"hostPort"`
			Protocol      string      `yaml:"protocol"`
		}{
			ContainerPort: "${{ components.openSearch.nodePorts.rest }}",
			HostPort:      9200,
			Protocol:      "TCP",
		})
	}

	// Add OpenSearch Dashboards if enabled
	if config.Components.OpenSearchDashboards.Enabled {
		mappings = append(mappings, struct {
			ContainerPort interface{} `yaml:"containerPort"`
			HostPort      int         `yaml:"hostPort"`
			Protocol      string      `yaml:"protocol"`
		}{
			ContainerPort: "${{ components.openSearchDashboards.nodePorts.http }}",
			HostPort:      5601,
			Protocol:      "TCP",
		})
	}

	// Add MySQL port mapping (always include, even if disabled, so users can see what will be mapped)
	mappings = append(mappings, struct {
		ContainerPort interface{} `yaml:"containerPort"`
		HostPort      int         `yaml:"hostPort"`
		Protocol      string      `yaml:"protocol"`
	}{
		ContainerPort: "${{ components.mysql.nodePorts.mysql }}",
		HostPort:      3306,
		Protocol:      "TCP",
	})

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
				case "openSearch":
					switch portName {
					case "rest":
						value = config.Components.OpenSearch.NodePorts.Rest
					default:
						return fmt.Errorf("unknown openSearch port: %s", portName)
					}
				case "openSearchDashboards":
					switch portName {
					case "http":
						value = config.Components.OpenSearchDashboards.NodePorts.Http
					default:
						return fmt.Errorf("unknown openSearchDashboards port: %s", portName)
					}
				case "mysql":
					switch portName {
					case "mysql":
						value = config.Components.MySQL.NodePorts.MySQL
					default:
						return fmt.Errorf("unknown mysql port: %s", portName)
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

	// Validate OpenSearch configuration
	if c.Components.OpenSearch.Enabled {
		if c.Components.OpenSearch.Namespace == "" {
			c.Components.OpenSearch.Namespace = "opensearch"
		}
		if c.Components.OpenSearch.NodePorts.Rest <= 0 {
			return errors.New("opensearch rest node port must be greater than 0")
		}
		// Set default for index management if not explicitly set
		if !c.Components.OpenSearch.IndexManagement.Enabled {
			c.Components.OpenSearch.IndexManagement.Enabled = true
		}
	}

	// Validate OpenSearch Dashboards configuration
	if c.Components.OpenSearchDashboards.Enabled {
		if c.Components.OpenSearchDashboards.Namespace == "" {
			c.Components.OpenSearchDashboards.Namespace = "opensearch"
		}
		if c.Components.OpenSearchDashboards.NodePorts.Http <= 0 {
			return errors.New("opensearch dashboards http node port must be greater than 0")
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

	// Validate MySQL component configuration
	if c.Components.MySQL.Enabled {
		if c.Components.MySQL.Namespace == "" {
			c.Components.MySQL.Namespace = "mysql"
		}
		if c.Components.MySQL.ChartVersion == "" {
			return errors.New("mysql chart version must be specified when enabled")
		}
		if c.Components.MySQL.Database == "" {
			return errors.New("mysql database name must be specified when enabled")
		}
		if c.Components.MySQL.NodePorts.MySQL < 30000 || c.Components.MySQL.NodePorts.MySQL > 32767 {
			return errors.New("mysql nodeport must be in range 30000-32767")
		}
		// Validate CPU format (e.g., "500m", "1")
		if c.Components.MySQL.Resources.CPU != "" {
			cpuRegex := regexp.MustCompile(`^[0-9]+m?$`)
			if !cpuRegex.MatchString(c.Components.MySQL.Resources.CPU) {
				return errors.New("mysql cpu resource must be in valid format (e.g., 500m, 1)")
			}
		}
		// Validate memory format (e.g., "1Gi", "512Mi")
		if c.Components.MySQL.Resources.Memory != "" {
			memoryRegex := regexp.MustCompile(`^[0-9]+[KMGT]i$`)
			if !memoryRegex.MatchString(c.Components.MySQL.Resources.Memory) {
				return errors.New("mysql memory resource must be in valid format (e.g., 1Gi, 512Mi)")
			}
		}
		// Validate persistence size if enabled
		if c.Components.MySQL.Persistence.Enabled {
			if c.Components.MySQL.Persistence.Size == "" {
				return errors.New("mysql persistence size must be specified when persistence is enabled")
			}
			persistenceSizeRegex := regexp.MustCompile(`^[0-9]+[KMGT]i$`)
			if !persistenceSizeRegex.MatchString(c.Components.MySQL.Persistence.Size) {
				return errors.New("mysql persistence size must be in valid format (e.g., 8Gi, 10Gi)")
			}
		}
	}

	return nil
}

// CreateDefaultConfig creates a default configuration for Kind environments
func CreateDefaultConfig() *KindEnvConfig {
	config := &KindEnvConfig{}

	// Cluster section
	config.Cluster.Name = "kindenv-default" // Will be overridden by project name if available
	config.Cluster.CreateIfNotExists = true

	// Components section
	config.Components.Temporal.Enabled = false
	config.Components.Temporal.Namespace = "temporal"
	config.Components.Temporal.ChartVersion = "0.62.0"
	config.Components.Temporal.NodePorts.Web = 30080
	config.Components.Temporal.NodePorts.Frontend = 30733

	config.Components.Redis.Enabled = false
	config.Components.Redis.NodePorts.Redis = 30679
	config.Components.Redis.ChartVersion = "17.3.7"
	config.Components.Redis.Auth.Enabled = false

	config.Components.Dapr.Enabled = false
	config.Components.Dapr.ChartVersion = "1.15.3"
	config.Components.Dapr.NodePorts.Dashboard = 30479
	config.Components.Dapr.LogLevel = "debug"
	config.Components.Dapr.Mtls.Enabled = false
	config.Components.Dapr.Ha.Enabled = false

	config.Components.OpenSearch.Enabled = true
	config.Components.OpenSearch.Namespace = "opensearch"
	config.Components.OpenSearch.Version = "2.17.1"
	config.Components.OpenSearch.NodePorts.Rest = 30920
	config.Components.OpenSearch.Security.Disabled = true
	config.Components.OpenSearch.IndexManagement.Enabled = true

	config.Components.OpenSearchDashboards.Enabled = true
	config.Components.OpenSearchDashboards.Namespace = "opensearch"
	config.Components.OpenSearchDashboards.Version = "2.17.1"
	config.Components.OpenSearchDashboards.NodePorts.Http = 30601

	config.Components.MySQL.Enabled = true
	config.Components.MySQL.Namespace = "mysql"
	config.Components.MySQL.ChartVersion = "9.4.6"
	config.Components.MySQL.Database = "mysql"
	config.Components.MySQL.NodePorts.MySQL = 30306
	config.Components.MySQL.Resources.CPU = "500m"
	config.Components.MySQL.Resources.Memory = "1Gi"
	config.Components.MySQL.Persistence.Enabled = false
	config.Components.MySQL.Persistence.Size = "8Gi"

	config.Components.TemporalWorkerOperator.Enabled = false
	config.Components.TemporalWorkerOperator.ChartVersion = "0.1.46-dev"
	config.Components.TemporalWorkerOperator.TemporalNamespace = "default"

	config.Components.IndicesOperator.Enabled = false
	config.Components.IndicesOperator.ChartVersion = "0.1.79-dev"

	config.Components.MetricsServer.Enabled = true
	config.Components.MetricsServer.ChartVersion = "3.10.0"

	// Images section
	config.Images.SkipPull = false
	config.Images.DockerHub.Username = ""
	config.Images.DockerHub.Password = ""
	config.Images.UseAwsEcr = true
	config.Images.AWS.Region = "eu-west-1"
	config.Images.AWS.EcrRegistry = "992979781608.dkr.ecr.eu-west-1.amazonaws.com"
	config.Images.AWS.Profile = ""
	config.Images.UseHarbor = true // Harbor is used for third-party components (MySQL, Redis, OpenSearch, etc.)
	config.Images.HarborRegistry = "harbor.shieldfis.com"

	// Secrets section
	config.Secrets.MySQL.Enabled = true
	config.Secrets.MySQL.Name = "kvv2-mysql"
	config.Secrets.MySQL.Namespace = "default"
	config.Secrets.MySQL.Username = "root"
	config.Secrets.MySQL.Password = "password"

	// Generate default port mappings based on enabled components
	config.Cluster.MapPorts = generateDefaultPortMappings(config)

	return config
}
