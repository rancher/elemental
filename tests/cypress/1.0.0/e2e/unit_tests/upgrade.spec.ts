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
import '~/support/functions';
import { Elemental } from '~/support/elemental';
import 'cypress-file-upload';
import filterTests from '~/support/filterTests.js';


Cypress.config();
describe('Upgrade tests', () => {
  const channelName        = "mychannel"
  const clusterName        = "mycluster"
  const checkK3s: RegExp   = /k3s/
  const elemental          = new Elemental();
  const elementalUIVersion = Cypress.env('elemental_ui_version')
  const elementalUser      = "elemental-user"
  const k8sVersion         = Cypress.env('k8s_version')
  const topLevelMenu       = new TopLevelMenu();
  const uiAccount          = Cypress.env('ui_account');
  const uiPassword         = "rancherpassword"
  const upgradeChannelList = Cypress.env('upgrade_channel_list')
  const upgradeImage       = Cypress.env('upgrade_image')

  beforeEach(() => {
    (uiAccount == "user") ? cy.login(elementalUser, uiPassword) : cy.login();
    cy.visit('/');

    // Open the navigation menu
    topLevelMenu.openIfClosed();

    // Click on the Elemental's icon
    elemental.accessElementalMenu(); 
  });
  filterTests(['upgrade'], () => {
    it('Upgrade one node (different methods if rke2 or k3s)', () => {
      // K3s cluster upgraded with OS Image
      // RKE2 cluster upgraded with OS version channel
      cy.get('.nav')
        .contains('Advanced')
        .click();
      cy.get('.nav')
        .contains('Update Groups')
        .click();
      cy.clickButton('Create');
      cy.get('.primaryheader')
        .contains('Update Group: Create');
      cy.typeValue({label: 'Name', value: 'upgrade'});
      cy.contains('Target Cluster')
        .click();
      cy.contains(clusterName)
        .click();
      cy.typeValue({label: 'OS Image', value: upgradeImage});
      cy.clickButton('Create');
      // Status changes a lot right after the creation so let's wait 10 secondes
      // before checking
      cy.wait(10000);
      cy.get('[data-testid="sortable-cell-0-0"]')
        .contains('Active');

      // Workaround to avoid sporadic issue with Upgrade
      // https://github.com/rancher/elemental/issues/410
      // Restart fleet agent inside downstream cluster
      topLevelMenu.openIfClosed();
      cy.contains(clusterName)
        .click();
      cy.contains('Workload')
        .click();
      cy.contains('Pods')
        .click();
      cy.get('.header-buttons > :nth-child(2)')
        .click();
      cy.wait(20000);
      cy.get('.shell-body')
        .type('kubectl scale deployment/fleet-agent -n cattle-fleet-system --replicas=0{enter}');
      cy.get('.shell-body')
        .type('kubectl scale deployment/fleet-agent -n cattle-fleet-system --replicas=1{enter}');

      // Check if the node reboots to apply the upgrade
      topLevelMenu.openIfClosed();
      elemental.accessElementalMenu();
      cy.clickNavMenu(["Dashboard"]);
      cy.clickButton('Manage Elemental Clusters');
      cy.get('.title')
        .contains('Clusters');
      cy.contains(clusterName)
        .click();
      cy.get('.primaryheader')
        .contains('Active');
      cy.get('.primaryheader')
        .contains('Updating', {timeout: 240000});
      cy.get('.primaryheader')
        .contains('Active', {timeout: 240000});
    });
  });
});
