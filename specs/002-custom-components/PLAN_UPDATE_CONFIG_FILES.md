# Plan Update: Configuration File Mounting

**Date**: 2026-01-30  
**Feature**: Custom Components for KindEnv  
**Update Type**: Feature Enhancement (Config File Mounting)

## Summary of Changes

The specification has been updated to include custom configuration file mounting capabilities based on clarifications from the user. This enhancement allows developers to mount inline configuration files into custom component containers without rebuilding images.

## What Was Added

### 1. New User Story (P6)

**User Story 6 - Mount Custom Configuration Files**
- Mount custom config files with inline YAML/JSON/XML contents
- Multiple config files per component supported
- ConfigMap auto-generation from inline contents
- Read-only mounts at specified paths
- Override behavior with warnings for existing files

### 2. New Functional Requirements (FR-021 through FR-029)

- **FR-021**: Support mounting custom configuration files with inline contents
- **FR-022**: Automatically create Kubernetes ConfigMaps from inline config contents
- **FR-023**: Support multiple config files per component as an array
- **FR-024**: Mount config files with read-only permissions (0644)
- **FR-025**: Log warnings when mounted files override image files
- **FR-026**: Validate config file specifications (name, path, contents required)
- **FR-027**: Preserve formatting and special characters in contents
- **FR-028**: Detect and report duplicate mount path errors
- **FR-029**: Update ConfigMaps and restart pods when contents change

### 3. New Entity: ConfigFile

**Attributes**:
- `name` (string): Filename (used as ConfigMap key)
- `path` (string): Absolute mount path in container
- `contents` (string): Inline file contents (YAML/JSON/XML/etc)

**Validation Rules**:
- Name required, no directory separators
- Path must be absolute (start with /)
- Contents cannot be empty
- Individual file limit: 1MB
- Total limit per component: 1MB
- No duplicate filenames or mount paths

**Relationships**:
- Belongs to CustomComponent
- Generates Kubernetes ConfigMap
- Mounted as read-only volume

### 4. Updated CustomComponent Entity

Added `configFiles []ConfigFile` field to CustomComponent struct.

### 5. Technical Implementation Decisions

Based on clarification session 2026-01-30:

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Config source | Inline contents with ConfigMap generation | Self-contained, Kubernetes-native, simplest |
| Mount path conflicts | Override with warning | Matches K8s behavior, logs warning for visibility |
| Sensitive data | ConfigMaps only (no sensitive flag) | Secrets already supported via env secretKeyRef |
| File permissions | Read-only, 0644 | Security best practice, matches ConfigMap defaults |
| Multiple files | Yes, array of configs | Real-world need, consistent with env/ports pattern |

### 6. ConfigMap Generation Pattern

```yaml
customComponents:
  - name: my-app
    image: myregistry/my-app:latest
    configFiles:
      - name: application.yaml
        path: /config/application.yaml
        contents: |
          server:
            port: 8080
          database:
            host: mysql
```

Generates:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-app-config
  namespace: default
data:
  application.yaml: |
    server:
      port: 8080
    database:
      host: mysql
