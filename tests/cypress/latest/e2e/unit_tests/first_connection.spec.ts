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

import filterTests from '~/support/filterTests.js';
import * as cypressLib from '@rancher-ecp-qa/cypress-library';
import { qase } from 'cypress-qase-reporter/dist/mocha';
import { isRancherManagerVersion } from '~/support/utils';

filterTests(['main', 'upgrade'], () => {
  Cypress.config();
  describe('First login on Rancher', () => {
    qase(46,
      it('Log in and accept terms and conditions', () => {
        cypressLib.firstLogin();
      })
    );
    // We need to enable prerelease versions to install the elemental dev operator
    it('Enable Helm Chart Prerelease versions', () => {
      cy.login();
      cy.visit('/');
      cy.getBySel('nav_header_showUserMenu').click();
      if (isRancherManagerVersion('2.10')) {
        cy.getBySel('user-menu-dropdown').contains('Preferences').click();
      } else {
        cy.contains('Preferences').click();
      }
      cy.clickButton('Include Prerelease Versions');
      cypressLib.burgerMenuToggle();
      cy.getBySel('side-menu').contains('Home').click();
    });
  })
});
