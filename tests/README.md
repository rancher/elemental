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

## `elemental_operator.spec.ts`

- **Describe:** Install Elemental Operator
    - **It:** Add local chartmuseum repo
    - **It:** Install latest dev Elemental operator
    - **It:** Install latest stable Elemental operator

## `elemental_plugin.spec.ts`

- **Describe:** Install Elemental plugin
    - **It:** Add elemental-ui repo
    - **It:** Add rancher-ui-plugin-charts repo
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

## `reset.spec.ts`

- **Describe:** Reset testing
    - **It:** Enable reset in machine inventory
    - **It:** Reset node by deleting the cluster
    - **It:** Create Elemental cluster

## `upgrade-operator.spec.ts`

- **Describe:** Elemental operator upgrade tests
    - **It:** Add elemental-operator dev repo
    - **It:** Upgrade Elemental operator
    - **It:** Check Elemental UI after upgrade

## `upgrade-ui-extension.spec.ts`

- **Describe:** UI extension upgrade tests
    - **It:** Add elemental-ui dev repo
    - **It:** Upgrade Elemental UI extension
    - **It:** Check Elemental UI after upgrade

## `upgrade.spec.ts`

- **Describe:** Upgrade tests
    - **It:** Delete stable channel for RKE2 upgrade
    - **It:** Add dev channel for RKE2 upgrade
    - **It:** Check OS Versions
    - **It:** Upgrade one node different methods if rke2 or k3s
    - **It:** Cannot create two upgrade groups targeting the same cluster
    - **It:** Delete OS Versions Channels

## `user.spec.ts`

- **Describe:** User role testing
    - **It:** Create standard user
    - **It:** Create elemental user
    - **It:** Elemental user should access the OS management menu
    - **It:** Standard user should not access the OS management menu

# Tests description for e2e

## `airgap_test.go`

- **Describe:** E2E - Build the airgap archive
    - **It:** Execute the script to build the archive
- **Describe:** E2E - Deploy K3S/Rancher in airgap environment
    - **It:** Create the rancher-manager machine
      -  **By:** Updating the default network configuration
      -  **By:** Creating the Rancher Manager VM
    - **It:** Install K3S/Rancher in the rancher-manager machine
      -  **By:** Sending the archive file into the rancher server
      -  **By:** Deploying airgap infrastructure by executing the deploy script
      -  **By:** Getting the kubeconfig file of the airgap cluster
      -  **By:** Installing kubectl
      -  **By:** Installing CertManager
      -  **By:** Installing Rancher
      -  **By:** Installing Elemental Operator

## `app_test.go`

- **Describe:** E2E - Install a simple application
    - **It:** Install HelloWorld application
      -  **By:** Installing application
- **Describe:** E2E - Checking a simple application
    - **It:** Check HelloWorld application
      -  **By:** Scaling the deployment to the number of nodes
      -  **By:** Waiting for deployment to be rollout
      -  **By:** Checking application

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
      -  **By:** Downloading MachineRegistration file
      -  **By:** Configuring iPXE boot script for network installation
      -  **By:** Installing node +h
      -  **By:** Checking SeedImage cloud-config on +h
    - **It:** Add the nodes in Rancher Manager
      -  **By:** Checking that node +h+ is available in Rancher
      -  **By:** Checking cluster state
      -  **By:** Incrementing number of nodes in +poolType+ pool
      -  **By:** Waiting for known cluster state before adding the nodes
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
    - **It:** Deploy a new cluster
      -  **By:** Creating a cluster
      -  **By:** Creating cluster selectors
      -  **By:** Adding MachineRegistration
    - **It:** Configure Libvirt if needed
      -  **By:** Starting default network

## `install_test.go`

- **Describe:** E2E - Install Rancher Manager
    - **It:** Install upstream K8s cluster
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
    - **It:** Install Rancher Manager
      -  **By:** Configuring kubectl to use Rancher admin user
      -  **By:** Workaround for upgrade test, restart Fleet controller and agent
    - **It:** Install Elemental Operator if needed
      -  **By:** Installing Operator for CLI tests

