GIT_COMMIT ?= $(shell git rev-parse HEAD)
GIT_COMMIT_SHORT ?= $(shell git rev-parse --short HEAD)
GIT_TAG ?= $(shell git describe --abbrev=0 --tags 2>/dev/null || echo "v0.0.0" )
TAG ?= ${GIT_TAG}-${GIT_COMMIT_SHORT}
REPO?=ttl.sh/elemental-ci
IMAGE=${REPO}:${GIT_TAG}
ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
SUDO?=sudo
FRAMEWORK_PACKAGES?=meta/cos-light
CLOUD_CONFIG_FILE?="iso/config"
MANIFEST_FILE?="iso/manifest.yaml"
# This are the default images already in the dockerfile but we want to be able to override them
OPERATOR_IMAGE?=quay.io/costoolkit/elemental-operator-ci:latest
REGISTER_IMAGE?=quay.io/costoolkit/elemental-register-ci:latest
SYSTEM_AGENT_IMAGE?=rancher/system-agent:v0.2.9
TOOL_IMAGE?=quay.io/costoolkit/elemental-cli-ci:latest
# Used to know if this is a release or just a normal dev build
RELEASE_TAG?=false

# Set tag based on release status for ease of use
ifeq ($(RELEASE_TAG), "true")
FINAL_TAG=$(GIT_TAG)
else
FINAL_TAG=$(TAG)
endif

.PHONY: clean
clean:
	rm -rf build

# Build elemental docker images
.PHONY: build
build:
	@DOCKER_BUILDKIT=1 docker build -f Dockerfile.image \
		--target default \
		--build-arg IMAGE_TAG=${FINAL_TAG} \
		--build-arg IMAGE_COMMIT=${GIT_COMMIT} \
		--build-arg IMAGE_REPO=${REPO} \
		--build-arg OPERATOR_IMAGE=${OPERATOR_IMAGE} \
		--build-arg REGISTER_IMAGE=${REGISTER_IMAGE} \
		--build-arg SYSTEM_AGENT_IMAGE=${SYSTEM_AGENT_IMAGE} \
		--build-arg TOOL_IMAGE=${TOOL_IMAGE} \
		-t ${REPO}:${FINAL_TAG} \
		.
	@DOCKER_BUILDKIT=1 docker push ${REPO}:${FINAL_TAG}

.PHONY: dump_image
dump_image:
	@mkdir -p build
	@docker save ${REPO}:${FINAL_TAG} -o build/elemental_${FINAL_TAG}.tar

# Build iso with the elemental image as base
.PHONY: iso
iso:
ifeq ($(CLOUD_CONFIG_FILE),"iso/config")
	@echo "No CLOUD_CONFIG_FILE set, using the default one at ${CLOUD_CONFIG_FILE}"
else
	@cp ${CLOUD_CONFIG_FILE} iso/config
endif
ifeq ($(MANIFEST_FILE),"iso/manifest.yaml")
	@echo "No MANIFEST_FILE set, using the default one at ${MANIFEST_FILE}"
else
	@cp ${MANIFEST_FILE} iso/config
endif
	@mkdir -p build
	@DOCKER_BUILDKIT=1 docker build -f Dockerfile.iso \
		--target default \
		--build-arg OS_IMAGE=${REPO}:${FINAL_TAG} \
		--build-arg TOOL_IMAGE=${TOOL_IMAGE} \
		--build-arg ELEMENTAL_VERSION=${FINAL_TAG} \
		--build-arg CLOUD_CONFIG_FILE=${CLOUD_CONFIG_FILE} \
		--build-arg MANIFEST_FILE=${MANIFEST_FILE} \
		-t iso:${FINAL_TAG} .
	@DOCKER_BUILDKIT=1 docker run --rm -v $(PWD)/build:/mnt \
		iso:${FINAL_TAG} \
		--config-dir . \
		--debug build-iso \
		-o /mnt \
		-n elemental-${FINAL_TAG} \
		--overlay-iso overlay dir:rootfs
	@echo "INFO: ISO available at build/elemental-${FINAL_TAG}.iso"

# Build an iso with the OBS base containers
.PHONY: remote_iso
proper_iso:
ifeq ($(CLOUD_CONFIG_FILE),"iso/config")
	@echo "No CLOUD_CONFIG_FILE set, using the default one at ${CLOUD_CONFIG_FILE}"
