SHELL := /usr/bin/env bash

# Include project specific values file
# Requires the following variables:
# - IMAGE_REGISTRY
# - IMAGE_REPOSITORY
# - IMAGE_NAME
# - VERSION_MAJOR
# - VERSION_MINOR
include project.mk

# Validate variables in project.mk exist
ifndef IMAGE_REGISTRY
$(error IMAGE_REGISTRY is not set; check project.mk file)
endif
ifndef IMAGE_REPOSITORY
$(error IMAGE_REPOSITORY is not set; check project.mk file)
endif
ifndef IMAGE_NAME
$(error IMAGE_NAME is not set; check project.mk file)
endif
ifndef VERSION_MAJOR
$(error VERSION_MAJOR is not set; check project.mk file)
endif
ifndef VERSION_MINOR
$(error VERSION_MINOR is not set; check project.mk file)
endif

# Generate version and tag information from inputs
COMMIT_NUMBER=$(shell git rev-list `git rev-list --parents HEAD | egrep "^[a-f0-9]{40}$$"`..HEAD --count)
BUILD_DATE=$(shell date -u +%Y-%m-%d)
CURRENT_COMMIT=$(shell git rev-parse --short=8 HEAD)
VERSION_FULL=v$(VERSION_MAJOR).$(VERSION_MINOR).$(COMMIT_NUMBER)-$(BUILD_DATE)-$(CURRENT_COMMIT)

BINFILE=build/_output/bin/dedicated-admin-operator
MAINPACKAGE=./cmd/manager
GOENV=GOOS=linux GOARCH=amd64 CGO_ENABLED=0
GOFLAGS=-gcflags="all=-trimpath=${GOPATH}" -asmflags="all=-trimpath=${GOPATH}"


TESTTARGETS := $(shell go list -e ./... | egrep -v "/(vendor)/")
# ex, -v
TESTOPTS := 

.PHONY: clean
clean:
	rm -rf ./build/_output

.PHONY: build
build:
	docker build . -f build/ci-operator/Dockerfile -t $(IMAGE_REGISTRY)/$(IMAGE_REPOSITORY)/$(IMAGE_NAME):$(VERSION_FULL)

.PHONY: push
push:
	docker push $(IMAGE_REGISTRY)/$(IMAGE_REPOSITORY)/$(IMAGE_NAME):$(VERSION_FULL)

.PHONY: gocheck
gocheck: ## Lint code
	gofmt -s -l $(shell go list -f '{{ .Dir }}' ./... ) | grep ".*\.go"; if [ "$$?" = "0" ]; then gofmt -s -d $(shell go list -f '{{ .Dir }}' ./... ); exit 1; fi
	go vet ./cmd/... ./pkg/...

.PHONY: gobuild
gobuild: gocheck gotest ## Build binary
	${GOENV} go build ${GOFLAGS} -o ${BINFILE} ${MAINPACKAGE}

.PHONY: gotest
gotest:
	go test $(TESTOPTS) $(TESTTARGETS)

.PHONY: env
env:
	@echo OPERATOR_NAME=$(OPERATOR_NAME)
	@echo OPERATOR_NAMESPACE=$(OPERATOR_NAMESPACE)
	@echo OPERATOR_VERSION=$(VERSION_FULL)
