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
IMAGE_VERSION ?= latest
RCLONE_VERSION = $(shell grep "github.com/rclone/rclone" go.mod | awk '{print $$2}' | sed 's/v//')
LDFLAGS = -X ${PKG}/pkg/rclone.driverVersion=${IMAGE_VERSION} -X ${PKG}/pkg/rclone.gitCommit=${GIT_COMMIT} -X ${PKG}/pkg/rclone.buildDate=${BUILD_DATE} -X ${PKG}/pkg/rclone.rcloneVersion=${RCLONE_VERSION}
EXT_LDFLAGS = -s -w -extldflags "-static"
# Use a custom version for E2E tests if we are testing in CI
ifdef CI
ifndef PUBLISH
override IMAGE_VERSION := e2e-$(GIT_COMMIT)
endif
endif
IMAGENAME ?= csi-driver-rclone
REGISTRY ?= ghcr.io/veloxpack
IMAGE_TAG = $(REGISTRY)/$(IMAGENAME):$(IMAGE_VERSION)
IMAGE_TAG_LATEST = $(REGISTRY)/$(IMAGENAME):latest

HELM_CHARTS_PATH = charts

# Output type of docker buildx build
OUTPUT_TYPE ?= docker

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

.PHONY: apply-patches
apply-patches:
	@echo "Applying vendor patches..."
	@bash scripts/apply-vendor-patches.sh

.PHONY: vendor-sync
vendor-sync:
	@echo "Syncing vendor directory..."
	go mod vendor
	@$(MAKE) apply-patches

.PHONY: rclone
rclone: apply-patches
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -a -ldflags "${LDFLAGS} ${EXT_LDFLAGS}" -mod vendor -o bin/${ARCH}/rcloneplugin ./cmd/rcloneplugin

.PHONY: container-build
container-build:
	docker buildx build --pull --load \
		--platform linux/$(ARCH) \
		--provenance=false --sbom=false \
		-t $(IMAGE_TAG) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) --build-arg BUILD_DATE=$(BUILD_DATE) \
		--build-arg DRIVER_VERSION=$(IMAGE_VERSION) .

.PHONY: container-build-multiarch
container-build-multiarch:
	docker buildx build --pull --output=type=image \
		--platform linux/amd64,linux/arm64 \
		--provenance=false --sbom=false \
		-t $(IMAGE_TAG) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) --build-arg BUILD_DATE=$(BUILD_DATE) \
		--build-arg DRIVER_VERSION=$(IMAGE_VERSION) .

.PHONY: container
container:
	@MULTIARCH_VALUE=$${MULTIARCH:-true}; \
	if [ "$$MULTIARCH_VALUE" = "true" ]; then \
		echo "🏗️ Building multi-architecture image (default)..."; \
	else \
		echo "🏗️ Building single-architecture image..."; \
	fi; \
	docker buildx rm container-builder || true; \
	docker buildx create --use --name=container-builder; \
	# enable qemu for arm64 build
	docker run --privileged --rm tonistiigi/binfmt --uninstall qemu-aarch64; \
	docker run --rm --privileged tonistiigi/binfmt --install all; \
	if [ "$$MULTIARCH_VALUE" = "true" ]; then \
		$(MAKE) container-build-multiarch; \
	else \
		$(MAKE) container-build; \
	fi

.PHONY: push
push:
	docker push $(IMAGE_TAG)

.PHONY: push-latest
push-latest:
	docker tag $(IMAGE_TAG) $(IMAGE_TAG_LATEST)
	docker push $(IMAGE_TAG_LATEST)

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix

.PHONY: lint-config
lint-config: golangci-lint ## Verify golangci-lint linter configuration
	$(GOLANGCI_LINT) config verify

.PHONY: golangci-lint
golangci-lint: $(LOCALBIN) ## Download golangci-lint locally if necessary
	@if [ ! -f $(GOLANGCI_LINT) ]; then \
		echo "Downloading golangci-lint $(GOLANGCI_LINT_VERSION)"; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(LOCALBIN) $(GOLANGCI_LINT_VERSION); \
	fi

.PHONY: helm-lint
helm-lint:
	helm lint ${HELM_CHARTS_PATH} --strict

.PHONY: helm-validate
helm-validate:
	helm template test ${HELM_CHARTS_PATH}

.PHONY: minikube-start
minikube-start:
	minikube start \
		--memory=4096 \
		--cpus=2 \
		--disk-size=20g \
		--kubernetes-version=1.34.0

.PHONY: minikube-stop
minikube-stop:
	minikube stop

.PHONY: minikube-delete
minikube-delete:
	minikube delete
