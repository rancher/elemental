# End-To-End Elemental tests quick description

## install_test.go (E2E - Install Rancher)
**Install Rancher Manager by:**
- Installing K3s
- Starting K3s
- Waiting for K3s to be started
- Installing CertManager
- Installing Rancher
- Installing Elemental Operator

## configure_test.go (E2E - Configure test)
**Configure Rancher and libvirt by:**
- Creating a new cluster
- Creating cluster selector
- Adding MachineRegistration
- Starting HTTP server for network installation
- Starting default network

## bootstrap_test.go (E2E - Bootstrapping node)
**Install node and add it in Rancher by:**
- Checking if VM name is set
- Configuring iPXE boot script for network installation
- Configuring emulated TPM if needed
- Creating and installing VM
- Checking that the VM is available in Rancher
- Increasing 'quantity' node of predefined cluster
- Restarting the VM to add it in the cluster
- Checking VM connection
- Checking cluster state
- Checking cluster version
- Rebooting the VM and checking that cluster is still healthy after

## upgrade_test.go (E2E - Upgrading node)
**Upgrade node by:**
- Checking if VM name is set
- Checking if upgrade type is set
- Triggering Upgrade in Rancher with <upgradeType>
- Checking VM upgrade
- Cleaning upgrade orders

<upgradeType> can be of type:
- managedOSVersionName
- osImage
- manual
