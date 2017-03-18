.PHONY: all clean version

NAME ?= dbgen
ORG := jhaynie
PKG := $(ORG)/$(NAME)

SHELL := /bin/bash
BASEDIR := $(shell echo $${PWD})
BUILD := $(shell git rev-parse HEAD | cut -c1-8)
SRC := $(shell find . -type f -name '*.go' -not -path './vendor/*' -not -path './.git/*' -not -path './hack/*')
PKGMAIN := main.go
VERSION := $(shell cat $(BASEDIR)/VERSION)

L="-X=github.com/$(PKG)/cmd/main.Build=$(BUILD) -X=github.com/$(PKG)/cmd/main.Version=$(VERSION)"

all: version build osx

version:
	@echo "version: $(VERSION) build: $(BUILD) package: $(PKG)"

version-short:
	@echo $(VERSION)

fmt:
	@gofmt -s -l -w $(SRC)

vet:
	@for i in `find . -type d -not -path './hack' -not -path './hack/*' -not -path './vendor/*' -not -path './.git/*' -not -path './cmd' -not -path './.*' -not -path './build/*' -not -path './backup' -not -path './vendor' -not -path '.' -not -path './build' -not -path './etc' -not -path './etc/*' -not -path './pkg' | sed 's/^\.\///g'`; do go vet github.com/$(BASEPKG)/$$i; done

test:
	go test -v ./pkg/orm/...

linux: build
	@docker run --rm -v $(GOPATH):/go -w /go/src/github.com/$(BASEPKG) golang:latest go build -ldflags $(L) -o build/linux/$(NAME)-linux-$(VERSION) $(PKGMAIN)

alpine: build
	@docker run --rm -v $(GOPATH):/go -w /go/src/github.com/$(BASEPKG) jhaynie/golang-alpine go build -ldflags $(L) -o build/alpine/$(NAME)-alpine-$(VERSION) $(PKGMAIN)

osx: build
	@go build -ldflags $(L) -o build/osx/$(NAME)-osx-$(VERSION) $(PKGMAIN)

build: version protoc

protoc:
	@docker run --rm -v $(BASEDIR):/app -w /app znly/protoc --go_out=plugins=grpc:pkg/orm --proto_path=pkg/orm pkg/orm/*.proto

compilegen:
	@docker run --rm -v $(BASEDIR):/app -v $(GOPATH):/go -w /app znly/protoc --go_out=plugins=grpc:gen/schema --proto_path=gen/schema --proto_path=/go/src gen/schema/*.proto

testgen:
	go test -v ./gen/...
