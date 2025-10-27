#!/bin/bash

# Copyright 2025 Veloxpack
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

set -xe

PROJECT_ROOT=$(git rev-parse --show-toplevel)
DRIVER="rclone"

install_ginkgo () {
    go install github.com/onsi/ginkgo/v2/ginkgo@latest
}

setup_e2e_binaries() {
    # download k8s external e2e binary
    curl -sL https://dl.k8s.io/release/v1.31.0/kubernetes-test-linux-amd64.tar.gz --output e2e-tests.tar.gz
    tar -xvf e2e-tests.tar.gz && rm e2e-tests.tar.gz

    export EXTRA_HELM_OPTIONS="--set driver.name=$DRIVER.csi.veloxpack.io --set controller.name=csi-$DRIVER-controller --set node.name=csi-$DRIVER-node --set feature.enableInlineVolume=true"

    # test on rclone driver
    sed -i.bak "s/rclone.csi.veloxpack.io/$DRIVER.csi.veloxpack.io/g" deploy/example/storageclass.yaml

    # install csi driver
    mkdir -p /tmp/csi
    cp deploy/example/storageclass.yaml /tmp/csi/storageclass.yaml

    # Deploy rclone CSI driver
    make e2e-bootstrap

    # Setup test backend (MinIO or similar for testing)
    # make install-test-backend
}

print_logs() {
    echo "print out driver logs ..."
    kubectl logs -l app=csi-rclone-controller -n veloxpack --tail=100 || true
    kubectl logs -l app=csi-rclone-node -n veloxpack --tail=100 || true
}

install_ginkgo
setup_e2e_binaries
trap print_logs EXIT

ginkgo -p --progress --v -focus="External.Storage.*$DRIVER.csi.veloxpack.io" \
       -skip='\[Disruptive\]|\[Feature:VolumeSnapshotDataSource\]|\[Feature:ExpandInUsePersistentVolumes\]|volume-expand|volume expansion|snapshot' kubernetes/test/bin/e2e.test  -- \
       -storage.testdriver=$PROJECT_ROOT/test/external-e2e/testdriver.yaml \
       --kubeconfig=$KUBECONFIG
