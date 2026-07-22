FROM golang:1.23-alpine AS build

RUN apk add --no-cache gcc musl-dev

WORKDIR /src
COPY go.mod ./
COPY *.go ./

RUN CGO_ENABLED=1 go build ./...

FROM build AS test

RUN CGO_ENABLED=1 go test -race ./...
RUN go vet ./...
