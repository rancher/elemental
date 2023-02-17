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

import { TopLevelMenu } from '~/cypress/support/toplevelmenu';
import '~/cypress/support/functions';
import { Elemental } from '~/cypress/support/elemental';
import filterTests from '~/cypress/support/filterTests.js';

filterTests(['main'], () => {
  Cypress.config();
  describe('Menu testing', () => {
    const topLevelMenu   = new TopLevelMenu();
    const elemental      = new Elemental();
    const ui_account     = Cypress.env('ui_account');
    const elemental_user = "elemental-user"
    const ui_password    = "rancherpassword"
  
    beforeEach(() => {
      (ui_account == "user") ? cy.login(elemental_user, ui_password) : cy.login();
      cy.visit('/');
    });
  
    it('Check Elemental logo', () => {
      topLevelMenu.openIfClosed();
  
      // Elemental's icon should appear in the side menu
        elemental.elementalIcon().should('exist');
    });
    
    it('Check Elemental menu', () => {
      topLevelMenu.openIfClosed();
  
      // Elemental's icon should appear in the side menu
        elemental.elementalIcon().should('exist');
  
        // Click on the Elemental's icon
        elemental.accessElementalMenu(); 
  
      // Check Elemental's side menu
      elemental.checkElementalNav();
    });
  });
}); 
