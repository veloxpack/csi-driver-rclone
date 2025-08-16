#!/bin/bash

# Package the Helm chart
set -e

CHART_DIR="latest/csi-driver-rclone"
CHART_NAME="csi-driver-rclone"
VERSION="v0.0.0"

echo "Packaging Helm chart..."

# Create the package
helm package $CHART_DIR

# Move the package to the charts directory
mv ${CHART_NAME}-${VERSION}.tgz charts/

echo "Chart packaged successfully: charts/${CHART_NAME}-${VERSION}.tgz"

# Update the index
echo "Updating chart index..."
helm repo index charts/ --url https://veloxpack.github.io/csi-driver-rclone/charts

echo "Chart index updated: charts/index.yaml"
