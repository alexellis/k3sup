Version := $(shell git describe --tags --dirty)
GitCommit := $(shell git rev-parse HEAD)
LDFLAGS := "-s -w -X github.com/alexellis/k3sup/cmd.Version=$(Version) -X github.com/alexellis/k3sup/cmd.GitCommit=$(GitCommit)"
PLATFORM := $(shell ./hack/platform-tag.sh)
SOURCE_DIRS = cmd pkg main.go
.PHONY: all

.PHONY: build
build:
	go build

.PHONY: fmtcheck
fmt: ## Checks for style violation using gofmt
	@gofmt -l -s $(SOURCE_DIRS) | grep ".*\.go"; if [ "$$?" = "0" ]; then echo "Run gofmt -w -s ./path" && exit 1; fi

.PHONY: test
test:
	CGO_ENABLED=0 go test $(shell go list ./... | grep -v /vendor/|xargs echo) -cover

.PHONY: dist
dist:
	mkdir -p bin
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -mod=vendor -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/k3sup
	GO111MODULE=on CGO_ENABLED=0 GOOS=darwin go build -mod=vendor -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/k3sup-darwin
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -mod=vendor -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/k3sup-armhf
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -mod=vendor -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/k3sup-arm64
	GO111MODULE=on CGO_ENABLED=0 GOOS=windows go build -mod=vendor -ldflags $(LDFLAGS) -a -installsuffix cgo -o bin/k3sup.exe
