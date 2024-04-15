/*
Copyright © 2022 - 2024 SUSE LLC

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
import { slowCypressDown } from 'cypress-slow-down'

// slow down each command by 500ms
slowCypressDown(500)

Cypress.config();
describe('Machine inventory testing', () => {
  const clusterName          = "mycluster"
  const elementalUser        = "elemental-user"
  const hwLabels             = ["TotalCPUThread", "TotalMemory", "CPUModel",
                               "CPUVendor", "NumberBlockDevices", "NumberNetInterface",
                               "CPUVendorTotalCPUCores"]
  const k8sDownstreamVersion = Cypress.env('k8s_downstream_version');
  const proxy                = "http://172.17.0.1:3128"
  const uiAccount            = Cypress.env('ui_account');
  const uiPassword           = "rancherpassword"
  const vmNumber             = 3;
  
  beforeEach(() => {
    (uiAccount == "user") ? cy.login(elementalUser, uiPassword) : cy.login();
    cy.visit('/');

    // Open the navigation menu
    cypressLib.burgerMenuToggle();

    // Click on the Elemental's icon
    cypressLib.accesMenu('OS Management');
  });

  filterTests(['main'], () => {
    qase(28,
      it('Check that machine inventory has been created', () => {
        cy.clickNavMenu(["Inventory of Machines"]);
        cy.contains('Namespace: fleet-default')
        for (let i = 0; i < vmNumber; i++) {
          cy.getBySel(`sortable-cell-${i}-0`)
            .contains('Active')
            .should('exist');
          cy.getBySel(`sortable-cell-${i}-1`)
            .contains(`node-00${i + 1}`)
            .should('exist');
        }
      })
    );

    qase(29,
      it('Check we can see our embedded hardware labels', () => {
        cy.clickNavMenu(["Inventory of Machines"]);
        cy.contains('node-001')
          .click()
        cy.checkMachInvLabel('machine-registration', 'myInvLabel1', 'myInvLabelValue1', true);
        for (const key in hwLabels) {
          cy.clickNavMenu(["Inventory of Machines"]);
          // Because of the double .parent() call we have to keep this ugly chain
          // TODO: try to use something more elegant later!
          // eslint-disable-next-line cypress/unsafe-to-chain-command
          cy.get('.table-options-group > .btn > .icon')
            .click()
            .parent()
            .parent()
            .contains(hwLabels[key])
            .click({force: true})
          cy.contains(hwLabels[key]);
        }
      })
    );
  });
  
  filterTests(['main', 'upgrade'], () => {
    qase(30,
      it('Create Elemental cluster', () => {
        utils.createCluster(clusterName, k8sDownstreamVersion, proxy);
      })
    );
  });
  
  filterTests(['main', 'upgrade'], () => {
    it('Check Elemental cluster status', () => {
      cypressLib.burgerMenuToggle();
      cypressLib.checkClusterStatus(clusterName, 'Active', 600000);
      cypressLib.burgerMenuToggle();
      cy.contains(clusterName)
        .click();
    })
  });
});
