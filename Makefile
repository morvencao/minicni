# Image URL to use all building/pushing image targets;
# Use your own docker registry and image name for dev/test by overridding the
# IMAGE_REPO, IMAGE_NAME and RELEASE_TAG environment variable.
IMAGE_REPO ?= quay.io/morvencao
IMAGE_NAME ?= install-minicni
CNI_NAME ?= minicni
VERSION ?= $(shell date +v%Y%m%d)-$(shell git describe --match=$(git rev-parse --short=8 HEAD) --tags --always --dirty)
RELEASE_VERSION ?= $(shell cat ./pkg/version/version.go | grep "Version =" | awk '{ print $$3}' | tr -d '"')

PWD := $(shell pwd)
BASE_DIR := $(shell basename $(PWD))

# Github host to use for checking the source tree;
# Override this variable ue with your own value if you're working on forked repo.
GIT_HOST ?= github.com/morvencao

LOCAL_OS := $(shell uname)
ifeq ($(LOCAL_OS),Linux)
    TARGET_OS ?= linux
    XARGS_FLAGS="-r"
else ifeq ($(LOCAL_OS),Darwin)
    TARGET_OS ?= darwin
    XARGS_FLAGS=
else
    $(error "This system's OS $(LOCAL_OS) isn't recognized/supported")
endif

ARCH := $(shell uname -m)
LOCAL_ARCH := "amd64"
ifeq ($(ARCH),x86_64)
    LOCAL_ARCH="amd64"
else ifeq ($(ARCH),ppc64le)
    LOCAL_ARCH="ppc64le"
else ifeq ($(ARCH),s390x)
    LOCAL_ARCH="s390x"
else
    $(error "This system's ARCH $(ARCH) isn't recognized/supported")
endif

TESTARGS_DEFAULT := "-v"
export TESTARGS ?= $(TESTARGS_DEFAULT)

FINDFILES=find . \( -path ./.git -o -path ./.github \) -prune -o -type f

XARGS = xargs -0 $(XARGS_FLAGS)
CLEANXARGS = xargs $(XARGS_FLAGS)

all: fmt lint test build image

############################################################
# format section
############################################################

fmt: format-go

format-go:
	@$(FINDFILES) -name '*.go' \( ! \( -name '*.gen.go' -o -name '*.pb.go' \) \) -print0 | $(XARGS) goimports -w -local $(GIT_HOST)

############################################################
# lint section
############################################################

lint: lint-go lint-scripts lint-yaml lint-dockerfiles

lint-go:
	@$(FINDFILES) -name '*.go' \( ! \( -name '*.gen.go' -o -name '*.pb.go' \) \) -print0 | $(XARGS) ./scripts/lint_go.sh

lint-scripts:
	@$(FINDFILES) -name '*.sh' -print0 | $(XARGS) shellcheck

lint-yaml:
	@$(FINDFILES) \( -name '*.yml' -o -name '*.yaml' \) -print0 | $(XARGS) grep -L -e "{{" | $(CLEANXARGS) yamllint -c ./config/.yamllint.yml

lint-dockerfiles:
	@$(FINDFILES) -name 'Dockerfile*' -print0 | $(XARGS) hadolint -c ./config/.hadolint.yml

############################################################
# test section
############################################################

test:
	@echo "Running the tests for the $(CNI_NAME)..."
	@go test $(TESTARGS) ./...

############################################################
# build section
############################################################

build:
	@echo "Building the $(CNI_NAME) on $(LOCAL_ARCH)..."
	@GOARCH=$(LOCAL_ARCH) ./scripts/gobuild.sh build/_output/bin/$(CNI_NAME) ./cmd

############################################################
# image section
############################################################

image: build-image push-image

build-image: build
	@echo "Building the $(IMAGE_NAME) docker image..."
	@docker build -t $(IMAGE_REPO)/$(IMAGE_NAME):$(RELEASE_VERSION) -f deployments/Dockerfile.install-minicni .

push-image: build-image
	@echo "Pushing the $(IMAGE_NAME) docker image..."
	@docker push $(IMAGE_REPO)/$(IMAGE_NAME):$(RELEASE_VERSION)

############################################################
# clean section
############################################################
clean:
	@rm -rf build/_output

.PHONY: all fmt lint test build image clean
