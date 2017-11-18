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
	# all       - runs verify, test, build
	# build     - builds targets
	# generate  - generates mocks and assets files
	# verify    - verifies generated files & scripts
	# test      - runs all tests

all: verify generate test build

build: go_build

generate:
	./hack/update-codegen.sh

verify: go_fmt go_vet go_dep

go_vet:
	go vet $$(go list ./pkg/...)
#go vet $$(go list ./pkg/... ./cmd/...)

go_fmt:
	@set -e; \
	GO_FMT=$$(git ls-files *.go | grep -v 'vendor/' | xargs gofmt -d); \
	if [ -n "$${GO_FMT}" ] ; then \
		echo "Please run go fmt"; \
		echo "$$GO_FMT"; \
		exit 1; \
	fi

go_dep:
	dep ensure -no-vendor -dry-run -v

test:
	go test $$(go list ./pkg/...)
#go test $$(go list ./pkg/... ./cmd/...)

go_build:
	CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -a -tags netgo -ldflags '-w -X main.version=$(CI_COMMIT_TAG) -X main.commit=$(CI_COMMIT_SHA) -X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)' -o k8s_subject_access_delegation_linux_amd64  .
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -tags netgo -ldflags '-w -X main.version=$(CI_COMMIT_TAG) -X main.commit=$(CI_COMMIT_SHA) -X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)' -o k8s_subject_access_delegation_darwin_amd64 .
