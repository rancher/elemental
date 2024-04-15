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

import { Elemental } from '~/support/elemental';
import '~/support/commands';
import 'cypress-file-upload';
import filterTests from '~/support/filterTests.js';
import * as cypressLib from '@rancher-ecp-qa/cypress-library';
import { qase } from 'cypress-qase-reporter/dist/mocha';
import * as utils from "~/support/utils";
import { slowCypressDown } from 'cypress-slow-down'

// slow down each command by 500ms
slowCypressDown(500)


Cypress.config();
describe('UI extension upgrade tests', () => {
  const elemental     = new Elemental();

  beforeEach(() => {
    // Elemental-user can not be used here because it does not have access to the local cluster
    cy.login();
    cy.visit('/');

    // Open the navigation menu 
    cypressLib.burgerMenuToggle();
  });

  filterTests(['upgrade'], () => {
    // Enable only with K3S because still too much flaky with RKE2
    if (utils.isK8sVersion('k3s')) {
      it('Add elemental-ui dev repo', () => {
        cypressLib.addRepository('elemental-ui', 'https://github.com/rancher/elemental-ui.git', 'git', 'gh-pages');
      });

      qase(56,
        it('Upgrade Elemental UI extension', () => {
          cy.contains('Extensions')
            .click();
          cy.getBySel('btn-available')
            .click();
          cy.getBySel('extension-card-elemental')
            .contains('elemental')
          cy.getBySel('extension-card-install-btn-elemental')
            .click();
          cy.getBySel('install-ext-modal-install-btn')
            .click();
          // Sometimes the reload button is not displayed and it breaks the test...
          // Adding a sleep command waiting a more elegant solution
          //cy.contains('Extensions changed - reload required', {timeout: 100000});
          //cy.clickButton('Reload');
          // eslint-disable-next-line cypress/no-unnecessary-waiting
          cy.wait(120000);
          cy.reload();
          cy.getBySel('extension-card-uninstall-btn-elemental')
        })
      );
      qase(58,
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
        })
      );
    };
  });
});
