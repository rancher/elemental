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
import '~/support/functions';
import filterTests from '~/support/filterTests.js';

Cypress.config();
describe('SeedImage testing', () => {
  const isoToTest     = Cypress.env('iso_to_test');
  const seedImageFile = "myseedimage.yaml"
  const topLevelMenu  = new TopLevelMenu();

  beforeEach(() => {
    cy.login();
    cy.visit('/');

    // Open the navigation menu
    topLevelMenu.openIfClosed();
  });

  filterTests(['main'], () => {
    it('Create SeedImage with custom base image', () => {
      if (typeof isoToTest !== 'undefined') {
        cy.exec(`sed -i "s|baseImage:.*|baseImage: ${isoToTest}|g" fixtures/${seedImageFile}`);
      }
        cy.contains('local')
        .click();
      cy.get('.header-buttons > :nth-child(1)')
        .click();
      cy.clickButton('Read from File');
      cy.get('input[type="file"]')
        .attachFile({filePath: seedImageFile});
      cy.clickButton('Import');
      cy.get('.badge-state')
        .contains('Active');
      cy.clickButton('Close');
      cy.exec('ls');
    });
  });
});