endif
ifeq ($(MANIFEST_FILE),"iso/manifest.yaml")
	@echo "No MANIFEST_FILE set, using the default one at ${MANIFEST_FILE}"
else
	@cp ${MANIFEST_FILE} iso/config
endif
	@mkdir -p build
	@DOCKER_BUILDKIT=1 docker build -f Dockerfile.iso \
		--target default \
		--build-arg CLOUD_CONFIG_FILE=${CLOUD_CONFIG_FILE} \
		--build-arg MANIFEST_FILE=${MANIFEST_FILE} \
		-t iso:latest .
	@DOCKER_BUILDKIT=1 docker run --rm -v $(PWD)/build:/mnt \
		iso:latest \
		--config-dir . \
		--debug build-iso \
		-o /mnt \
		-n elemental-${FINAL_TAG} \
		--overlay-iso overlay dir:rootfs
	@echo "INFO: ISO available at build/elemental-${FINAL_TAG}.iso"

.PHONY: extract_kernel_init_squash
extract_kernel_init_squash:
	isoinfo -x /rootfs.squashfs -R -i build/elemental-${FINAL_TAG}.iso > build/elemental-${FINAL_TAG}.squashfs
	isoinfo -x /boot/kernel -R -i build/elemental-${FINAL_TAG}.iso > build/elemental-${FINAL_TAG}-kernel
	isoinfo -x /boot/initrd -R -i build/elemental-${FINAL_TAG}.iso > build/elemental-${FINAL_TAG}-initrd

.PHONY: ipxe
ipxe:
	@mkdir -p build
	echo "#!ipxe" > build/elemental-${FINAL_TAG}.ipxe
	echo "set arch amd64" >> build/elemental-${FINAL_TAG}.ipxe
ifeq ($(RELEASE_TAG), "true")
	echo "set url https://github.com/rancher/elemental/releases/download/${FINAL_TAG}" >> build/elemental-${FINAL_TAG}.ipxe
else
	echo "set url tftp://10.0.2.2/${TAG}" >> build/elemental-${FINAL_TAG}.ipxe
endif
	echo "set kernel elemental-${FINAL_TAG}-kernel" >> build/elemental-${FINAL_TAG}.ipxe
	echo "set initrd elemental-${FINAL_TAG}-initrd" >> build/elemental-${FINAL_TAG}.ipxe
	echo "set rootfs elemental-${FINAL_TAG}.squashfs" >> build/elemental-${FINAL_TAG}.ipxe
	echo "set iso    elemental-${FINAL_TAG}.iso" >> build/elemental-${FINAL_TAG}.ipxe  #not used anymore, check if we can boot from iso directly with sanboot?
	echo "# set config http://example.com/machine-config" >> build/elemental-${FINAL_TAG}.ipxe
	echo "# set cmdline extra.values=1" >> build/elemental-${FINAL_TAG}.ipxe
	echo "initrd \$${url}/\$${initrd}"  >> build/elemental-${FINAL_TAG}.ipxe
	echo "chain --autofree --replace \$${url}/\$${kernel} initrd=\$${initrd} ip=dhcp rd.cos.disable root=live:\$${url}/\$${rootfs} stages.initramfs[0].commands[0]=\"curl -k \$${config} > /run/initramfs/live/livecd-cloud-config.yaml\" console=tty1 console=ttyS0 \$${cmdline}"  >> build/elemental-${FINAL_TAG}.ipxe

.PHONY: build_all
build_all: build iso extract_kernel_init_squash ipxe

.PHONY: docs
docs:
	mkdocs build

_FW_CMD=apk add curl && ( curl -L https://raw.githubusercontent.com/rancher-sandbox/cOS-toolkit/master/scripts/get_luet.sh | sh ) && luet install --system-target /framework -y $(FRAMEWORK_PACKAGES) && rm -rf /framework/var/luet
update-cos-framework:
	@echo "Cleanup generated files"
	$(SUDO) rm -rf $(ROOT_DIR)/framework/cos
	docker run --rm --entrypoint /bin/sh \
		-v $(ROOT_DIR)/framework/cos:/framework \
		alpine -c \
		"$(_FW_CMD)"
	$(SUDO) chown -R $$(id -u) $(ROOT_DIR)/framework/cos
