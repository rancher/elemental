# End-To-End Elemental tests quick description

## install_test.go (E2E - Install Rancher)
**Install and configure Rancher and libvirt by:**
- Installing K3s
- Starting K3s
- Waiting for K3s to be started
- Installing CertManager
- Installing Rancher
- Installing Elemental Operator
- Creating a new cluster
- Adding MachineRegistration
- Starting HTTP server for network installation
- Starting libvirtd
- Starting default network

## bootstrap_test.go (E2E - Bootstrapping node)
**Install node and add it in Rancher by:**
- Checking if VM name is set
- Configuring iPXE boot script for network installation
- Creating and installing VM
- Checking that the VM is available in Rancher
- Adding server role to predefined cluster
- Restarting the VM
- Checking that the VM is added in the cluster
- Checking VM ssh connection

## upgrade_test.go (2E - Upgrading node)
**Upgrade node by:**
- Checking if VM name is set
- Checking if upgrade type is set
- Triggering Upgrade in Rancher with <upgradeType> | Triggering Manual Upgrade
- Checking VM upgrade
- Cleaning upgrade orders
