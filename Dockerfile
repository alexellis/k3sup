FROM golang:1.12 as builder

WORKDIR /opt

COPY go.mod go.sum ./
RUN go mod download

ADD ./ ./
RUN go test

ENV GOOS linux
ENV GOARCH amd64
ENV CGO_ENABLED=0

ARG VERSION
RUN go build -v -ldflags "-s -w -X github.com/alexellis/k3sup/pkg/cmd.Version=$VERSION -X github.com/alexellis/k3sup/pkg/cmd.GitCommit=$COMMIT" -o k3sup

FROM alpine:latest

EXPOSE 5000
ENTRYPOINT ["k3sup"]
CMD ["--help"]

RUN apk --no-cache add ca-certificates
COPY --from=builder /opt/k3sup /usr/local/bin/k3sup
