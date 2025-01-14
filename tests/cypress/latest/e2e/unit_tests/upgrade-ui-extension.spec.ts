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

describe('UI extension upgrade tests', () => {
  const elemental = new Elemental();

  beforeEach(() => {
    cy.login();
    cy.visit('/');
    cypressLib.burgerMenuToggle();
  });

  filterTests(['upgrade'], () => {
    // 1 - Enable only with K3S because still too much flaky with RKE2
    // 2 - TODO: Remove rancher 2-10 condition later
    // UI extension upgrade cannot be upgraded with rancher manager 2.10 yet
    // because we have only one version so far
    if (utils.isK8sVersion('k3s') && !utils.isRancherManagerVersion('2.10')) {
      it('Add elemental-ui dev repo', () => {
        cypressLib.addRepository('elemental-ui', 'https://github.com/rancher/elemental-ui.git', 'git', 'gh-pages');
      });

      qase(56,
        it('Upgrade Elemental UI extension', () => {
          cy.contains('Extensions').click();
          cy.getBySel('btn-available').click();
          cy.getBySel('extension-card-elemental').contains('elemental');
          cy.getBySel('extension-card-install-btn-elemental').click();
          cy.getBySel('install-ext-modal-install-btn').click();
          cy.wait(120000); // eslint-disable-line cypress/no-unnecessary-waiting
          cy.reload();
          cy.getBySel('extension-card-uninstall-btn-elemental');
      }));

      qase(58,
        it('Check Elemental UI after upgrade', () => {
          cy.viewport(1920, 1080);
          cypressLib.checkNavIcon('elemental').should('exist');
          cypressLib.accesMenu('OS Management');
          elemental.checkElementalNav();
          // Check Elemental's main page
          // TODO: Could be improve to check everything
          cy.get('[data-testid="card-registration-endpoints"]').contains('1');
          cy.get('[data-testid="card-inventory-of-machines"]').contains('1');
          cy.get('[data-testid="card-clusters"]').contains('1');
          cy.get('[data-testid="machine-reg-block"]').contains('machine-registration');
      }));
    }
  });
});
