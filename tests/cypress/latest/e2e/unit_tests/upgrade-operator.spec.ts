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

import { Elemental } from '~/support/elemental';
import '~/support/commands';
import 'cypress-file-upload';
import filterTests from '~/support/filterTests.js';
import * as cypressLib from '@rancher-ecp-qa/cypress-library';
import { qase } from 'cypress-qase-reporter/dist/mocha';
import * as utils from '~/support/utils';

Cypress.config();
describe('Elemental operator upgrade tests', () => {
  const elemental = new Elemental();

  beforeEach(() => {
    cy.login();
    cy.visit('/');
    cypressLib.burgerMenuToggle();
  });

  filterTests(['upgrade'], () => {
    if (utils.isK8sVersion('k3s') && !utils.isRancherManagerVersion('2.7')) {
      if (!utils.isOperatorVersion('marketplace')) {
        it('Add elemental-operator dev repo', () => {
          cypressLib.addRepository('elemental-operator', `${Cypress.env('chartmuseum_repo')}:8080`, 'helm', 'none');
        });
      } else {
        qase(55,
          it('Upgrade Elemental operator', () => {
            cy.contains('local').click();
            cy.get('.nav').contains('Apps').click();
            cy.contains('Elemental', { timeout: 30000 }).click();
            cy.contains('Charts: Elemental', { timeout: 30000 });
            cy.clickButton('Upgrade');
            cy.contains('.header > .title', 'elemental-operator');
            cy.clickButton('Next');
            cy.clickButton('Upgrade');
            cy.contains('SUCCESS: helm', { timeout: 120000 });
            cy.contains('Installed App: elemental-operator Pending-Upgrade', { timeout: 120000 });
            cy.contains('Installed App: elemental-operator Deployed', { timeout: 120000 });
        }));

        qase(58,
          it('Check Elemental UI after upgrade', () => {
            cy.viewport(1920, 1080);
            cypressLib.checkNavIcon('elemental').should('exist');
            cypressLib.accesMenu('OS Management');
            elemental.checkElementalNav();
            cy.get('[data-testid="card-registration-endpoints"]').contains('1');
            cy.get('[data-testid="card-inventory-of-machines"]').contains('1');
            cy.get('[data-testid="card-clusters"]').contains('1');
            cy.get('[data-testid="machine-reg-block"]').contains('machine-registration');
            cy.clickNavMenu(['Advanced', 'OS Version Channels']);
            cy.get('.main-row').contains('Active elemental-channel', { timeout: 60000 });
        }));
      }
    }
  });
});
