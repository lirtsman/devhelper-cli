# Feature Specification: Custom Components for KindEnv

**Feature Branch**: `002-custom-components`  
**Created**: 2026-01-30  
**Status**: Draft  
**Input**: User description: "Add ability for developers to deploy custom services with configurable images, commands, and environment variables"

## Clarifications

### Session 2026-01-30

- Q: How should configuration files be provided to custom components? → A: Inline contents with automatic ConfigMap generation
- Q: When a custom config file is mounted at a path that already exists in the container, what should happen? → A: Override with warning
- Q: Should the system support marking configs as sensitive for encryption? → A: No, use ConfigMaps only (sensitive data uses existing secretKeyRef pattern)
- Q: What file permissions should be set on mounted config files? → A: Read-only, default user (0644)
- Q: Can a custom component mount multiple configuration files? → A: Yes, array of configs

## Configuration Overview

**Minimal Required Configuration**:
```yaml
customComponents:
  - name: my-app        # Required
    image: nginx:latest # Required
```

**Optional Fields with Automatic Defaults**:
- `enabled: true` (can set to false to disable)
- `namespace: default` (can specify different namespace)
- `replicas: 1` (can increase for multiple instances)
- `resources:` CPU/memory limits (defaults: 100m/128Mi requests, 500m/512Mi limits)

**Optional Fields (No Defaults)**:
- `command:` Override container entrypoint
- `args:` Container arguments
- `env:` Environment variables (direct values or secret references)
- `ports:` Port mappings for external access
- `configFiles:` Custom configuration files to mount
- `labels:` Custom Kubernetes labels (standard labels auto-generated)
- `annotations:` Custom Kubernetes annotations

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Deploy Custom Application with Basic Configuration (Priority: P1)

As a developer, I want to deploy my custom Spring Boot application into the kindenv cluster by specifying just the image and environment variables, so that I can test my application alongside standard components like MySQL and OpenSearch.

**Why this priority**: This is the core functionality that enables developers to add any custom service to their local development environment. Without this, developers cannot test custom applications in the cluster.

**Independent Test**: Can be fully tested by adding a simple custom component configuration (image + basic env vars) to kindenv.yaml, running `kindenv start`, and verifying the custom service is deployed and running in the cluster using `kubectl get pods`.

**Acceptance Scenarios**:

1. **Given** a kindenv.yaml with no custom components, **When** I add a custom component with image specification, **Then** the component is deployed to the cluster with the specified image
2. **Given** a custom component configuration with environment variables, **When** I start the kindenv, **Then** the deployed pod contains all specified environment variables
3. **Given** a custom component with direct value environment variables, **When** the service starts, **Then** the environment variables are available to the application with correct values

---

### User Story 2 - Connect Custom Application to Existing Services via Secrets (Priority: P2)

As a developer, I want to configure my custom application to connect to MySQL and OpenSearch using secrets from the cluster, so that my application can access databases and search engines without hardcoding credentials.

**Why this priority**: Most real-world applications need to connect to databases and other services. Using secrets ensures proper credential management and mirrors production patterns.

**Independent Test**: Can be tested by configuring a custom component with secretKeyRef environment variables, deploying it, and verifying the pod receives the secret values by inspecting the container environment.

**Acceptance Scenarios**:

1. **Given** MySQL secrets exist in the cluster, **When** I configure a custom component with environment variables from secretKeyRef, **Then** the deployed pod receives the secret values
2. **Given** multiple secret references in environment variables, **When** the component is deployed, **Then** all secret values are correctly injected into the container
3. **Given** a non-existent secret is referenced, **When** deployment is attempted, **Then** appropriate error message indicates the missing secret

---

### User Story 3 - Configure Custom Command and Arguments (Priority: P3)

As a developer, I want to override the default container command and provide custom arguments, so that I can run my application with specific startup parameters or wrapper scripts.

**Why this priority**: While less common than basic deployment, custom commands enable advanced scenarios like running migration scripts, custom entrypoints, or development-specific configurations.

**Independent Test**: Can be tested by configuring a custom component with a command array, deploying it, and verifying via `kubectl describe pod` that the container runs with the specified command.

**Acceptance Scenarios**:

1. **Given** a custom component with command specified, **When** the component is deployed, **Then** the container runs with the custom command instead of the image default
2. **Given** a custom component with both command and args, **When** deployed, **Then** the container executes the command with the provided arguments
3. **Given** a component with args but no command, **When** deployed, **Then** the container uses the default image command with the custom args

