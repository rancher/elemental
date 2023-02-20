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

Cypress.config();
describe('Deploy application in fresh Elemental Cluster', () => {
  const topLevelMenu = new TopLevelMenu();
  beforeEach(() => {
    cy.login();
    cy.visit('/');
  });
  
  it('Deploy CIS Benchmark application', () => {
    topLevelMenu.openIfClosed();
    cy.contains('myelementalcluster').click();
    cy.contains('Apps').click();
    cy.contains('Charts').click();
    cy.contains('CIS Benchmark').click();
    cy.contains('.name-logo-install', 'CIS Benchmark', {timeout:20000});
    cy.clickButton('Install');
    cy.clickButton('Next');
    cy.clickButton('Install');
    cy.contains('SUCCESS: helm upgrade', {timeout:30000});
    cy.reload;
    cy.contains('CIS Benchmark');
  });

  it('Remove CIS Benchmark application', () => {
    topLevelMenu.openIfClosed();
    cy.contains('myelementalcluster').click();
    cy.contains('Apps').click();
    cy.contains('Installed Apps').click();
    cy.contains('.title', 'Installed Apps', {timeout:20000});
    cy.get('.ns-dropdown > .icon').click().type('cis-operator');
    cy.contains('cis-operator').click();
    cy.get('.ns-dropdown > .icon-chevron-up').click();
    cy.get('[width="30"] > .checkbox-outer-container').click();
    cy.clickButton('Delete');
    cy.confirmDelete();
    cy.contains('SUCCESS: helm uninstall', {timeout:30000});
    cy.contains('.apps', 'CIS Benchmark').should('not.exist');
  });
});
