build:
	go build

test:
	go test -cpu 4 -v -race

# Get the build dependencies
build_dep:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.5.0

# Do source code quality checks
check:
	golangci-lint run
	go vet ./...