## `logs_test.go`

- **Describe:** E2E - Getting logs node
    - **It:** Get the upstream cluster logs
      -  **By:** Downloading and executing tools to generate logs
      -  **By:** Collecting additionals logs with kubectl commands
      -  **By:** Collecting proxy log and make sure traffic went through it

## `multi-cluster_test.go`

- **Describe:** E2E - Bootstrapping nodes
    - **It:** Configure Libvirt
      -  **By:** Starting default network
    - **It:** Configure and create ISO image
      -  **By:** Adding MachineRegistration
      -  **By:** Downloading MachineRegistration file
      -  **By:** Creating ISO from SeedImage
    - **It:** Downloading ISO built by SeedImage
    - **It:** Create clusters and deploy nodes
      -  **By:** Creating cluster +createdClusterName
      -  **By:** Creating cluster selector for cluster +createdClusterName
      -  **By:** Installing node +h+ on cluster +createdClusterName
      -  **By:** Restarting +h+ to add it in cluster +createdClusterName
      -  **By:** Checking +h+ SSH connection
      -  **By:** Waiting for cluster +createdClusterName+ to be Active
      -  **By:** Waiting for cluster +c+ to be Active

## `reset_test.go`

- **Describe:** E2E - Test the reset feature
    - **It:** Reset one node in the cluster
      -  **By:** Configuring reset at MachineInventory level
      -  **By:** Deleting and removing the node from the cluster
      -  **By:** Checking that MachineInventory is deleted
      -  **By:** Checking that MachineInventory is back after the reset
      -  **By:** Checking cluster state

## `seedImage_test.go`

- **Describe:** E2E - Creating ISO image
    - **It:** Configure and create ISO image
      -  **By:** Adding SeedImage
      -  **By:** Setting emulated TPM to +strconv.FormatBoolemulateTPM
    - **It:** Download ISO built by SeedImage

## `suite_test.go`

*No test defined!*

## `ui_test.go`

- **Describe:** E2E - Bootstrap node for UI
    - **It:** Configure libvirt and bootstrap a node
      -  **By:** Downloading MachineRegistration
      -  **By:** Starting default network
      -  **By:** Configuring iPXE boot script for network installation
      -  **By:** Installing node +h
    - **It:** Add the nodes in Rancher Manager
      -  **By:** Restarting +h+ to add it in the cluster
      -  **By:** Checking +h+ SSH connection
      -  **By:** Checking that TPM is correctly configured on +h
      -  **By:** Checking OS version on +h

## `uninstall-operator_test.go`

- **Describe:** E2E - Uninstall Elemental Operator
    - **It:** Uninstall Elemental Operator
      -  **By:** Testing cluster resource availability BEFORE operator uninstallation
      -  **By:** Uninstalling Operator via Helm
      -  **By:** Testing cluster resource availability AFTER operator uninstallation
      -  **By:** Checking that Elemental resources are gone
      -  **By:** Checking that Elemental operator CRDs cannot be reinstalled
      -  **By:** Deleting cluster resource
      -  **By:** Removing finalizers from MachineInventory/Machine
      -  **By:** Testing cluster resource unavailability
    - **It:** Re-install Elemental Operator
      -  **By:** Installing Operator via Helm
      -  **By:** Creating a dumb MachineRegistration
      -  **By:** Creating cluster
      -  **By:** Testing cluster resource availability

## `upgrade_test.go`

- **Describe:** E2E - Upgrading Elemental Operator
    - **It:** Upgrade operator
- **Describe:** E2E - Upgrading Rancher Manager
    - **It:** Upgrade Rancher Manager
- **Describe:** E2E - Upgrading node
    - **It:** Upgrade node
      -  **By:** Checking if upgrade type is set
      -  **By:** Checking OS version on +h+ before upgrade
      -  **By:** Triggering Upgrade in Rancher with +upgradeType
      -  **By:** Checking VM upgrade on +h
      -  **By:** Checking OS version on +h+ after upgrade
      -  **By:** Checking cluster state after upgrade

