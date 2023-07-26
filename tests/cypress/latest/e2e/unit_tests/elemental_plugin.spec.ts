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
import { isUIVersion, isRancherManagerVersion } from '../../support/utils';
import * as cypressLib from '@rancher-ecp-qa/cypress-library';
import { qase } from 'cypress-qase-reporter/dist/mocha';

filterTests(['main', 'upgrade'], () => {
  Cypress.config();
  describe('Install Elemental plugin', () => {
  
    beforeEach(() => {
      cy.login();
      cy.visit('/');
    });
  
    qase(11,
      it('Add elemental-ui repo', () => {
        !isUIVersion('stable') ? cypressLib.addRepository('elemental-ui', 'https://github.com/rancher/elemental-ui.git', 'git') : null;
      })
    );
    
    qase(12,
      it('Enable extension support', () => {
        cypressLib.burgerMenuOpenIfClosed();
        isUIVersion('stable') ? cypressLib.enableExtensionSupport(true) : cypressLib.enableExtensionSupport(false, isRancherManagerVersion("head"));
      })
    );
  
    qase(13,
      it('Install Elemental plugin', () => {
        cypressLib.burgerMenuOpenIfClosed();
        // TODO: create a function to install any plugin and not elemental only
        cy.contains('Extensions')
          .click();
        cy.contains('elemental');
        cy.get('.plugin')
          .contains('Install')
          .click();
        cy.clickButton('Install');
        cy.contains('Installing');
        cy.contains('Extensions changed - reload required', {timeout: 40000});
        cy.clickButton('Reload');
        cy.get('.plugins')
          .children()
          .should('contain', 'elemental')
          .and('contain', 'Uninstall');
      })
    );
  });
});
