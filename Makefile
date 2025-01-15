.DEFAULT_GOAL := default
VERSION := $(shell git describe --tags | sed 's/^v//' | rev | cut -d - -f 2- | rev)
COMMIT := $(shell git log -1 --format='%H')

TAGS := $(strip netgo)
LD_FLAGS := -s -w \
	-X github.com/cosmos/cosmos-sdk/version.Name=sentinel \
	-X github.com/cosmos/cosmos-sdk/version.AppName=dvpnx \
	-X github.com/cosmos/cosmos-sdk/version.Version=${VERSION} \
	-X github.com/cosmos/cosmos-sdk/version.Commit=${COMMIT} \
	-X github.com/cosmos/cosmos-sdk/version.BuildTags=${TAGS}

.PHONY: benchmark
benchmark:
	@go test -bench -mod=readonly -v ./...

.PHONY: build
build:
	go build -ldflags="${LD_FLAGS}" -mod=readonly -tags="${TAGS}" -trimpath \
		-o ./bin/dvpnx main.go

.PHONY: clean
clean:
	rm -rf ./bin ./vendor

.PHONY: default
default: clean build

.PHONY: install
install:
	go build -ldflags="${LD_FLAGS}" -mod=readonly -tags="${TAGS}" -trimpath \
		-o "${GOPATH}/bin/dvpnx" main.go

.PHONY: build-image
build-image:
	@docker build --compress --file Dockerfile --force-rm --tag dvpnx .

.PHONY: go-lint
go-lint:
	@golangci-lint run --fix

.PHONY: test
test:
	@go test -cover -mod=readonly -v ./...

.PHONY: tools
tools:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50.1
