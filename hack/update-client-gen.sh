#!/bin/bash

# The only argument this script should ever be called with is '--verify-only'

set -o errexit
set -o nounset
set -o pipefail

PACKAGE=github.com/joshvanl/k8s-subject-access-delegation
REPO_ROOT=$GOPATH/src/${PACKAGE}
BINDIR=${REPO_ROOT}/bin

# Generate the internal clientset (pkg/client/clientset_generated/internalclientset)
${BINDIR}/client-gen "$@" \
	      --input-base "${PACKAGE}/pkg/apis/" \
	      --input "authz/" \
	      --clientset-path "${PACKAGE}/pkg/client/" \
	      --clientset-name internalclientset \
	      --go-header-file "${GOPATH}/src/github.com/kubernetes/repo-infra/verify/boilerplate/boilerplate.go.txt"
# Generate the versioned clientset (pkg/client/clientset_generated/clientset)
${BINDIR}/client-gen "$@" \
		  --input-base "${PACKAGE}/pkg/apis/" \
		  --input "authz/v1alpha1" \
	      --clientset-path "${PACKAGE}/pkg/" \
	      --clientset-name "client" \
	      --go-header-file "${GOPATH}/src/github.com/kubernetes/repo-infra/verify/boilerplate/boilerplate.go.txt"
# generate lister
${BINDIR}/lister-gen "$@" \
		  --input-dirs="${PACKAGE}/pkg/apis/authz" \
	      --input-dirs="${PACKAGE}/pkg/apis/authz/v1alpha1" \
	      --output-package "${PACKAGE}/pkg/listers" \
	      --go-header-file "${GOPATH}/src/github.com/kubernetes/repo-infra/verify/boilerplate/boilerplate.go.txt"
# generate informer
${BINDIR}/informer-gen "$@" \
	      --go-header-file "${GOPATH}/src/github.com/kubernetes/repo-infra/verify/boilerplate/boilerplate.go.txt" \
	      --input-dirs "${PACKAGE}/pkg/apis/authz" \
	      --input-dirs "${PACKAGE}/pkg/apis/authz/v1alpha1" \
	      --internal-clientset-package "${PACKAGE}/pkg/client/internalclientset" \
	      --versioned-clientset-package "${PACKAGE}/pkg/client" \
	      --listers-package "${PACKAGE}/pkg/listers" \
	      --output-package "${PACKAGE}/pkg/informers"
