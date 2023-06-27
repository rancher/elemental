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

import { TopLevelMenu } from '~/support/toplevelmenu';
import { Elemental } from '~/support/elemental';
import '~/support/commands';
import filterTests from '~/support/filterTests.js';
import * as utils from "~/support/utils";

Cypress.config();
describe('Machine registration testing', () => {
  const elemental     = new Elemental();
  const elementalUser = "elemental-user"
  const topLevelMenu  = new TopLevelMenu();
  const uiAccount     = Cypress.env('ui_account');
  const uiPassword    = "rancherpassword"

  before(() => {
    (uiAccount == "user") ? cy.login(elementalUser, uiPassword) : cy.login();
    cy.visit('/');

    // Open the navigation menu
    topLevelMenu.openIfClosed();

    // Click on the Elemental's icon
    elemental.accessElementalMenu(); 

    // Create OS Version Channels from which to build the ISO
    // No need to create one if we test stable because it is already created
    // by the elemental-operator
    utils.isOperatorVersion('dev') ? cy.addOsVersionChannel({channelVersion:'dev'}): null;
    utils.isOperatorVersion('staging') ? cy.addOsVersionChannel({channelVersion: 'staging'}): null;
  });

  beforeEach(() => {
    (uiAccount == "user") ? cy.login(elementalUser, uiPassword) : cy.login();
    cy.visit('/');

    // Open the navigation menu
    topLevelMenu.openIfClosed();

    // Click on the Elemental's icon
    elemental.accessElementalMenu(); 
    
    // Delete all files previously downloaded
    cy.exec('rm cypress/downloads/*', {failOnNonZeroExit: false});

    // Delete all existing machine registrations
    cy.getBySel('manage-reg-btn')
      .click();
    cy.get('.outlet > header').contains('Registration Endpoints');
    cy.get('body').then(($body) => {
      if (!$body.text().includes('There are no rows to show.')) {
        cy.deleteAllResources();
      };
    });
  });

  filterTests(['main'], () => {
    it('Create machine registration with default options', () => {
      cy.createMachReg({machRegName: 'default-options-test'});
    });

    it('Create machine registration with labels and annotations', () => {
      cy.createMachReg({machRegName: 'labels-annotations-test',
        checkLabels: true,
        checkAnnotations: true});
    });

    it('Delete machine registration', () => {
      cy.createMachReg({machRegName: 'delete-test'});
      cy.deleteMachReg({machRegName: 'delete-test'});
    });

    it('Edit a machine registration with edit config button', () => {
      cy.createMachReg({machRegName: 'edit-config-test'});
      cy.editMachReg({machRegName: 'edit-config-test',
        addLabel: true,
        addAnnotation: true });
      cy.getBySel('form-save')
        .contains('Save')
        .click();

      // Check that we can see our label and annotation in the YAML
      cy.checkMachRegLabel({machRegName: 'edit-config-test',
        labelName: 'myLabel1',
        labelValue: 'myLabelValue1'});
      cy.checkMachRegAnnotation({machRegName: 'edit-config-test',
        annotationName: 'myAnnotation1',
        annotationValue: 'myAnnotationValue1'});
    });

    it('Edit a machine registration with edit YAML button', () => {
      cy.createMachReg({machRegName: 'edit-yaml-test'});
      cy.editMachReg({machRegName: 'edit-yaml-test',
        addLabel: true,
        addAnnotation: true,
        withYAML: true });
      cy.getBySel('action-button-async-button')
        .contains('Save')
        .click();

      // Check that we can see our label and annotation in the YAML
      cy.checkMachRegLabel({machRegName: 'edit-yaml-test',
        labelName: 'myLabel1',
        labelValue: 'myLabelValue1'});
      cy.checkMachRegAnnotation({machRegName: 'edit-yaml-test',
        annotationName: 'myAnnotation1',
        annotationValue: 'myAnnotationValue1'});
    });

    it('Clone a machine registration', () => {
      cy.createMachReg({machRegName: 'clone-test',
        checkLabels: true,
        checkAnnotations: true});
      cy.contains('clone-test')
        .click();
      cy.get('div.actions > .role-multi-action')
        .click()
      cy.contains('li', 'Clone')
        .click();
      cy.getBySel('name-ns-description-name')
        .type('cloned-machine-reg')
      cy.getBySel('form-save')
        .contains('Create')
        .click();
      cy.contains('.masthead', 'Registration Endpoint: cloned-machine-regActive')
        .should('exist');
      
      // Check that we got the same label and annotation in both machine registration
      cy.checkMachRegLabel({machRegName: 'cloned-machine-reg',
        labelName: 'myLabel1',
        labelValue: 'myLabelValue1'});
      cy.contains('cloned-machine-reg')
        .click();
      cy.checkMachRegAnnotation({machRegName: 'cloned-machine-reg',
        annotationName: 'myAnnotation1',
        annotationValue: 'myAnnotationValue1'});
      cy.contains('cloned-machine-reg')
        .click();
    });

    it('Download Machine registration YAML', () => {
      cy.createMachReg({machRegName: 'download-yaml-test'});
      cy.contains('download-yaml-test')
        .click();
      cy.get('div.actions > .role-multi-action')
        .click()
      cy.contains('li', 'Download YAML')
        .click();
      cy.verifyDownload('download-yaml-test.yaml');
    });

    it('Check machine registration label name size', () => {
      cy.checkLabelSize({sizeToCheck: 'name'});
    });

    it('Check machine registration label value size', () => {
      cy.checkLabelSize({sizeToCheck: 'value'});
    });

  // This test must stay the last one because we use this machine registration when we test adding a node.
  // It also tests using a custom cloud config by using read from file button.
    it('Create Machine registration we will use to test adding a node', () => {
      cy.createMachReg({machRegName: 'machine-registration',
        checkInventoryLabels: true,
        checkInventoryAnnotations: true,
        checkIsoBuilding: true,
        customCloudConfig: 'custom_cloud-config.yaml',
        checkDefaultCloudConfig: false});
      cy.checkMachInvLabel({machRegName: 'machine-registration',
        labelName: 'myInvLabel1',
        labelValue: 'myInvLabelValue1'});
    });
  });

  // In the upgrade test we test the ISO building feature 
  // and boot from the stable ISO, mainly because we can
  // only select stable ISO for now.
  // We will move the test to the standard scenario later
  filterTests(['upgrade'], () => {
    it('Create Machine registration we will use to test adding a node', () => {
      cy.createMachReg({machRegName: 'machine-registration',
        checkInventoryLabels: true,
        checkInventoryAnnotations: true,
        checkIsoBuilding: true,
        customCloudConfig: 'custom_cloud-config.yaml',
        checkDefaultCloudConfig: false});
      cy.checkMachInvLabel({machRegName: 'machine-registration',
        labelName: 'myInvLabel1',
        labelValue: 'myInvLabelValue1'});
    });
  });
});
