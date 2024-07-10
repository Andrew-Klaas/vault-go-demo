# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:latest

ENV GOOS linux
ENV GOARCH amd64
ADD . /go/src/github.com/Andrew-Klaas/vault-go-demo
WORKDIR /go/src/github.com/Andrew-Klaas/vault-go-demo

COPY go.mod ./
COPY go.sum ./
RUN go mod download
RUN go mod tidy

RUN --mount=type=ssh \
  go get github.com/Andrew-Klaas/vault-go-demo
RUN go get github.com/hashicorp/hcl/hcl/ast
RUN go get github.com/cenkalti/backoff/v3
RUN go get github.com/hashicorp/vault/api
RUN go get github.com/lib/pq

RUN env GOOS=linux GOARCH=amd64 go install /go/src/github.com/Andrew-Klaas/vault-go-demo

# Run the outyet command by default when the container starts.
ENTRYPOINT /go/bin/vault-go-demo

# Document that the service listens on port 8080.
EXPOSE 9090

#docker build --platform linux/amd64 -t aklaas2/vault-go-demo-oauth .;docker push aklaas2/vault-go-demo-oauth:latest


# syntax=docker/dockerfile:1

# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
# FROM golang:1.16-alpine
# WORKDIR /go/src/github.com/Andrew-Klaas/vault-go-demo-non-tokenization
# ADD . /go/src/github.com/Andrew-Klaas/vault-go-demo-non-tokenization

# COPY go.mod ./
# COPY go.sum ./
# RUN go mod download
# # RUN go get github.com/hashicorp/hcl
# # RUN go get github.com/cenkalti/backoff
# # RUN go get github.com/hashicorp/vault/api
# # RUN go get github.com/lib/pq
# RUN go get github.com/Andrew-Klaas/vault-go-demo-non-tokenization
# RUN go install /go/src/github.com/Andrew-Klaas/vault-go-demo-non-tokenization


# # Run the outyet command by default when the container starts.
# ENTRYPOINT /go/bin/vault-go-demo-non-tokenization

# # Document that the service listens on port 8080.
# EXPOSE 9090

#docker build -t aklaas2/vault-go-demo-non-tokenization .;docker push aklaas2/vault-go-demo-non-tokenization:latest