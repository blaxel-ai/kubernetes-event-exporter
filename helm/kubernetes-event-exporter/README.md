# kubernetes-event-exporter

A Helm chart for deploying Kubernetes Event Exporter to forward Kubernetes events to various sinks including AWS EventBridge.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+

## Development Prerequisites

- Python 3.6+ (for documentation generation)
- PyYAML (`pip install pyyaml`)

## Installation

### Add the repository (when published)

```bash
helm repo add kubernetes-event-exporter https://blaxel-ai.github.io/kubernetes-event-exporter
helm repo update
```

### Install from local directory

```bash
helm install event-exporter ./helm/kubernetes-event-exporter -n monitoring --create-namespace
```

### Install on EKS with Pod Identity (Recommended)

For EKS clusters, we recommend using Pod Identity for secure AWS authentication:

```bash
# See examples/eks-pod-identity-setup.md for complete setup
helm install event-exporter ./helm/kubernetes-event-exporter \
  -n monitoring --create-namespace \
  -f examples/values-eks-pod-identity.yaml
```

### Install with custom values

```bash
helm install event-exporter ./helm/kubernetes-event-exporter \
  -n monitoring \
  --create-namespace \
  -f my-values.yaml
```

## Configuration

### Basic Configuration

```yaml
# Set the cluster name
cluster:
  name: "my-cluster"

# Configure AWS credentials
aws:
  enabled: true
  region: "us-east-1"
  credentials:
    accessKeyId: "YOUR_ACCESS_KEY"
    secretAccessKey: "YOUR_SECRET_KEY"
    sessionToken: "OPTIONAL_SESSION_TOKEN"  # For temporary credentials
```

### Using EKS Pod Identity (Recommended for EKS)

```yaml
aws:
  enabled: true
  region: "us-east-1"
  # No credentials needed - uses Pod Identity
  
serviceAccount:
  create: true
  # No annotations needed for Pod Identity
```

See [eks-pod-identity-setup.md](examples/eks-pod-identity-setup.md) for detailed setup instructions.

### Using an existing AWS Secret

```yaml
aws:
  enabled: true
  region: "us-east-1"
  existingSecret: "my-aws-secret"
  secretKeys:
    accessKeyId: "aws-access-key-id"
    secretAccessKey: "aws-secret-access-key"
    sessionToken: "aws-session-token"  # Optional
```

### Configure Event Routing

```yaml
# Route all events to EventBridge and stdout
routes:
  - match:
      - receiver: "eventbridge"
  - match:
      - receiver: "stdout"

# Enable/disable receivers
receivers:
  eventbridge:
    enabled: true
  stdout:
    enabled: true
    deDot: true
```

### Resource Configuration

```yaml
resources:
  requests:
    memory: "128Mi"
    cpu: "100m"
  limits:
    memory: "256Mi"
    cpu: "500m"
```

### Using an existing ConfigMap

```yaml
existingConfigMap: "my-event-exporter-config"
```

## Values

