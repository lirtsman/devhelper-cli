package kindenv

// volume.go provides functionality for generating Kubernetes volume and volumeMount
// specifications for mounting ConfigMaps as volumes in pods.

// generateVolumes generates volume specifications for a custom component's config files
func generateVolumes(component *CustomComponent) ([]VolumeSpec, error) {
	if len(component.ConfigFiles) == 0 {
		return nil, nil
	}

	// Generate a single volume for all config files (ConfigMap contains all files)
	configMapName := component.Name + "-config"
	volume := VolumeSpec{
		Name: configMapName + "-volume",
		ConfigMap: &ConfigMapVolumeSource{
			Name:        configMapName,
			DefaultMode: 0644, // Read-only, 0644 permissions
		},
	}

	return []VolumeSpec{volume}, nil
}

// generateVolumeMounts generates volumeMount specifications for mounting config files
func generateVolumeMounts(component *CustomComponent) ([]VolumeMountSpec, error) {
	if len(component.ConfigFiles) == 0 {
		return nil, nil
	}

	volumeName := component.Name + "-config-volume"
	var volumeMounts []VolumeMountSpec

	// Generate a volumeMount for each config file using subPath
	for _, configFile := range component.ConfigFiles {
		volumeMount := VolumeMountSpec{
			Name:      volumeName,
			MountPath: configFile.Path,
			SubPath:   configFile.Name, // Mount individual file using subPath
			ReadOnly:  true,            // ConfigMaps are read-only
		}
		volumeMounts = append(volumeMounts, volumeMount)
	}

	return volumeMounts, nil
}
