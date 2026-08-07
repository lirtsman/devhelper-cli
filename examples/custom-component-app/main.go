package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// AppConfig represents the application configuration
type AppConfig struct {
	Server struct {
		Port int `yaml:"port" json:"port"`
	} `yaml:"server" json:"server"`
	Database struct {
		Host string `yaml:"host" json:"host"`
		Port int    `yaml:"port" json:"port"`
	} `yaml:"database" json:"database"`
}

func main() {
	// Read configuration from environment or config file
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "/config/application.yaml"
	}

	// For simplicity, we'll use environment variables
	// In a real app, you'd parse the YAML file
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	log.Printf("Starting server on port %s", port)
	log.Printf("Database host: %s", dbHost)
	log.Printf("Config file location: %s", configFile)

	// Check if config file exists
	if _, err := os.Stat(configFile); err == nil {
		log.Printf("Config file found at: %s", configFile)
	} else {
		log.Printf("Config file not found at: %s (using environment variables)", configFile)
	}

	// Simple HTTP server
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/config", handleConfig)
	http.HandleFunc("/env", handleEnv)

	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Server listening on http://0.0.0.0:%s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"message": "Hello from custom component app!",
		"time":    time.Now().Format(time.RFC3339),
		"version": "1.0.0",
	}
	json.NewEncoder(w).Encode(response)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]interface{}{
		"status":  "healthy",
		"service": "custom-component-app",
		"time":    time.Now().Format(time.RFC3339),
	}
	json.NewEncoder(w).Encode(response)
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "/config/application.yaml"
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"config_file": configFile,
		"exists":      false,
	}

	if file, err := os.Open(configFile); err == nil {
		defer file.Close()
		response["exists"] = true
		if contents, err := io.ReadAll(file); err == nil {
			response["contents"] = string(contents)
		}
	}

	json.NewEncoder(w).Encode(response)
}

func handleEnv(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Return relevant environment variables (excluding secrets)
	env := make(map[string]string)
	for _, key := range []string{
		"SERVER_PORT",
		"DB_HOST",
		"DB_PORT",
		"APP_ENV",
		"CONFIG_FILE",
	} {
		if val := os.Getenv(key); val != "" {
			env[key] = val
		}
	}

	response := map[string]interface{}{
		"environment_variables": env,
	}
	json.NewEncoder(w).Encode(response)
}
