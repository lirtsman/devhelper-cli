package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	dapr "github.com/dapr/go-sdk/client"
)

var (
	configMutex sync.RWMutex
	configStore = make(map[string]interface{}) // A map to store configurations by their key
)

// LoadConfig loads configuration items from a Dapr configuration store and stores them in the global config map.
func LoadConfig[T any](daprClient dapr.Client, configStoreName string, configKey string) (*T, error) {
	log.Printf("Attempting to load configuration for key '%s' from store '%s'", configKey, configStoreName)

	ctx := context.Background()

	// Retrieve the configuration stored in a single key
	item, err := daprClient.GetConfigurationItem(ctx, configStoreName, configKey)
	if err != nil {
		log.Printf("Error retrieving configuration for key '%s' from store '%s': %v", configKey, configStoreName, err)
		return nil, fmt.Errorf("failed to load configuration: %v", err)
	}

	log.Printf("Retrieved configuration for key '%s': %s", configKey, item.Value)

	// Unmarshal the configuration into the provided type
	var config T
	if err := json.Unmarshal([]byte(item.Value), &config); err != nil {
		log.Printf("Failed to unmarshal configuration for key '%s': %v", configKey, err)
		return nil, fmt.Errorf("failed to unmarshal configuration: %v", err)
	}

	// Store the config in the global config map
	configMutex.Lock()
	configStore[configKey] = config
	configMutex.Unlock()

	log.Printf("Configuration loaded and stored for key '%s'", configKey)
	return &config, nil
}

// SubscribeToConfigUpdates subscribes to updates for a configuration key in Dapr and updates the global config map.
func SubscribeToConfigUpdates[T any](daprClient dapr.Client, configStoreName string, configKey string) error {
	log.Printf("Subscribing to configuration updates for key '%s' from store '%s'", configKey, configStoreName)

	ctx := context.Background()

	// Subscribe to configuration updates
	_, err := daprClient.SubscribeConfigurationItems(ctx, configStoreName, []string{configKey}, func(id string, items map[string]*dapr.ConfigurationItem) {
		configMutex.Lock()
		defer configMutex.Unlock()

		for _, item := range items {
			log.Printf("Received configuration update for key '%s': %s", configKey, item.Value)

			// Unmarshal the configuration into the provided type
			var config T
			if err := json.Unmarshal([]byte(item.Value), &config); err != nil {
				log.Printf("Failed to unmarshal configuration update for key '%s': %v", configKey, err)
				continue
			}

			// Update the current configuration in the map
			configStore[configKey] = config
			log.Printf("Configuration updated for key '%s'", configKey)
		}
	})

	if err != nil {
		log.Printf("Failed to subscribe to configuration updates for key '%s' from store '%s': %v", configKey, configStoreName, err)
		return fmt.Errorf("failed to subscribe to configuration updates: %v", err)
	}

	log.Printf("Successfully subscribed to configuration updates for key '%s' from store '%s'", configKey, configStoreName)
	return nil
}

// GetConfig retrieves the current configuration for the specified key.
func GetConfig[T any](configKey string) *T {
	log.Printf("Attempting to retrieve configuration for key '%s'", configKey)

	configMutex.RLock()
	defer configMutex.RUnlock()

	if config, ok := configStore[configKey]; ok {
		typedConfig := config.(T)
		log.Printf("Configuration found for key '%s'", configKey)
		return &typedConfig
	}

	log.Printf("Configuration for key '%s' has not been loaded yet", configKey)
	return nil // Return nil if the config is not found
}
