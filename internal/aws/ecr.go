package aws

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

// ECRClient defines the interface for AWS ECR operations
type ECRClient interface {
	// GetAuthorizationToken gets an ECR authorization token
	GetAuthorizationToken() (string, string, string, error)

	// CreateDockerConfig creates a Docker config.json for ECR authentication
	CreateDockerConfig(registry, username, password string) (string, error)
}

type ecrClientImpl struct {
	client *ecr.Client
	region string
}

// NewECRClient creates a new AWS ECR client
func NewECRClient(region string) (ECRClient, error) {
	// Create AWS SDK config
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create ECR client
	client := ecr.NewFromConfig(cfg)

	return &ecrClientImpl{
		client: client,
		region: region,
	}, nil
}

// GetAuthorizationToken gets an ECR authorization token
func (e *ecrClientImpl) GetAuthorizationToken() (string, string, string, error) {
	// Get authorization token
	result, err := e.client.GetAuthorizationToken(context.Background(), &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get ECR authorization token: %w", err)
	}

	if len(result.AuthorizationData) == 0 {
		return "", "", "", fmt.Errorf("no ECR authorization data returned")
	}

	// Get token from response
	authData := result.AuthorizationData[0]
	token := *authData.AuthorizationToken
	endpoint := *authData.ProxyEndpoint

	// Decode token
	decodedToken, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to decode ECR token: %w", err)
	}

	// Token is in format "username:password"
	parts := strings.SplitN(string(decodedToken), ":", 2)
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("invalid ECR token format")
	}

	return endpoint, parts[0], parts[1], nil
}

// DockerConfigJSON represents the Docker config.json structure
type DockerConfigJSON struct {
	Auths map[string]DockerConfigAuth `json:"auths"`
}

// DockerConfigAuth represents the auth field in Docker config.json
type DockerConfigAuth struct {
	Auth string `json:"auth"`
}

// CreateDockerConfig creates a Docker config.json for ECR authentication
func (e *ecrClientImpl) CreateDockerConfig(registry, username, password string) (string, error) {
	// Create auth string (base64 encoded username:password)
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))

	// Create config structure
	config := DockerConfigJSON{
		Auths: map[string]DockerConfigAuth{
			registry: {
				Auth: auth,
			},
		},
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Docker config: %w", err)
	}

	return string(jsonBytes), nil
}
