/*
Copyright © 2022 - 2023 SUSE LLC

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
interface hardwareLabels {
  [key: string]: string;
};

const hwLabels: hardwareLabels = {
  'CPUModel': '${System Data/CPU/Model}',
  'CPUVendor': '${System Data/CPU/Vendor}',
  'NumberBlockDevices': '${System Data/Block Devices/Number Devices}',
  'NumberNetInterface': '${System Data/Network/Number Interfaces}',
  'CPUVendorTotalCPUCores': '${System Data/CPU/Total Cores}',
  'TotalCPUThread': '${System Data/CPU/Total Threads}',
  'TotalMemory': '${System Data/Memory/Total Physical Bytes}'
};

// Generic commands
// ////////////////

Cypress.Commands.overwrite('type', (originalFn, subject, text, options = {}) => {
  options.delay = 100;
  return originalFn(subject, text, options);
});

// Add a delay between command without using cy.wait()
// https://github.com/cypress-io/cypress/issues/249#issuecomment-443021084
const COMMAND_DELAY = 1000;

for (const command of ['visit', 'click', 'trigger', 'type', 'clear', 'reload', 'contains']) {
  Cypress.Commands.overwrite(command, (originalFn, ...args) => {
    const origVal = originalFn(...args);

    return new Promise((resolve) => {
      setTimeout(() => {
        resolve(origVal);
      }, COMMAND_DELAY);
    });
  });
}; 

// Machine registration commands
// ///////////////////////////// 

// Create a machine registration
Cypress.Commands.add('createMachReg', (
  machRegName,
  namespace='fleet-default',
  checkLabels=false,
  checkAnnotations=false,
  checkInventoryLabels=false,
  checkInventoryAnnotations=false,
  checkIsoBuilding=false,
  customCloudConfig='',
  checkDefaultCloudConfig=true ) => {
  cy.clickNavMenu(["Dashboard"]);
  cy.getBySel('button-create-registration-endpoint')
    .click();
  cy.getBySel('name-ns-description-name')
    .type(machRegName);

  if (customCloudConfig != '') {
    cy.get('input[type="file"]')
      .attachFile({filePath: customCloudConfig});
  }

  checkLabels ? cy.addMachRegLabel('myLabel1', 'myLabelValue1') : null;
  checkAnnotations? cy.addMachRegAnnotation('myAnnotation1', 'myAnnotationValue1') : null;
  checkInventoryLabels ? cy.addMachInvLabel('myInvLabel1', 'myInvLabelValue1') : null;
  checkInventoryAnnotations ? cy.addMachInvAnnotation('myInvAnnotation1', 'myInvAnnotationValue1') : null;

  cy.getBySel('form-save')
    .contains('Create')
    .click();

  // Make sure the machine registration is created and active
  cy.contains('.masthead', 'Registration Endpoint: '+ machRegName + 'Active')
    .should('exist');

  // Check the namespace
  cy.contains('.masthead', 'Namespace: '+ namespace)
    .should('exist');

  // Make sure there is an URL registration in the Registration URL block
  cy.getBySel('registration-url')
    .contains(/https:\/\/.*elemental\/registration/);

  // Test ISO building feature
  if (checkIsoBuilding) {
    // Build the ISO according to the elemental operator version
    // Most of the time, it uses the latest dev version but sometimes
    // before releasing, we want to test staging/stable artifacts 
    cy.getBySel('select-os-version-build-iso')
      .click();
    // Never build from dev ISO in upgrade scenario
    if (utils.isCypressTag('upgrade')) {
      // Stable operator version is hardcoded for now
      // Will try to improve it in next version
      if (utils.isOperatorVersion('staging')) {
        // In rare case, we might want to test upgrading from staging to dev
        utils.isUpgradeOsChannel('dev') ? cy.contains('ISO x86_64 (unstable)').click(): null;
      } else {
          cy.contains('ISO x86_64 v1.2.3')
          .click();
      }
    } else if (utils.isOperatorVersion('registry.suse.com')) {
      cy.contains('ISO x86_64 v1.2.3')
        .click();
    } else {
      cy.contains('ISO x86_64 (unstable)')
        .click();
    }
    cy.getBySel('build-iso-btn')
      .click();
    cy.getBySel('build-iso-btn')
      .get('.icon-spin');
    // Download button is disabled while ISO is building
    cy.getBySel('download-iso-btn').should(($input) => {
      expect($input).to.have.attr('disabled')
    })
    // Download button is enabled once ISO building done
    cy.getBySel('download-iso-btn', { timeout: 600000 }).should(($input) => {
      expect($input).to.not.have.attr('disabled')
    })
    cy.getBySel('download-iso-btn')
      .click()
    cy.verifyDownload('.iso', { contains:true, timeout: 180000, interval: 5000 });
  }
  
  // Try to download the registration file and check it
  cy.getBySel('download-btn')
    .click();
  cy.verifyDownload(machRegName + '_registrationURL.yaml');
  cy.contains('Saving')
    .should('not.exist');

  // Check Cloud configuration
  // TODO: Maybe the check may be improved in one line
  if (checkDefaultCloudConfig) {
    cy.getBySel('yaml-editor-code-mirror')
      .should('include.text','config:')
      .should('include.text','cloud-config:')
      .should('include.text','users:')
      .should('include.text','- name: root')
      .should('include.text','passwd: root')
      .should('include.text','elemental:')
      .should('include.text','install:')
      .should('include.text','device: /dev/nvme0n1')
      .should('include.text','poweroff: true');
  }

  // Check label and annotation in YAML
  // For now, we can only check in YAML because the fields are disabled and we cannot check their content
  // It looks like we can use shadow DOM to catch it but too complicated for now
  cy.contains('Registration Endpoint')
    .click();
  checkLabels ? cy.checkMachRegLabel(machRegName, 'myLabel1', 'myLabelValue1') : null;
  checkAnnotations ? cy.checkMachRegAnnotation(machRegName, 'myAnnotation1', 'myAnnotationValue1') : null;
});

// Add Label to machine registration
Cypress.Commands.add('addMachRegLabel', (labelName, labelValue) => {
  cy.getBySel('labels-and-annotations-block')
    .contains('Registration Endpoint')
    .click();
  cy.get('[data-testid="add-label-mach-reg"] > .footer > .btn')
    .click();
  cy.get('[data-testid="add-label-mach-reg"] > .kv-container > .kv-item.key')
    .type(labelName);
  cy.get('[data-testid="add-label-mach-reg"] > .kv-container > .kv-item.value')
    .type(labelValue);
});

// Add Annotation to machine registration
Cypress.Commands.add('addMachRegAnnotation', (annotationName, annotationValue) => {
  cy.getBySel('labels-and-annotations-block')
    .contains('Registration Endpoint')
    .click();
  cy.get('[data-testid="add-annotation-mach-reg"] > .footer > .btn')
    .click();
  cy.get('[data-testid="add-annotation-mach-reg"] > .kv-container > .kv-item.key')
    .type(annotationName);
  cy.get('[data-testid="add-annotation-mach-reg"] > .kv-container > .kv-item.value')
    .type(annotationValue);
});

// Add Label to machine inventory
Cypress.Commands.add('addMachInvLabel', (labelName, labelValue, useHardwareLabels=true) => {
  cy.getBySel('labels-and-annotations-block')
    .contains('Inventory of Machines')
    .click();
  cy.get('[data-testid="add-label-mach-inv"] > .footer > .btn')
    .click();
  cy.get('[data-testid="add-label-mach-inv"] > .kv-container > .kv-item.key').type(labelName);
  cy.get('[data-testid="add-label-mach-inv"] > .kv-container > .kv-item.value').type(labelValue);
  if (useHardwareLabels) {
    let nthChildIndex = 7;
    for (const key in hwLabels) {
      cy.get('[data-testid="add-label-mach-inv"] > .footer > .btn')
        .click();
      cy.get(`[data-testid="add-label-mach-inv"] > .kv-container > :nth-child(${nthChildIndex}) > input`).type(key);
      // Following condition could be removed when we will release next Elemental UI (> 1.2.0)
      if (utils.isUIVersion('dev')) {
        cy.get(`[data-testid="add-label-mach-inv"] > .kv-container > :nth-child(${nthChildIndex + 1}) 
          > .value-container > [data-testid="text-area-auto-grow"]`).type(hwLabels[key], {parseSpecialCharSequences: false});
      } else {
        cy.get(`[data-testid="add-label-mach-inv"] > .kv-container > :nth-child(${nthChildIndex + 1})
          > .no-resize`).type(hwLabels[key], {parseSpecialCharSequences: false});
      };
      nthChildIndex += 3;
    };
  };
});

// Add Annotation to machine inventory
Cypress.Commands.add('addMachInvAnnotation', (annotationName, annotationValue) => {
  cy.getBySel('labels-and-annotations-block')
    .contains('Inventory of Machines')
    .click();
  cy.clickButton('Add Annotation');
  cy.get('[data-testid="add-annotation-mach-inv"] > .kv-container > .kv-item.key')
    .type(annotationName);
  cy.get('[data-testid="add-annotation-mach-inv"] > .kv-container > .kv-item.value')
    .type(annotationValue);
});

// Check machine inventory label in YAML
Cypress.Commands.add('checkMachInvLabel', (machRegName, labelName, labelValue, afterBoot=false, userHardwareLabels=true) => {
  if (afterBoot == false ) {
    cy.contains(machRegName)
      .click();
    cy.get('div.actions > .role-multi-action')
      .click()
    cy.contains('li', 'Edit YAML')
      .click();
    cy.contains('Registration Endpoint: '+ machRegName)
      .should('exist');
    cy.getBySel('yaml-editor-code-mirror')
      .contains(labelName + ': ' + labelValue);
    if (userHardwareLabels) {
      for (const key in hwLabels) {
        cy.getBySel('yaml-editor-code-mirror')
          .contains(key +': ' + hwLabels[key]);
      };
    };
    cy.clickButton('Cancel');
  } else {
    cy.getBySel('yaml-editor-code-mirror')
      .contains(labelName + ': ' + labelValue);
    if (userHardwareLabels) {
      for (const key in hwLabels) {
        cy.getBySel('yaml-editor-code-mirror')
          .contains(key +': ');
      };
    };
  }
});

// Check machine registration label in YAML
Cypress.Commands.add('checkMachRegLabel', (machRegName, labelName, labelValue) => {
  cy.contains(machRegName)
    .click();
  cy.get('div.actions > .role-multi-action')
    .click()
  cy.contains('li', 'Edit YAML')
    .click();
  cy.contains('Registration Endpoint: '+ machRegName)
    .should('exist');
  cy.getBySel('yaml-editor-code-mirror')
    .contains(labelName + ': ' + labelValue);
  cy.clickButton('Cancel');
});

// Check machine registration annotation in YAML
Cypress.Commands.add('checkMachRegAnnotation', ( machRegName, annotationName, annotationValue) => {
  cy.contains(machRegName)
    .click();
  cy.get('div.actions > .role-multi-action')
    .click()
  cy.contains('li', 'Edit YAML')
    .click();
  cy.contains('Registration Endpoint: '+ machRegName)
    .should('exist');
  cy.getBySel('yaml-editor-code-mirror')
    .contains(annotationName + ': ' + annotationValue);
  cy.clickButton('Cancel');
});

// Edit a machine registration
Cypress.Commands.add('editMachReg', ( machRegName, addLabel=false, addAnnotation=false, withYAML=false) => {
  cy.contains(machRegName)
    .click();
  // Select the 3dots button and edit configuration
  cy.get('div.actions > .role-multi-action')
    .click()
  if (withYAML) {
    cy.contains('li', 'Edit YAML')
      .click();
    cy.contains('metadata').as('meta')
    cy.get('@meta').click(0,0)
    cy.get('@meta').type('{end}{enter}  labels:{enter}  myLabel1: myLabelValue1');
    cy.contains('metadata').as('meta')
    cy.get('@meta').click(0,0)
    cy.get('@meta').type('{end}{enter}  annotations:{enter}  myAnnotation1: myAnnotationValue1');
  } else {
    cy.contains('li', 'Edit Config')
      .click();
    addLabel ? cy.addMachRegLabel('myLabel1', 'myLabelValue1' ) : null;
    addAnnotation ? cy.addMachRegAnnotation('myAnnotation1', 'myAnnotationValue1') : null;
  }
});

// Delete a machine registration
Cypress.Commands.add('deleteMachReg', (machRegName) => {
  cy.contains('Registration Endpoint')
    .click();
  /*  This code cannot be used anymore for now because of
      https://github.com/rancher/elemental/issues/714
      As it is not a blocker, we need to bypass it.
      Instead of selecting resource to delete by name
      we select all resources.
  cy.contains(machRegName)
    .parent()
    .parent()
    .click();
  */
  cy.get('[width="30"] > .checkbox-outer-container')
    .click();
  cy.getBySel('sortable-table-promptRemove')
    .contains('Delete')
    .click();
  cy.confirmDelete();
  // Timeout should fix this issue https://github.com/rancher/elemental/issues/643
  cy.contains(machRegName, {timeout: 20000})
    .should('not.exist')
});

