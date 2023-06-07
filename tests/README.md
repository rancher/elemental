# Tests description for cypress/1.0.0/e2e/unit_tests

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
    - **It:** Check we can see our embedded hardware labels
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
    - **It:** Upgrade one node different methods if rke2 or k3s)

## `user.spec.ts`

- **Describe:** User role testing
    - **It:** Create elemental user
    - **It:** Create standard user
    - **It:** Elemental user should access the OS management menu
    - **It:** Standard user should not access the OS management menu

# Tests description for cypress/latest/e2e/unit_tests

## `advanced_filtering.spec.ts`

- **Describe:** Advanced filtering testing
    - **It:** Create fake machine inventories
    - **It:** Two machine inventories should appear by filtering on test-filter
    - **It:** One machine inventory should appear by filtering on test-filter-one
    - **It:** No machine inventory should appear by filtering on test-bad-filter
    - **It:** Delete all fake machine inventories

## `deploy_app.spec.ts`

- **Describe:** Deploy application in fresh Elemental Cluster
    - **It:** Deploy Alerting Drivers application
    - **It:** Remove Alerting Drivers application

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
    - **It:** Check we can see our embedded hardware labels
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
    - **It:** Check machine registration label name size
    - **It:** Check machine registration label value size
    - **It:** Create Machine registration we will use to test adding a node
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

## `seed_image.spec.ts`

- **Describe:** SeedImage testing
    - **It:** Create SeedImage with custom base image

## `upgrade.spec.ts`

- **Describe:** Upgrade tests
    - **It:** Create an OS Version Channels
    - **It:** Check OS Versions
    - **It:** Upgrade one node different methods if rke2 or k3s)
    - **It:** Cannot create two upgrade groups targeting the same cluster
    - **It:** Delete OS Versions
    - **It:** Delete OS Versions Channels

## `user.spec.ts`

- **Describe:** User role testing
    - **It:** Create elemental user
    - **It:** Create standard user
    - **It:** Elemental user should access the OS management menu
    - **It:** Standard user should not access the OS management menu

# Tests description for e2e

## `backup-restore_test.go`

- **Describe:** E2E - Install Backup/Restore Operator
    - **It:** Install Backup/Restore Operator
      -  **By:** Configuring Chart repository
      -  **By:** Installing rancher-backup-operator
      -  **By:** Waiting for rancher-backup-operator pod
- **Describe:** E2E - Test Backup/Restore
    - **It:** Do a backup
      -  **By:** Adding a backup resource
      -  **By:** Checking that the backup has been done
    - **It:** Do a restore
      -  **By:** Deleting some Elemental resources
      -  **By:** Adding a restore resource
      -  **By:** Checking that the restore has been done
      -  **By:** Checking cluster state after restore

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
      -  **By:** Removing the ISO

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
      -  **By:** Installing RKE2
      -  **By:** Configuring hardened cluster
      -  **By:** Starting RKE2
      -  **By:** Waiting for RKE2 to be started
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

## `uninstall-operator_test.go`

- **Describe:** E2E - Uninstall Elemental Operator
    - **It:** Uninstall Elemental Operator
      -  **By:** Testing cluster resource availability BEFORE operator uninstallation
      -  **By:** Uninstalling Operator via Helm
      -  **By:** Testing cluster resource availability AFTER operator uninstallation
      -  **By:** Checking that Elemental resources are gone
      -  **By:** Deleting cluster resource
      -  **By:** Testing cluster resource unavailability
    - **It:** Re-install Elemental Operator
      -  **By:** Installing Operator via Helm
      -  **By:** Creating a dumb MachineRegistration
      -  **By:** Creating cluster
      -  **By:** Testing cluster resource availability

## `upgrade_test.go`

- **Describe:** E2E - Upgrading Elemental Operator
    - **It:** Upgrade operator
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

