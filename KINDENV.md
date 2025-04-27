# Kind-based Environment Setup

The `kindenv` command provisions a local Kind-based Kubernetes development environment with all necessary components for Shield application development.

## Overview

This command creates and manages a Kind Kubernetes cluster with:

- [Temporal](https://temporal.io/) server
- [Dapr](https://dapr.io/) runtime
- [Redis](https://redis.io/)
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
    webPort: 8080
    webNodePort: 30080
    frontendPort: 7233
    frontendNodePort: 30733
  redis:
    enabled: true
    port: 6379
    nodePort: 30679
    image: bitnami/redis:7.0.5-debian-11-r7
    auth:
      enabled: false
images:
  skipPull: false
  dockerHub:
    username: ""
    password: ""
  useAwsEcr: false
  aws:
    region: "eu-west-1"
    ecrRegistry: "992979781608.dkr.ecr.eu-west-1.amazonaws.com"
    serviceAccount: "ecr-pull-service-account"
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
- `images.aws.serviceAccount`: Kubernetes service account name for ECR credentials

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
   - `bitnami`: For Redis

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