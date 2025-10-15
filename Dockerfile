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

# Build the manager binary
FROM golang:1.25 AS builder
ARG TARGETOS
ARG TARGETARCH
ARG RCLONE_BACKEND_MODE=all

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/rcloneplugin/ cmd/rcloneplugin
COPY pkg/ pkg/

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
    go build -a \
    -ldflags="-s -w -extldflags '-static'" \
    -trimpath \
    -tags "netgo ${RCLONE_BACKEND_MODE}" \
    -o rcloneplugin \
    cmd/rcloneplugin/main.go

# Use alpine as base image to package the rcloneplugin binary with rclone
FROM alpine:3.22.2
WORKDIR /

# Install required dependencies
RUN apk add --no-cache ca-certificates fuse3 tzdata && \
    rm -rf /var/cache/apk/* /tmp/*

COPY --from=builder /workspace/rcloneplugin .

ENTRYPOINT ["/rcloneplugin"]
