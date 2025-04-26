package helm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

// Client defines the interface for Helm operations
type Client interface {
	// Repository operations
	AddRepository(name, url string) error
	UpdateRepositories() error

	// Chart operations
	Install(releaseName, chartName, namespace string, valuesYaml string, timeout time.Duration) error
	Upgrade(releaseName, chartName, namespace string, valuesYaml string, timeout time.Duration) error
	Uninstall(releaseName, namespace string) error

	// Status operations
	IsReleaseInstalled(releaseName, namespace string) (bool, error)
}

type clientImpl struct {
	settings *cli.EnvSettings
}

// NewClient creates a new Helm client
func NewClient() (Client, error) {
	settings := cli.New()
	return &clientImpl{
		settings: settings,
	}, nil
}

// addRepository adds a Helm repository
func (c *clientImpl) AddRepository(name, url string) error {
	// Check if repository file exists
	repoFile := c.settings.RepositoryConfig

	// Ensure the repository directory exists
	err := os.MkdirAll(filepath.Dir(repoFile), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create repository directory: %w", err)
	}

	// Load or create the repository file
	r, err := repo.LoadFile(repoFile)
	if err != nil {
		if os.IsNotExist(err) {
			r = repo.NewFile()
		} else {
			return fmt.Errorf("failed to load repository file: %w", err)
		}
	}

	// Check if repository already exists
	for _, entry := range r.Repositories {
		if entry.Name == name {
			// Repository already exists, check if URL matches
			if entry.URL == url {
				return nil
			}
			// Repository exists but URL doesn't match, update it
			entry.URL = url
			return r.WriteFile(repoFile, 0644)
		}
	}

	// Add the new repository entry
	newRepo := &repo.Entry{
		Name: name,
		URL:  url,
	}

	r.Add(newRepo)

	// Save the repository file
	if err := r.WriteFile(repoFile, 0644); err != nil {
		return fmt.Errorf("failed to write repository file: %w", err)
	}

	// Update the repository
	return c.updateRepository(name)
}

// updateRepository updates a specific Helm repository
func (c *clientImpl) updateRepository(name string) error {
	repoFile := c.settings.RepositoryConfig

	// Load the repository file
	r, err := repo.LoadFile(repoFile)
	if err != nil {
		return fmt.Errorf("failed to load repository file: %w", err)
	}

	// Find the repository entry
	var repoEntry *repo.Entry
	for _, entry := range r.Repositories {
		if entry.Name == name {
			repoEntry = entry
			break
		}
	}

	if repoEntry == nil {
		return fmt.Errorf("repository '%s' not found", name)
	}

	// Create a repository client
	chartRepo, err := repo.NewChartRepository(repoEntry, getter.All(c.settings))
	if err != nil {
		return fmt.Errorf("failed to create chart repository client: %w", err)
	}

	// Update the repository
	if _, err := chartRepo.DownloadIndexFile(); err != nil {
		return fmt.Errorf("failed to download index file for repository '%s': %w", name, err)
	}

	return nil
}

// UpdateRepositories updates all Helm repositories
func (c *clientImpl) UpdateRepositories() error {
	repoFile := c.settings.RepositoryConfig

	// Load the repository file
	r, err := repo.LoadFile(repoFile)
	if err != nil {
		return fmt.Errorf("failed to load repository file: %w", err)
	}

	// Update each repository
	for _, entry := range r.Repositories {
		if err := c.updateRepository(entry.Name); err != nil {
			return err
		}
	}

	return nil
}

// newActionConfig creates a new action configuration
func (c *clientImpl) newActionConfig(namespace string) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)

	// Initialize the action configuration
	if err := actionConfig.Init(c.settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), func(format string, args ...interface{}) {
		// This is a logger function that can be customized
		// fmt.Printf(format, args...)
	}); err != nil {
		return nil, fmt.Errorf("failed to initialize Helm action configuration: %w", err)
	}

	return actionConfig, nil
}

// parseChart parses a chart reference
func (c *clientImpl) parseChart(chartRef string) (*chart.Chart, error) {
	// Handle repository prefix
	if strings.Contains(chartRef, "/") {
		parts := strings.Split(chartRef, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid chart reference format: %s", chartRef)
		}

		repoName := parts[0]
		chartName := parts[1]

		// Get chart location from repository
		repoFile := c.settings.RepositoryConfig
		r, err := repo.LoadFile(repoFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load repository file: %w", err)
		}

		// Find the repository
		var repoEntry *repo.Entry
		for _, entry := range r.Repositories {
			if entry.Name == repoName {
				repoEntry = entry
				break
			}
		}

		if repoEntry == nil {
			return nil, fmt.Errorf("repository '%s' not found", repoName)
		}

		// Download the chart
		chartPath, err := c.downloadChart(repoEntry.URL, chartName)
		if err != nil {
			return nil, err
		}

		return loader.Load(chartPath)
	}

	// Handle local chart path
	return loader.Load(chartRef)
}

