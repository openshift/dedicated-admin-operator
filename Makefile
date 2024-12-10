SHELL := /usr/bin/env bash

# Include shared Makefiles
include project.mk
include standard.mk
include functions.mk

default: generate-syncset gobuild

# Extend Makefile after here

# Build the docker image
.PHONY: docker-build
docker-build:
	$(MAKE) build

# Push the docker image
.PHONY: docker-push
docker-push:
	$(MAKE) push

.PHONY: generate-syncset
generate-syncset:
	if [ "${IN_DOCKER_CONTAINER}" == "true" ]; then \
		docker run --rm -v `pwd -P`:`pwd -P` quay.io/app-sre/python:2.7.15 /bin/sh -c "cd `pwd`; ls -la /var/lib/jenkins/workspace/openshift-dedicated-admin-operator-gh-pr-check/scripts/generate_syncset.py /var/lib/jenkins/workspace/openshift-dedicated-admin-operator-gh-pr-check/scripts /var/lib/jenkins/workspace/openshift-dedicated-admin-operator-gh-pr-check /var/lib/jenkins/workspace/openshift-dedicated-admin-operator-gh-pr-check /var/lib/jenkins/workspace /var/lib/jenkins/;pip install oyaml; `pwd`/${GEN_SYNCSET}"; \
	else \
		${GEN_SYNCSET}; \
	fi

