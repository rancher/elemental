/*
Copyright © 2022 - 2026 SUSE LLC

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

Cypress.config();
describe('User role testing', () => {
  const elemental = new Elemental();
  const elementalUser = "elemental-user"
  const stdUser = "std-user"
  const uiPassword = "rancherpassword"

  beforeEach(() => {
    cy.visit('/');
  });

  filterTests(['main', 'upgrade'], () => {
    qase(15,
      it('Create standard user', () => {
        // User without the elemental-administrator role
        cy.login();
        cypressLib.burgerMenuToggle();
        cypressLib.createUser(stdUser, uiPassword);
      })
    );

    qase(14,
      it('Create elemental user', () => {
        // User with the elemental-administrator role
        cy.login();
        cypressLib.burgerMenuToggle();
        cypressLib.createUser(elementalUser, uiPassword, 'Elemental Administrator');
      })
    );
  });

  filterTests(['main'], () => {
    qase(47,
      it('Elemental user should access the OS management menu', () => {
        cy.login(elementalUser, uiPassword);
        cy.getBySel('banner-title').contains('Welcome to Rancher');
        cypressLib.burgerMenuToggle();
        cypressLib.checkNavIcon('elemental').should('exist');
        cypressLib.accesMenu('OS Management');
        elemental.checkElementalNav();
    }));

    qase(48,
      it('Standard user should not access the OS management menu', () => {
        cy.login(stdUser, uiPassword);
        cy.getBySel('banner-title').contains('Welcome to Rancher');
        cypressLib.burgerMenuToggle();
        cypressLib.checkNavIcon('elemental').should('exist');
        cypressLib.accesMenu('OS Management');
        cy.getBySel('elemental-icon').should('exist');
        cy.getBySel('elemental-description-text').contains('Elemental is a software stack').should('exist');
        cy.getBySel('warning-not-install-or-no-schema').should('exist');
    }));
  });
});