// Machine Inventory commands
// /////////////////////////

// Import machine inventory
Cypress.Commands.add('importMachineInventory', (machineInventoryFile, machineInventoryName) => {
  cy.clickNavMenu(["Inventory of Machines"]);
  cy.getBySel('masthead-create-yaml')
    .click();
  cy.clickButton('Read from File');
  cy.get('input[type="file"]')
    .attachFile({filePath: machineInventoryFile});
  cy.getBySel('action-button-async-button')
    .contains('Create')
    .click();
  cy.contains('Creating')
    .should('not.exist');
  cy.contains(machineInventoryName)
    .should('exist');
});

Cypress.Commands.add('checkFilter', (filterName, testFilterOne, testFilterTwo, shouldNotMatch) => {
  cy.clickNavMenu(["Inventory of Machines"]);
  cy.clickButton("Add Filter");
  cy.get('.advanced-search-box').type(filterName);
  cy.get('.bottom-block > .role-primary').click();
  (testFilterOne) ? cy.contains('test-filter-one').should('exist') : cy.contains('test-filter-one').should('not.exist');
  (testFilterTwo) ? cy.contains('test-filter-two').should('exist') : cy.contains('test-filter-two').should('not.exist');
  (shouldNotMatch) ? cy.contains('shouldnotmatch').should('exist') : cy.contains('shouldnotmatch').should('not.exist');
});

