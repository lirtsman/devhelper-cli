# Research: RabbitMQ Support for KindEnv

**Feature**: `003-rabbitmq-kindenv`  
**Date**: 2026-02-05  
**Phase**: 0 (Research & Analysis)

## Research Overview

This document captures the research findings for adding RabbitMQ support to the kindenv command, following the same architectural pattern established by MySQL integration (spec 001-mysql8-kindenv).

## Research Questions Addressed

### 1. Bitnami RabbitMQ Helm Chart Selection

**Question**: Which RabbitMQ deployment method is most appropriate for kindenv integration?

**Research Findings**:

1. **Bitnami RabbitMQ Helm Chart** (SELECTED)
   - **Version**: 11.x+ (latest stable)
   - **Repository**: `https://charts.bitnami.com/bitnami` with bitnamilegacy registry
   - **Pros**:
     - Consistent with existing MySQL and Redis integration patterns
     - Well-maintained by Bitnami with regular security updates
     - Supports bitnamilegacy image repository for ECR compatibility
     - Simple configuration via Helm values
     - Includes management plugin with UI
     - Supports both standalone and clustered deployments
   - **Cons**:
     - Opinionated defaults may need customization
     - Larger image size compared to alpine-based images

2. **RabbitMQ Official Kubernetes Operator**
   - **Pros**: Native Kubernetes integration, advanced features
   - **Cons**: More complex setup, overkill for single-node dev environments, additional CRDs

3. **Custom Kubernetes Manifests**
   - **Pros**: Full control over configuration
   - **Cons**: High maintenance burden, no standardized upgrade path

**Decision**: Use Bitnami RabbitMQ Helm chart (version 11.x+)

**Rationale**: 
- Maintains consistency with existing component integrations (MySQL, Redis, Temporal)
- Bitnami charts are production-ready and well-tested
- Built-in support for bitnamilegacy registry required for ECR integration
- Minimal configuration required for single-node dev setup
- Strong community support and documentation

### 2. Port Exposure Strategy

**Question**: How should RabbitMQ services be exposed for local development access?

**Research Findings**:

RabbitMQ requires two distinct ports:
- **AMQP Protocol**: Port 5672 (client connections)
- **Management UI**: Port 15672 (HTTP management interface)

**Options Evaluated**:

1. **NodePort with Kind Cluster Port Mapping** (SELECTED)
   - **AMQP**: NodePort 30672 → Kind maps to host port 5672
   - **Management UI**: NodePort 31672 → Kind maps to host port 15672
   - **Pros**:
     - Direct localhost access without port-forwarding
     - Persistent across pod restarts
     - Consistent with MySQL pattern (port 3306)
     - No manual intervention required
   - **Cons**:
     - Requires cluster recreation if port mappings change
     - Limited to NodePort range (30000-32767)

2. **Ingress Controller**
   - **Pros**: HTTP-based routing, TLS termination
   - **Cons**: Requires ingress controller installation, complex for dev environment, AMQP protocol not HTTP-based

3. **kubectl port-forward**
   - **Pros**: Simple, no cluster config changes
   - **Cons**: Manual process, not persistent, separate process for each port

**Decision**: NodePort with Kind cluster port mapping

