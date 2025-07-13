#!/bin/bash

# Kubernetes Event Exporter Helm Chart Installation Script

set -e

NAMESPACE=${NAMESPACE:-monitoring}
RELEASE_NAME=${RELEASE_NAME:-event-exporter}
VALUES_FILE=${VALUES_FILE:-""}

echo "Installing Kubernetes Event Exporter..."
echo "Namespace: $NAMESPACE"
echo "Release Name: $RELEASE_NAME"

# Create namespace if it doesn't exist
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# Install or upgrade the chart
if [ -n "$VALUES_FILE" ]; then
    echo "Using values file: $VALUES_FILE"
    helm upgrade --install $RELEASE_NAME . \
        --namespace $NAMESPACE \
        --values $VALUES_FILE
else
    echo "Using default values"
    helm upgrade --install $RELEASE_NAME . \
        --namespace $NAMESPACE
fi

echo ""
echo "Installation complete!"
echo ""
echo "To check the status:"
echo "  kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=kubernetes-event-exporter"
echo ""
echo "To view logs:"
echo "  kubectl logs -n $NAMESPACE -l app.kubernetes.io/name=kubernetes-event-exporter"
echo ""
echo "To uninstall:"
echo "  helm uninstall $RELEASE_NAME -n $NAMESPACE" 