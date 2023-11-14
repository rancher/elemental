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
ifeq ($(TEST_TYPE),multi_cli)
	GINKGO_TIMEOUT=10800
endif

extract_kernel_init_squash:
	@VAR=$(ISO); \
	INITRD_FILE=$$(isoinfo -R -i $(ISO) -find -type f -name initrd -print 2>/dev/null); \
	KERNEL_FILE=$$(isoinfo -R -i $(ISO) -find -type f -name kernel -print 2>/dev/null); \
	[[ -z "$${KERNEL_FILE}" ]] && KERNEL_FILE=$$(isoinfo -R -i $(ISO) -find -type f -name linux -print 2>/dev/null); \
	isoinfo -x /rootfs.squashfs -R -i $(ISO) > $${VAR/\.iso/.squashfs} 2>/dev/null; \
	isoinfo -x $${INITRD_FILE} -R -i $(ISO) > $${VAR/\.iso/-initrd} 2>/dev/null; \
	isoinfo -x $${KERNEL_FILE} -R -i $(ISO) > $${VAR/\.iso/-kernel} 2>/dev/null

ipxe:
	@VAR=$(notdir $(ISO)); \
	echo "#!ipxe" > $(IPXE); \
	echo "set arch amd64" >> $(IPXE); \
	echo "set url tftp://10.0.2.2/$${VAR##*/}" >> $(IPXE); \
	echo "set kernel $${VAR/\.iso/-kernel}" >> $(IPXE); \
	echo "set initrd $${VAR/\.iso/-initrd}" >> $(IPXE); \
	echo "set rootfs $${VAR/\.iso/.squashfs}" >> $(IPXE); \
	echo "# set config http://example.com/machine-config" >> $(IPXE); \
	echo "# set cmdline cos.setup=/run/initramfs/live/cloud-config extra.values=1" >> $(IPXE); \
	echo "initrd \$${url}/\$${initrd}" >> $(IPXE); \
	echo "chain --autofree --replace \$${url}/\$${kernel} initrd=\$${initrd} ip=dhcp rd.cos.disable root=live:\$${url}/\$${rootfs} stages.initramfs[0].commands[0]=\"curl -k \$${config} > /run/initramfs/live/livecd-cloud-config.yaml\" console=tty1 console=ttyS0 \$${cmdline}" >> $(IPXE)

deps: 
	go install -mod=mod github.com/onsi/ginkgo/v2/ginkgo
	go install -mod=mod github.com/onsi/gomega
	go mod tidy

# Generate tests description file
generate-readme:
	@./scripts/generate-readme > README.md

# E2E tests
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
	sudo ./scripts/deploy-chartmuseum

e2e-install-rancher: deps
	ginkgo --label-filter install -r -v ./e2e

e2e-iso-image: deps
	ginkgo --label-filter iso-image -r -v ./e2e

e2e-multi-cluster: deps
	ginkgo --timeout $(GINKGO_TIMEOUT)s --label-filter multi-cluster -r -v ./e2e

e2e-reset: deps
	ginkgo --label-filter reset -r -v ./e2e

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
