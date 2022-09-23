FROM golang:1.18-alpine

# Install git
RUN set -ex; \
    apk update; \
    apk add --no-cache git

# Set working directory
WORKDIR /go/src/github.com/apolsh/yapr-gophermart/integration-tests/

# Run tests
CMD CGO_ENABLED=0 go test ./... -v -coverprofile cover.out
#ENTRYPOINT ["CGO_ENABLED=0", "go", "test", "-v", "./...", "-coverprofile", "cover.out"]