# KEDA Quick Start Guide

**KEDA (Kubernetes Event Driven Autoscaling)** enables event-driven autoscaling for your applications running in Kind environments. This guide shows you how to configure and use KEDA for autoscaling based on external event sources.

## Overview

KEDA extends Kubernetes with event-driven autoscaling capabilities:

- **Scale to Zero**: Applications can scale down to zero replicas when idle
- **50+ Event Sources**: RabbitMQ, Kafka, Prometheus, MySQL, Redis, and more
- **Standard Kubernetes**: Works with existing Deployments and StatefulSets
- **Custom Metrics**: Expose custom metrics to Horizontal Pod Autoscaler

## Quick Start

### 1. Enable KEDA in Configuration

Edit your `kindenv.yaml`:

```yaml
components:
  keda:
    enabled: true
    namespace: keda
    chartVersion: 2.16.0
```

### 2. Start Your Environment

```bash
# Start with KEDA enabled
devhelper-cli kindenv start

# Or skip KEDA temporarily
devhelper-cli kindenv start --skip-keda
```

### 3. Verify KEDA Installation

```bash
# Check status
devhelper-cli kindenv status

# Verify KEDA pods are running
kubectl get pods -n keda

# Expected output:
# NAME                                      READY   STATUS    RESTARTS   AGE
# keda-operator-xxxxxxxxxx-xxxxx           1/1     Running   0          2m
# keda-metrics-apiserver-xxxxxxxxx-xxxxx   1/1     Running   0          2m
```

### 4. View KEDA Custom Resource Definitions

```bash
# List available CRDs
kubectl get crd | grep keda

# Expected output:
# scaledobjects.keda.sh
# scaledjobs.keda.sh
# triggerauthentications.keda.sh
# clustertriggerauthentications.keda.sh
```

## Configuration Options

### Default Configuration

```yaml
components:
  keda:
    enabled: false        # Disabled by default (opt-in)
    namespace: keda       # Kubernetes namespace for KEDA
    chartVersion: 2.16.0  # Stable KEDA Helm chart version
```

### Custom Chart Version

Use a specific KEDA version for environment parity:

```yaml
components:
  keda:
    enabled: true
    namespace: keda
    chartVersion: 2.19.0  # Latest version
```

### Custom Namespace

Install KEDA in a different namespace:

```yaml
components:
  keda:
    enabled: true
    namespace: autoscaling  # Custom namespace
    chartVersion: 2.16.0
```

## ScaledObject Examples

### Example 1: RabbitMQ Queue-Based Autoscaling

Scale a consumer application based on RabbitMQ queue depth:

```yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: rabbitmq-consumer-scaler
  namespace: default
spec:
  scaleTargetRef:
    name: rabbitmq-consumer  # Your deployment name
  minReplicaCount: 0         # Scale to zero when queue is empty
  maxReplicaCount: 10        # Maximum replicas
  pollingInterval: 30        # Check every 30 seconds
  cooldownPeriod: 300        # Wait 5 minutes before scaling down
  triggers:
    - type: rabbitmq
      metadata:
        protocol: amqp
        queueName: tasks
        mode: QueueLength
        value: "5"  # Scale up when queue has 5+ messages
        host: amqp://user:password@rabbitmq.rabbitmq.svc.cluster.local:5672/
```

**Apply the ScaledObject:**

```bash
kubectl apply -f rabbitmq-scaledobject.yaml

# Check scaling status
kubectl get scaledobject -n default
kubectl describe scaledobject rabbitmq-consumer-scaler -n default
```

### Example 2: RabbitMQ with TriggerAuthentication

Use secrets for RabbitMQ credentials:

**Step 1: Create TriggerAuthentication**

```yaml
apiVersion: keda.sh/v1alpha1
kind: TriggerAuthentication
metadata:
  name: rabbitmq-auth
  namespace: default
spec:
  secretTargetRef:
    - parameter: host
      name: kvv2-rabbitmq  # Existing RabbitMQ secret
      key: connection-string
```

**Step 2: Create ScaledObject**

```yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: rabbitmq-consumer-scaler-secure
  namespace: default
spec:
  scaleTargetRef:
    name: rabbitmq-consumer
  minReplicaCount: 1
  maxReplicaCount: 20
  triggers:
    - type: rabbitmq
      metadata:
        queueName: tasks
        mode: QueueLength
        value: "10"
      authenticationRef:
        name: rabbitmq-auth
```

### Example 3: MySQL Query-Based Autoscaling

Scale based on MySQL query results:

```yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: mysql-query-scaler
  namespace: default
spec:
  scaleTargetRef:
    name: data-processor
  minReplicaCount: 1
  maxReplicaCount: 8
  triggers:
    - type: mysql
      metadata:
        queryValue: "5"
        query: "SELECT COUNT(*) FROM pending_jobs WHERE status='pending'"
        connectionStringFromEnv: "MYSQL_CONNECTION_STRING"
```