| Parameter | Description | Default |
|-----------|-------------|---------|
| `affinity` | See values.yaml | `{}` |
| `annotations` | Deployment annotations | `{}` |
| `aws.credentials.accessKeyId` | See values.yaml | `""` |
| `aws.credentials.secretAccessKey` | See values.yaml | `""` |
| `aws.credentials.sessionToken` | See values.yaml | `""` |
| `aws.enabled` | See values.yaml | `true` |
| `aws.existingSecret` | Reference an existing secret | `""` |
| `aws.region` | See values.yaml | `"eu-west-1"` |
| `aws.secretKeys.accessKeyId` | See values.yaml | `"AWS_ACCESS_KEY_ID"` |
| `aws.secretKeys.secretAccessKey` | See values.yaml | `"AWS_SECRET_ACCESS_KEY"` |
| `aws.secretKeys.sessionToken` | See values.yaml | `"AWS_SESSION_TOKEN"` |
| `cluster.name` | The name of the service account to use. | `"event-exporter-demo"` |
| `config.kubeBurst` | See values.yaml | `500` |
| `config.kubeQPS` | See values.yaml | `100` |
| `config.logFormat` | See values.yaml | `"json"` |
| `config.logLevel` | See values.yaml | `"debug"` |
| `config.maxEventAgeSeconds` | See values.yaml | `60` |
| `config.metricsNamePrefix` | See values.yaml | `"event_exporter_"` |
| `eventbridge.detailType` | See values.yaml | `"ExecutionPlane Event"` |
| `eventbridge.enabled` | See values.yaml | `true` |
| `eventbridge.eventBusName` | See values.yaml | `"default"` |
| `eventbridge.source` | See values.yaml | `"executionPlane"` |
| `existingConfigMap` | ConfigMap name (if you want to use an existing one) | `""` |
| `extraEnvVars` | Additional environment variables | `[]` |
| `extraVolumeMounts` | - name: MY_ENV_VAR value: "my-value" Additional volume mounts | `[]` |
| `extraVolumes` | Additional volumes | `[]` |
| `fullnameOverride` | See values.yaml | `""` |
| `image.pullPolicy` | See values.yaml | `"IfNotPresent"` |
| `image.repository` | See values.yaml | `"ghcr.io/blaxel-ai/kubernetes-event-exporter"` |
| `image.tag` | See values.yaml | `"latest-preview"` |
| `imagePullSecrets` | See values.yaml | `[]` |
| `nameOverride` | See values.yaml | `""` |
| `nodeSelector` | See values.yaml | `{}` |
| `podAnnotations.prometheus.io/path` | See values.yaml | `"/metrics"` |
| `podAnnotations.prometheus.io/port` | See values.yaml | `"2112"` |
| `podAnnotations.prometheus.io/scrape` | See values.yaml | `"true"` |
| `podSecurityContext.runAsNonRoot` | See values.yaml | `true` |
| `podSecurityContext.seccompProfile.type` | See values.yaml | `"RuntimeDefault"` |
| `rbac.create` | Create ClusterRole and ClusterRoleBinding | `true` |
| `receivers.eventbridge.enabled` | See values.yaml | `true` |
| `receivers.stdout.deDot` | See values.yaml | `true` |
| `receivers.stdout.enabled` | See values.yaml | `true` |
| `replicaCount` | Default values for kubernetes-event-exporter. | `1` |
| `routes` | Route configuration | See values.yaml |
| `securityContext.allowPrivilegeEscalation` | See values.yaml | `false` |
| `securityContext.capabilities.drop` | See values.yaml | See values.yaml |
| `serviceAccount.annotations` | Annotations to add to the service account | `{}` |
| `serviceAccount.create` | Specifies whether a service account should be created | `true` |
| `serviceAccount.name` | The name of the service account to use. | `""` |
| `tolerations` | See values.yaml | `[]` |
## Uninstallation

```bash
helm uninstall event-exporter -n monitoring
```

## Troubleshooting

### Check pod status
```bash
kubectl get pods -n monitoring -l app.kubernetes.io/name=kubernetes-event-exporter
```

### View logs
```bash
kubectl logs -n monitoring -l app.kubernetes.io/name=kubernetes-event-exporter
```

### Verify AWS credentials
```bash
kubectl get secret -n monitoring <release-name>-kubernetes-event-exporter-aws -o yaml
```

## Security Considerations

1. **AWS Credentials**: 
   - **EKS**: Use Pod Identity (recommended) or IRSA - no static credentials needed
   - **Other**: Use Kubernetes secrets, never hardcode credentials
   - See [eks-pod-identity-setup.md](examples/eks-pod-identity-setup.md) for EKS setup
2. **RBAC**: The chart creates a ClusterRole with read access to all resources. Review and adjust permissions as needed
3. **Network Policies**: Consider implementing network policies to restrict egress traffic

## Development

### Update Values Documentation

After modifying `values.yaml`, regenerate the documentation:
```bash
make docs
# or
./update-docs.sh
```

### Test the chart
```bash
helm lint ./helm/kubernetes-event-exporter
helm template event-exporter ./helm/kubernetes-event-exporter
```

### Dry run installation
```bash
helm install event-exporter ./helm/kubernetes-event-exporter --dry-run --debug
``` 