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

import '~/support/commands';
import filterTests from '~/support/filterTests.js';
import * as cypressLib from '@rancher-ecp-qa/cypress-library';
import { qase } from 'cypress-qase-reporter/dist/mocha';
import { isCypressTag, isRancherManagerVersion } from '~/support/utils';

filterTests(['main', 'upgrade'], () => {
  Cypress.config();
  describe('Install Elemental Operator', () => {
  
    beforeEach(() => {
      cy.login();
      cy.visit('/');
      cypressLib.burgerMenuToggle();
    });

    // Install the dev operator in the main scenario for Rancher 2.8.x
    if (isCypressTag('main') && isRancherManagerVersion('2.8')){
      qase(11,
        it('Add local chartmuseum repo', () => {
          cypressLib.addRepository('elemental-operator', Cypress.env('chartmuseum_repo')+':8080', 'helm', 'none');
        })
      );
    };
  
    if (isRancherManagerVersion('2.8')) {
      qase(13,
        it('Install Elemental operator', () => {
          cy.contains('local')
            .click();
          cy.get('.nav').contains('Apps')
            .click();
          if (isCypressTag('main')) {
            cy.contains('.item.has-description.color1', 'Elemental', {timeout:30000})
              .click();
          } else {
            cy.contains('Elemental', {timeout:30000})
              .click();
          }
          cy.contains('Charts: Elemental', {timeout:30000});
          cy.clickButton('Install');
          cy.contains('.outer-container > .header', 'Elemental');
          cy.clickButton('Next');
          // Workaround for https://github.com/rancher/rancher/issues/43379
          if (isCypressTag('upgrade')) {
            cy.get('[data-testid="string-input-channel.repository"]')
              .type('registry.suse.com/rancher/elemental-teal-channel')
            cy.get('[data-testid="string-input-channel.tag"]')
              .type('1.3.5')
          }
          cy.clickButton('Install');
          cy.contains('SUCCESS: helm', {timeout:120000});
          cy.reload;
          cy.contains('Only User Namespaces') // eslint-disable-line cypress/unsafe-to-chain-command
            .click()
            .type('cattle-elemental-system{enter}{esc}');
          cy.get('.outlet').contains('Deployed elemental-operator cattle-elemental-system', {timeout: 120000});
        })
      );
    };
  });
});
