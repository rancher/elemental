# End-To-End Elemental UI Tests
This document is only a summarize of existing tests, more details are in the internal documentation.

## install_test.go (E2E - Install Rancher)
**Install Rancher Manager by:**
- Installing K3s
- Starting K3s
- Waiting for K3s to be started
- Installing CertManager
- Installing Rancher
- Installing Elemental Operator

## Execute basics Cypress tests (mainly the ones which do not need additional system)
**Execute first basics UI tests using Cypress:**

**1. Menu**
- Log in to Rancher Dashboard and accept terms and conditions
- Check that Elemental logo appears in the navigation menu
- Go in the Elemental screen and Check that all items appear

**2. Machine registration**
- Create machine registration with default options
- Create machine registration in custom namespace 
- Create machine registration with labels and annotations 
- Create machine registration with custom cloud-config
- Delete machine registration
- Edit a machine registration with edit config button
- Edit a machine registration with edit YAML button
- Clone a machine registration
- Download Machine registration YAML

## configure_test.go (E2E - Configure test)
**Configure Rancher and libvirt by:**
- Adding MachineRegistration
- Starting HTTP server for network installation
- Starting default network
- Create a VM using IPXE with latest assets

## Execute basics Cypress tests (mainly the ones which do not need additional system)
**Execute advanced UI tests using Cypress:**

  **1. Machine selector testing**
  - Testing selector without any rule
  - Testing selector with unmatching
  - Testing selector with matching rule

  **2. Machine inventory testing**
  - Check that machine inventory has been created
  - Create Elemental cluster
  - Check Elemental cluster status

  **3. Upgrade tests**
  - Create an OS Version Channels
  - Check OS Versions
  - Upgrade one node with OS Image Upgrades
