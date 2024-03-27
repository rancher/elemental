ROOT_DIR:=$(realpath $(PWD)/..)
IPXE:=$(ROOT_DIR)/install.ipxe
ISO:=$(shell file -Ls $(ROOT_DIR)/*.iso 2>/dev/null | awk -F':' '/boot sector/ { print $$1 }')

# Define Ginkgo timeout for the tests
GINKGO_TIMEOUT?=3600
ifdef VM_NUMBERS
	ifeq ($(shell expr $(VM_NUMBERS) \> 50), 1)
		GINKGO_TIMEOUT=7200
	endif
endif
ifeq ($(TEST_TYPE),multi)
	GINKGO_TIMEOUT=10800
endif

extract_kernel_init_squash:
	@./scripts/get-boot-files-for-pxe $(ISO)

ipxe:
	@VAR=$(notdir $(ISO)); \
	echo "#!ipxe" > $(IPXE); \
	echo "set arch amd64" >> $(IPXE); \
	echo "set url tftp://10.0.2.2/$${VAR##*/}" >> $(IPXE); \
	echo "set kernel $${VAR/\.iso/-linux}" >> $(IPXE); \
	echo "set initrd $${VAR/\.iso/-initrd}" >> $(IPXE); \
	echo "set rootfs $${VAR/\.iso/-squashfs}" >> $(IPXE); \
	echo "# set config http://example.com/machine-config" >> $(IPXE); \
	echo "# set cmdline cos.setup=/run/initramfs/live/cloud-config extra.values=1" >> $(IPXE); \
	echo "initrd \$${url}/\$${initrd}" >> $(IPXE); \
	echo "chain --autofree --replace \$${url}/\$${kernel} initrd=\$${initrd} ip=dhcp rd.cos.disable root=live:\$${url}/\$${rootfs} stages.initramfs[0].commands[0]=\"curl -k \$${config} > /run/initramfs/live/livecd-cloud-config.yaml\" console=tty1 console=ttyS0 \$${cmdline}" >> $(IPXE)

deps: 
	@go install -mod=mod github.com/onsi/ginkgo/v2/ginkgo
	@go install -mod=mod github.com/onsi/gomega
	@go mod tidy

# Generate tests description file
generate-readme:
	@./scripts/generate-readme > README.md

# Qase commands
create-qase-run: deps
	@go run qase/qase_cmd.go -create

delete-qase-run: deps
	@go run qase/qase_cmd.go -delete

publish-qase-run: deps
	@go run qase/qase_cmd.go -publish

# E2E tests
e2e-airgap-rancher: deps
	ginkgo --label-filter airgap-rancher -r -v ./e2e

e2e-bootstrap-node: deps
	ginkgo --timeout $(GINKGO_TIMEOUT)s --label-filter bootstrap -r -v ./e2e

e2e-backup-restore: deps
	ginkgo --label-filter test-backup-restore -r -v ./e2e

e2e-check-app: deps
	ginkgo --label-filter check-app -r -v ./e2e

e2e-configure-rancher: deps
	ginkgo --label-filter configure -r -v ./e2e

e2e-get-logs: deps
	ginkgo --label-filter logs -r -v ./e2e

e2e-install-app: deps
	ginkgo --label-filter install-app -r -v ./e2e

e2e-install-backup-restore: deps
	ginkgo --label-filter install-backup-restore -r -v ./e2e

e2e-install-chartmuseum:
	sudo ./scripts/deploy-chartmuseum $(OPERATOR_REPO)

e2e-install-rancher: deps
	ginkgo --label-filter install -r -v ./e2e

e2e-iso-image: deps
	ginkgo --label-filter iso-image -r -v ./e2e

e2e-multi-cluster: deps
	ginkgo --timeout $(GINKGO_TIMEOUT)s --label-filter multi-cluster -r -v ./e2e

e2e-reset: deps
	ginkgo --label-filter reset -r -v ./e2e

e2e-prepare-archive: deps
	ginkgo --label-filter prepare-archive -r -v ./e2e
	
e2e-ui-rancher: deps
	ginkgo --label-filter ui -r -v ./e2e

e2e-uninstall-operator:
	ginkgo --label-filter uninstall-operator -r -v ./e2e

e2e-upgrade-node: deps
	ginkgo --label-filter upgrade-node -r -v ./e2e

e2e-upgrade-operator: deps
	ginkgo --label-filter upgrade-operator -r -v ./e2e

e2e-upgrade-rancher-manager: deps
	ginkgo --label-filter upgrade-rancher-manager -r -v ./e2e

start-cypress-tests:
	@./scripts/start-cypress-tests
