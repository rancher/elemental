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
import { slowCypressDown } from 'cypress-slow-down'
import { isRancherManagerVersion } from '~/support/utils';

// slow down each command by 500ms
slowCypressDown(500)

filterTests(['main'], () => {
  Cypress.config();
  describe('OS versions testing', () => {
    const elementalUser        = "elemental-user"
    const uiAccount            = Cypress.env('ui_account');
    const uiPassword           = "rancherpassword"
    let htmlSelector;
  
    beforeEach(() => {
      (uiAccount == "user") ? cy.login(elementalUser, uiPassword) : cy.login();
      cy.visit('/');
      cypressLib.burgerMenuToggle();
      cypressLib.accesMenu('OS Management');
    });

    it('Check In Sync column status', () => {
      cy.clickNavMenu(["Advanced", "OS Versions"]);
      (isRancherManagerVersion('2.9')) ? htmlSelector = 'sortable-table-list-container' : htmlSelector = 'cluster-list-container';
      cy.getBySel(htmlSelector)
        .should('not.contain', 'Unavailable');
      cy.getBySel(htmlSelector)
        .contains('Type')
        .click();
      cy.getBySel('sortable-table-0-action-button')
        .click();
      cy.contains('Edit YAML')
        .click();
      cy.contains('annotations').as('anno')
      cy.get('@anno').click(0,0)
      cy.get('@anno').type('{end}{enter}  elemental.cattle.io/channel-no-longer-in-sync: \'true\'');
      cy.getBySel('action-button-async-button')
        .contains('Save')
        .click();
      cy.getBySel(htmlSelector)
        .contains('Type')
        .click();
      cy.getBySel('sortable-cell-0-4')
        .should('contain', 'Out of sync');
    });

    it('Out of sync OS version should appear deprecated', () => {
      cy.createMachReg('sample-machine-reg');
      cy.contains('sample-machine-reg')
        .click();
      cy.getBySel('select-media-type-build-media')
        .click();
      cy.contains('Raw')
        .click();
      cy.getBySel('select-os-version-build-media')
        .click();
      cy.contains(new RegExp('OS.*deprecated'))
    });

  });
});
