/*
Copyright © 2022 - 2025 SUSE LLC

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
import * as utils from "~/support/utils";

filterTests(['main'], () => {
  Cypress.config();
  
  describe('Seed images menu testing', () => {
    const elementalUser = "elemental-user";
    const uiAccount = Cypress.env('ui_account');
    const uiPassword = "rancherpassword";
    const selectors = {
      sortableTableRow: 'sortable-table-0-row',
    };
    const login = uiAccount === "user" ? () => cy.login(elementalUser, uiPassword) : () => cy.login();

    beforeEach(() => {
      login();
      cy.visit('/');
      cypressLib.burgerMenuToggle();
      cypressLib.accesMenu('OS Management');
    });

    it('Download from seed images menu', () => {
      // Delete all files previously downloaded
      cy.exec('rm -f cypress/downloads/*', { failOnNonZeroExit: false });
      cy.clickNavMenu(["Advanced", "Seed Images"]);
      cy.getBySel(selectors.sortableTableRow).contains('Download').click();
      if (utils.isBootType('iso')) {
        cy.verifyDownload('.iso', { contains: true, timeout: 300000, interval: 5000 });
      } else {
        // .img still needed for rancher 2.8
        let extension = 'img';
        !utils.isRancherManagerVersion('2.8') ? extension = 'raw' : null;
        cy.verifyDownload('.'+extension, { contains: true, timeout: 300000, interval: 5000 });
      }
      // The following line will replace the confition just above when we will remove 2.8 tests
      //cy.verifyDownload(isBootType('iso') ? '.iso' : '.raw', { contains: true, timeout: 300000, interval: 5000 });
      if (!utils.isRancherManagerVersion('2.8') && utils.isUIVersion('dev')) {
        cy.getBySel('download-checksum-btn-list').click();
        cy.verifyDownload('.sh256', { contains: true, timeout: 60000, interval: 5000 });
      }
    });
  });
});
