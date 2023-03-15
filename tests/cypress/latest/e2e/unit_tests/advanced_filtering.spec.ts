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

filterTests(['main'], () => {
  Cypress.config();
  describe('Advanced filtering testing', () => {
    const elemental     = new Elemental();
    const elementalUser = "elemental-user"
    const uiAccount     = Cypress.env('ui_account');
    const uiPassword    = "rancherpassword"
    const topLevelMenu  = new TopLevelMenu();
  
    beforeEach(() => {
      (uiAccount == "user") ? cy.login(elementalUser, uiPassword) : cy.login();
      cy.visit('/');
  
      // Open the navigation menu
      topLevelMenu.openIfClosed();
  
      // Click on the Elemental's icon
      elemental.accessElementalMenu(); 
    });
  
    it('Create fake machine inventories', () => {
      // TODO: Use loop?
      cy.importMachineInventory({machineInventoryFile: 'machine_inventory_1.yaml',
        machineInventoryName: 'test-filter-one'});
      cy.importMachineInventory({machineInventoryFile: 'machine_inventory_2.yaml',
        machineInventoryName: 'test-filter-two'});
      cy.importMachineInventory({machineInventoryFile: 'machine_inventory_3.yaml',
        machineInventoryName: 'shouldnotmatch'});
    });
  
    it('Two machine inventories should appear by filtering on test-filter', () => {
      // Only test-filter-one and test-filter-two should appear with test-filter as filter
      cy.checkFilter({filterName: 'test-filter',
        testFilterOne: true,
        testFilterTwo: true,
        shouldNotMatch: false});
    });
  
    it('One machine inventory should appear by filtering on test-filter-one', () => {
      // Only test-filter-one should appear with test-filter-one as filter
      cy.checkFilter({filterName: 'test-filter-one',
        testFilterOne: true,
        testFilterTwo: false,
        shouldNotMatch: false});
    });
  
    it('No machine inventory should appear by filtering on test-bad-filter', () => {
      // This test will also serve as no regression test for https://github.com/rancher/elemental-ui/issues/41
      cy.checkFilter({filterName: 'test-bad-filter',
        testFilterOne: false,
        testFilterTwo: false,
        shouldNotMatch: false});
      cy.contains('There are no rows which match your search query.')
    });
  
    it('Delete all fake machine inventories', () => {
      cy.clickNavMenu(["Inventory of Machines"]);
      cy.get('[width="30"] > .checkbox-outer-container > .checkbox-container > .checkbox-custom')
        .click();
      cy.clickButton('Actions');
      cy.get('.tooltip-inner > :nth-child(1) > .list-unstyled > :nth-child(3)')
        .click();
      cy.confirmDelete();
    });
  });
});