```

And mounts using:
- Volume: ConfigMap volume with defaultMode 0644
- VolumeMount: subPath mount for each file at specified path

## Files Updated

1. **spec.md**:
   - Added Clarifications section with Q&A
   - Added User Story 6 (config file mounting)
   - Added 5 new edge cases for config files
   - Added FR-021 through FR-029
   - Added ConfigFile entity
   - Updated CustomComponent entity to include configFiles
   - Added SC-011 through SC-013 success criteria

2. **research.md**:
   - Added section 11: Configuration File Mounting
   - Documented ConfigMap vs Secrets decision
   - Documented mount path conflict handling
   - Documented ConfigMap size limits
   - Documented update strategy (delete + recreate)

3. **data-model.md**:
   - Added ConfigFile entity with full struct definition
   - Added validation functions for ConfigFile
   - Added ConfigMap generation logic
   - Added volume mount generation logic
   - Updated CustomComponent to include configFiles field
   - Updated validation to check for duplicate paths/names and size limits
   - Updated entity relationships diagram

4. **contracts/custom-component-schema.yaml**:
   - Added configFiles array schema
   - Added name, path, contents field definitions
   - Added validation patterns and constraints
   - Updated complete example to include config files

5. **contracts/custom-component-api-interface.go**:
   - Added ConfigMapManager interface
   - Added ConfigMapSpec, VolumeSpec, VolumeMountSpec types
   - Added ConfigMapVolumeSource type
   - Updated DeploymentSpec to include Volumes and VolumeMounts

6. **contracts/deployment-template.yaml**:
   - Added ConfigMap resource template
   - Added volumes section with ConfigMap volume
   - Added volumeMounts section with subPath mounting
   - Updated concrete Spring Boot example with ConfigMap and mounts

7. **quickstart.md**:
   - Added "Mounting Configuration Files" section to TOC
   - Added comprehensive config file mounting examples
   - Added multiple files example
   - Added config + secrets combination example
   - Added update workflow instructions
   - Added config file best practices
   - Added config file troubleshooting section
   - Updated configuration reference
   - Updated FAQ with config file questions

## Implementation Impact

### New Components Required

1. **ConfigMap Management**:
   - ConfigMap creation/update/delete logic
   - ConfigMap existence checks
   - ConfigMap content validation

2. **Volume Generation**:
   - Generate volume specs from configFiles
   - Generate volumeMount specs with subPath
   - Handle mount path conflict detection

3. **Validation**:
   - Config file field validation (name, path, contents)
   - Duplicate path/name detection
   - Size limit validation (1MB total)
   - Path format validation (must be absolute)

4. **Update Handling**:
   - Detect ConfigMap content changes
   - Delete old ConfigMap
   - Create new ConfigMap
   - Trigger pod restart (rollout restart)

### Testing Additions

- Unit tests for ConfigFile validation
- Unit tests for ConfigMap generation
- Unit tests for volume/volumeMount generation
- Integration tests for ConfigMap creation and mounting
- Integration tests for config file updates
- Table-driven tests for various config file scenarios

### Documentation Additions

- Config file mounting quickstart guide
- ConfigMap troubleshooting section
- Config file best practices
- Security guidance (don't put secrets in configs)

## Success Criteria Updates

Added 3 new measurable outcomes:

- **SC-011**: Developers can mount custom configuration files and verify contents in under 1 minute
- **SC-012**: Config file changes reflected in pods within 30 seconds of cluster restart
- **SC-013**: Support at least 10 config files per component without performance degradation

## Edge Cases Identified

5 new edge cases for config file handling:

1. Config file mount path conflicts with existing container directory
2. Invalid YAML/JSON syntax in config file contents
3. Multiple config files mounted to the same path
4. Very large config file contents (>1MB)
5. Config file names with special characters or multi-level paths

## Next Steps

The plan update is complete. All design artifacts have been updated with config file mounting:

- ✅ Specification updated with clarifications and new requirements
- ✅ Research document updated with ConfigMap decision
- ✅ Data model updated with ConfigFile entity
- ✅ Contracts updated with schema and interface definitions
- ✅ Deployment templates updated with ConfigMap and volume examples
- ✅ Quickstart guide updated with comprehensive config file examples

**Ready for**: `/speckit.tasks` to break down implementation including config file mounting features.

**Estimated Additional Scope**:
- New files: +2-3 files (configmap.go, configmap_test.go, volume.go)
- Modified files: +3-4 files (config.go, validation.go, deployment.go)
- Test files: +2-3 files (configmap_test.go, volume_test.go, integration tests)
- Additional LOC: ~500-700 lines (config file handling, validation, ConfigMap generation)

Total feature scope is now **Large** (was Medium-Large before config files).
