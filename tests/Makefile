# Used by integration tests
export COS_USER?=root
export COS_PASS?=ros

export COS_HOST?=192.168.122.50:22
export COS_TIMEOUT?=360

BOX_IMAGE?=os2
BOX?=
ISO?=
ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

export CLOUD_INIT_ISO?=$(ROOT_DIR)/build/ci.iso

clean: clean_vm_from_iso
	(vagrant destroy -f) 2> /dev/null || true
	(vagrant box remove $(BOX_IMAGE)) 2> /dev/null || true
	rm -rf build-box build || true

add:
	vagrant box add --force $(BOX_IMAGE) $(BOX)
	
.PHONY: up
up: build/ci.iso
	@echo "Cloud init iso at $(CLOUD_INIT_ISO)"
	ls -liah $(CLOUD_INIT_ISO)
	vagrant up $(BOX_IMAGE) --provider=libvirt

deps: 
	go get github.com/onsi/ginkgo/v2/ginkgo
	go get github.com/onsi/gomega/...

integration-tests: clean add up deps run

run:
	ginkgo -r ./smoke

.PHONY: install
install:
	ginkgo --label-filter setup -r ./install

build/ci.iso:
	./prepare_cloudinit

vbox-tests: build/ci.iso iso-to-vm
	$(MAKE) run

build-box:
	mkdir build-box

# targets to run tests on VBox (GHA macOS runners)
build-box/sda.vdi: build-box
	VBoxManage createmedium disk --filename build-box/sda.vdi --size 40000

iso-to-vm: build-box/sda.vdi
	VBoxManage createvm --name "test" --register
	VBoxManage modifyvm "test" --memory 10240 --cpus 3
	VBoxManage modifyvm "test" --nic1 nat --boot1 disk --boot2 dvd --natpf1 "guestssh,tcp,,2222,,22"
	VBoxManage storagectl "test" --name "sata controller" --add sata --portcount 2 --hostiocache off
	VBoxManage storageattach "test" --storagectl "sata controller" --port 0 --device 0 --type hdd --medium build-box/sda.vdi
	VBoxManage storageattach "test" --storagectl "sata controller" --port 1 --device 0 --type dvddrive --medium $(ISO)
	VBoxManage startvm "test" --type headless
	$(MAKE) install
	VBoxManage controlvm "test" poweroff &>/dev/null || true
	VBoxManage storageattach 'test' --storagectl 'sata controller' --port 1 --device 0 --type dvddrive --medium emptydrive --forceunmount
	VBoxManage storageattach "test" --storagectl "sata controller" --port 1 --device 0 --type dvddrive --medium $(CLOUD_INIT_ISO)
	VBoxManage startvm "test" --type headless

clean_vm_from_iso:
	VBoxManage controlvm "test" poweroff &>/dev/null || true
	VBoxManage unregistervm "test" --delete &>/dev/null || true
	VBoxManage closemedium disk sda.vdi --delete &>/dev/null || true

installer-tests: build-box/sda.vdi
	VBoxManage createvm --name "test" --register
	VBoxManage modifyvm "test" --memory 10240 --cpus 3
	VBoxManage modifyvm "test" --nic1 nat --boot1 disk --boot2 dvd --natpf1 "guestssh,tcp,,2222,,22"
	VBoxManage storagectl "test" --name "sata controller" --add sata --portcount 2 --hostiocache off
	VBoxManage storageattach "test" --storagectl "sata controller" --port 0 --device 0 --type hdd --medium build-box/sda.vdi
	VBoxManage storageattach "test" --storagectl "sata controller" --port 1 --device 0 --type dvddrive --medium $(ISO)
	VBoxManage startvm "test" --type headless
	ginkgo --label-filter "!setup" -r ./install
	VBoxManage controlvm "test" poweroff &>/dev/null || true

# e2e CI tests
install-rancher:
	./scripts/install-rancher

configure-rancher-step-01:
	./scripts/configure-rancher-step-01

install-ros-node:
	# Force the execution with a tty, needed to follow VM installation
	script -e -c ./scripts/install-ros-node

configure-rancher-step-02:
	./scripts/configure-rancher-step-02

add-node-in-cluster:
	./scripts/add-node-in-cluster

e2e-bootstraping-node-with-rancher: install-rancher configure-rancher-step-01 install-ros-node configure-rancher-step-02 add-node-in-cluster
