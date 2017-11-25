PACKAGE_NAME := k8s-subject-access-delegation
PATH_NAME := github.com/joshvanl/$(PACKAGE_NAME)
API_PATH := $(PATH_NAME)/pkg/apis/authz

# A temporary directory to store generator executors in
BINDIR ?= bin
GOPATH ?= $HOME/go
MOCKDIR = pkg/subject_access_delegation/mocks
CLIENTGo = k8s.io/client-go/kubernetes
CLIENTGoCore = $(CLIENTGo)/typed/core/v1
CLIENTGoRbac = $(CLIENTGo)/typed/rbac/v1

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

generate: go_codegen go_mock

verify: go_fmt go_vet go_dep

go_vet:
	go vet $$(go list ./pkg/... ./cmd/...)

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
	go test $$(go list ./pkg/... ./cmd/...)

go_codegen:
	./hack/update-codegen.sh

go_build:
	CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -a -tags netgo -ldflags '-w -X main.version=$(CI_COMMIT_TAG) -X main.commit=$(CI_COMMIT_SHA) -X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)' -o k8s_subject_access_delegation_linux_amd64  .
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -tags netgo -ldflags '-w -X main.version=$(CI_COMMIT_TAG) -X main.commit=$(CI_COMMIT_SHA) -X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)' -o k8s_subject_access_delegation_darwin_amd64 .

go_mock:
	mkdir -p $(MOCKDIR)
	mockgen -imports .=github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces -package=mocks -source=pkg/subject_access_delegation/interfaces/interfaces.go -destination $(MOCKDIR)/subject_access_delegation.go
	mockgen -destination=pkg/subject_access_delegation/mocks/kubernetes.go -package=mocks -source=vendor/k8s.io/client-go/kubernetes/clientset.go Interface
	#mockgen doesn't support these embedded interfaces
	mockgen $(CLIENTGoCore) CoreV1Interface > $(MOCKDIR)/core_v1.go
	mockgen $(CLIENTGoCore) ServiceAccountInterface  > $(MOCKDIR)/service_account_v1.go
	mockgen $(CLIENTGoCore) PodInterface  > $(MOCKDIR)/pod_v1.go
	mockgen $(CLIENTGoRbac) RoleBindingInterface  > $(MOCKDIR)/rolebinding_v1.go
	mockgen $(CLIENTGoRbac) RoleInterface  > $(MOCKDIR)/role_v1.go
	mockgen $(CLIENTGoRbac) RbacV1Interface  > $(MOCKDIR)/rbac_v1.go
	sed -i 's/mock_v1/mocks/g' $(MOCKDIR)/core_v1.go
	sed -i 's/mock_v1/mocks/g' $(MOCKDIR)/service_account_v1.go
	sed -i 's/mock_v1/mocks/g' $(MOCKDIR)/pod_v1.go
	sed -i 's/mock_v1/mocks/g' $(MOCKDIR)/rolebinding_v1.go
	sed -i 's/mock_v1/mocks/g' $(MOCKDIR)/role_v1.go
	sed -i 's/mock_v1/mocks/g' $(MOCKDIR)/rbac_v1.go
