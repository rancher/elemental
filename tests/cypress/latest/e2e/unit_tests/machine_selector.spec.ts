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

import { TopLevelMenu } from '~/support/toplevelmenu';
import { Elemental } from '~/support/elemental';
import '~/support/commands';
import filterTests from '~/support/filterTests.js';

filterTests(['main'], () => {
  Cypress.config();
  describe('Machine selector testing', () => {
    const elemental     = new Elemental();
    const elementalUser = "elemental-user"
    const k8sVersion    = Cypress.env('k8s_version');
    const topLevelMenu  = new TopLevelMenu();
    const uiAccount     = Cypress.env('ui_account');
    const uiPassword    = "rancherpassword"
  
    beforeEach(() => {
      (uiAccount == "user") ? cy.login(elementalUser, uiPassword) : cy.login();
      cy.visit('/');
  
      // Open the navigation menu
      topLevelMenu.openIfClosed();
  
      // Click on the Elemental's icon
      elemental.accessElementalMenu(); 
  
      // Go to the cluster creation page
      elemental.accessClusterMenu(); 
    });
  
    it('Testing selector without any rule', () => {
      cy.contains('.banner', 'Matches all 1 existing Inventory of Machines')
        .should('exist');
    });
  
    it('Testing selector with unmatching rule', () => {
      //cy.clickButton('Add Rule');
      // TODO: Cannot use the clickButton here, I do not know why yet 
      cy.get('[cluster="[provisioning.cattle.io.cluster: undefined]"]')
        .contains('Add Rule')
        .click();
      cy.get('[data-testid="input-match-expression-values-0"] > input')
        .click()
        .type('wrong');
      cy.contains('.banner', 'Matches no existing Inventory of Machines')
        .should('exist');
    });
  
    it('Testing selector with matching rule', () => {
      //cy.clickButton('Add Rule');
      // TODO: Cannot use the clickButton here, I do not know why yet 
      cy.get('[cluster="[provisioning.cattle.io.cluster: undefined]"]')
        .contains('Add Rule')
        .click();
      cy.get('[data-testid="input-match-expression-key-0"]')
        .click()
      cy.contains('myInvLabel1')
        .click();
      cy.get('[data-testid="input-match-expression-values-0"] > input')
        .click()
        .type('myInvLabelValue1');
      cy.contains('.banner', 'Matches all 1 existing Inventory of Machines')
        .should('exist');
    });
  });
}); 
