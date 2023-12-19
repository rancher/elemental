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

import { Elemental } from '~/support/elemental';
import '~/support/commands';
import 'cypress-file-upload';
import filterTests from '~/support/filterTests.js';
import * as cypressLib from '@rancher-ecp-qa/cypress-library';
import { qase } from 'cypress-qase-reporter/dist/mocha';
import * as utils from "~/support/utils";


Cypress.config();
describe('Elemental operator upgrade tests', () => {
  const elemental = new Elemental();

  beforeEach(() => {
    // Elemental-user can not be used here because it does not have access to the local cluster
    cy.login();
    cy.visit('/');

    // Open the navigation menu 
    cypressLib.burgerMenuToggle();
  });

  filterTests(['upgrade'], () => {
    // Enable only with K3S because still too much flaky with RKE2
    if (utils.isK8sVersion('k3s') && !utils.isRancherManagerVersion('2.7')) {
      it('Add elemental-operator dev repo', () => {
        cypressLib.addRepository('elemental-operator', Cypress.env('chartmuseum_repo')+':8080', 'helm', 'none');
      });

      qase(888,
        it('Upgrade Elemental operator', () => {
          cy.contains('local')
            .click();
          cy.get('.nav').contains('Apps')
            .click();
          cy.contains('.item.has-description.color1', 'Elemental', {timeout:30000})
            .click();
          cy.contains('Charts: Elemental', {timeout:30000});
          cy.clickButton('Upgrade');
          cy.contains('.header > .title', 'elemental-operator');
          cy.clickButton('Next');
          cy.get('[data-testid="string-input-channel.repository"]')
            .clear()
          cy.get('[data-testid="string-input-channel.repository"]')
            .type('rancher/elemental-channel')
          cy.get('[data-testid="string-input-channel.tag"]')
            .clear()
          cy.get('[data-testid="string-input-channel.tag"]')
            .type('1.5.0')
          cy.clickButton('Upgrade');
          cy.contains('SUCCESS: helm', {timeout:120000});
          cy.contains('Installed App: elemental-operator Pending-Upgrade', {timeout:120000});
          cy.contains('Installed App: elemental-operator Deployed', {timeout:120000});
        })
      );

      qase(889,
        it('Check Elemental UI after upgrade', () => {
          cy.viewport(1920, 1080);
          // Elemental's icon should appear in the side menu
          cypressLib.checkNavIcon('elemental')
            .should('exist');

          // Click on the Elemental's icon
          cypressLib.accesMenu('OS Management');

          // Check Elemental's side menu
          elemental.checkElementalNav();

          // Check Elemental's main page
          // TODO: Could be improve to check everything
          cy.get('[data-testid="card-registration-endpoints"]')
            .contains('1');
          cy.get('[data-testid="card-inventory-of-machines"]')
            .contains('1');
          cy.get('[data-testid="card-clusters"]')
            .contains('1');
          cy.get('[data-testid="machine-reg-block"]')
            .contains('machine-registration');
          // Check OS Versions Channel
          cy.clickNavMenu(["Advanced", "OS Version Channels"]);
          cy.get('.main-row')
            .contains('Active elemental-channel', {timeout: 60000});
        })
      );
    };
  });
});