---

### User Story 4 - Expose Custom Service Ports (Priority: P4)

As a developer, I want to expose specific ports from my custom service to the host machine, so that I can access my application's web interface or API from my local browser or testing tools.

**Why this priority**: Port exposure enables interaction with custom services, but basic deployment and configuration are more critical for initial MVP.

**Independent Test**: Can be tested by configuring a custom component with port mappings, starting kindenv, and accessing the service from the host machine via localhost and the mapped port.

**Acceptance Scenarios**:

1. **Given** a custom component with containerPort and hostPort mapping, **When** kindenv starts, **Then** the service is accessible from the host machine on the specified port
2. **Given** multiple port mappings for a single component, **When** deployed, **Then** all ports are correctly mapped and accessible
3. **Given** a port conflict with existing components, **When** configuration is validated, **Then** an error indicates the conflicting port

---

### User Story 5 - Configure Resource Limits for Custom Components (Priority: P5)

As a developer, I want to set CPU and memory limits for my custom components, so that I can control resource consumption and prevent any single service from overwhelming my local machine.

**Why this priority**: Resource management is important for stability but can be added after basic deployment works. Default resource limits can be applied initially.

**Independent Test**: Can be tested by configuring a custom component with resource requests and limits, deploying it, and verifying via `kubectl describe pod` that the resource constraints are applied.

**Acceptance Scenarios**:

1. **Given** a custom component with CPU and memory limits specified, **When** deployed, **Then** the pod respects the specified resource constraints
2. **Given** no resource limits specified, **When** deployed, **Then** default resource limits are applied
3. **Given** resource limits exceed node capacity, **When** deployment is attempted, **Then** the pod remains in pending state with appropriate error message

---

### User Story 6 - Mount Custom Configuration Files (Priority: P6)

As a developer, I want to mount custom configuration files into my application container, so that I can provide application-specific settings (application.yaml, logging.xml, etc.) without rebuilding the container image.

**Why this priority**: Configuration file mounting is a common pattern for customizing application behavior in development environments. While less critical than basic deployment, it significantly improves developer experience by avoiding image rebuilds for config changes.

**Independent Test**: Can be tested by configuring a custom component with inline configuration file contents, deploying it, and verifying the file exists at the specified mount path with correct contents using `kubectl exec`.

**Acceptance Scenarios**:

1. **Given** a custom component with inline config file specified, **When** deployed, **Then** the config file is mounted at the specified path with the provided contents
2. **Given** multiple config files specified for a component, **When** deployed, **Then** all config files are mounted at their respective paths
3. **Given** a config file mounted at a path that exists in the container image, **When** deployed, **Then** the mounted file overrides the image file and a warning is logged
4. **Given** a config file with multiline YAML contents, **When** deployed, **Then** the file preserves formatting and special characters
5. **Given** a config file is updated in kindenv.yaml, **When** cluster is restarted, **Then** the ConfigMap is updated and pod receives new contents

---

### Edge Cases

- What happens when a custom component image cannot be pulled (invalid registry, missing credentials)?
- How does the system handle when multiple custom components have the same name?
- What happens when a secret reference points to a non-existent secret?
- How does the system handle when custom component ports conflict with predefined component ports?
- What happens when environment variable values contain special characters or multiline strings?
- How does the system validate resource limit formats (CPU, memory)?
- What happens when a custom component namespace doesn't exist?
- What happens when a config file mount path conflicts with an existing container directory (not just a file)?
- How does the system handle config files with invalid YAML/JSON syntax in contents?
- What happens when multiple config files are mounted to the same path?
- How does the system handle very large config file contents (e.g., >1MB)?
- What happens when config file names contain special characters or paths with multiple directories?

## Requirements *(mandatory)*

### Functional Requirements

#### Required Fields
- **FR-001**: System MUST require name and image fields for custom components
- **FR-002**: System MUST support specifying container image in format [registry/]repository[:tag]
- **FR-007**: System MUST validate custom component configuration before deployment
- **FR-008**: System MUST create Kubernetes deployments for enabled custom components

#### Optional Fields with Defaults
- **FR-010**: System MUST support optional namespace specification (default: "default")
- **FR-011**: System MUST support optional replica count (default: 1)
- **FR-012**: System MUST apply default resource limits if not specified (default: 100m CPU, 128Mi memory requests; 500m CPU, 512Mi memory limits)
- **FR-014**: System MUST support optional enabled flag (default: true)

