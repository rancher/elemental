# Used by integration tests
export COS_USER?=root
export COS_PASS?=ros

export COS_HOST?=192.168.122.50:22
export COS_TIMEOUT?=360

BOX_IMAGE?=elemental
BOX?=
ISO?=
ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

export CLOUD_INIT_ISO?=$(ROOT_DIR)/build/ci.iso

clean:
	VBoxManage controlvm "test" poweroff &>/dev/null || true
	VBoxManage unregistervm "test" --delete &>/dev/null || true
	VBoxManage closemedium disk sda.vdi --delete &>/dev/null || true
	rm -rf build-box build || true

deps: 
	go install -mod=mod github.com/onsi/ginkgo/v2/ginkgo
	go install -mod=mod github.com/onsi/gomega
	go mod tidy

# Targets to run tests on VBox (GHA macOS runners)
build-box/sda.vdi:
	mkdir build-box
	VBoxManage createmedium disk --filename build-box/sda.vdi --size 40000

# Test installation from ISO image
iso-tests: build-box/sda.vdi deps
	VBoxManage createvm --name "test" --register
	VBoxManage modifyvm "test" --firmware efi --memory 10240 --cpus 3
	VBoxManage modifyvm "test" --nic1 nat --boot1 disk --boot2 dvd --natpf1 "guestssh,tcp,,2222,,22"
	VBoxManage storagectl "test" --name "sata controller" --add sata --portcount 2 --hostiocache off
	VBoxManage storageattach "test" --storagectl "sata controller" --port 0 --device 0 --type hdd --medium build-box/sda.vdi
	VBoxManage storageattach "test" --storagectl "sata controller" --port 1 --device 0 --type dvddrive --medium $(ISO)
	VBoxManage startvm "test" --type headless
	ginkgo --label-filter iso -r -v ./install
	ginkgo -r -v ./smoke

# Test installation from container image
container-tests: build-box/sda.vdi deps
	VBoxManage createvm --name "test" --register
	VBoxManage modifyvm "test" --firmware efi --memory 10240 --cpus 3
	VBoxManage modifyvm "test" --nic1 nat --boot1 disk --boot2 dvd --natpf1 "guestssh,tcp,,2222,,22"
	VBoxManage storagectl "test" --name "sata controller" --add sata --portcount 2 --hostiocache off
	VBoxManage storageattach "test" --storagectl "sata controller" --port 0 --device 0 --type hdd --medium build-box/sda.vdi
	VBoxManage storageattach "test" --storagectl "sata controller" --port 1 --device 0 --type dvddrive --medium $(ISO)
	VBoxManage startvm "test" --type headless
	ginkgo --label-filter container -r -v ./install
	ginkgo -r -v ./smoke

# Generate tests description file
generate-readme:
	@./scripts/generate-readme > README.md

# E2E tests
e2e-install-rancher: deps
	ginkgo --label-filter install -r -v ./e2e
e2e-configure-rancher: deps
	ginkgo --label-filter configure -r -v ./e2e
e2e-ui-rancher: deps
	ginkgo --label-filter ui -r -v ./e2e
e2e-bootstrap-node: deps
	ginkgo --label-filter bootstrap -r -v ./e2e
e2e-upgrade-node: deps
	ginkgo --label-filter upgrade -r -v ./e2e
start-cypress-tests:
	@./scripts/start-cypress-tests
e2e-get-logs: deps
	ginkgo --label-filter logs -r -v ./e2e
