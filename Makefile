PACKAGE_NAME := k8s-subject-access-delegation
PATH_NAME := github.com/joshvanl/$(PACKAGE_NAME)
API_PATH := $(PATH_NAME)/pkg/apis/authz

# A temporary directory to store generator executors in
BINDIR ?= bin
GOPATH ?= $HOME/go
HACK_DIR ?= hack

# A list of all types.go files in pkg/apis
TYPES_FILES = $(shell find pkg/apis -name types.go)

help:
	# all       - runs verify, build targets
	# build     - builds targets
	# generate  - generates mocks and assets files
	# verify    - verifies generated files & scripts

all: generate verify build

build:
	go build

generate:
	./hack/update-codegen.sh

verify: go_fmt go_vet

go_vet:
	go vet $$(go list ./pkg/... ./cmd/...)

go_test:
	go test $$(go list ./pkg/... ./cmd/...)

go_fmt:
	@set -e; \
	GO_FMT=$$(git ls-files *.go | grep -v 'vendor/' | xargs gofmt -d); \
	if [ -n "$${GO_FMT}" ] ; then \
		echo "Please run go fmt"; \
		echo "$$GO_FMT"; \
		exit 1; \
	fi
