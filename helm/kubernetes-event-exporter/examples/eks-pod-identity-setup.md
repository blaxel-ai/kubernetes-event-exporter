# EKS Pod Identity Setup for Kubernetes Event Exporter

This guide explains how to set up Kubernetes Event Exporter on Amazon EKS using Pod Identity for AWS authentication.

## Why Pod Identity?

EKS Pod Identity is the recommended approach for AWS authentication because:
- ✅ No static credentials needed
- ✅ Automatic credential rotation
- ✅ Simpler than IRSA (no service account annotations)
- ✅ Better security through temporary credentials
- ✅ Works seamlessly with AWS SDK

## Prerequisites

1. EKS cluster with Pod Identity enabled
2. AWS CLI configured with appropriate permissions
3. Helm 3.x installed

## Setup Steps

### 1. Create IAM Policy

Create an IAM policy for EventBridge access:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "events:PutEvents"
      ],
      "Resource": "arn:aws:events:*:*:event-bus/default"
    }
  ]
}
```

Save this as `eventbridge-policy.json` and create the policy:

```bash
aws iam create-policy \
  --policy-name KubernetesEventExporterPolicy \
  --policy-document file://eventbridge-policy.json
```

### 2. Create IAM Role

Create an IAM role for the Pod Identity:

```bash
# Create the role with the Pod Identity trust policy
aws iam create-role \
  --role-name kubernetes-event-exporter \
  --assume-role-policy-document '{
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Principal": {
          "Service": "pods.eks.amazonaws.com"
        },
        "Action": [
          "sts:AssumeRole",
          "sts:TagSession"
        ]
      }
    ]
  }'

# Attach the policy to the role
aws iam attach-role-policy \
  --role-name kubernetes-event-exporter \
  --policy-arn arn:aws:iam::$(aws sts get-caller-identity --query Account --output text):policy/KubernetesEventExporterPolicy
```

### 3. Install the Helm Chart

First, add the repository (once it's published):

```bash
helm repo add kubernetes-event-exporter https://blaxel-ai.github.io/kubernetes-event-exporter
helm repo update
```

Install the chart:

```bash
helm install event-exporter kubernetes-event-exporter/kubernetes-event-exporter \
  --namespace monitoring \
  --create-namespace \
  -f values-eks-pod-identity.yaml
```

### 4. Create Pod Identity Association

After the chart is installed, create the Pod Identity association:

```bash
# Get the service account name
SA_NAME=$(kubectl get sa -n monitoring -l app.kubernetes.io/name=kubernetes-event-exporter -o jsonpath='{.items[0].metadata.name}')

# Create the association
aws eks create-pod-identity-association \
  --cluster-name YOUR_CLUSTER_NAME \
  --namespace monitoring \
  --service-account $SA_NAME \
  --role-arn arn:aws:iam::$(aws sts get-caller-identity --query Account --output text):role/kubernetes-event-exporter
```

### 5. Restart the Pod

The pod needs to be restarted to pick up the Pod Identity:

```bash
kubectl rollout restart deployment/event-exporter-kubernetes-event-exporter -n monitoring
```

## Verification

### Check Pod Identity is Working

```bash
# Check the pod has AWS credentials
kubectl exec -n monitoring deployment/event-exporter-kubernetes-event-exporter -- env | grep AWS_

# You should see:
# AWS_REGION=us-east-1
# AWS_CONTAINER_CREDENTIALS_FULL_URI=http://...
# AWS_CONTAINER_AUTHORIZATION_TOKEN_FILE=/var/run/secrets/pods.eks.amazonaws.com/serviceaccount/eks-pod-identity-token
```

### Check Logs

```bash
kubectl logs -n monitoring -l app.kubernetes.io/name=kubernetes-event-exporter
```

## Troubleshooting

### Pod Identity Not Working

1. **Check Pod Identity Agent**:
   ```bash
   kubectl get pods -n kube-system | grep eks-pod-identity-agent
   ```

2. **Verify Association**:
   ```bash
   aws eks list-pod-identity-associations \
     --cluster-name YOUR_CLUSTER_NAME \
     --namespace monitoring
   ```

3. **Check IAM Role Trust Policy**:
   ```bash
   aws iam get-role --role-name kubernetes-event-exporter
   ```

### Events Not Reaching EventBridge

1. **Check IAM Permissions**:
   - Verify the role has EventBridge PutEvents permission
   - Check the resource ARN matches your event bus

2. **Enable Debug Logging**:
   ```yaml
   config:
     logLevel: debug
   ```

3. **Test Manually**:
   ```bash
   # Exec into the pod and test AWS access
   kubectl exec -it -n monitoring deployment/event-exporter-kubernetes-event-exporter -- sh
   # Inside the pod:
   aws events put-events --entries Source=test,DetailType=test,Detail='{}'
   ```

## Comparison with Other Methods

| Method | Pros | Cons |
|--------|------|------|
| **Pod Identity** | Simple setup, No SA annotations, Native EKS feature | Requires EKS 1.24+ |
| **IRSA** | Widely supported, Mature | Complex setup, Requires SA annotations |
| **Static Credentials** | Works anywhere | Security risk, Manual rotation |
| **Instance Profile** | No pod config needed | Over-permissive, All pods get access |

## Security Best Practices

1. **Use Least Privilege**: Only grant PutEvents permission to the specific event bus
2. **Separate Roles**: Use different roles for different environments
3. **Enable CloudTrail**: Monitor EventBridge API calls
4. **Resource Restrictions**: Limit the role to specific event buses:
   ```json
   "Resource": "arn:aws:events:us-east-1:123456789012:event-bus/my-bus"
   ```

## Example Values File

See [values-eks-pod-identity.yaml](./values-eks-pod-identity.yaml) for a complete example configuration.
