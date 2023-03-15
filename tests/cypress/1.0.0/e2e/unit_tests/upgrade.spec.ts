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
      // TODO: REMOVE 'IF' BLOCK AFTER NEXT STABLE VERSION (> 1.0.0)
      // OSversion channel was not integrated in the UI before 1.1.0
      if ( elementalUIVersion == '1.0.0' ) {
        cy.exec(`sed -i 's/# namespace: fleet-default/namespace: fleet-default/g' assets/managedOSVersionChannel.yaml`);
        cy.exec(`sed -i 's@image: %UPGRADE_CHANNEL_LIST%@image: ${upgradeChannelList}@g' assets/managedOSVersionChannel.yaml`);
        cy.get('.nav')
          .contains('Advanced')
          .click();
        cy.get('.nav')
          .contains('OS Version Channels')
          .click();
        cy.clickButton('Create from YAML');
        // Wait needed to avoid crash with the upload
        cy.wait(2000);
        cy.get('input[type="file"]')
          .attachFile({filePath: '../../assets/managedOSVersionChannel.yaml'});
        // Wait needed to avoid crash with the upload
        cy.wait(2000);
        // Wait for os-versions to be printed, that means the upload is done
        cy.clickButton('Create');
        // The new resource must be active
        cy.contains('Active');
      } else {
        cy.get('.nav')
          .contains('Advanced')
          .click();
        cy.get('.nav')
          .contains('OS Version Channels')
          .click();
        cy.clickButton('Create');
        cy.typeValue({label: 'Name', value: channelName});
        cy.typeValue({label: 'Image registry path', value: upgradeChannelList});
        cy.clickButton('Create');
      }
    });

    it('Check OS Versions', () => {
      cy.get('.nav')
        .contains('Advanced')
        .click();
      cy.get('.nav')
        .contains('OS Versions')
        .click();
      cy.contains('Active dev', {timeout: 120000});
      cy.contains('Active stable');
      cy.contains('Active staging');
    });

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
      // TODO: REMOVE FIRST 'IF' BLOCK AFTER NEXT STABLE VERSION (> 1.0.0)
      // Following 'if' code will be removed once we get a new stable version
      if (elementalUIVersion == '1.0.0') {
        cy.typeValue({label: 'OS Image', value: upgradeImage});
      } else {
          if (checkK3s.test(k8sVersion)) {
            cy.contains('Use image from registry')
              .click();
            cy.typeValue({label: 'Image path', value: upgradeImage});
          } else {
            cy.contains('Use Managed OS version')
              .click();
            cy.get(':nth-child(4) > .labeled-select')
              .contains('Managed OS version')
              .click();
            cy.contains('staging')
              .click();
          };
      }; 
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

    it('Delete OS Versions', () => {
      cy.get('.nav').contains('Advanced')
        .click();
      cy.get('.nav')
        .contains('OS Versions')
        .click();
      cy.contains('dev')
        .parent()
        .parent()
        .click();
      cy.clickButton('Delete');
      cy.confirmDelete();
      cy.contains('dev')
        .should('not.exist');
    });

    it('Delete OS Versions Channels', () => {
      cy.get('.nav')
        .contains('Advanced')
        .click();
      cy.get('.nav')
        .contains('OS Version Channels')
        .click();
      cy.deleteAllResources();
      cy.get('.nav')
        .contains('Advanced')
        .click();
      cy.get('.nav')
        .contains('OS Versions')
        .click();
      cy.contains('There are no rows to show');
    });
  });
});
