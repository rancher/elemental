import { TopLevelMenu } from '~/cypress/support/toplevelmenu';
import '~/cypress/support/functions';
import { Elemental } from '~/cypress/support/elemental';

Cypress.config();
describe('Install Elemental plugin', () => {
  const topLevelMenu = new TopLevelMenu();
  const elemental = new Elemental();
  const elemental_ui_version = Cypress.env('elemental_ui_version');

  beforeEach(() => {
    cy.login();
    cy.visit('/');
  });

  it('Add elemental-ui repo', () => {
    topLevelMenu.openIfClosed();
    cy.contains('local').click();
    cy.addHelmRepo({repoName: 'elemental-ui', repoUrl: 'https://github.com/rancher/elemental-ui.git', repoType: 'git'});
  });
  
  it('Enable extension support', () => {
    topLevelMenu.openIfClosed();
    cy.contains('Extensions').click();
    cy.clickButton('Enable');
    cy.contains('Enable Extension Support?')
    cy.contains('Add the Rancher Extension Repository').click();
    cy.clickButton('OK');
    cy.contains('No Extensions installed', {timeout: 40000});
  });

  it('Install Elemental plugin', () => {
    topLevelMenu.openIfClosed();
    cy.contains('Extensions').click();
    cy.contains('elemental');
    cy.get('.plugin').contains('Install').click();
    cy.contains('Install Extension elemental');
    if (elemental_ui_version != 'latest') {
      cy.get('.labeled-select').click();
      cy.contains(elemental_ui_version).click();
    }
    cy.clickButton('Install');
    cy.contains('Installing');
    cy.contains('Extensions changed - reload required', {timeout: 40000});
    cy.clickButton('Reload');
    cy.get('.plugins')
      .children()
      .should('contain', 'elemental')
      .and('contain', 'Uninstall')
  });
});
