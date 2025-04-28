package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ShieldFC-RD/devhelper-cli/examples/temporal-worker/utils"
	dapr "github.com/dapr/go-sdk/client"
	"github.com/go-redis/redis/v8"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// Config represents the worker configuration structure
type Config struct {
	Environment string `json:"environment"`
	LogLevel    string `json:"logLevel"`
	Features    struct {
		EnableLogging bool `json:"enableLogging"`
		EnableMetrics bool `json:"enableMetrics"`
		EnableRetries bool `json:"enableRetries"`
	} `json:"features"`
	MaxConcurrentTasks int `json:"maxConcurrentTasks"`
	BatchSize          int `json:"batchSize"`
	RetryCount         int `json:"retryCount"`
}

// loadConfigWithDapr attempts to load configuration from Dapr configuration store
func loadConfigWithDapr(workerName string) (*Config, error) {
	// Default configuration
	config := &Config{
		Environment:        "dev",
		LogLevel:           "info",
		MaxConcurrentTasks: 5,
		BatchSize:          100,
		RetryCount:         3,
	}
	config.Features.EnableLogging = true
	config.Features.EnableMetrics = true
	config.Features.EnableRetries = true

	// Try to initialize Dapr client
	daprClient, err := dapr.NewClient()
	if err != nil {
		fmt.Printf("Warning: Failed to create Dapr client: %v\n", err)
		fmt.Println("Using default configuration...")
		return config, nil
	}
	defer daprClient.Close()

	// Configuration store name
	configStoreName := getEnv("CONFIG_STORE_NAME", "configstore")

	// Use the worker name as the configuration key
	configKey := workerName

	// Try to load configuration from Dapr
	loadedConfig, err := utils.LoadConfig[Config](daprClient, configStoreName, configKey)
	if err != nil {
		fmt.Printf("Warning: Failed to load configuration from Dapr: %v\n", err)
		fmt.Println("Using default configuration...")
		return config, nil
	}

	// Subscribe to configuration updates
	err = utils.SubscribeToConfigUpdates[Config](daprClient, configStoreName, configKey)
	if err != nil {
		fmt.Printf("Warning: Failed to subscribe to configuration updates: %v\n", err)
	} else {
		fmt.Println("Subscribed to configuration updates")
	}

	fmt.Printf("Loaded configuration from Dapr for worker: %s\n", workerName)
	return loadedConfig, nil
}

// loadConfig tries to load from Redis (legacy) or falls back to Dapr
func loadConfig(workerName string) (*Config, error) {
	// Check if we should use Dapr for configuration
	useDapr := getEnv("USE_DAPR_CONFIG", "false")
	if strings.ToLower(useDapr) == "true" {
		return loadConfigWithDapr(workerName)
	}

	// Default configuration
	config := &Config{
		Environment:        "dev",
		LogLevel:           "info",
		MaxConcurrentTasks: 5,
		BatchSize:          100,
		RetryCount:         3,
	}
	config.Features.EnableLogging = true
	config.Features.EnableMetrics = true
	config.Features.EnableRetries = true

	// Try to get configuration from Redis (legacy)
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	defer redisClient.Close()

	// Check if Redis is available
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("Warning: Redis connection failed: %v\n", err)
		fmt.Println("Using default configuration...")
		return config, nil
	}

	// Use the expected key format: configuration||{worker-name}
	redisKey := fmt.Sprintf("configuration||%s", workerName)
	configJSON, err := redisClient.Get(ctx, redisKey).Result()
	if err != nil {
		if err == redis.Nil {
			fmt.Printf("Configuration not found in Redis for key: %s\n", redisKey)
			fmt.Println("Using default configuration...")
			return config, nil
		}
		return nil, fmt.Errorf("failed to get config from Redis: %w", err)
	}

	// Parse configuration JSON
	if err := json.Unmarshal([]byte(configJSON), config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	fmt.Printf("Loaded configuration from Redis for worker: %s\n", workerName)
	return config, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func main() {
	fmt.Println("Starting Temporal worker...")

	// Worker name should match the one in your tw.yaml metadata.name
	// This is used to fetch the configuration
	workerName := getEnv("WORKER_NAME", "temporal-worker")

	// Load configuration (from Dapr, Redis, or defaults)
	config, err := loadConfig(workerName)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Print loaded configuration
	configJSON, _ := json.MarshalIndent(config, "", "  ")
	fmt.Printf("Worker configuration:\n%s\n", string(configJSON))

	// Set up Temporal client
	temporalOptions := client.Options{
		HostPort: getEnv("TEMPORAL_HOST_PORT", "localhost:7233"),
	}

	// Extract worker name to get the task queue name
	// Assuming format: "temporal-{queue}-worker"
	taskQueue := "sample"
	parts := strings.Split(workerName, "-")
	if len(parts) > 1 && len(parts) < 4 {
		// Extract the middle part if it's in the format "temporal-{queue}-worker"
		taskQueue = parts[1]
	}

	fmt.Printf("Using task queue: %s\n", taskQueue)

	// Create Temporal client
	temporalClient, err := client.Dial(temporalOptions)
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer temporalClient.Close()

	// Create worker
	w := worker.New(temporalClient, taskQueue, worker.Options{
		MaxConcurrentActivityExecutionSize: config.MaxConcurrentTasks,
	})

	// Register workflow and activity functions
	w.RegisterWorkflow(SampleWorkflow)
	w.RegisterActivity(SampleActivity)

	// Start worker
	err = w.Start()
	if err != nil {
		log.Fatalf("Failed to start worker: %v", err)
	}

	// Handle graceful shutdown
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)
	<-shutdownChan

	fmt.Println("Shutting down worker...")
	w.Stop()
	fmt.Println("Worker stopped")
}
