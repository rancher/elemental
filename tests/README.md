# Tests description for cypress/e2e/unit_tests

## `advanced_filtering.spec.ts`

- **Describe:** Advanced filtering testing
    - **It:** Create fake machine inventories
    - **It:** Two machine inventories should appear by filtering on test-filter
    - **It:** One machine inventory should appear by filtering on test-filter-one
    - **It:** No machine inventory should appear by filtering on test-bad-filter
    - **It:** Delete all fake machine inventories

## `deploy_app.spec.ts`

- **Describe:** Deploy application in fresh Elemental Cluster
    - **It:** Deploy CIS Benchmark application
    - **It:** Remove CIS Benchmark application

## `elemental_plugin.spec.ts`

- **Describe:** Install Elemental plugin
    - **It:** Add elemental-ui repo
    - **It:** Enable extension support
    - **It:** Install Elemental plugin

## `first_connection.spec.ts`

- **Describe:** First login on Rancher
    - **It:** Log in and accept terms and conditions

## `machine_inventory.spec.ts`

- **Describe:** Machine inventory testing
    - **It:** Check that machine inventory has been created
    - **It:** Create Elemental cluster
    - **It:** Check Elemental cluster status

## `machine_registration.spec.ts`

- **Describe:** Machine registration testing
    - **It:** Create machine registration with default options
    - **It:** Create machine registration with labels and annotations
    - **It:** Delete machine registration
    - **It:** Edit a machine registration with edit config button
    - **It:** Edit a machine registration with edit YAML button
    - **It:** Clone a machine registration
    - **It:** Download Machine registration YAML
    - **It:** Create Machine registration we will use to test adding a node

## `machine_selector.spec.ts`

- **Describe:** Machine selector testing
    - **It:** Testing selector without any rule
    - **It:** Testing selector with unmatching rule
    - **It:** Testing selector with matching rule

## `menu.spec.ts`

- **Describe:** Menu testing
    - **It:** Check Elemental logo
    - **It:** Check Elemental menu

## `upgrade.spec.ts`

- **Describe:** Upgrade tests
    - **It:** Create an OS Version Channels
    - **It:** Check OS Versions
    - **It:** Delete OS Versions
    - **It:** Delete OS Versions Channels
    - **It:** Upgrade one node with OS Image Upgrades

## `user.spec.ts`

- **Describe:** User role testing
    - **It:** Create elemental user
    - **It:** Create standard user
    - **It:** Elemental user should access the OS management menu
    - **It:** Standard user should not access the OS management menu

# Tests description for e2e

## `bootstrap_test.go`

- **Describe:** E2E - Bootstrapping node
    - **It:** Provision the node
      -  **By:** Setting emulated TPM to +strconv.FormatBoolemulateTPM)
      -  **By:** Downloading installation config file
      -  **By:** Configuring iPXE boot script for network installation
      -  **By:** Adding registration file to ISO
      -  **By:** Installing node +h
    - **It:** Add the node in Rancher Manager
      -  **By:** Checking that node +h+ is available in Rancher
      -  **By:** Ensuring that the cluster is in healthy state
      -  **By:** Increasing quantity node of predefined cluster
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
      -  **By:** Starting K3s
      -  **By:** Waiting for K3s to be started
      -  **By:** Configuring Private CA
      -  **By:** Installing CertManager
      -  **By:** Installing Rancher
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
      -  **By:** Showing OS version before upgrade
      -  **By:** Triggering Upgrade in Rancher with +upgradeType
      -  **By:** Triggering Manual Upgrade
      -  **By:** Checking VM upgrade
      -  **By:** Showing OS version after upgrade
      -  **By:** Cleaning upgrade orders

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

