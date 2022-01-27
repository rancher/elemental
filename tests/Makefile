# Used by integration tests
export COS_USER?=root
export COS_PASS?=ros

export COS_HOST?=192.168.122.50:22
export COS_TIMEOUT?=360

ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
export CLOUD_INIT_ISO?=$(ROOT_DIR)/build/ci.iso

.PHONY: up
up: build/ci.iso
	@echo "Cloud init iso at $(CLOUD_INIT_ISO)"
	ls -liah $(CLOUD_INIT_ISO)
	vagrant up os2 --provider=libvirt

deps: 
	go get github.com/onsi/ginkgo/v2/ginkgo
	go get github.com/onsi/gomega/...

integration-tests: up deps
	ginkgo -r ./smoke

build/ci.iso:
	./prepare_cloudinit