### Example 4: Cron-Based Autoscaling

Scale up during business hours:

```yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: business-hours-scaler
  namespace: default
spec:
  scaleTargetRef:
    name: web-app
  minReplicaCount: 1
  maxReplicaCount: 1
  triggers:
    - type: cron
      metadata:
        timezone: UTC
        start: 0 8 * * 1-5   # 8 AM Monday-Friday
        end: 0 18 * * 1-5     # 6 PM Monday-Friday
        desiredReplicas: "5"
```

### Example 5: Prometheus Metrics-Based Autoscaling

Scale based on custom Prometheus metrics:

```yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: prometheus-scaler
  namespace: default
spec:
  scaleTargetRef:
    name: api-service
  minReplicaCount: 2
  maxReplicaCount: 15
  triggers:
    - type: prometheus
      metadata:
        serverAddress: http://prometheus.monitoring.svc.cluster.local:9090
        metricName: http_requests_per_second
        query: sum(rate(http_requests_total[1m]))
        threshold: "100"
```

## Common Use Cases

### Use Case 1: Background Job Processing

**Scenario**: Process jobs from a RabbitMQ queue with automatic scaling.

**Setup**:
1. Enable RabbitMQ in `kindenv.yaml`
2. Enable KEDA in `kindenv.yaml`
3. Deploy your worker application
4. Apply RabbitMQ ScaledObject (see Example 1)

**Result**: Workers automatically scale based on queue depth, scaling to zero when idle.

### Use Case 2: Event Stream Processing

**Scenario**: Process events from Kafka topics with load-based scaling.

**Setup**:
```yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: kafka-consumer-scaler
  namespace: default
spec:
  scaleTargetRef:
    name: event-processor
  triggers:
    - type: kafka
      metadata:
        bootstrapServers: kafka.kafka.svc.cluster.local:9092
        consumerGroup: my-group
        topic: events
        lagThreshold: "50"
```

### Use Case 3: Database Connection Pool Management

**Scenario**: Scale database workers based on pending transactions.

**Benefits**:
- Efficient resource usage
- Automatic response to load spikes
- Scale to zero during idle periods

## Integration with Existing Components

### RabbitMQ Integration

KEDA works seamlessly with RabbitMQ already configured in kindenv:

```bash
# RabbitMQ is available at
# AMQP: amqp://localhost:5672/
# Management UI: http://localhost:15672

# Connection string for KEDA
amqp://user:password@rabbitmq.rabbitmq.svc.cluster.local:5672/
```

### MySQL Integration

Scale based on MySQL workload:

```bash
# MySQL is available at
# Host: mysql.mysql.svc.cluster.local
# Port: 3306

# Connection string format
mysql://user:password@mysql.mysql.svc.cluster.local:3306/database
```

## Monitoring and Debugging

### Check ScaledObject Status

```bash
# List all ScaledObjects
kubectl get scaledobjects -A

# Detailed status
kubectl describe scaledobject <name> -n <namespace>

# Check HPA created by KEDA
kubectl get hpa -A
```

### View KEDA Operator Logs

```bash
# KEDA operator logs
kubectl logs -n keda deployment/keda-operator -f

# KEDA metrics server logs
kubectl logs -n keda deployment/keda-metrics-apiserver -f
```

### Debug Scaling Issues

```bash
# Check current replica count
kubectl get deployment <deployment-name> -n <namespace>

# View scaling events
kubectl get events -n <namespace> --sort-by='.lastTimestamp'

# Check HPA status
kubectl describe hpa <hpa-name> -n <namespace>
```

## Troubleshooting

### Issue: ScaledObject Not Triggering

**Symptoms**: Deployment doesn't scale despite trigger conditions being met.

**Solutions**:
1. Verify KEDA operator is running:
   ```bash
   kubectl get pods -n keda
   ```

2. Check ScaledObject events:
   ```bash
   kubectl describe scaledobject <name> -n <namespace>
   ```

3. Verify trigger authentication:
   ```bash
   kubectl get triggerauthentication -n <namespace>
   ```

4. Check KEDA operator logs for errors:
   ```bash
   kubectl logs -n keda deployment/keda-operator | grep ERROR
   ```

### Issue: Scale to Zero Not Working

**Symptoms**: Pods don't scale down to zero when idle.

**Solutions**:
1. Ensure `minReplicaCount: 0` is set in ScaledObject
2. Check `cooldownPeriod` - scaling down might be delayed
3. Verify trigger returns zero when idle
4. Check if HPA has minimum replica constraints

### Issue: Authentication Failures

**Symptoms**: Cannot connect to event source (RabbitMQ, MySQL, etc.)

