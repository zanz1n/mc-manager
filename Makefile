ifneq ($(wildcard .env),)
include .env
endif

SHELL := /usr/bin/env bash -o errexit -o pipefail -o nounset

DEBUG ?= 0

PREFIX ?= mc-
SUFIX ?=

BINS = manager
DIR ?= bin
TMP ?= tmp

GO ?= go

VERSION ?= release-$(shell git rev-parse HEAD | head -c8)

GOMODPATH := github.com/zanz1n/mc-manager
LDFLAGS := -X $(GOMODPATH)/config.Version=$(VERSION)

ifeq ($(DEBUG), 1)
SUFIX += -debug
else
LDFLAGS += -s -w
endif

OS := $(if $(GOOS),$(GOOS),$(shell GOTOOLCHAIN=local $(GO) env GOOS))
ARCH := $(if $(GOARCH),$(GOARCH),$(shell GOTOOLCHAIN=local $(GO) env GOARCH))

ifeq ($(ARCH), amd64)
UNAME_ARCH := x86_64
else ifeq ($(ARCH), arm64)
UNAME_ARCH := aarch64
endif

ifeq ($(OS), windows)
SUFIX += .exe
endif

default: test all

all: $(addprefix build-, $(BINS))

run-%: build-%
ifneq ($(OS), $(shell GOTOOLCHAIN=local $(GO) env GOOS))
	$(error when running GOOS must be equal to the current os)
else ifneq ($(ARCH), $(shell GOTOOLCHAIN=local $(GO) env GOARCH))
	$(error when running GOARCH must be equal to the current cpu arch)
else ifneq ($(OUTPUT),)
	$(OUTPUT)
else
	$(DIR)/$(PREFIX)$*-$(OS)-$(UNAME_ARCH)$(SUFIX)
endif

build-%: $(DIR) generate
ifneq ($(OUTPUT),)
	GOOS=$(OS) GOARCH=$(ARCH) $(GO) build -ldflags "$(LDFLAGS)" \
	-o $(OUTPUT) $(GOMODPATH)/cmd/$*
else
	GOOS=$(OS) GOARCH=$(ARCH) $(GO) build -ldflags "$(LDFLAGS)" \
	-o $(DIR)/$(PREFIX)$*-$(OS)-$(UNAME_ARCH)$(SUFIX) $(GOMODPATH)/cmd/$*
endif
ifneq ($(POST_BUILD_CHMOD),)
	chmod $(POST_BUILD_CHMOD) $(DIR)/$(PREFIX)$*-$(OS)-$(UNAME_ARCH)$(SUFIX)
endif

$(DIR):
	mkdir $(DIR)

TESTFLAGS = -v -race

ifeq ($(SHORTTESTS), 1)
TESTFLAGS += -short
endif

ifeq ($(NOTESTCACHE), 1)
TESTFLAGS += -count=1
endif

test: generate
ifneq ($(SKIPTESTS), 1)
	$(GO) test ./... $(TESTFLAGS)
else
    $(warning skipped tests)
endif

deps:
	$(GO) install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

update: deps
	$(GO) mod tidy
	$(GO) get -u ./...
	$(GO) mod tidy

NATIVE_OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
NATIVE_ARCH := $(shell uname -m)

ifeq ($(NATIVE_ARCH), aarch64)
PROTOC_ARCH := aarch_64
else
PROTOC_ARCH := $(NATIVE_ARCH)
endif

PROTOC := $(TMP)/protoc-$(NATIVE_OS)-$(NATIVE_ARCH)

PROTOC_INCLUDE := -I api/proto -I $(PROTOC)/include

$(PROTOC):
	$(info Downloading protoc)

	mkdir -p $(PROTOC)

	LATEST=$$(curl \
	--silent "https://api.github.com/repos/protocolbuffers/protobuf/releases/latest" | \
	grep '"tag_name":' | \
	sed -E 's/.*"([^"]+)".*/\1/'); \
	curl -fsSL -o $(PROTOC).zip \
	https://github.com/protocolbuffers/protobuf/releases/download/$$LATEST/protoc-$${LATEST:1}-$(NATIVE_OS)-$(PROTOC_ARCH).zip;

	rm -rf $(PROTOC)

	unzip -q $(PROTOC).zip -d $(PROTOC)
	rm -f $(PROTOC).zip

proto-generate: $(PROTOC) deps
	rm -f internal/pb/*pb.go

	$(PROTOC)/bin/protoc $(PROTOC_INCLUDE) \
	--go_out=./internal/pb --go_opt=paths=source_relative \
	--go-grpc_out=./internal/pb --go-grpc_opt=paths=source_relative \
	./api/proto/*.proto

sqlc-generate:
	find internal/db ! -name '*_conv.go' ! -name '.gitignore' -type f -exec rm -f {} +
	sqlc generate

generate: proto-generate

fmt:
	go fmt ./...
	buf format -w

debug:
	@echo DEBUG = $(DEBUG)
	@echo DIR = $(DIR)
	@echo NATIVE_ARCH = $(NATIVE_ARCH)
	@echo NATIVE_OS = $(NATIVE_OS)
	@echo BINNAME = $(PREFIX)%-$(OS)-$(UNAME_ARCH)$(SUFIX)
	@echo GOMODPATH = $(GOMODPATH)
	@echo VERSION = $(VERSION)
	@echo BINS = $(BINS)
	@echo LDFLAGS = $(LDFLAGS)
