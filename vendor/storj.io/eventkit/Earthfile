VERSION 0.6
FROM golang:1.26
WORKDIR /go/eventkit

lint:
    RUN --mount=type=cache,target=/root/.cache/go-build \
        --mount=type=cache,target=/go/pkg/mod \
        go install honnef.co/go/tools/cmd/staticcheck@2025.1
    RUN --mount=type=cache,target=/root/.cache/go-build \
        --mount=type=cache,target=/go/pkg/mod \
        go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.3
    RUN --mount=type=cache,target=/root/.cache/go-build \
        --mount=type=cache,target=/go/pkg/mod \
        go install github.com/storj/ci/check-cross-compile@latest
    COPY . .

    RUN golangci-lint run --timeout=3m
    RUN staticcheck ./...
    RUN check-cross-compile ./...

test:
   COPY . .
   RUN --mount=type=cache,target=/root/.cache/go-build \
       --mount=type=cache,target=/go/pkg/mod \
       go test -race ./...

check-format:
   COPY . .
   RUN bash -c 'mkdir build || true'
   RUN bash -c '[[ $(git status --short) == "" ]] || (echo "Before formatting, please commit all your work!!! (Formatter will format only last commit)" && exit -1)'
   RUN git show --name-only --pretty=format: | grep ".go" | xargs -n1 gofmt -s -w
   RUN git diff > build/format.patch
   SAVE ARTIFACT build/format.patch

format:
   LOCALLY
   COPY +check-format/format.patch build/format.patch
   RUN git apply --allow-empty build/format.patch
   RUN git status
