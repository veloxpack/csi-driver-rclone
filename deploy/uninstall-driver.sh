#!/bin/bash

# Copyright 2025 Veloxpack.io
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Example usage:
#   ./uninstall-driver.sh main
#   ./uninstall-driver.sh main metrics
#   ./uninstall-driver.sh main local
#   ./uninstall-driver.sh main local-metrics

set -euo pipefail

ver="main"
if [[ "$#" -gt 0 ]]; then
  ver="$1"
fi

repo="https://github.com/veloxpack/csi-driver-rclone//deploy"
if [[ "$#" -gt 1 && "$2" == *"local"* ]]; then
  echo "Using local deploy manifests..."
  repo="./deploy"
fi

echo "Uninstalling RCLONE CSI driver, version: $ver ..."

# Determine which overlay to use
if [[ "$#" -gt 1 && "$2" == *"metrics"* ]]; then
  overlay="overlays/metrics"
  echo "Detected metrics overlay..."
else
  overlay="overlays/default"
fi

# Delete manifests
if [[ "$repo" == "./deploy"* ]]; then
  echo "Deleting local manifests via Kustomize: $repo/$overlay"
  kubectl delete -k "$repo/$overlay" --ignore-not-found
else
  echo "Deleting remote manifests via Kustomize: $repo/$overlay?ref=$ver"
  kubectl delete -k "$repo/$overlay?ref=$ver" --ignore-not-found
fi

echo "RCLONE CSI driver uninstalled successfully."
