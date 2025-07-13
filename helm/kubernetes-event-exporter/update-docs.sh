#!/bin/bash

# Script to update Helm chart documentation
# Run this after making changes to values.yaml

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "🔄 Updating Helm chart documentation..."

# Change to chart directory
cd "$SCRIPT_DIR"

# Run documentation generation
if command -v make >/dev/null 2>&1; then
    make docs
else
    # Direct script execution
    ./scripts/generate-values-docs.py
fi

echo ""
echo "✅ Documentation updated!"
echo ""
echo "Please review the changes and commit them:"
echo "  git add README.md"
echo "  git commit -m 'docs: Update Helm values documentation'" 