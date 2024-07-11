DOCKER?=docker
ELEMENTAL_OS_IMAGE?=docker.io/local/elemental-os:dev 
ELEMENTAL_ISO_IMAGE?=docker.io/local/elemental-iso:dev 
ELEMENTAL_REGISTER?=docker.io/local/elemental-register:dev 
ELEMENTAL_TOOLKIT?=docker.io/local/elemental-toolkit:dev 

RANCHER_SYSTEM_AGENT_VERSION?=v0.3.4

.PHONY: build-dev-os
build-dev-os: 
	$(DOCKER) build \
			--build-arg ELEMENTAL_TOOLKIT=$(ELEMENTAL_TOOLKIT) \
			--build-arg ELEMENTAL_REGISTER=$(ELEMENTAL_REGISTER) \
			--build-arg RANCHER_SYSTEM_AGENT_VERSION=$(RANCHER_SYSTEM_AGENT_VERSION) \
			-t $(ELEMENTAL_OS_IMAGE) \
			-f Dockerfile.dev.os .

.PHONY: build-dev-iso
build-dev-iso: build-dev-os
	$(DOCKER) build \
			--build-arg ELEMENTAL_OS_IMAGE=$(ELEMENTAL_OS_IMAGE) \
			-t $(ELEMENTAL_ISO_IMAGE) \
			-f Dockerfile.dev.iso .

.PHONY: kind-load-dev-iso
kind-load-dev-iso: build-dev-iso
	kind load docker-image $(ELEMENTAL_ISO_IMAGE) --name operator-e2e
