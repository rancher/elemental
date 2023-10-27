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

import '~/support/commands';
import filterTests from '~/support/filterTests.js';
import * as utils from "~/support/utils";
import * as cypressLib from '@rancher-ecp-qa/cypress-library';
import { qase } from 'cypress-qase-reporter/dist/mocha';

Cypress.config();
describe('Machine registration testing', () => {
  const elementalUser = "elemental-user"
  const uiAccount     = Cypress.env('ui_account');
  const uiPassword    = "rancherpassword"

  before(() => {
    (uiAccount == "user") ? cy.login(elementalUser, uiPassword) : cy.login();
    cy.visit('/');

    // Open the navigation menu
    cypressLib.burgerMenuToggle();

    // Click on the Elemental's icon
    cypressLib.accesMenu('OS Management');

    // In upgrade scenario, we need to add extra channels
    // KEEP THIS COMMENTED FOR NOW
    //if (utils.isCypressTag('upgrade')) {
    //  if (utils.isOperatorVersion('registry.suse.com')) {
    //    utils.isUpgradeOsChannel('dev') ? cy.addOsVersionChannel('dev') : cy.addOsVersionChannel('staging');
    //  } else if (utils.isOperatorVersion('staging')) {
    //    cy.addOsVersionChannel('dev');
    //  } else {
    //    cy.addOsVersionChannel('stable');
    //  }
    //}
  });

  beforeEach(() => {
    cy.viewport(1920, 1080);
    (uiAccount == "user") ? cy.login(elementalUser, uiPassword) : cy.login();
    cy.visit('/');

    // Open the navigation menu
    cypressLib.burgerMenuToggle();

    // Click on the Elemental's icon
    cypressLib.accesMenu('OS Management');
    
    // Delete all files previously downloaded
    cy.exec('rm cypress/downloads/*', {failOnNonZeroExit: false});

    // Delete all existing machine registrations
    cy.getBySel('manage-reg-btn')
      .click();
    cy.get('.outlet > header').contains('Registration Endpoints');
    cy.get('body').then(($body) => {
      if (!$body.text().includes('There are no rows to show.')) {
        cy.deleteAllResources();
      }
    });
  });

  filterTests(['main'], () => {
    qase(4,
      it('Create machine registration with default options', () => {
        cy.createMachReg('default-options-test');
      })
    );

    qase(6,
      it('Create machine registration with labels and annotations', () => {
        cy.createMachReg('labels-annotations-test', 'fleet-default', true, true);
      })
    );

    qase(17,
      it('Delete machine registration', () => {
        cy.createMachReg('delete-test');
        cy.deleteMachReg('delete-test');
      })
    );

    qase(49,
      it('Edit a machine registration with edit config button', () => {
        cy.createMachReg('edit-config-test');
        cy.editMachReg('edit-config-test', true, true);
        cy.getBySel('form-save')
          .contains('Save')
          .click();

        // Check that we can see our label and annotation in the YAML
        cy.checkMachRegLabel('edit-config-test', 'myLabel1', 'myLabelValue1');
        cy.checkMachRegAnnotation('edit-config-test', 'myAnnotation1', 'myAnnotationValue1');
      })
    );

    qase(50,
      it('Edit a machine registration with edit YAML button', () => {
        cy.createMachReg('edit-yaml-test');
        cy.editMachReg('edit-yaml-test', true, true, true);
        cy.getBySel('action-button-async-button')
          .contains('Save')
          .click();

        // Check that we can see our label and annotation in the YAML
        cy.checkMachRegLabel('edit-yaml-test', 'myLabel1', 'myLabelValue1');
        cy.checkMachRegAnnotation('edit-yaml-test', 'myAnnotation1', 'myAnnotationValue1');
      })
    );

    qase(18,
      it('Clone a machine registration', () => {
        cy.createMachReg('clone-test', 'fleet-default', true, true);
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
        cy.checkMachRegLabel('cloned-machine-reg','myLabel1', 'myLabelValue1');
        cy.contains('cloned-machine-reg')
          .click();
        cy.checkMachRegAnnotation('cloned-machine-reg', 'myAnnotation1', 'myAnnotationValue1');
        cy.contains('cloned-machine-reg')
          .click();
      })
    );

    qase(19,
      it('Download Machine registration YAML', () => {
        cy.createMachReg('download-yaml-test');
        cy.contains('download-yaml-test')
          .click();
        cy.get('div.actions > .role-multi-action')
          .click()
        cy.contains('li', 'Download YAML')
          .click();
        cy.verifyDownload('download-yaml-test.yaml');
      })
    );

    qase(51,
      it('Check machine registration label name size', () => {
        cy.checkLabelSize('name');
      })
    );

    qase(52,
      it('Check machine registration label value size', () => {
        cy.checkLabelSize('value');
      })
    );

  // This test must stay the last one because we use this machine registration when we test adding a node.
  // It also tests using a custom cloud config by using read from file button.
    qase([8,20,5],
      it('Create Machine registration we will use to test adding a node', () => {
        cy.createMachReg('machine-registration',
          'fleet-default',
          //checkLabels
          false,
          //checkAnnotations
          false,
          //checkInventoryLabels
          true,
          //checkInventoryAnnotations
          true,
          //checkIsoBuilding
          true,
          utils.isK8sVersion('k3s') ? 'custom_cloud-config_with_reset.yaml' : 'custom_cloud-config.yaml',
          //checkDefaultCloudConfig
          false);
        cy.checkMachInvLabel('machine-registration', 'myInvLabel1', 'myInvLabelValue1', false);
      })
    );
  });

  // In the upgrade test we test the ISO building feature 
  // and boot from the stable ISO, mainly because we can
  // only select stable ISO for now.
  // We will move the test to the standard scenario later
  filterTests(['upgrade'], () => {
    qase([8,20,5],
      it('Create Machine registration we will use to test adding a node', () => {
        cy.createMachReg('machine-registration',
        'fleet-default',
        //checkLabels
        false,
        //checkAnnotations
        false,
        //checkInventoryLabels
        true,
        //checkInventoryAnnotations
        true,
        //checkIsoBuilding
        true,
        'custom_cloud-config.yaml',
        //checkDefaultCloudConfig
        false);
        cy.checkMachInvLabel('machine-registration', 'myInvLabel1', 'myInvLabelValue1', false);
      })
    );
  });
});
