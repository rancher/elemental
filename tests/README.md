# Tests description for cypress/e2e/unit_tests

# Tests description for e2e

## `bootstrap_test.go`

- **Describe:** E2E - Bootstrapping node
    - **It:** Provision the node
      -  **By:** Setting emulated TPM to +strconv.FormatBoolemulateTPM)
      -  **By:** Downloading installation config file
      -  **By:** Configuring iPXE boot script for network installation
      -  **By:** Adding registration file to ISO
      -  **By:** Installing node +h
    - **It:** Add the nodes in Rancher Manager
      -  **By:** Checking that node +h+ is available in Rancher
      -  **By:** Checking cluster state
      -  **By:** Incrementing number of nodes in +poolType+ pool
      -  **By:** Waiting for known cluster state before adding the nodes)
      -  **By:** Restarting +h+ to add it in the cluster
      -  **By:** Checking +h+ SSH connection
      -  **By:** Checking that TPM is correctly configured on +h
      -  **By:** Checking OS version on +h
      -  **By:** Configuring kubectl command on node +h
      -  **By:** Checking kubectl command on +h
      -  **By:** Checking cluster agent on +h
      -  **By:** Checking cluster state
      -  **By:** Checking cluster version on +h
      -  **By:** Rebooting +h
      -  **By:** Checking cluster agent on +h
      -  **By:** Checking cluster state after reboot

## `configure_test.go`

- **Describe:** E2E - Configure test
    - **It:** Configure Rancher and libvirt
      -  **By:** Creating a new cluster
      -  **By:** Creating cluster selectors
      -  **By:** Adding MachineRegistration
      -  **By:** Starting default network

## `install_test.go`

- **Describe:** E2E - Install Rancher Manager
    - **It:** Install Rancher Manager
      -  **By:** Installing K3s
      -  **By:** Configuring hardened cluster
      -  **By:** Starting K3s
      -  **By:** Waiting for K3s to be started
      -  **By:** Configuring Kubeconfig file
      -  **By:** Configuring Private CA
      -  **By:** Installing CertManager
      -  **By:** Installing Rancher
      -  **By:** Configuring kubectl to use Rancher admin user
      -  **By:** Workaround for upgrade test, restart Fleet controller and agent
      -  **By:** Installing Elemental Operator

## `logs_test.go`

- **Describe:** E2E - Getting logs node
    - **It:** Get the upstream cluster logs
      -  **By:** Downloading and executing tools to generate logs
      -  **By:** Collecting additionals logs with kubectl commands
      -  **By:** Collecting proxy log and make sure traffic went through it

## `suite_test.go`

*No test defined!*

## `ui_test.go`

- **Describe:** E2E - Bootstrap node for UI
    - **It:** Configure libvirt and bootstrap a node
      -  **By:** Downloading MachineRegistration
      -  **By:** Starting default network
      -  **By:** Configuring iPXE boot script for network installation
      -  **By:** Adding VM in default network
      -  **By:** Creating and installing VM
      -  **By:** Checking that the VM is available in Rancher
      -  **By:** Restarting the VM to add it in the cluster
      -  **By:** Checking VM connection

## `upgrade_test.go`

- **Describe:** E2E - Upgrading node
    - **It:** Upgrade node
      -  **By:** Checking if upgrade type is set
      -  **By:** Checking OS version on +h+ before upgrade
      -  **By:** Triggering Upgrade in Rancher with +upgradeType
      -  **By:** Checking VM upgrade on +h
      -  **By:** Checking OS version on +h+ after upgrade
      -  **By:** Checking cluster state after upgrade

# Tests description for install

## `install_suite_test.go`

*No test defined!*

## `install_test.go`

- **Describe:** Elemental Installation tests
  - **Context:** From ISO
    - **It:** can install
  - **Context:** From container
    - **It:** can install
    - **It:** has customization applied
      -  **By:** Checking we booted from the installed OS)
      -  **By:** Checking config file was run)

# Tests description for smoke

## `smoke_suite_test.go`

*No test defined!*

## `smoke_test.go`

- **Describe:** Elemental Smoke tests
  - **Context:** First boot
    - **It:** fmt.Sprintfstarts successfully %s on boot, unit)
    - **It:** has default mounts
    - **It:** has default cmdline
    - **It:** has the user added via cloud-init

