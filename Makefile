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

CMDS=rcloneplugin
DEPLOY_FOLDER = ./deploy
CMDS=rcloneplugin
PKG = github.com/veloxpack/csi-driver-rclone
GINKGO_FLAGS = -ginkgo.v
GO111MODULE = on
GOPATH ?= $(shell go env GOPATH)
## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint
GOBIN ?= $(GOPATH)/bin
DOCKER_CLI_EXPERIMENTAL = enabled

# Architecture configuration - defaults to host architecture
ARCH ?= $(shell go env GOARCH)

export GOPATH GOBIN GO111MODULE DOCKER_CLI_EXPERIMENTAL

GIT_COMMIT = $(shell git rev-parse HEAD)
BUILD_DATE = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
IMAGE_VERSION ?= v4.13.0
LDFLAGS = -X ${PKG}/pkg/rclone.driverVersion=${IMAGE_VERSION} -X ${PKG}/pkg/rclone.gitCommit=${GIT_COMMIT} -X ${PKG}/pkg/rclone.buildDate=${BUILD_DATE}
EXT_LDFLAGS = -s -w -extldflags "-static"
# Use a custom version for E2E tests if we are testing in CI
ifdef CI
ifndef PUBLISH
override IMAGE_VERSION := e2e-$(GIT_COMMIT)
endif
endif
IMAGENAME ?= rcloneplugin
REGISTRY ?= andyzhangx
REGISTRY_NAME ?= $(shell echo $(REGISTRY) | sed "s/.azurecr.io//g")
IMAGE_TAG = $(REGISTRY)/$(IMAGENAME):$(IMAGE_VERSION)
IMAGE_TAG_LATEST = $(REGISTRY)/$(IMAGENAME):latest

E2E_HELM_OPTIONS ?= --set image.rclone.repository=$(REGISTRY)/$(IMAGENAME) --set image.rclone.tag=$(IMAGE_VERSION) --set image.rclone.pullPolicy=Always --set feature.enableInlineVolume=true --set externalSnapshotter.enabled=true --set controller.runOnControlPlane=true
E2E_HELM_OPTIONS += ${EXTRA_HELM_OPTIONS}

# Output type of docker buildx build
OUTPUT_TYPE ?= docker

ALL_ARCH.linux = arm64 amd64 ppc64le
ALL_OS_ARCH = linux-arm64 linux-arm-v7 linux-amd64 linux-ppc64le
GOLANGCI_LINT_VERSION ?=  v2.5.0

.EXPORT_ALL_VARIABLES:

all: rclone

.PHONY: unit-test
unit-test:
	go test -covermode=count -coverprofile=profile.cov ./pkg/... -v

.PHONY: local-build-push
local-build-push: rclone
	docker build -t $(LOCAL_USER)/rcloneplugin:latest .
	docker push $(LOCAL_USER)/rcloneplugin

.PHONY: rclone
rclone:
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -a -ldflags "${LDFLAGS} ${EXT_LDFLAGS}" -o bin/${ARCH}/rcloneplugin ./cmd/rcloneplugin

.PHONY: rclone-armv7
rclone-armv7:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -a -ldflags "${LDFLAGS} ${EXT_LDFLAGS}" -o bin/arm/v7/rcloneplugin ./cmd/rcloneplugin

.PHONY: container-build
container-build:
	docker buildx build --pull --output=type=$(OUTPUT_TYPE) --platform="linux/$(ARCH)" \
		--provenance=false --sbom=false \
		-t $(IMAGE_TAG)-linux-$(ARCH) --build-arg ARCH=$(ARCH) .

.PHONY: container-linux-armv7
container-linux-armv7:
	docker buildx build --pull --output=type=$(OUTPUT_TYPE) --platform="linux/arm/v7" \
		--provenance=false --sbom=false \
		-t $(IMAGE_TAG)-linux-arm-v7 --build-arg ARCH=arm/v7 .

.PHONY: container
container:
	docker buildx rm container-builder || true
	docker buildx create --use --name=container-builder
	# enable qemu for arm64 build
	# https://github.com/docker/buildx/issues/464#issuecomment-741507760
	docker run --privileged --rm tonistiigi/binfmt --uninstall qemu-aarch64
	docker run --rm --privileged tonistiigi/binfmt --install all
	for arch in $(ALL_ARCH.linux); do \
		ARCH=$${arch} $(MAKE) rclone; \
		ARCH=$${arch} $(MAKE) container-build; \
	done
	$(MAKE) rclone-armv7
	$(MAKE) container-linux-armv7

.PHONY: push
push:
ifdef CI
	docker manifest create --amend $(IMAGE_TAG) $(foreach osarch, $(ALL_OS_ARCH), $(IMAGE_TAG)-${osarch})
	docker manifest push --purge $(IMAGE_TAG)
	docker manifest inspect $(IMAGE_TAG)
else
	docker push $(IMAGE_TAG)
endif

.PHONY: push-latest
push-latest:
ifdef CI
	docker manifest create --amend $(IMAGE_TAG_LATEST) $(foreach osarch, $(ALL_OS_ARCH), $(IMAGE_TAG)-${osarch})
	docker manifest push --purge $(IMAGE_TAG_LATEST)
	docker manifest inspect $(IMAGE_TAG_LATEST)
else
	docker tag $(IMAGE_TAG) $(IMAGE_TAG_LATEST)
	docker push $(IMAGE_TAG_LATEST)
endif

.PHONY: install-helm
install-helm:
	curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash

.PHONY: e2e-bootstrap
e2e-bootstrap: install-helm
	OUTPUT_TYPE=registry $(MAKE) container push
	helm install csi-driver-rclone ./charts/latest/csi-driver-rclone --namespace default --wait --timeout=15m -v=5 --debug \
		${E2E_HELM_OPTIONS} \
		--set controller.logLevel=8 \
		--set node.logLevel=8

.PHONY: e2e-teardown
e2e-teardown:
	helm delete csi-driver-rclone --namespace default

.PHONY: e2e-test
e2e-test:
	if [ ! -z "$(EXTERNAL_E2E_TEST)" ]; then \
		bash ./test/external-e2e/run.sh;\
	else \
		go test -v -timeout=0 ./test/e2e ${GINKGO_FLAGS};\
	fi

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter
	$(GOLANGCI_LINT) run --no-config ./cmd/... ./pkg/...

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --no-config --fix ./cmd/... ./pkg/...

.PHONY: lint-config
lint-config: golangci-lint ## Verify golangci-lint linter configuration
	$(GOLANGCI_LINT) config verify

.PHONY: golangci-lint
golangci-lint: $(LOCALBIN) ## Download golangci-lint locally if necessary
	@if [ ! -f $(GOLANGCI_LINT) ]; then \
		echo "Downloading golangci-lint $(GOLANGCI_LINT_VERSION)"; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(LOCALBIN) $(GOLANGCI_LINT_VERSION); \
	fi
