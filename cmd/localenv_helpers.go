/*
Copyright © 2023 Shield

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"
)

// Check if a Temporal namespace exists
func isTemporalNamespaceExist(namespace string) bool {
	cmd := exec.Command("temporal", "operator", "namespace", "describe", namespace)
	return cmd.Run() == nil
}

// Helper functions for determining if components are required
func getDaprRequirement(configLoaded bool, configValue bool, skipFlag bool) bool {
	if configLoaded {
		return configValue
	}
	return !skipFlag
}

func getTemporalRequirement(configLoaded bool, configValue bool, skipFlag bool) bool {
	if configLoaded {
		return configValue
	}
	return !skipFlag
}

func getDaprDashboardRequirement(configLoaded bool, configValue bool, skipFlag bool) bool {
	if configLoaded {
		return configValue
	}
	return !skipFlag
}

func tryStartDashboard(command string, port int, logFile *os.File) bool {
	return tryStartDashboardWithTimeout(command, port, logFile, 3*time.Second)
}

func tryStartDashboardWithTimeout(command string, port int, logFile *os.File, timeout time.Duration) bool {
	dashboardCmd := exec.Command(command, "dashboard", "-p", strconv.Itoa(port), "--address", "0.0.0.0")

	// Redirect output to null device or log file
	if logFile == nil {
		devNull, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
		dashboardCmd.Stdout = devNull
		dashboardCmd.Stderr = devNull
	} else {
		dashboardCmd.Stdout = logFile
		dashboardCmd.Stderr = logFile
	}

	// Start the dashboard in a goroutine
	resultChan := make(chan error, 1)
	go func() {
		resultChan <- dashboardCmd.Run()
	}()

	// Wait a bit for the dashboard to start
	time.Sleep(timeout)

	// Check if the process exited quickly (indicating failure)
	select {
	case err := <-resultChan:
		// If we get here, the process exited before our timeout
		if err != nil {
			fmt.Printf("Dashboard failed to start: %v\n", err)
		}
		return false
	default:
		// No error yet, process is still running
		return true
	}
}
