Version := $(shell git describe --tags --dirty)
# Version := "dev"
GitCommit := $(shell git rev-parse HEAD)
LDFLAGS := "-s -w -X github.com/alexellis/k3sup/cmd.Version=$(Version) -X github.com/alexellis/k3sup/cmd.GitCommit=$(GitCommit)"
export GO111MODULE=on
SOURCE_DIRS = cmd pkg main.go

.PHONY: all
all: gofmt test dist hash

.PHONY: test
test:
	CGO_ENABLED=0 go test $(shell go list ./... | grep -v /vendor/|xargs echo) -cover

.PHONY: gofmt
gofmt: 
	@test -z $(shell gofmt -l -s $(SOURCE_DIRS) ./ | tee /dev/stderr) || (echo "[WARN] Fix formatting issues with 'make fmt'" && exit 1)

.PHONY: dist
dist:
	mkdir -p bin/
	rm -rf bin/k3sup*
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags $(LDFLAGS) -installsuffix cgo -o bin/k3sup
	GOARM=7 GOARCH=arm CGO_ENABLED=0 GOOS=linux go build -a -ldflags $(LDFLAGS) -installsuffix cgo -o bin/k3sup-armhf
	GOARCH=arm64 CGO_ENABLED=0 GOOS=linux go build -a -ldflags $(LDFLAGS) -installsuffix cgo -o bin/k3sup-arm64
	CGO_ENABLED=0 GOOS=darwin go build -a -ldflags $(LDFLAGS) -installsuffix cgo -o bin/k3sup-darwin
	GOARCH=arm64 CGO_ENABLED=0 GOOS=darwin go build -a -ldflags $(LDFLAGS) -installsuffix cgo -o bin/k3sup-darwin-arm64
	GOOS=windows CGO_ENABLED=0 go build -a -ldflags $(LDFLAGS) -installsuffix cgo -o bin/k3sup.exe

.PHONY: hash
hash:
	rm -rf bin/*.sha256 && ./hack/hashgen.sh
