.PHONY: help docs lint package install test

# Default target
help:
	@echo "Available targets:"
	@echo "  docs        - Generate values documentation from values.yaml"
	@echo "  lint        - Run helm lint"
	@echo "  package     - Package the Helm chart"
	@echo "  install     - Install the chart locally"
	@echo "  test        - Run helm template to test rendering"

# Generate documentation
docs:
	@echo "Generating values documentation..."
	@./scripts/generate-values-docs.py

# Lint the chart
lint:
	helm lint .

# Package the chart
package:
	helm package .

# Install the chart
install:
	helm upgrade --install event-exporter . \
		--namespace monitoring \
		--create-namespace

# Test template rendering
test:
	helm template event-exporter . --namespace monitoring 