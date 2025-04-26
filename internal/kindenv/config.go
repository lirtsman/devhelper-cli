package kindenv

import (
	"errors"
	"fmt"
	"os"

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
			ContainerPort int    `yaml:"containerPort"`
			HostPort      int    `yaml:"hostPort"`
			Protocol      string `yaml:"protocol"`
		} `yaml:"mapPorts"`
	} `yaml:"cluster"`
	Components struct {
		Temporal struct {
			Enabled          bool   `yaml:"enabled"`
			Namespace        string `yaml:"namespace"`
			WebPort          int    `yaml:"webPort"`
			WebNodePort      int    `yaml:"webNodePort"`
			FrontendPort     int    `yaml:"frontendPort"`
			FrontendNodePort int    `yaml:"frontendNodePort"`
		} `yaml:"temporal"`
		Redis struct {
			Enabled  bool   `yaml:"enabled"`
			Port     int    `yaml:"port"`
			NodePort int    `yaml:"nodePort"`
			Image    string `yaml:"image"`
			Auth     struct {
				Enabled bool `yaml:"enabled"`
			} `yaml:"auth"`
		} `yaml:"redis"`
		CertManager struct {
			Enabled bool   `yaml:"enabled"`
			Version string `yaml:"version"`
		} `yaml:"certManager"`
		Dapr struct {
			Enabled  bool   `yaml:"enabled"`
			Version  string `yaml:"version"`
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
	config.Cluster.Name = "kind"
	config.Cluster.CreateIfNotExists = true

	if configPath == "" {
		return config, nil
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

	return config, nil
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
		if c.Components.Temporal.WebPort <= 0 {
			return errors.New("temporal web port must be greater than 0")
		}
		if c.Components.Temporal.FrontendPort <= 0 {
			return errors.New("temporal frontend port must be greater than 0")
		}
	}

	// Validate Redis configuration
	if c.Components.Redis.Enabled {
		if c.Components.Redis.Port <= 0 {
			return errors.New("redis port must be greater than 0")
		}
		if c.Components.Redis.Image == "" {
			c.Components.Redis.Image = "redis:6.2"
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