**Solutions**:
1. Verify secret exists and contains correct data:
   ```bash
   kubectl get secret <secret-name> -n <namespace> -o yaml
   ```

2. Check TriggerAuthentication references correct secret:
   ```bash
   kubectl describe triggerauthentication <name> -n <namespace>
   ```

3. Test connection manually from a pod:
   ```bash
   kubectl run test-pod --rm -it --image=alpine -- sh
   # Test connection to event source
   ```

### Issue: Chart Version Not Found

**Symptoms**: Helm install fails with "chart version not found"

**Solutions**:
1. Update Helm repositories:
   ```bash
   helm repo update
   ```

2. Check available versions:
   ```bash
   helm search repo kedacore/keda --versions
   ```

3. Use a valid chart version in `kindenv.yaml`:
   ```yaml
   components:
     keda:
       chartVersion: 2.16.0  # Known stable version
   ```

## Advanced Configuration

### Custom Helm Values

For advanced KEDA configuration, you can extend the installation after deployment:

```bash
# Add custom values
helm upgrade keda kedacore/keda \
  --namespace keda \
  --set resources.operator.limits.cpu=2 \
  --set resources.operator.limits.memory=2Gi
```

### Multiple Triggers

Scale based on multiple conditions:

```yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: multi-trigger-scaler
  namespace: default
spec:
  scaleTargetRef:
    name: processor
  triggers:
    - type: rabbitmq
      metadata:
        queueName: high-priority
        value: "5"
        host: amqp://rabbitmq.rabbitmq.svc.cluster.local:5672/
    - type: rabbitmq
      metadata:
        queueName: low-priority
        value: "20"
        host: amqp://rabbitmq.rabbitmq.svc.cluster.local:5672/
```

### ScaledJob for Batch Processing

Use ScaledJob for one-time job execution:

```yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledJob
metadata:
  name: batch-processor
  namespace: default
spec:
  jobTargetRef:
    template:
      spec:
        containers:
          - name: processor
            image: myapp:latest
        restartPolicy: Never
  pollingInterval: 30
  maxReplicaCount: 10
  triggers:
    - type: rabbitmq
      metadata:
        queueName: batch-jobs
        value: "1"
        host: amqp://rabbitmq.rabbitmq.svc.cluster.local:5672/
```

## Best Practices

### 1. Start Conservative

- Begin with higher `minReplicaCount` (e.g., 1-2)
- Use moderate `maxReplicaCount` values
- Tune after observing actual load patterns

### 2. Set Appropriate Cooldown

```yaml
spec:
  cooldownPeriod: 300  # 5 minutes - prevents rapid scaling
  pollingInterval: 30   # Check every 30 seconds
```

### 3. Use TriggerAuthentication

Always use TriggerAuthentication for credentials instead of hardcoding in ScaledObject.

### 4. Monitor Resource Usage

Track CPU and memory usage to set appropriate resource limits:

```bash
kubectl top pods -n <namespace>
```

### 5. Test Scaling Behavior

```bash
# Generate load to trigger scaling
# Example: Send messages to RabbitMQ queue
# Observe scaling behavior

kubectl get hpa -w  # Watch scaling in real-time
```

## Reference

### Supported Scalers

KEDA supports 50+ event sources including:

- **Message Queues**: RabbitMQ, Kafka, Azure Queue, AWS SQS
- **Databases**: MySQL, PostgreSQL, MongoDB, Redis
- **Cloud Services**: AWS CloudWatch, Azure Monitor, GCP Pub/Sub
- **Metrics**: Prometheus, Datadog, New Relic
- **HTTP**: HTTP endpoint polling
- **Cron**: Time-based scaling

Full list: https://keda.sh/docs/scalers/

### Useful Commands

```bash
# List all KEDA resources
kubectl get scaledobjects,scaledjobs,triggerauthentications -A

# Watch scaling activity
kubectl get hpa -A -w

# Remove a ScaledObject
kubectl delete scaledobject <name> -n <namespace>

# Get KEDA version
kubectl get deployment keda-operator -n keda -o jsonpath='{.spec.template.spec.containers[0].image}'
```

### Documentation Links

- **KEDA Documentation**: https://keda.sh/docs/
- **Scaler Reference**: https://keda.sh/docs/scalers/
- **GitHub Repository**: https://github.com/kedacore/keda
- **Samples**: https://github.com/kedacore/samples

## Next Steps

1. **Enable KEDA** in your `kindenv.yaml`
2. **Deploy your application** with appropriate resource requests/limits
3. **Create a ScaledObject** matching your use case
4. **Monitor scaling behavior** and tune parameters
5. **Add TriggerAuthentication** for production use

For more examples and detailed scaler documentation, visit https://keda.sh/docs/