Cypress.Commands.add('checkLabelSize', (sizeToCheck) => {
  cy.clickNavMenu(["Dashboard"]);
  cy.getBySel('button-create-registration-endpoint')
    .click();
  sizeToCheck == "name" ? cy.addMachInvLabel('labeltoolonggggggggggggggggggggggggggggggggggggggggggggggggggggg', 'mylabelvalue', false) : null;
  sizeToCheck == "value" ? cy.addMachInvLabel('mylabelname', 'valuetoolonggggggggggggggggggggggggggggggggggggggggggggggggggggg', false) : null;
  // A banner should appear alerting you about the size exceeded
  // Following condition could be removed when we will release next Elemental UI (> 1.2.0)
  utils.isUIVersion('dev') ? cy.get('[data-testid="banner-content"]') : cy.get('.banner > span');
  // Create button should be disabled
  cy.getBySel('form-save').should(($input) => {
    expect($input).to.have.attr('disabled')
  })
});

// OS Versions commands
// ////////////////////

// Add an OS version channel
Cypress.Commands.add('addOsVersionChannel', (channelVersion) => {
  let channelRepo = `registry.opensuse.org/isv/rancher/elemental/${channelVersion}/containers/rancher/elemental-channel:latest`;
  if (channelVersion == "stable") {
    channelRepo = 'registry.suse.com/rancher/elemental-channel:latest';
  }
  cy.clickNavMenu(["Advanced", "OS Version Channels"]);
  cy.getBySel('masthead-create')
    .contains('Create')
    .click();
  cy.getBySel('name-ns-description-name')
    .type(channelVersion + "-channel");
  cy.getBySel('os-version-channel-path')
    .type(channelRepo);
  cy.getBySel('form-save')
    .contains('Create')
    .click();
  // Status changes a lot right after the creation so let's wait 10 secondes
  // before checking
  // eslint-disable-next-line cypress/no-unnecessary-waiting
  cy.wait(10000);
  // Make sure the new channel is in Active state
  cy.contains("Active "+channelVersion+"-channel", {timeout: 50000});
});
