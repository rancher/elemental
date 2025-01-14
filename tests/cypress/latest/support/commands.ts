/*
Copyright © 2022 - 2025 SUSE LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
import 'cypress-file-upload';
import * as utils from "~/support/utils";

// Global
interface HardwareLabels {
  [key: string]: string;
}

const hwLabels: HardwareLabels = {
  'CPUModel': '${System Data/CPU/Model}',
  'CPUVendor': '${System Data/CPU/Vendor}',
  'NumberBlockDevices': '${System Data/Block Devices/Number Devices}',
  'NumberNetInterface': '${System Data/Network/Number Interfaces}',
  'CPUVendorTotalCPUCores': '${System Data/CPU/Total Cores}',
  'TotalCPUThread': '${System Data/CPU/Total Threads}',
  'TotalMemory': '${System Data/Memory/Total Physical Bytes}'
};

// Selectors
const selectors = {
  createButton: 'button-create-registration-endpoint',
  nameNsDescription: 'name-ns-description-name',
  formSave: 'form-save',
  registrationUrl: 'registration-url',
  selectOsVersionBuildMedia: 'select-os-version-build-media',
  selectMediaTypeBuildMedia: 'select-media-type-build-media',
  buildMediaBtn: 'build-media-btn',
  downloadMediaBtn: 'download-media-btn',
  yamlEditor: 'yaml-editor-code-mirror',
  labelsAndAnnotationsBlock: 'labels-and-annotations-block',
  addLabelMachReg: '[data-testid="add-label-mach-reg"]',
  addAnnotationMachReg: '[data-testid="add-annotation-mach-reg"]',
  addLabelMachInv: '[data-testid="add-label-mach-inv"]',
  addAnnotationMachInv: '[data-testid="add-annotation-mach-inv"]',
  downloadBtn: 'download-btn',
  mastheadCreateYaml: 'masthead-create-yaml',
  actionButtonAsync: 'action-button-async-button',
  bannerContent: 'banner-content',
  osVersionChannelPath: 'os-version-channel-path',
  mastheadCreate: 'masthead-create',
  nameNsDescriptionName: 'name-ns-description-name',
  sortableTablePromptRemove: 'sortable-table-promptRemove',
  inputKvItemKey: (index: number) => `[data-testid="input-kv-item-key-${index}"]`,
  kvItemValue: (index: number) => `[data-testid="kv-item-value-${index}"] > .value-container > [data-testid="value-multiline"]`,
  kvContainer: '[data-testid="add-label-mach-inv"] > .kv-container > :nth-child',
  textAreaAutoGrow: '[data-testid="text-area-auto-grow"]',
  //osVersionChannelActiveText: (channelVersion: string) => `Active ${channelVersion}-channel`,
};

// Machine registration commands
// ///////////////////////////// 

// Create a machine registration
Cypress.Commands.add('createMachReg', (
  machRegName: string,
  namespace: string = 'fleet-default',
  checkLabels: boolean = false,
  checkAnnotations: boolean = false,
  checkInventoryLabels: boolean = false,
  checkInventoryAnnotations: boolean = false,
  checkIsoBuilding: boolean = false,
  customCloudConfig: string = '',
  checkDefaultCloudConfig: boolean = true
) => {
  cy.clickNavMenu(["Dashboard"]);
  cy.getBySel(selectors.createButton).click();
  cy.getBySel(selectors.nameNsDescription).type(machRegName);

  if (customCloudConfig) {
    cy.get('input[type="file"]').attachFile({ filePath: customCloudConfig });
  }

  checkLabels && cy.addMachRegLabel('myLabel1', 'myLabelValue1');
  checkAnnotations && cy.addMachRegAnnotation('myAnnotation1', 'myAnnotationValue1');
  checkInventoryLabels && cy.addMachInvLabel('myInvLabel1', 'myInvLabelValue1');
  checkInventoryAnnotations && cy.addMachInvAnnotation('myInvAnnotation1', 'myInvAnnotationValue1');

  cy.getBySel(selectors.formSave)
    .contains('Create')
    .click();

  // Make sure the machine registration is created and active
  cy.contains('.masthead', 'Registration Endpoint: ' + machRegName + 'Active')
    .should('exist');
  // Check the namespace
  cy.contains('.masthead', 'Namespace: ' + namespace)
  .should('exist');

  // Make sure there is a URL registration in the Registration URL block
  cy.getBySel(selectors.registrationUrl)
    .contains(/https:\/\/.*elemental\/registration/);

  // Test ISO building feature
  if (checkIsoBuilding) {
    // Build the ISO according to the elemental operator version
    // Most of the time, it uses the latest dev version but sometimes
    // before releasing, we want to test staging/stable artifacts 

    if (utils.isCypressTag('upgrade') || utils.isUIVersion('stable')) {
      cy.getBySel(selectors.selectOsVersionBuildMedia).click();
    } else {
      cy.getBySel(selectors.selectMediaTypeBuildMedia).click();
      cy.contains(utils.isBootType('raw') ? 'Raw' : 'Iso').click();
      cy.getBySel(selectors.selectOsVersionBuildMedia).click();
    }

    // Never build from dev ISO in upgrade scenario
    if (utils.isCypressTag('upgrade')) {
      // Stable operator version is hardcoded for now
      // Will try to improve it in next version
      if (utils.isOperatorVersion('staging')) {
        // In rare case, we might want to test upgrading from staging to dev
        if (utils.isUpgradeOsChannel('dev')) {
          cy.contains('(unstable)').click();
        }
      } else if (utils.isOperatorVersion('marketplace')) {
        cy.contains(Cypress.env('os_version_install')).click();
      } else {
          if (utils.isBootType('raw')) {
            cy.contains(new RegExp(`OS.*${Cypress.env('container_stable_os_version')}`)).click();
          } else {
            cy.contains(new RegExp(`ISO.*${Cypress.env('iso_stable_os_version')}`)).click();
          }
      }
    } else if (utils.isOperatorVersion('registry.suse.com') || utils.isOperatorVersion('marketplace')) {
      cy.contains(Cypress.env('os_version_install')).click();
    // Sometimes we want to test dev/staging operator version with stable OS version
    } else if ( utils.isOsVersion('stable') && utils.isOperatorVersion('dev') || utils.isOperatorVersion('staging')) {
      cy.contains(Cypress.env('iso_stable_os_version')).click();
    } else {
      cy.contains('(unstable)').click();
    }

    cy.getBySel(selectors.buildMediaBtn).click();
    cy.getBySel(selectors.buildMediaBtn).get('.icon-spin');
    // Download button is disabled while ISO is building
    cy.getBySel(selectors.downloadMediaBtn).should('have.attr', 'disabled');
    // Download button is enabled once ISO building done
    cy.getBySel(selectors.downloadMediaBtn, { timeout: 600000 }).should('not.have.attr', 'disabled');
    cy.getBySel(selectors.downloadMediaBtn).click();

    if (utils.isBootType('raw')) {
      // .img will be removed in next elemental UI, only .raw will be available
      let extension = 'img';
      !utils.isRancherManagerVersion('2.8') ? extension = 'raw' : null;
      cy.verifyDownload('.'+extension, { contains: true, timeout: 300000, interval: 5000 });
    } else {
      cy.verifyDownload('.iso', { contains: true, timeout: 300000, interval: 5000 });
    }

    // Check we can download the checksum file (only in dev UI for Rancher 2.9 / 2.10)
    if (!utils.isRancherManagerVersion('2.8') && utils.isUIVersion('dev')) {
      cy.getBySel('download-checksum-btn').click();
      cy.verifyDownload('.sh256', { contains: true, timeout: 60000, interval: 5000 });
    }
  }

  // Try to download the registration file and check it
  cy.getBySel(selectors.downloadBtn).click();
  cy.verifyDownload(`${machRegName}_registrationURL.yaml`);
  cy.contains('Saving').should('not.exist');

  // Check Cloud configuration
  if (checkDefaultCloudConfig) {
    const cloudConfigChecks = [
      'config:', 'cloud-config:', 'users:', '- name: root', 'passwd: root',
      'elemental:', 'install:', 'device-selector:', '- key: Name', 'operator: In',
      'values:', '- /dev/sda', '- /dev/vda', '- /dev/nvme0', '- key: Size',
      'operator: Gt', 'values:', '- 25Gi', 'reboot: true', 'snapshotter:',
      'type: btrfs', 'reset:', 'reboot: true', 'reset-oem: true', 'reset-persistent: true'
    ];

    cloudConfigChecks.forEach(text => cy.getBySel(selectors.yamlEditor).should('include.text', text));
  }

  // Check label and annotation in YAML
  cy.contains('Registration Endpoint').click();
  checkLabels && cy.checkMachRegLabel(machRegName, 'myLabel1', 'myLabelValue1');
  checkAnnotations && cy.checkMachRegAnnotation(machRegName, 'myAnnotation1', 'myAnnotationValue1');
});

// Add Label to machine registration
Cypress.Commands.add('addMachRegLabel', (labelName: string, labelValue: string) => {
  cy.getBySel(selectors.labelsAndAnnotationsBlock)
    .contains('Registration Endpoint')
    .click();
  cy.get(`${selectors.addLabelMachReg} > .footer > .btn`).click();
  cy.get(`${selectors.addLabelMachReg} > .kv-container > .kv-item.key`).type(labelName);
  cy.get(`${selectors.addLabelMachReg} > .kv-container > .kv-item.value`).type(labelValue);
});

// Add Annotation to machine registration
Cypress.Commands.add('addMachRegAnnotation', (annotationName: string, annotationValue: string) => {
  cy.getBySel(selectors.labelsAndAnnotationsBlock)
    .contains('Registration Endpoint')
    .click();
  cy.get(`${selectors.addAnnotationMachReg} > .footer > .btn`).click();
  cy.get(`${selectors.addAnnotationMachReg} > .kv-container > .kv-item.key`).type(annotationName);
  cy.get(`${selectors.addAnnotationMachReg} > .kv-container > .kv-item.value`).type(annotationValue);
});

// Add Label to machine inventory
Cypress.Commands.add('addMachInvLabel', (labelName: string, labelValue: string, useHardwareLabels: boolean = true) => {
  cy.getBySel(selectors.labelsAndAnnotationsBlock)
    .contains('Inventory of Machines')
    .click();
  cy.get(`${selectors.addLabelMachInv} > .footer > .btn`).click();
  cy.get(`${selectors.addLabelMachInv} > .kv-container > .kv-item.key`).type(labelName);
  cy.get(`${selectors.addLabelMachInv} > .kv-container > .kv-item.value`).type(labelValue);

  if (useHardwareLabels) {
    const isRancher28 = utils.isRancherManagerVersion('2.8');
    let index = isRancher28 ? 7 : 1;

    for (const key in hwLabels) {
      cy.get(`${selectors.addLabelMachInv} > .footer > .btn`).click();
      const keySelector = isRancher28 ? `${selectors.kvContainer}(${index}) > input` : selectors.inputKvItemKey(index);
      const valueSelector = isRancher28 ? `${selectors.kvContainer}(${index + 1}) > .value-container > ${selectors.textAreaAutoGrow}` : selectors.kvItemValue(index);

      cy.get(keySelector).type(key);
      cy.get(valueSelector).type(hwLabels[key], { parseSpecialCharSequences: false });

      index += isRancher28 ? 3 : 1;
    }
  }
});

// Add Annotation to machine inventory
Cypress.Commands.add('addMachInvAnnotation', (annotationName: string, annotationValue: string) => {
  cy.getBySel(selectors.labelsAndAnnotationsBlock)
    .contains('Inventory of Machines')
    .click();
  cy.clickButton('Add Annotation');
  cy.get(`${selectors.addAnnotationMachInv} > .kv-container > .kv-item.key`).type(annotationName);
  cy.get(`${selectors.addAnnotationMachInv} > .kv-container > .kv-item.value`).type(annotationValue);
});

// Check machine inventory label in YAML
Cypress.Commands.add('checkMachInvLabel', (machRegName: string, labelName: string, labelValue: string, afterBoot: boolean = false, useHardwareLabels: boolean = true) => {
  if (!afterBoot) {
    cy.contains(machRegName).click();
    cy.get('div.actions > .role-multi-action').click();
    cy.contains('li', 'Edit YAML').click();
    cy.contains(`Registration Endpoint: ${machRegName}`).should('exist');
    cy.getBySel(selectors.yamlEditor).contains(`${labelName}: ${labelValue}`);

    if (useHardwareLabels) {
      for (const key in hwLabels) {
        cy.getBySel(selectors.yamlEditor).contains(`${key}: ${hwLabels[key]}`);
      }
    }

    cy.clickButton('Cancel');
  } else {
    cy.getBySel(selectors.yamlEditor).contains(`${labelName}: ${labelValue}`);

    if (useHardwareLabels) {
      for (const key in hwLabels) {
        cy.getBySel(selectors.yamlEditor).contains(`${key}: `);
      }
    }
  }
});

// Check machine registration label in YAML
Cypress.Commands.add('checkMachRegLabel', (machRegName: string, labelName: string, labelValue: string) => {
  cy.contains(machRegName).click();
  cy.get('div.actions > .role-multi-action').click();
  cy.contains('li', 'Edit YAML').click();
  cy.contains('Registration Endpoint: ' + machRegName).should('exist');
  cy.getBySel(selectors.yamlEditor).contains(`${labelName}: ${labelValue}`);
  cy.clickButton('Cancel');
});

// Check machine registration annotation in YAML
Cypress.Commands.add('checkMachRegAnnotation', (machRegName: string, annotationName: string, annotationValue: string) => {
  cy.contains(machRegName).click();
  cy.get('div.actions > .role-multi-action').click();
  cy.contains('li', 'Edit YAML').click();
  cy.contains('Registration Endpoint: ' + machRegName).should('exist');
  cy.getBySel(selectors.yamlEditor).contains(`${annotationName}: ${annotationValue}`);
  cy.clickButton('Cancel');
});

// Edit a machine registration
Cypress.Commands.add('editMachReg', (machRegName: string, addLabel: boolean = false, addAnnotation: boolean = false, withYAML: boolean = false) => {
  cy.contains(machRegName).click();
  cy.get('div.actions > .role-multi-action').click();

  if (withYAML) {
    cy.contains('li', 'Edit YAML').click();
    cy.contains('metadata').as('meta');
    cy.get('@meta').click(0, 0);
    cy.get('@meta').type('{end}{enter}  labels:{enter}  myLabel1: myLabelValue1');
    cy.contains('metadata').as('meta');
    cy.get('@meta').click(0, 0);
    cy.get('@meta').type('{end}{enter}  annotations:{enter}  myAnnotation1: myAnnotationValue1');
  } else {
    cy.contains('li', 'Edit Config').click();
    addLabel && cy.addMachRegLabel('myLabel1', 'myLabelValue1');
    addAnnotation && cy.addMachRegAnnotation('myAnnotation1', 'myAnnotationValue1');
  }
});

// Delete a machine registration
Cypress.Commands.add('deleteMachReg', (machRegName: string) => {
  cy.contains('Registration Endpoint').click();
  cy.get('[width="30"] > .checkbox-outer-container').click();
  cy.getBySel(selectors.sortableTablePromptRemove).contains('Delete').click();
  cy.confirmDelete();
  cy.contains(machRegName, { timeout: 20000 }).should('not.exist');
});

// Machine Inventory commands
// /////////////////////////

// Import machine inventory
Cypress.Commands.add('importMachineInventory', (machineInventoryFile: string, machineInventoryName: string) => {
  cy.clickNavMenu(["Inventory of Machines"]);
  cy.getBySel(selectors.mastheadCreateYaml).click();
  cy.clickButton('Read from File');
  cy.get('input[type="file"]').attachFile({ filePath: machineInventoryFile });
  cy.getBySel(selectors.actionButtonAsync).contains('Create').click();
  cy.contains('Creating').should('not.exist');
  cy.contains(machineInventoryName).should('exist');
});

Cypress.Commands.add('checkFilter', (filterName: string, testFilterOne: boolean, testFilterTwo: boolean, shouldNotMatch: boolean) => {
  cy.clickNavMenu(["Inventory of Machines"]);
  cy.clickButton('Add Filter');
  cy.get('.advanced-search-box').type(filterName);
  cy.get('.bottom-block > .role-primary').click();
  testFilterOne ? cy.contains('test-filter-one').should('exist') : cy.contains('test-filter-one').should('not.exist');
  testFilterTwo ? cy.contains('test-filter-two').should('exist') : cy.contains('test-filter-two').should('not.exist');
  shouldNotMatch ? cy.contains('shouldnotmatch').should('exist') : cy.contains('shouldnotmatch').should('not.exist');
});

// Check label size
Cypress.Commands.add('checkLabelSize', (sizeToCheck: string) => {
  cy.clickNavMenu(["Dashboard"]);
  cy.getBySel(selectors.createButton).click();

  if (sizeToCheck === "name") {
    cy.addMachInvLabel('labeltoolonggggggggggggggggggggggggggggggggggggggggggggggggggggg', 'mylabelvalue', false);
  } else if (sizeToCheck === "value") {
    cy.addMachInvLabel('mylabelname', 'valuetoolonggggggggggggggggggggggggggggggggggggggggggggggggggggg', false);
  }

  // A banner should appear alerting you about the size exceeded
  //cy.get(selectors.bannerContent);
  cy.getBySel(selectors.bannerContent);

  // Create button should be disabled
  cy.getBySel(selectors.formSave).should('have.attr', 'disabled');
});

// OS Versions commands
// ////////////////////

// Add an OS version channel
Cypress.Commands.add('addOsVersionChannel', (channelVersion: string) => {
  let channelRepo = "";

  switch (channelVersion) {
    case "stable":
      channelRepo = 'registry.suse.com/rancher/elemental-channel/sl-micro:6.0-baremetal';
      break;
    case "dev":
      channelRepo = 'registry.opensuse.org/isv/rancher/elemental/dev/containers/rancher/elemental-unstable-channel:latest';
      break;
    default:
      cy.log("Channel not found");
      return;
  }

  cy.clickNavMenu(["Advanced", "OS Version Channels"]);
  cy.getBySel(selectors.mastheadCreate).contains('Create').click();
  cy.getBySel(selectors.nameNsDescriptionName).type(`${channelVersion}-channel`);
  cy.getBySel(selectors.osVersionChannelPath).type(channelRepo);
  cy.getBySel(selectors.formSave).contains('Create').click();

  // Status changes a lot right after the creation so let's wait 10 seconds
  // before checking
  // eslint-disable-next-line cypress/no-unnecessary-waiting
  cy.wait(10000);

  // Make sure the new channel is in Active state
  cy.contains(new RegExp('Active.*'+channelVersion+'-channel'), { timeout: 50000 });
});