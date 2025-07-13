# Publishing Helm Charts

This document describes how the Helm chart is automatically published via GitHub Actions.

## Overview

The Helm chart is automatically packaged and published to GitHub Pages whenever:
1. Code is pushed to the `main` branch (development versions)
2. Code is pushed to the `develop` branch (development versions with `-dev-` prefix)
3. A new release is created on GitHub (stable versions)

## How It Works

### 1. Automatic Publishing

The Helm chart publishing is integrated into the main CI/CD workflows:

- **`.github/workflows/build.yml`**: 
  - On push to main: creates a development version with the commit SHA (e.g., `0.1.0-abc1234`)
  - On push to develop: creates a development version with `-dev-` prefix (e.g., `0.1.0-dev-abc1234`)
- **`.github/workflows/release.yml`**: On release, creates a stable version matching the release tag (e.g., `1.0.0`)

### 2. GitHub Pages Setup

The charts are hosted on GitHub Pages at:
```
https://<owner>.github.io/<repo-name>
```

## Initial Setup (One-time)

The gh-pages branch is automatically created on the first workflow run. You only need to:

1. **Push to main or develop branch**:
   - The workflow will automatically create the gh-pages branch
   - It will initialize the Helm repository structure

2. **Enable GitHub Pages** (after the first workflow run):
   - Go to Settings → Pages
   - Source: Deploy from a branch
   - Branch: `gh-pages` / `/ (root)`
   - Click Save

3. **Wait for deployment**:
   - GitHub Pages takes a few minutes to activate
   - Check the Actions tab to see the workflow progress

## Publishing a New Version

### Development Versions (Automatic)

Push changes to the `helm/` directory on either `main` or `develop`:

```bash
git add helm/
git commit -m "feat: Update Helm chart"

# For main branch versions (0.1.0-abc1234)
git push origin main

# For develop branch versions (0.1.0-dev-abc1234)
git push origin develop
```

### Stable Versions (Via Release)

1. Create a new release on GitHub:
   - Go to Releases → Create a new release
   - Tag version: `v1.0.0` (must start with 'v')
   - Release title: `v1.0.0`
   - Describe the changes
   - Publish release

2. The workflow will automatically:
   - Update Chart.yaml with the version
   - Package the chart
   - Publish to the Helm repository

## Using the Published Chart

Once published, users can install the chart:

```bash
# Add the repository
helm repo add kubernetes-event-exporter https://<owner>.github.io/<repo-name>
helm repo update

# Search for charts
helm search repo kubernetes-event-exporter

# Install the chart
helm install event-exporter kubernetes-event-exporter/kubernetes-event-exporter
```

## Troubleshooting

### Chart not appearing in repository

1. Check the workflow run in Actions tab
2. Ensure GitHub Pages is enabled
3. Wait 5-10 minutes for GitHub Pages to update
4. Check https://<owner>.github.io/<repo-name>/index.yaml

### Version conflicts

The workflow automatically handles versioning:
- Release versions override any existing version
- Main branch builds append commit SHA to prevent conflicts

### GitHub Pages not working

1. Ensure the `gh-pages` branch exists
2. Check Settings → Pages is configured correctly
3. Look for deployment status in Actions → Deployments

## Version Strategy

- **Stable releases**: Match git tags (e.g., `v1.0.0` → `1.0.0`)
- **Main branch builds**: Base version + commit SHA (e.g., `0.1.0-abc1234`)
- **Develop branch builds**: Base version + `-dev-` + commit SHA (e.g., `0.1.0-dev-abc1234`)

## Security Notes

- AWS credentials should NEVER be committed to the Helm chart
- Use Kubernetes secrets or external secret management
- The chart supports both methods - see values.yaml 