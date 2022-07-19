GIT_COMMIT ?= $(shell git rev-parse HEAD)
GIT_COMMIT_SHORT ?= $(shell git rev-parse --short HEAD)
GIT_TAG ?= $(shell git describe --abbrev=0 --tags 2>/dev/null || echo "v0.0.0" )
REPO?=quay.io/costoolkit/elemental-ci
IMAGE=${REPO}:${GIT_TAG}
ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
SUDO?=sudo
FRAMEWORK_PACKAGES?=meta/cos-light
CLOUD_CONFIG_FILE?="iso/config"

.PHONY: clean
clean:
	rm -rf build

# Build an iso with the OBS base containers
.PHONY: remote_iso
proper_iso:
ifeq ($(CLOUD_CONFIG_FILE),"iso/config")
	@echo "No CLOUD_CONFIG_FILE set, using the default one at ${CLOUD_CONFIG_FILE}"
endif
	@mkdir -p build
	@DOCKER_BUILDKIT=1 docker build -f Dockerfile.iso \
		--target default \
		--build-arg CLOUD_CONFIG_FILE=${CLOUD_CONFIG_FILE} \
		-t elemental/iso:latest .
	@DOCKER_BUILDKIT=1 docker run --rm -v $(PWD)/build:/mnt \
		elemental/iso:latest \
		--debug build-iso \
		-o /mnt \
		--squash-no-compression \
		-n elemental-${TAG} \
		--overlay-iso overlay dir:rootfs
	@echo "INFO: ISO available at build/elemental-${TAG}.iso"

.PHONY: extract_kernel_init_squash
	isoinfo -x /rootfs.squashfs -R -i dist/artifacts/elemental-${TAG}.iso > build/output.squashfs
	isoinfo -x /boot/kernel.xz -R -i dist/artifacts/elemental-${TAG}.iso > build/output-kernel
	isoinfo -x /boot/rootfs.xz -R -i dist/artifacts/elemental-${TAG}.iso > build/output-initrd

.PHONY: docs
docs:
	mkdocs build

deps:
	go install -mod=mod github.com/onsi/ginkgo/v2/ginkgo@latest
	go get github.com/onsi/gomega/...

integration-tests: 
	$(MAKE) -C tests/ integration-tests

_FW_CMD=apk add curl && ( curl -L https://raw.githubusercontent.com/rancher-sandbox/cOS-toolkit/master/scripts/get_luet.sh | sh ) && luet install --system-target /framework -y $(FRAMEWORK_PACKAGES) && rm -rf /framework/var/luet
update-cos-framework:
	@echo "Cleanup generated files"
	$(SUDO) rm -rf $(ROOT_DIR)/framework/cos
	docker run --rm --entrypoint /bin/sh \
		-v $(ROOT_DIR)/framework/cos:/framework \
		alpine -c \
		"$(_FW_CMD)"
	$(SUDO) chown -R $$(id -u) $(ROOT_DIR)/framework/cos
