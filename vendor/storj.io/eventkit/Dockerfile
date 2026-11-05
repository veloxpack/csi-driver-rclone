FROM --platform=$BUILDPLATFORM golang:1.26 AS build
WORKDIR /usr/src/eventkit
COPY . .
ARG TARGETOS TARGETARCH
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -o /out/ ./eventkitd ./eventkitd-bigquery

FROM gcr.io/distroless/static-debian12 AS eventkitd
COPY --from=build /out/eventkitd /usr/local/bin/eventkitd
ENTRYPOINT ["/usr/local/bin/eventkitd"]

FROM gcr.io/distroless/static-debian12 AS eventkitd-bigquery
COPY --from=build /out/eventkitd-bigquery /usr/local/bin/eventkitd-bigquery
ENTRYPOINT ["/usr/local/bin/eventkitd-bigquery"]
