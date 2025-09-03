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

import '~/support/commands';
import 'cypress-file-upload';
import filterTests from '~/support/filterTests.js';
import * as utils from '~/support/utils';
import * as cypressLib from '@rancher-ecp-qa/cypress-library';
import { qase } from 'cypress-qase-reporter/dist/mocha';

describe('Upgrade tests', () => {
  const channelName = 'mychannel';
  const clusterName = 'mycluster';
  const elementalUser = 'elemental-user';
  const uiAccount = Cypress.env('ui_account');
  const uiPassword = 'rancherpassword';
  const upgradeImage = Cypress.env('upgrade_image');
  const login = () => (uiAccount === 'user' ? cy.login(elementalUser, uiPassword) : cy.login());

  beforeEach(() => {
    login();
    cy.visit('/');
    cy.viewport(1920, 1080);
    cypressLib.burgerMenuToggle();
    cypressLib.accesMenu('OS Management');
  });

  filterTests(['upgrade'], () => {
    // Add dev OS Version Channel if stable operator is installed
    // because we do not update the operator in RKE2 UI test so far
    // Only RKE2 tests use os version channel
    if (utils.isK8sVersion('rke2')) {
      it('Delete stable channel for RKE2 upgrade', () => {
        cy.clickNavMenu(['Advanced', 'OS Version Channels']);
        cy.deleteAllResources();
      });

      it('Add staging channel for RKE2 upgrade', () => {
        cy.addOsVersionChannel('staging');
      });
    }

    if (utils.isK8sVersion('rke2')) {
      qase(33,
        it('Check OS Versions', () => {
          cy.clickNavMenu(['Advanced', 'OS Versions']);
          cy.contains(new RegExp('Active.*unstable-iso'), { timeout: 120000 });
      }));
    }

    qase(34,
      it('Upgrade one node (different methods if rke2 or k3s)', () => {
        cypressLib.burgerMenuToggle();
        cypressLib.checkClusterStatus(clusterName, 'Active', 600000);
        cypressLib.burgerMenuToggle();
        cypressLib.accesMenu('OS Management');
        // K3s cluster upgraded with OS Image
        // RKE2 cluster upgraded with OS version channel
        // Marketplace test uses OS version channel
        cy.clickNavMenu(['Advanced', 'Update Groups']);
        cy.getBySel('masthead-create').contains('Create').click();
        cy.get('.masthead').contains('Update Group: Create');
        cy.getBySel('name-ns-description-name').type(channelName);
        cy.contains('Target Cluster');
        cy.getBySel('cluster-target').click();
        cy.get('.vs__dropdown-menu').contains(clusterName).click();

        if (utils.isK8sVersion('k3s') && !utils.isOperatorVersion('marketplace')) {
          cy.getBySel('upgrade-choice-selector').parent().contains('Use image from registry').click();
          cy.getBySel('os-image-box').type(upgradeImage);
        } else {
          cy.getBySel('upgrade-choice-selector').parent().contains('Use Managed OS Version').click();
          cy.getBySel('os-version-box').click();
          if (utils.isOperatorVersion('marketplace')) {
            cy.getBySel('os-version-box').parents().contains(Cypress.env('os_version_target')).click();
          } else {
            cy.getBySel('os-version-box').parents().contains('unstable').click();
          }
        }

        cy.getBySel('form-save').contains('Create').click();
        // Status changes a lot right after the creation so let's wait 10 secondes
        // before checking
        cy.wait(10000); // eslint-disable-line cypress/no-unnecessary-waiting
        cy.getBySel('sortable-cell-0-0').contains('Active');
        // Check if the node reboots to apply the upgrade
        cypressLib.burgerMenuToggle();
        cypressLib.accesMenu('OS Management');
        cy.clickNavMenu(['Dashboard']);
        cy.getBySel('card-clusters').contains('Manage Elemental Clusters').click();
        cy.get('.title').contains('Clusters');
        cy.get('.outlet').contains(clusterName).click();
        cy.get('.top').contains('Updating', { timeout: 420000 });
        cy.get('.top').contains('Active', { timeout: 720000 });
    }));

    qase(35,
      it('Cannot create two upgrade groups targeting the same cluster', () => {
        cy.clickNavMenu(['Advanced', 'Update Groups']);
        cy.getBySel('masthead-create').contains('Create').click();
        cy.get('.masthead').contains('Update Group: Create');
        cy.getBySel('cluster-target').click();
        // As there is already an upgrade group targeting the cluster,
        // the cluster should not be available in the dropdown
        cy.get('#vs4__listbox').should('not.contain', clusterName);
    }));

    qase(76,
      it('Delete Upgrade Group', () => {
        cy.clickNavMenu(['Advanced', 'Update Groups']);
        cy.deleteAllResources();
        cy.contains('There are no rows to show');
    }));

    qase(37,
      it('Delete OS Versions Channels', () => {
        cy.clickNavMenu(['Advanced', 'OS Version Channels']);
        cy.deleteAllResources();
        cy.clickNavMenu(['Advanced', 'OS Versions']);
        cy.contains('There are no rows to show');
    }));
  });
});