#### Optional Fields (No Defaults)
- **FR-003**: System MUST support optional environment variables with direct string values
- **FR-004**: System MUST support optional environment variables with secretKeyRef (secret name and key)
- **FR-005**: System MUST support optional command override as an array of strings
- **FR-006**: System MUST support optional args as an array of strings
- **FR-009**: System MUST support optional port mappings from container ports to host ports
- **FR-015**: System MUST support optional labels and annotations
- **FR-021**: System MUST support optional custom configuration files with inline contents

#### Behavior Requirements
- **FR-013**: System MUST provide meaningful error messages when custom component configuration is invalid
#### Deployment and Lifecycle Requirements
- **FR-016**: Custom components MUST be deployed after infrastructure components (MySQL, OpenSearch) are ready
- **FR-017**: System MUST support checking readiness and liveness of custom components (Future enhancement - not included in initial MVP)
- **FR-018**: System MUST display status of custom components in `kindenv status` command
- **FR-019**: System MUST clean up custom component deployments when `kindenv stop` is executed
- **FR-020**: System MUST support multiple custom components in a single kindenv.yaml file

#### Configuration File Mounting Requirements
- **FR-022**: System MUST automatically create Kubernetes ConfigMaps from inline config file contents
- **FR-023**: System MUST support multiple config files per custom component as an array
- **FR-024**: System MUST mount config files with read-only permissions (0644) by default
- **FR-025**: System MUST log a warning when a mounted config file overrides an existing file in the container image
- **FR-026**: System MUST validate config file specifications (name, path, contents are required)
- **FR-027**: System MUST preserve formatting and special characters in config file contents
- **FR-028**: System MUST detect and report errors when multiple config files specify the same mount path
- **FR-029**: System MUST update ConfigMaps and restart pods when config file contents change in kindenv.yaml

### Key Entities

- **CustomComponent**: Represents a user-defined service to be deployed in the cluster
  - Attributes: name, enabled, image, command, args, env, ports, resources, namespace, replicas, labels, annotations, configFiles
  - Relationships: Can reference secrets from the cluster, can be deployed to specific namespace, can have multiple config files

- **EnvVar (Environment Variable)**: Represents an environment variable for the custom component
  - Attributes: name, value (direct), valueFrom (secret reference)
  - Relationships: Belongs to a CustomComponent, may reference a Secret

- **PortMapping**: Represents port exposure configuration
  - Attributes: containerPort, hostPort, protocol, nodePort
  - Relationships: Belongs to a CustomComponent

- **ResourceRequirements**: Represents CPU and memory constraints
  - Attributes: cpu (request/limit), memory (request/limit)
  - Relationships: Belongs to a CustomComponent

- **ConfigFile**: Represents a custom configuration file to be mounted in the container
  - Attributes: name (filename), path (mount path in container), contents (inline YAML/text)
  - Relationships: Belongs to a CustomComponent, generates a Kubernetes ConfigMap
  - Constraints: Mount path must be unique per component, contents are stored as ConfigMap with read-only permissions (0644)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Developers can deploy a custom component with basic image configuration in under 2 minutes by editing kindenv.yaml
- **SC-002**: Custom components can successfully connect to MySQL and OpenSearch using secret references
- **SC-003**: System correctly validates configuration and provides actionable error messages for 100% of common configuration mistakes (invalid image format, missing secrets, port conflicts)
- **SC-004**: Custom components are deployed within 30 seconds of running `kindenv start` on a typical development machine
- **SC-005**: Developers can access custom services from their host machine via mapped ports within 5 seconds of deployment completion
- **SC-006**: The `kindenv status` command displays accurate status (running/pending/error) for all custom components
- **SC-007**: 95% of developers can successfully deploy their first custom component without consulting documentation beyond a single example
- **SC-008**: Custom component deployments consume no more than 10% additional resources compared to the base kindenv setup
- **SC-009**: Custom components are cleanly removed with zero leftover resources when `kindenv stop` is executed
- **SC-010**: System supports at least 10 custom components deployed simultaneously without performance degradation
- **SC-011**: Developers can mount custom configuration files and verify contents in under 1 minute using kubectl exec
- **SC-012**: Config file changes in kindenv.yaml are reflected in running pods within 30 seconds of cluster restart
- **SC-013**: System supports at least 10 config files per component without performance degradation
