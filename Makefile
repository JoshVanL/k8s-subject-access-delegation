PACKAGE_NAME := k8s-subject-access-delegation
PATH_NAME := github.com/joshvanl/$(PACKAGE_NAME)
API_PATH := $(PATH_NAME)/pkg/apis/authz

# A temporary directory to store generator executors in
BINDIR ?= bin
GOPATH ?= $HOME/go
HACK_DIR ?= hack

# A list of all types.go files in pkg/apis
TYPES_FILES = $(shell find pkg/apis -name types.go)

all: generate build

build:
	go build

generate:
	./hack/update-codegen.sh

