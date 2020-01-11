Version := $(shell git describe --tags --dirty)
GitCommit := $(shell git rev-parse HEAD)
LDFLAGS := "-s -w -X github.com/alexellis/k3sup/pkg/cmd.Version=$(Version) -X github.com/alexellis/k3sup/pkg/cmd.GitCommit=$(GitCommit)"
PLATFORM := $(shell ./hack/platform-tag.sh)

.PHONY: all

.PHONY: build
build:
	go build

.PHONY: test
test:
	CGO_ENABLED=0 go test $(shell go list ./... | grep -v /vendor/|xargs echo) -cover

.PHONY: dist
dist:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -mod=vendor -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/k3sup
	CGO_ENABLED=0 GOOS=darwin go build -mod=vendor -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/k3sup-darwin
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -mod=vendor -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/k3sup-armhf
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -mod=vendor -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/k3sup-arm64
	CGO_ENABLED=0 GOOS=windows go build -mod=vendor -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/k3sup.exe
