# Custom Component App Example

A simple example application demonstrating how to deploy custom components in a Kind environment using `devhelper-cli kindenv`.

## Overview

This example includes:
- A simple Go HTTP server that demonstrates custom component features
- Configuration file mounting via ConfigMaps
- Environment variable configuration
- Port mapping for local access
- Resource limits configuration
- Integration with MySQL (via environment variables)

## Features Demonstrated

- ✅ Container image deployment
- ✅ Environment variable configuration
- ✅ Port mapping (NodePort)
- ✅ Resource limits (CPU and memory)
- ✅ Configuration file mounting (ConfigMaps)
- ✅ Multi-file configuration support

## Prerequisites

- Go 1.21 or later
- Docker or Podman
- Kind cluster (created via `devhelper-cli kindenv start`)
- kubectl configured to access the Kind cluster

## Quick Start

### 1. Build the Docker Image

```bash
# Build the image
docker build -t custom-component-app:latest .

# Load into Kind cluster
kind load docker-image custom-component-app:latest
```

Or using Podman:

```bash
# Build the image
podman build -t custom-component-app:latest .

# Load into Kind cluster
kind load image-archive $(podman save -o /tmp/custom-component-app.tar custom-component-app:latest) --name <your-cluster-name>
```

### 2. Configure kindenv.yaml

Copy the example `kindenv.yaml` to your project root or merge it with your existing configuration:

```bash
# Copy the example config
cp examples/custom-component-app/kindenv.yaml ./kindenv.yaml

# Or merge with your existing kindenv.yaml
```

### 3. Deploy the Component

```bash
# Start the Kind environment (will deploy the custom component)
devhelper-cli kindenv start

# Check status
devhelper-cli kindenv status

# View logs
kubectl logs -l app=custom-component-app -f
```

### 4. Test the Application

```bash
# Health check
curl http://localhost:30088/health

# Root endpoint
curl http://localhost:30088/

# View configuration
curl http://localhost:30088/config

# View environment variables
curl http://localhost:30088/env
```

## Application Endpoints

- `GET /` - Root endpoint with basic information
- `GET /health` - Health check endpoint
- `GET /config` - View mounted configuration file
- `GET /env` - View environment variables

## Configuration

The application supports configuration via:

1. **Environment Variables**: Set in `kindenv.yaml` under `env`
2. **Config Files**: Mounted from ConfigMaps at `/config/`

### Environment Variables

- `SERVER_PORT` - HTTP server port (default: 8080)
- `APP_ENV` - Application environment (development/production)
- `DB_HOST` - Database hostname
- `DB_PORT` - Database port
- `CONFIG_FILE` - Path to configuration file (default: /config/application.yaml)

### Config Files

- `/config/application.yaml` - Main application configuration
- `/config/logback.xml` - Logging configuration (example)

## Customization

### Modify the Application

Edit `main.go` to add your own endpoints and logic.

### Update Configuration

Edit `config/application.yaml` or modify the `configFiles` section in `kindenv.yaml`.

### Change Port Mapping

Update the `ports` section in `kindenv.yaml`:

```yaml
ports:
  - containerPort: 8080
    protocol: TCP
    nodePort: 30089  # Change to a different port
```

### Add MySQL Connection

Uncomment the MySQL secret references in `kindenv.yaml`:

```yaml
env:
  - name: DB_USERNAME
    valueFrom:
      secretKeyRef:
        name: mysql-secret
        key: username
  - name: DB_PASSWORD
    valueFrom:
      secretKeyRef:
        name: mysql-secret
        key: password
```

## Troubleshooting

### Image Not Found

If you see `ImagePullBackOff`:

```bash
# Verify image is loaded
docker images | grep custom-component-app

# Load into Kind
kind load docker-image custom-component-app:latest
```

### Port Already in Use

Change the `nodePort` in `kindenv.yaml` to a different value (30000-32767).

### Config File Not Found

Verify the ConfigMap was created:

```bash
kubectl get configmap custom-component-app-config
kubectl describe configmap custom-component-app-config
```

### View Pod Logs

```bash
# Get pod name
kubectl get pods -l app=custom-component-app

# View logs
kubectl logs <pod-name>

# Follow logs
kubectl logs -f <pod-name>
```

### Check Deployment Status

```bash
# Check deployment
kubectl get deployment custom-component-app

# Describe deployment
kubectl describe deployment custom-component-app

# Check pods
kubectl get pods -l app=custom-component-app
```

## Cleanup

```bash
# Stop the Kind environment (removes all components)
devhelper-cli kindenv stop

# Or delete just the custom component
kubectl delete deployment custom-component-app
kubectl delete service custom-component-app
kubectl delete configmap custom-component-app-config
```

## Next Steps

- Add more endpoints to the application
- Integrate with MySQL or OpenSearch
- Add health checks and readiness probes
- Scale the application (increase replicas)
- Add multiple custom components

For more information, see:
- [CUSTOM_COMPONENTS.md](../../CUSTOM_COMPONENTS.md) - User guide
- [Quick Start Guide](../../specs/002-custom-components/quickstart.md) - Detailed examples
