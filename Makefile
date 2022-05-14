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

docker-build:
	$(DOCKER_CMD) build -t $(DOCKER_TAG) $(DOCKER_OPTS) \
		--build-arg VERSION_PKG=$(VERSION_PKG) --build-arg VERSION=$(VERSION) --build-arg GIT_COMMIT=$(GIT_COMMIT) \
		.

define def-release-single
release-$1/$2:
	CGO_ENABLED=0 GOOS=$(1) GOARCH=$(2) go build -trimpath -o $(OUT_PATH)/$(PKG_NAME)-$(1)_$(2) \
	-ldflags "-X $(VERSION_PKG).version=$(VERSION) -X $(VERSION_PKG).gitCommit=$(GIT_COMMIT)"  \
	./cmd/*.go
	cp LICENSE $(OUT_PATH)/LICENSE
	tar czf $(OUT_PATH)/$(PKG_NAME)-$(1)_$(2).tar.gz -C $(OUT_PATH) $(PKG_NAME)-$(1)_$(2) LICENSE

release-all:: release-$1/$2

.PHONY: release-$1/$2
endef

$(foreach os, \
	$(OS_LIST), \
	$(foreach arch, \
		$(ARCH_LIST), \
		$(eval $(call def-release-single,$(os),$(arch))) \
	) \
)

clean:
	rm -r $(OUT_PATH)

.PHONY: build docker-build release-all clean
