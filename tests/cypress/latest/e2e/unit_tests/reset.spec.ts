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

filterTests(['main'], () => {
  Cypress.config();
  describe('Reset testing', () => {
    const clusterName   = "mycluster"
    const elementalUser = "elemental-user"
    const k8sVersion    = Cypress.env('k8s_version');
    const proxy         = "http://172.17.0.1:3128" 
    const uiAccount     = Cypress.env('ui_account');
    const uiPassword    = "rancherpassword"
  
    beforeEach(() => {
      (uiAccount == "user") ? cy.login(elementalUser, uiPassword) : cy.login();
      cy.visit('/');
      cypressLib.burgerMenuToggle();
      cypressLib.accesMenu('OS Management');
    });
    
    if (utils.isK8sVersion('rke')) {
      qase(54,
        it('Enable reset in machine inventory', () => {
          cy.clickNavMenu(["Inventory of Machines"]);
          cy.getBySel('sortable-table-0-action-button').click();
          cy.contains('Edit YAML').click();
          cy.contains('annotations').as('anno')
          cy.get('@anno').click(0,0)
          cy.get('@anno').type('{end}{enter}  elemental.cattle.io/resettable: \'true\'');
          cy.getBySel('action-button-async-button')
          .contains('Save')
          .click();
      }));
    };

    qase(['54', '55'],
      it('Reset node by deleting the cluster', () => {
        cy.viewport(1920, 1080);
        cy.getBySel('button-manage-elemental-clusters').click();
        cy.getBySel('sortable-cell-0-0').click();
        cy.clickButton('Delete');
        cy.getBySel('prompt-remove-input')
          .type('mycluster');
        cy.getBySel('prompt-remove-confirm-button').click();
        cypressLib.burgerMenuToggle();
        cypressLib.accesMenu('OS Management');
        cy.clickNavMenu(["Inventory of Machines"]);
        cy.contains('There are no rows to show.');
        cy.getBySel('sortable-table-0-row', { timeout: 180000 })
          .contains('Active', { timeout: 180000 });
    }));

    qase(30,
      it('Create Elemental cluster', () => {
        utils.createCluster(clusterName, k8sVersion, proxy);
      }));
  });
});
