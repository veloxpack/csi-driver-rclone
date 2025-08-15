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

set -euo pipefail

ver="main"
if [[ "$#" -gt 0 ]]; then
  ver="$1"
fi

repo="https://raw.githubusercontent.com/veloxpack/csi-driver-rclone/$ver/deploy"
if [[ "$#" -gt 1 ]]; then
  if [[ "$2" == *"local"* ]]; then
    echo "use local deploy"
    repo="./deploy"
  fi
fi

if [ $ver != "main" ]; then
  repo="$repo/$ver"
fi

echo "Installing RCLONE CSI driver, version: $ver ..."
kubectl apply -f $repo/rbac-csi-rclone.yaml
kubectl apply -f $repo/csi-rclone-driverinfo.yaml
kubectl apply -f $repo/csi-rclone-controller.yaml
kubectl apply -f $repo/csi-rclone-node.yaml

echo 'RCLONE CSI driver installed successfully.'
