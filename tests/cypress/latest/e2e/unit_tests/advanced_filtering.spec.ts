/*
Copyright © 2022 - 2024 SUSE LLC

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

filterTests(['main'], () => {
  describe('Advanced filtering testing', () => {
    const elementalUser = 'elemental-user';
    const uiAccount = Cypress.env('ui_account');
    const uiPassword = 'rancherpassword';
    const login = () => (uiAccount === 'user' ? cy.login(elementalUser, uiPassword) : cy.login());

    beforeEach(() => {
      login();
      cy.visit('/');
      cypressLib.burgerMenuToggle();
      cypressLib.accesMenu('OS Management');
    });

    qase(21,
      it('Create fake machine inventories', () => {
        const machineInventoryMap = new Map([
          ['machine_inventory_1', 'test-filter-one'],
          ['machine_inventory_2', 'test-filter-two'],
          ['machine_inventory_3', 'shouldnotmatch']
        ]);

        machineInventoryMap.forEach((value, key) => {
          cy.importMachineInventory(`${key}.yaml`, value);
        });
    }));

    qase(22,
      it('Two machine inventories should appear by filtering on test-filter', () => {
        cy.checkFilter('test-filter', true, true, false);
    }));

    qase(22,
      it('One machine inventory should appear by filtering on test-filter-one', () => {
        ['test-filter-one', 'Test-Filter-One'].forEach(filter => {
          cy.checkFilter(filter, true, false, false);
        });
    }));

    qase(23,
      it('No machine inventory should appear by filtering on test-bad-filter', () => {
        cy.checkFilter('test-bad-filter', false, false, false);
        cy.contains('There are no rows which match your search query.');
    }));

    qase(24,
      it('Delete all fake machine inventories', () => {
        cy.clickNavMenu(['Inventory of Machines']);
        cy.get('[width="30"] > .checkbox-outer-container > .checkbox-container > .checkbox-custom').click();
        cy.clickButton('Actions');
        cy.get('.tooltip-inner > :nth-child(1) > .list-unstyled > :nth-child(3)').click();
        cy.confirmDelete();
    }));
  });
});
