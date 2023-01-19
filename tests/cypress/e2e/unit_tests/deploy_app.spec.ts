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
