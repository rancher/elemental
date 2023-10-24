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
import filterTests from '~/support/filterTests.js';
import * as cypressLib from '@rancher-ecp-qa/cypress-library';
import { qase } from 'cypress-qase-reporter/dist/mocha';

filterTests(['main'], () => {
  Cypress.config();
  describe('Menu testing', () => {
    const elemental     = new Elemental();
    const elementalUser = "elemental-user"
    const uiAccount     = Cypress.env('ui_account');
    const uiPassword    = "rancherpassword"
  
    beforeEach(() => {
      (uiAccount == "user") ? cy.login(elementalUser, uiPassword) : cy.login();
      cy.visit('/');
      cypressLib.burgerMenuToggle();
    });
  
    qase(2,
      it('Check Elemental logo', () => {
        // Elemental's icon should appear in the side menu
        cypressLib.checkNavIcon('elemental')
          .should('exist');
      })
    );
    
    qase(3,
      it('Check Elemental menu', () => {
        // Elemental's icon should appear in the side menu
        cypressLib.checkNavIcon('elemental')
          .should('exist');
  
        // Click on the Elemental's icon
        cypressLib.accesMenu('OS Management');
  
        // Check Elemental's side menu
        elemental.checkElementalNav();
      })
    );
  });
}); 
