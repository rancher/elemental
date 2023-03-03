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
import { Elemental } from '~/cypress/support/elemental';
import '~/cypress/support/functions';
import filterTests from '~/cypress/support/filterTests.js';

Cypress.config();
describe('User role testing', () => {
  const elemental     = new Elemental();
  const elementalUser = "elemental-user"
  const stdUser       = "std-user"
  const topLevelMenu  = new TopLevelMenu();
  const uiPassword    = "rancherpassword"

  beforeEach(() => {
    cy.visit('/');
  });

  filterTests(['main', 'upgrade'], () => {
    it('Create elemental user', () => {
      // User with the elemental-administrator role
      cy.login();
      topLevelMenu.openIfClosed();
      cy.contains('Users & Authentication')
        .click();
      cy.contains('.title', 'Users')
        .should('exist');
      cy.clickButton('Create');
      cy.typeValue({label: 'Username', value: stdUser});
      cy.typeValue({label: 'New Password', value: uiPassword});
      cy.typeValue({label: 'Confirm Password', value: uiPassword});
      cy.clickButton('Create');
    });

    it('Create standard user', () => {
      // User without the elemental-administrator role
      cy.login();
      topLevelMenu.openIfClosed();
      cy.contains('Users & Authentication')
        .click();
      cy.contains('.title', 'Users')
        .should('exist');
      cy.clickButton('Create');
      cy.typeValue({label: 'Username', value: elementalUser});
      cy.typeValue({label: 'New Password', value: uiPassword});
      cy.typeValue({label: 'Confirm Password', value: uiPassword});
      cy.contains('Elemental Administrator')
        .click();
      cy.clickButton('Create');
    });
  });

  filterTests(['main'], () => {
    it('Elemental user should access the OS management menu', () => {
      cy.login(elementalUser, uiPassword);
      cy.get('[data-testid="banner-title"]')
        .contains('Welcome to Rancher');
      topLevelMenu.openIfClosed();
      elemental.elementalIcon().should('exist');
      elemental.accessElementalMenu();
      elemental.checkElementalNav();
    });

    it('Standard user should not access the OS management menu', () => {
      cy.login(stdUser, uiPassword);
      cy.get('[data-testid="banner-title"]')
        .contains('Welcome to Rancher');
      topLevelMenu.openIfClosed();
      elemental.elementalIcon().should('exist');
      elemental.accessElementalMenu();
      // User without appropriate role will get a specific page
      cy.contains('Elemental is a software stack');
    });
  });
});
