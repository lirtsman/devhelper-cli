# Kind-based Environment Setup

The `kindenv` command provisions a local Kind-based Kubernetes development environment with all necessary components for Shield application development.

## Overview

This command creates and manages a Kind Kubernetes cluster with:

- [Temporal](https://temporal.io/) server
- [Dapr](https://dapr.io/) runtime
- [Redis](https://redis.io/)
- [MySQL 8](https://www.mysql.com/) database
- Required infrastructure for Shield applications

## Prerequisites

- [Kind](https://kind.sigs.k8s.io/) - Kubernetes in Docker
- [kubectl](https://kubernetes.io/docs/tasks/tools/) - Kubernetes command-line tool
- [Helm](https://helm.sh/) - Kubernetes package manager
- Docker or Podman - Container engine
- AWS CLI (optional) - For AWS ECR integration

## Initial Setup

Before using the `kindenv` command for the first time, you should run the initialization command to set up the configuration file and required Helm repositories:

```bash
devhelper-cli kindenv init
```

This will create a default `kindenv.yaml` configuration file and set up all necessary Helm repositories.

## Configuration

The `kindenv` command uses a YAML configuration file (by default `kindenv.yaml` in the current directory):

```yaml
tools:
  podman:
    path: /opt/homebrew/bin/podman
    version: 5.4.1
  docker:
    path: /usr/bin/docker
    version: ""
  kind:
    path: /opt/homebrew/bin/kind
    version: ""
  kubectl:
    path: /opt/homebrew/bin/kubectl
    version: ""
  helm:
    path: /opt/homebrew/bin/helm
    version: ""
  aws:
    path: /opt/homebrew/bin/aws
    version: ""
cluster:
  name: kind
  createIfNotExists: true
  mapPorts:
    - containerPort: 80
      hostPort: 80
      protocol: TCP
    - containerPort: 443
      hostPort: 443
      protocol: TCP
components:
  temporal:
    enabled: true
    namespace: temporal
    chartVersion: "0.62.0"
    nodePorts:
      web: 30080
      frontend: 30733
  redis:
    enabled: true
    chartVersion: "17.3.7"
    nodePorts:
      redis: 30679
    auth:
      enabled: false
  mysql:
    enabled: true
    chartVersion: "9.4.6"
    database: "mysql"
    nodePorts:
      mysql: 30306
    resources:
      cpu: "500m"
      memory: "1Gi"
    persistence:
      enabled: false
      size: "8Gi"
images:
  skipPull: false
  dockerHub:
    username: ""
    password: ""
  useAwsEcr: false
  aws:
    region: "eu-west-1"
    ecrRegistry: "992979781608.dkr.ecr.eu-west-1.amazonaws.com"
secrets:
  mysql:
    enabled: true
    name: "kvv2-mysql-creds"
    namespace: "default"
    username: ""
    password: ""
```

### AWS ECR Configuration

The `kindenv` command supports pulling images from AWS Elastic Container Registry (ECR):

- `images.useAwsEcr`: Set to `true` to enable AWS ECR integration
- `images.aws.region`: AWS region for ECR (default: eu-west-1)
- `images.aws.ecrRegistry`: ECR registry URL (optional, can be auto-detected)

When AWS ECR integration is enabled, the command will:
1. Get AWS credentials using the AWS CLI
2. Create ECR pull secrets in relevant namespaces
3. Configure Kubernetes service accounts to use ECR credentials
4. Configure Helm releases to use ECR credentials

## Usage

### Initialize configuration

```bash
devhelper-cli kindenv init
```

This command performs two key setup tasks:

1. **Configuration file**: Generates a default `kindenv.yaml` configuration file with recommended settings for Kind, Temporal, Redis, Dapr, and cert-manager.

2. **Helm repositories**: Adds and updates the necessary Helm repositories:
   - `temporal`: For Temporal server
   - `dapr`: For Dapr runtime
   - `bitnami`: For Redis and MySQL

Options:
- `--output`, `-o`: Specify a different output path for the configuration file (default: `kindenv.yaml`)
- `--force`, `-f`: Force overwrite if the configuration file already exists
- `--skip-repos`: Skip adding and updating Helm repositories

After initializing, you can customize the generated configuration file to suit your needs.

### Start the environment

```bash
devhelper-cli kindenv start
```

Options:
- `--config`: Path to configuration file (default: `kindenv.yaml` in current directory)
- `--cluster-name`: Name of the Kind cluster (default: `kind`)
- `--operator-namespace`: Namespace for Temporal worker operator (default: `temporal-worker-operator-system`)
- `--skip-temporal`: Skip installing Temporal
- `--skip-dapr`: Skip installing Dapr
- `--skip-redis`: Skip installing Redis
- `--use-aws-ecr`: Enable AWS ECR integration
- `--verbose`: Enable verbose output

### Check environment status

```bash
devhelper-cli kindenv status
```

This command shows the status of the Kind cluster and all deployed components.

### Stop/delete the environment

```bash
devhelper-cli kindenv stop --delete
```

Options:
- `--delete`: Delete the Kind cluster completely (if not specified, the command just displays instructions)

## Typical Workflow

A typical workflow for using the `kindenv` command is:

1. **Initialize** the configuration and Helm repositories: `devhelper-cli kindenv init`
2. Optionally edit the generated `kindenv.yaml` file to customize settings
3. **Start** the environment: `devhelper-cli kindenv start`
4. Work with the environment
5. Check **status** when needed: `devhelper-cli kindenv status`
6. **Stop** or delete when done: `devhelper-cli kindenv stop --delete`

## Accessing Services

After starting the environment, services are accessible at:

- **Temporal Web UI**: http://localhost:8080 (configurable)
- **Temporal Frontend**: localhost:7233 (configurable)
- **Redis**: localhost:6379 (configurable)
- **MySQL**: localhost:3306 (configurable)

### MySQL Configuration

MySQL 8 is deployed using the Bitnami MySQL Helm chart. You can configure MySQL in the `components.mysql` section:

```yaml
components:
  mysql:
    enabled: true                    # Enable/disable MySQL installation
    chartVersion: "9.4.6"            # Helm chart version
    database: "mysql"                # Database name to create
    nodePorts:
      mysql: 30306                   # Kubernetes NodePort (30000-32767)
    resources:
      cpu: "500m"                    # CPU request (e.g., "500m", "1")
      memory: "1Gi"                  # Memory request (e.g., "1Gi", "512Mi")
    persistence:
      enabled: false                 # Enable persistent storage
      size: "8Gi"                    # Storage size when persistence enabled
    initScripts:                     # Optional: SQL initialization scripts
      init.sql: |                    # Scripts are executed in alphabetical order
        CREATE DATABASE IF NOT EXISTS myapp;
        USE myapp;
        CREATE TABLE IF NOT EXISTS users (
          id INT AUTO_INCREMENT PRIMARY KEY,
          username VARCHAR(255) NOT NULL,
          email VARCHAR(255) NOT NULL
        );
```

#### MySQL Credentials

MySQL credentials are managed through the `secrets.mysql` section:

```yaml
secrets:
  mysql:
    enabled: true                    # Enable MySQL secret creation
    name: "mysql-credentials"        # Kubernetes secret name
    namespace: "mysql"               # Namespace for the secret
    username: "root"                 # MySQL username
    password: "password"             # MySQL password
```

When `secrets.mysql.enabled` is `true`, the MySQL Helm chart will use the specified secret for authentication. Otherwise, default credentials (root/password) are used.

#### MySQL Persistence

By default, MySQL persistence is disabled for faster development cycles. To enable data persistence across restarts:

```yaml
components:
  mysql:
    persistence:
      enabled: true
      size: "8Gi"                    # Adjust size as needed
```

**Note**: When persistence is enabled, data will survive cluster restarts. When disabled, data is lost when the cluster is stopped.

#### MySQL Initialization Scripts

You can provide SQL initialization scripts that will be executed when MySQL first starts. Scripts can be provided either inline or as file paths:

**Inline Scripts:**
```yaml
components:
  mysql:
    initScripts:
      init.sql: |
        CREATE DATABASE IF NOT EXISTS myapp;
        USE myapp;
        CREATE TABLE IF NOT EXISTS users (
          id INT AUTO_INCREMENT PRIMARY KEY,
          username VARCHAR(255) NOT NULL,
          email VARCHAR(255) NOT NULL
        );
```

**File Paths (relative to kindenv.yaml location):**
```yaml
components:
  mysql:
    initScripts:
      init.sql: ./scripts/init.sql
      seed-data.sql: ./scripts/seed-data.sql
```

**Mixed (inline and file paths):**
```yaml
components:
  mysql:
    initScripts:
      init.sql: ./scripts/init.sql          # Load from file
      seed-data.sql: |                      # Inline content
        INSERT INTO myapp.users (username, email) VALUES
          ('admin', 'admin@example.com'),
          ('user1', 'user1@example.com');
```

**Important Notes**:
- Scripts are executed in alphabetical order by filename
- Scripts only run on the **first initialization** of MySQL (when the data directory is empty)
- If persistence is enabled and data already exists, scripts will **not** run again
- Use `.sql` extension for SQL scripts (`.sh` scripts are also supported but only run on primary nodes)
- Scripts are mounted to `/docker-entrypoint-initdb.d/` in the MySQL container
- File paths are resolved relative to the `kindenv.yaml` file location
- If a file path doesn't exist, it will be treated as inline content (with a warning)

#### Connecting to MySQL

After starting the environment, connect to MySQL using:

```bash
mysql -h localhost -P 3306 -u root -p
```

Or using the configured credentials:

```bash
mysql -h localhost -P 3306 -u <username> -p
```

The default password is `password` unless configured otherwise in `secrets.mysql.password`.

### Monitoring Stack (kube-prometheus-stack)

Deploy Prometheus Operator and Grafana for cluster monitoring.

**Configuration** (`kindenv.yaml`):

```yaml
components:
  monitoring:
    enabled: false                   # Enable monitoring stack (default: false)
    namespace: monitoring            # Kubernetes namespace (default: monitoring)
    chartVersion: "72.6.2"           # kube-prometheus-stack Helm chart version
    grafana:
      nodePort: 31300                # NodePort for Grafana (default: 31300, range: 30000-32767)
    prometheus:
      retention: "24h"               # Metrics retention period (default: 24h)
    resources:
      prometheus:
        cpu: "500m"                  # Prometheus CPU limit
        memory: "512Mi"              # Prometheus memory limit
      grafana:
        cpu: "200m"                  # Grafana CPU limit
        memory: "256Mi"              # Grafana memory limit
```

**Access**: When enabled, Grafana is available at `http://localhost:3000` (no login required).

**Skip flag**: `devhelper-cli kindenv start --skip-monitoring`

## Troubleshooting

### Cannot connect to services

If you cannot connect to services, verify:

1. The Kind cluster is running: `kind get clusters`
2. Services are deployed: `kubectl -n temporal get pods`
3. Port mappings are correct: `kubectl -n temporal get service`

### AWS ECR Authentication Issues

If you encounter ECR authentication issues:
1. Ensure AWS CLI is installed and configured: `aws sts get-caller-identity`
2. Check if your ECR permissions are correct: `aws ecr get-login-password`
3. Verify the ECR registry URL is correct in your config

### Resource limitations

Kind runs Kubernetes in Docker containers, which has resource limitations. Ensure your Docker/Podman has enough CPU and memory allocated. 