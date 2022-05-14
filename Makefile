OUT_PATH ?= _out
PKG_NAME ?= kexplain
OS_LIST ?= darwin linux
ARCH_LIST ?= amd64 arm64
GIT_COMMIT ?= $(shell git rev-parse --short HEAD)
VERSION ?= $(shell git describe --tags --dirty --always)
VERSION_PKG := kexplain/pkg/version

DOCKER_CMD ?= docker
DOCKER_TAG ?= kexplain
DOCKER_OPTS ?= ""

build:
	CGO_ENABLED=0 go build -trimpath -o $(OUT_PATH)/$(PKG_NAME) \
		-ldflags "-X $(VERSION_PKG).version=$(VERSION) -X $(VERSION_PKG).gitCommit=$(GIT_COMMIT)"  \
	 ./cmd/*.go

define build-single
CGO_ENABLED=0 GOOS=$(1) GOARCH=$(2) go build -trimpath -o $(OUT_PATH)/$(PKG_NAME)-$(1)_$(2) \
	-ldflags "-X $(VERSION_PKG).version=$(VERSION) -X $(VERSION_PKG).gitCommit=$(GIT_COMMIT)"  \
	./cmd/*.go
endef

build-all:
	$(foreach os, \
		$(OS_LIST), \
		$(foreach arch, \
			$(ARCH_LIST), \
			$(shell $(call build-single,$(os),$(arch))) \
		) \
	)

docker-build:
	$(DOCKER_CMD) build -t $(DOCKER_TAG) $(DOCKER_OPTS) \
		--build-arg VERSION_PKG=$(VERSION_PKG) --build-arg VERSION=$(VERSION) --build-arg GIT_COMMIT=$(GIT_COMMIT) \
		.

.PHONY: build build-all
