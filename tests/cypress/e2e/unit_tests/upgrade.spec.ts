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

import { TopLevelMenu } from '~/cypress/support/toplevelmenu';
import '~/cypress/support/functions';
import { Elemental } from '~/cypress/support/elemental';
import 'cypress-file-upload';
import filterTests from '~/cypress/support/filterTests.js';


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
    it('Create an OS Version Channels', () => {
      cy.clickNavMenu(["Advanced", "OS Version Channels"]);
      cy.contains('[data-testid="masthead-create"]', 'Create')
        .click();
      cy.get('[data-testid="name-ns-description-name"]')
        .type(channelName);
      cy.get('[data-testid="os-version-channel-path"]')
        .type(upgradeChannelList);
      cy.contains('[data-testid="form-save"]', 'Create')
        .click();
    });

    it('Check OS Versions', () => {
      cy.clickNavMenu(["Advanced", "OS Versions"]);
      cy.contains('[data-testid="sortable-table-0-row"]', 'Active dev', {timeout: 120000});
      cy.contains('[data-testid="sortable-table-1-row"]', 'Active stable');
      cy.contains('[data-testid="sortable-table-2-row"]', 'Active staging');
    });

    it('Upgrade one node (different methods if rke2 or k3s)', () => {
      // K3s cluster upgraded with OS Image
      // RKE2 cluster upgraded with OS version channel
      cy.clickNavMenu(["Advanced", "Update Groups"]);
      cy.contains('[data-testid="masthead-create"]', 'Create')
        .click();
      cy.get('.primaryheader')
        .contains('Update Group: Create');
      cy.get('[data-testid="name-ns-description-name"]')
        .type(channelName);
      cy.contains('Target Cluster')
      cy.get('[data-testid="cluster-target"]')
        .click();
      cy.contains(clusterName)
        .click();
      // TODO: REMOVE FIRST 'IF' BLOCK AFTER NEXT STABLE VERSION (> 1.0.0)
      // Following 'if' code will be removed once we get a new stable version
      if (elementalUIVersion == '1.0.0') {
        cy.typeValue({label: 'OS Image', value: upgradeImage});
      } else {
        if (!checkK3s.test(k8sVersion)) {
          cy.get('[data-testid="upgrade-choice-selector"]')
            .parent()
            .contains('Use image from registry')
            .click();
          cy.get('[data-testid="os-image-box"]')
            .type(upgradeImage)
        } else {
          cy.get('[data-testid="upgrade-choice-selector"]')
            .parent()
            .contains('Use Managed OS version')
            .click();
          cy.get('[data-testid="os-version-box"]')
            .click()
          cy.get('[data-testid="os-version-box"]')
            .parents()
            .contains('staging')
            .click();
        };
      };
      cy.contains('[data-testid="form-save"]', 'Create')
        .click();
      // Status changes a lot right after the creation so let's wait 10 secondes
      // before checking
      cy.wait(10000);
      cy.get('[data-testid="sortable-cell-0-0"]')
        .contains('Active');

      // Workaround to avoid sporadic issue with Upgrade
      // https://github.com/rancher/elemental/issues/410
      // Restart fleet agent inside downstream cluster
      topLevelMenu.openIfClosed();
      cy.get('[data-testid="side-menu"]')
        .contains(clusterName)
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
      cy.get('[data-testid="card-clusters"]')
        .contains('Manage Elemental Clusters')
        .click()
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

    it('Delete OS Versions', () => {
      cy.clickNavMenu(["Advanced", "OS Versions"]);
      cy.contains('dev')
        .parent()
        .parent()
        .click();
      cy.contains('[data-testid="sortable-table-promptRemove"]', 'Delete')
        .click()
      cy.confirmDelete();
      cy.contains('dev')
        .should('not.exist');
    });

    it('Delete OS Versions Channels', () => {
      cy.clickNavMenu(["Advanced", "OS Version Channels"]);
      cy.deleteAllResources();
      cy.clickNavMenu(["Advanced", "OS Versions"]);
      cy.contains('There are no rows to show');
    });
  });
});
