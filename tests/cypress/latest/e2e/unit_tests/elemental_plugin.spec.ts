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

import { Rancher } from '~/support/rancher';
import '~/support/commands';
import filterTests from '~/support/filterTests.js';

filterTests(['main', 'upgrade'], () => {
  Cypress.config();
  describe('Install Elemental plugin', () => {
    const rancher = new Rancher();
  
    beforeEach(() => {
      cy.login();
      cy.visit('/');
    });
  
    it('Add elemental-ui repo', () => {
      if ( Cypress.env('elemental_ui_version') != "stable") {
        rancher.addRepository('elemental-ui', 'https://github.com/rancher/elemental-ui.git', 'git');
      };
    });
    
    it('Enable extension support', () => {
      rancher.burgerMenuOpenIfClosed();
      cy.contains('Extensions')
        .click();
      // Make sure we are on the Extensions page
      cy.contains('.message-icon', 'Extension support is not enabled');
      cy.clickButton('Enable');
      cy.contains('Enable Extension Support?')
      if ( Cypress.env('elemental_ui_version') != "stable") {
        cy.contains('Add the Rancher Extension Repository')
          .click();
      }
      cy.clickButton('OK');
      cy.get('.tabs', {timeout: 40000})
        .contains('Installed Available Updates All');
    });
  
    it('Install Elemental plugin', () => {
      rancher.burgerMenuOpenIfClosed();
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
    });
  });
}); 