**Rationale**:
- Matches MySQL integration pattern for consistency
- Provides seamless localhost access (amqp://localhost:5672, http://localhost:15672)
- Zero configuration needed after initial setup
- Both AMQP and Management UI accessible simultaneously

### 3. Virtual Host Management

**Question**: How should RabbitMQ virtual hosts be configured for development environments?

**Research Findings**:

RabbitMQ virtual hosts provide namespace isolation similar to databases in MySQL.

**Virtual Host Concepts**:
- Default virtual host: "/" (root)
- Used to isolate applications, users, queues, and exchanges
- Can be created dynamically via management UI or API
- Recommended for multi-tenant or multi-application environments

**Options Evaluated**:

1. **Configurable Default Virtual Host** (SELECTED)
   - Allow users to specify default virtual host in `kindenv.yaml`
   - Default to "/" if not specified
   - Additional virtual hosts can be created manually via Management UI
   - **Pros**:
     - Simple configuration
     - Flexible for single or multi-app development
     - No complex vhost pre-provisioning logic
   - **Cons**:
     - Limited to one pre-configured vhost

2. **Multiple Virtual Hosts via Helm Values**
   - Pre-create multiple virtual hosts during installation
   - **Pros**: Fully automated multi-vhost setup
   - **Cons**: Complex configuration, rarely needed for dev environments

3. **No Virtual Host Configuration**
   - Always use default "/" vhost
   - **Pros**: Simplest approach
   - **Cons**: Inflexible, no isolation between test scenarios

**Decision**: Configurable default virtual host with "/" as default

**Rationale**:
- Provides flexibility without added complexity
- Users can specify custom vhost for isolation (e.g., "/dev", "/test")
- Additional vhosts can be created dynamically via Management UI
- Aligns with "simple defaults, advanced options available" principle

### 4. Management Plugin Configuration

**Question**: Should the RabbitMQ management plugin be enabled by default?

**Research Findings**:

RabbitMQ management plugin provides:
- Web-based UI for monitoring queues, exchanges, bindings
- REST API for automation and monitoring
- User management interface
- Message rate statistics and visualizations

**Options Evaluated**:

1. **Enabled by Default** (SELECTED)
   - Management plugin enabled in Helm chart configuration
   - UI accessible at http://localhost:15672
   - **Pros**:
     - Essential for development workflow
     - Provides visibility into message flows
     - Enables debugging and troubleshooting
     - Consistent with RabbitMQ Docker images (management-enabled)
   - **Cons**:
     - Slightly increased memory footprint (~50MB)

2. **Disabled by Default**
   - Users must manually enable if needed
   - **Pros**: Minimal resource usage
   - **Cons**: Poor developer experience, requires additional configuration

3. **Optional Plugin Configuration**
   - Make management plugin configurable
   - **Pros**: Maximum flexibility
   - **Cons**: Adds configuration complexity, most users want it enabled

**Decision**: Enable management plugin by default

**Rationale**:
- Management UI is essential for development and debugging
- Memory overhead is acceptable for dev environments (within 1Gi default limit)
- Consistent with industry standard (Docker official images include management)
- Improves developer experience significantly

### 5. Secret Management Strategy

**Question**: How should RabbitMQ credentials and sensitive data be managed?

**Research Findings**:

RabbitMQ requires several secrets:
- **Username**: Admin user (default: "user")
- **Password**: Admin password (default: "password")
- **Erlang Cookie**: Cluster authentication token (required even for single-node)

**Options Evaluated**:

1. **Dedicated Kubernetes Secret** (SELECTED)
   - Create `rabbitmq-credentials` secret in rabbitmq namespace
   - Store username, password, and erlang cookie
   - Pass secret name to Helm chart via values
   - **Pros**:
     - Follows MySQL secrets pattern
     - Supports future clustering (erlang cookie required)
     - Secure storage via Kubernetes secrets
     - Easy to rotate credentials
   - **Cons**:
     - Additional secret resource to manage

2. **Reuse Existing Secrets Structure**
   - Add rabbitmq credentials to existing `secrets.mysql` structure
   - **Pros**: Less configuration
   - **Cons**: Naming confusion, potential conflicts, not semantically correct

3. **Inline Credentials in Helm Values**
   - Pass credentials directly via Helm values
   - **Pros**: Simplest approach
   - **Cons**: Insecure, credentials visible in Helm history

**Decision**: Dedicated Kubernetes secret for RabbitMQ

**Rationale**:
- Follows established MySQL pattern for consistency
- Properly isolates RabbitMQ credentials
- Supports future clustering capabilities (erlang cookie)
- Secure by default (Kubernetes secret encryption)
- Configuration structure in `kindenv.yaml`:

```yaml
secrets:
  rabbitmq:
    enabled: true
    name: "rabbitmq-credentials"
    namespace: "rabbitmq"
    username: "user"
    password: "password"
    erlangCookie: "secretcookie"  # Auto-generated if not provided
```

### 6. Resource Defaults

**Question**: What resource limits should be set for RabbitMQ development environments?

**Research Findings**:

RabbitMQ resource requirements vary by workload:
- **Baseline**: ~200MB memory, minimal CPU for idle state
- **Active messaging**: 500MB-1GB memory, 0.5-1 CPU core
- **High throughput**: 2GB+ memory, 2+ CPU cores

**Options Evaluated**:

1. **Match MySQL Defaults** (SELECTED)
   - **CPU**: 500m (0.5 cores)
   - **Memory**: 1Gi (1 gibibyte)
   - **Storage**: 8Gi (if persistence enabled)
   - **Pros**:
     - Consistent with existing component defaults
     - Sufficient for typical dev workloads (thousands of messages/sec)
     - Proven to work well in dev environments
   - **Cons**:
     - May be excessive for very light usage

2. **Lower Resources** (250m CPU, 512Mi Memory)
   - **Pros**: More resource-efficient
   - **Cons**: May impact performance, message ingestion rate limits

3. **Higher Resources** (1 CPU, 2Gi Memory)
   - **Pros**: Better performance headroom
   - **Cons**: Wasteful for typical dev scenarios, slower startup

**Decision**: CPU 500m, Memory 1Gi, Storage 8Gi

**Rationale**:
- Consistency with MySQL resource defaults
- Provides good balance between performance and resource efficiency
- Handles typical development workloads (10,000+ messages/sec)
- Users can override via configuration if needed
- Aligns with Bitnami chart recommended values for single-node setup

### 7. Persistence Strategy

**Question**: Should RabbitMQ data persistence be enabled by default?

**Research Findings**:

RabbitMQ persistence affects:
- **Durable Queues**: Queue definitions survive restarts
- **Persistent Messages**: Message content written to disk
- **Configuration**: Virtual hosts, users, policies persist across restarts

**Persistence Trade-offs**:
- **Enabled**: 
  - Pro: Data survives pod restarts, useful for testing
  - Con: Slower startup (~30-60s additional), disk I/O overhead
- **Disabled**:
  - Pro: Fast startup (~60s), clean state each restart
  - Con: All data lost on pod termination

**Options Evaluated**:

1. **Disabled by Default, Optional Enable** (SELECTED)
   - Persistence disabled by default for fast iteration
   - Users can enable via configuration: `persistence.enabled: true`
   - **Pros**:
     - Fast startup for ephemeral dev environments
     - Clean state for each test run (predictable)
     - Opt-in for stateful testing scenarios
   - **Cons**:
     - Need to remember to enable if persistence required

2. **Enabled by Default**
   - All deployments use persistent volumes
   - **Pros**: Data preserved automatically
   - **Cons**: Slower startup, more complex cleanup

3. **No Persistence Option**
   - Always ephemeral
   - **Pros**: Simplest approach
   - **Cons**: Inflexible, can't test stateful scenarios

**Decision**: Disabled by default, configurable to enable

**Rationale**:
- Matches MySQL persistence pattern
- Optimizes for common dev workflow (fast iteration)
- Provides flexibility for stateful testing when needed
- Reduces storage requirements for default installations
- Configuration example:

```yaml
components:
  rabbitmq:
    persistence:
      enabled: false  # Default
      size: "8Gi"     # Used when enabled
```

### 8. Chart Version Selection

**Question**: Which Bitnami RabbitMQ Helm chart version should be used?

**Research Findings**:

Bitnami RabbitMQ chart versions:
- **11.x**: Latest stable, RabbitMQ 3.11+, Kubernetes 1.21+
- **10.x**: Previous stable, RabbitMQ 3.10, older Kubernetes support
- **9.x**: Legacy, RabbitMQ 3.9, deprecated

**Compatibility Matrix**:
| Chart Version | RabbitMQ Version | Kubernetes Min | Status |
|---------------|------------------|----------------|--------|
| 11.x | 3.11+ | 1.21+ | Current |
| 10.x | 3.10 | 1.19+ | Supported |
| 9.x | 3.9 | 1.16+ | Deprecated |

**Decision**: Use chart version 11.x (latest stable)

**Rationale**:
- RabbitMQ 3.11+ includes important features and security fixes
- Compatible with existing kindenv Kubernetes requirements (1.21+)
- Bitnami maintains active support for 11.x series
- Default configuration: `chartVersion: "11.0.0"` (example, actual version may vary)

## Best Practices Research

### RabbitMQ Development Environment Best Practices

1. **Management Plugin**: Always enable for development visibility
2. **Virtual Hosts**: Use default "/" for single-app, custom vhosts for isolation
3. **Persistence**: Disable for fast iteration, enable for stateful integration tests
4. **Resource Limits**: Set appropriate limits to prevent resource exhaustion
5. **Health Checks**: Implement both liveness and readiness probes
6. **Connection Pooling**: Applications should use connection pooling for efficiency

### Helm Chart Configuration Best Practices

1. **Image Registry**: Configure bitnamilegacy registry for ECR compatibility
2. **Service Type**: Use NodePort for dev, ClusterIP for production
3. **Replica Count**: Single replica (1) for development environments
4. **Clustering**: Disable for single-node setups to reduce complexity
5. **Plugins**: Enable only required plugins (management, prometheus)
6. **Security**: Set strong passwords, rotate erlang cookie periodically

### Integration Pattern Best Practices

1. **Configuration Validation**: Validate all inputs before Helm installation
2. **Error Handling**: Provide clear error messages with troubleshooting hints
3. **Status Reporting**: Include both infrastructure status (pods) and application status (RabbitMQ health)
4. **Idempotency**: Installation should be safe to run multiple times
5. **Cleanup**: Properly uninstall Helm release and clean up resources on stop

## Technology Stack Decisions

| Component | Selected Technology | Rationale |
|-----------|-------------------|-----------|
| Deployment Method | Bitnami Helm Chart 11.x | Consistency, maintenance, ECR support |
| Service Exposure | NodePort + Kind Port Mapping | Localhost access, persistent |
| Secret Management | Kubernetes Secrets | Secure, follows existing pattern |
| Image Registry | bitnamilegacy (Harbor/ECR) | ECR compatibility requirement |
| Configuration Format | YAML (kindenv.yaml) | Consistency with existing components |
| Validation | Go validation functions | Type-safe, reusable, testable |
| Logging | uber-go/zap structured logging | Existing standard in codebase |

## Alternatives Considered and Rejected

### 1. RabbitMQ Cluster Operator
- **Why Rejected**: Too complex for single-node dev environments, additional CRDs, not aligned with Helm-based approach

### 2. Docker Compose Integration
- **Why Rejected**: Requires Docker Compose installation, doesn't integrate with Kind cluster, separate from existing patterns

### 3. Manual kubectl Manifests
- **Why Rejected**: High maintenance burden, no standardized configuration, upgrade path unclear

### 4. StatefulSet Direct Deployment
- **Why Rejected**: Reinvents wheel, Bitnami chart provides production-ready configuration

## Implementation Complexity Assessment

**Complexity Level**: Medium (Similar to MySQL integration)

**Known Complexities**:
1. **Dual Port Exposure**: Managing two NodePort services (AMQP + Management UI)
2. **Virtual Host Configuration**: Validating and configuring virtual host names
3. **Erlang Cookie Management**: Generating and storing erlang cookie securely
4. **Health Checks**: Implementing comprehensive health checks for both AMQP and Management API
5. **Status Reporting**: Displaying connection info for both protocols

**Mitigation Strategies**:
1. Follow MySQL integration pattern closely
2. Use table-driven tests for validation logic
3. Implement clear error messages with troubleshooting steps
4. Create comprehensive quickstart guide with examples
5. Add integration tests for both AMQP and Management UI connectivity

## Open Questions Resolved

1. **Q**: Should we support RabbitMQ clustering?
   **A**: No, out of scope for initial implementation. Single-node sufficient for dev environments. Clustering can be added in future enhancement.

2. **Q**: Should we pre-create exchanges and queues?
   **A**: No, leave to applications to define their messaging topology. Provides maximum flexibility.

3. **Q**: Should we support RabbitMQ plugins (beyond management)?
   **A**: Not initially. Management plugin enabled by default. Additional plugins can be configured manually via Helm values if needed.

4. **Q**: Should we integrate with Prometheus for metrics?
   **A**: Out of scope for this feature. Can be added as future enhancement.

5. **Q**: Should we support AMQP over TLS?
   **A**: No, not needed for local development environment. Plain AMQP sufficient.

## References

- [Bitnami RabbitMQ Helm Chart Documentation](https://github.com/bitnami/charts/tree/main/bitnami/rabbitmq)
- [RabbitMQ Official Documentation](https://www.rabbitmq.com/documentation.html)
- [RabbitMQ Management Plugin Guide](https://www.rabbitmq.com/management.html)
- [RabbitMQ Virtual Hosts Documentation](https://www.rabbitmq.com/vhosts.html)
- [Kubernetes NodePort Services](https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport)
- [Kind Extra Port Mappings](https://kind.sigs.k8s.io/docs/user/configuration/#extra-port-mappings)
- [MySQL Integration Reference](../001-mysql8-kindenv/research.md) - Pattern reference

## Research Completion

**Status**: ✅ Complete  
**All NEEDS CLARIFICATION items resolved**: Yes  
**Ready for Phase 1 (Design)**: Yes

All research questions have been answered with clear decisions and rationale. The implementation approach is well-defined and follows established patterns from the MySQL integration.