// downloadChart downloads a chart from a repository
func (c *clientImpl) downloadChart(repoURL, chartName string) (string, error) {
	// Download chart implementation
	// For simplicity, assuming charts are installed from a repository using the repository name
	// Full implementation would require downloading the chart from the repository URL
	return "", fmt.Errorf("downloading charts directly is not implemented, please use repository/chartname format")
}

// parseValues parses values from YAML
func (c *clientImpl) parseValues(valuesYaml string) (map[string]interface{}, error) {
	if valuesYaml == "" {
		return nil, nil
	}

	providers := getter.All(c.settings)
	options := values.Options{
		StringValues: []string{},
		ValueFiles:   []string{},
		Values:       []string{valuesYaml},
	}

	return options.MergeValues(providers)
}

// Install installs a Helm chart
func (c *clientImpl) Install(releaseName, chartName, namespace string, valuesYaml string, timeout time.Duration) error {
	// Check if release is already installed
	installed, err := c.IsReleaseInstalled(releaseName, namespace)
	if err != nil {
		return err
	}

	if installed {
		return c.Upgrade(releaseName, chartName, namespace, valuesYaml, timeout)
	}

	// Get action configuration
	actionConfig, err := c.newActionConfig(namespace)
	if err != nil {
		return err
	}

	// Create install action
	client := action.NewInstall(actionConfig)
	client.ReleaseName = releaseName
	client.Namespace = namespace
	client.CreateNamespace = true

	if timeout > 0 {
		client.Timeout = timeout
	}

	// Parse chart
	chart, err := c.parseChart(chartName)
	if err != nil {
		return err
	}

	// Parse values
	values, err := c.parseValues(valuesYaml)
	if err != nil {
		return err
	}

	// Install the chart
	_, err = client.Run(chart, values)
	if err != nil {
		return fmt.Errorf("failed to install chart '%s': %w", chartName, err)
	}

	return nil
}

// Upgrade upgrades a Helm chart
func (c *clientImpl) Upgrade(releaseName, chartName, namespace string, valuesYaml string, timeout time.Duration) error {
	// Get action configuration
	actionConfig, err := c.newActionConfig(namespace)
	if err != nil {
		return err
	}

	// Create upgrade action
	client := action.NewUpgrade(actionConfig)
	client.Namespace = namespace

	if timeout > 0 {
		client.Timeout = timeout
	}

	// Parse chart
	chart, err := c.parseChart(chartName)
	if err != nil {
		return err
	}

	// Parse values
	values, err := c.parseValues(valuesYaml)
	if err != nil {
		return err
	}

	// Upgrade the chart
	_, err = client.Run(releaseName, chart, values)
	if err != nil {
		return fmt.Errorf("failed to upgrade chart '%s': %w", chartName, err)
	}

	return nil
}

// Uninstall uninstalls a Helm release
func (c *clientImpl) Uninstall(releaseName, namespace string) error {
	// Check if release exists
	installed, err := c.IsReleaseInstalled(releaseName, namespace)
	if err != nil {
		return err
	}

	if !installed {
		return nil // Release not installed
	}

	// Get action configuration
	actionConfig, err := c.newActionConfig(namespace)
	if err != nil {
		return err
	}

	// Create uninstall action
	client := action.NewUninstall(actionConfig)

	// Uninstall the release
	_, err = client.Run(releaseName)
	if err != nil {
		return fmt.Errorf("failed to uninstall release '%s': %w", releaseName, err)
	}

	return nil
}

// IsReleaseInstalled checks if a Helm release is installed
func (c *clientImpl) IsReleaseInstalled(releaseName, namespace string) (bool, error) {
	// Get action configuration
	actionConfig, err := c.newActionConfig(namespace)
	if err != nil {
		return false, err
	}

	// Create list action
	client := action.NewList(actionConfig)
	client.Filter = releaseName
	client.Namespace = namespace

	// Get releases
	releases, err := client.Run()
	if err != nil {
		return false, fmt.Errorf("failed to check if release '%s' is installed: %w", releaseName, err)
	}

	// Check if release exists
	for _, release := range releases {
		if release.Name == releaseName {
			return true, nil
		}
	}

	return false, nil
}
