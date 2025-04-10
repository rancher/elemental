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
import filterTests from '~/support/filterTests.js';
import * as cypressLib from '@rancher-ecp-qa/cypress-library';
import { qase } from 'cypress-qase-reporter/dist/mocha';
import { isRancherManagerVersion } from '../../support/utils';

filterTests(['main'], () => {
  describe('Deploy application in fresh Elemental Cluster', () => {
    const clusterName = 'mycluster';

    beforeEach(() => {
      cy.login();
      cy.visit('/');
      cypressLib.burgerMenuToggle();
      cy.viewport(1920, 1080);
    });

    qase(31,
      it('Deploy Alerting Drivers application', () => {
        let myAppToInstall;
        if (isRancherManagerVersion('2.11') || isRancherManagerVersion('rancher:head')) {
          myAppToInstall = 'Cerbos'
        } else {
          myAppToInstall = 'Alerting Drivers';
        }
        cypressLib.checkClusterStatus(clusterName, 'Active', 600000);
        cypressLib.burgerMenuToggle();
        // eslint-disable-next-line cypress/no-unnecessary-waiting
        cy.wait(20000);
        isRancherManagerVersion('2.8') && cypressLib.burgerMenuToggle();
        cypressLib.checkClusterStatus(clusterName, 'Active', 600000);
        if (!isRancherManagerVersion('2.8')) {
          cy.get('.main-panel').contains(clusterName).click();
        } else {
          cy.contains(clusterName).click();
        }
        cy.get('.nav').contains('Apps').click();
        cy.contains('Charts').click();
        cy.contains(myAppToInstall, { timeout: 30000 }).click();
        cy.contains('.name-logo-install', myAppToInstall, { timeout: 30000 });
        cy.clickButton('Install');
        if (isRancherManagerVersion('2.11') || isRancherManagerVersion('rancher:head')) {
          cy.contains('.top > .title', myAppToInstall)
        } else {
          cy.contains('.outer-container > .header', myAppToInstall);
        }
        cy.clickButton('Next');
        cy.clickButton('Install');
        cy.contains('SUCCESS: helm install', { timeout: 120000 });
        cy.reload();
        if (isRancherManagerVersion('2.11') || isRancherManagerVersion('rancher:head')) {
          cy.contains(new RegExp('Deployed.*cerbos'));
        } else { 
          cy.contains(new RegExp('Deployed.*rancher-alerting-drivers'))  
        }
      }));

    qase(32,
      it('Remove Alerting Drivers application', () => {
        let myAppToInstall;
        if (isRancherManagerVersion('2.11') || isRancherManagerVersion('rancher:head')) {
          myAppToInstall = 'cerbos'
        } else {
          myAppToInstall = 'rancher-alerting-drivers';
        }
        cypressLib.checkClusterStatus(clusterName, 'Active', 600000);
        if (!isRancherManagerVersion('2.8')) {
          cy.get('.main-panel').contains(clusterName).click();
        } else {
          cy.contains(clusterName).click();
        }
        cy.get('.nav').contains('Apps').click();
        cy.contains('Installed Apps').click();
        cy.contains('.title', 'Installed Apps', { timeout: 20000 });
        cy.contains(myAppToInstall);
        cy.get('[width="30"] > .checkbox-outer-container').click();
        cy.get('.outlet').getBySel('sortable-table-promptRemove').click();
        cy.confirmDelete();
        cy.contains('SUCCESS: helm uninstall', { timeout: 60000 });
        cy.contains('.apps', myAppToInstall).should('not.exist');
      }));
  });
});
