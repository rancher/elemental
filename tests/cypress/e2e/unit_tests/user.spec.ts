import { TopLevelMenu } from '~/cypress/support/toplevelmenu';
import '~/cypress/support/functions';
import { Elemental } from '~/cypress/support/elemental';

Cypress.config();
describe('User role testing', () => {
  const topLevelMenu   = new TopLevelMenu();
  const elemental      = new Elemental();
  const std_user       = "std-user"
  const elemental_user = "elemental-user"
  const ui_password    = "rancherpassword"

  beforeEach(() => {
    cy.visit('/');
  });

  it('Create elemental user', () => {
    // User with the elemental-administrator role
    cy.login();
    topLevelMenu.openIfClosed();
    cy.contains('Users & Authentication').click();
    cy.contains('.title', 'Users').should('exist');
    cy.clickButton('Create');
    cy.typeValue({label: 'Username', value: std_user});
    cy.typeValue({label: 'New Password', value: ui_password});
    cy.typeValue({label: 'Confirm Password', value: ui_password});
    cy.clickButton('Create');
  });

  it('Create standard user', () => {
    // User without the elemental-administrator role
    cy.login();
    topLevelMenu.openIfClosed();
    cy.contains('Users & Authentication').click();
    cy.contains('.title', 'Users').should('exist');
    cy.clickButton('Create');
    cy.typeValue({label: 'Username', value: elemental_user});
    cy.typeValue({label: 'New Password', value: ui_password});
    cy.typeValue({label: 'Confirm Password', value: ui_password});
    cy.contains('Elemental Administrator').click();
    cy.clickButton('Create');
  });

  it('Elemental user should access the OS management menu', () => {
    cy.login(elemental_user, ui_password);
    cy.get('[data-testid="banner-title"]').contains('Welcome to Rancher');
    topLevelMenu.openIfClosed();
    elemental.elementalIcon().should('exist');
    elemental.accessElementalMenu();
    elemental.checkElementalNav();
  });

  it('Standard user should not access the OS management menu', () => {
    cy.login(std_user, ui_password);
    cy.get('[data-testid="banner-title"]').contains('Welcome to Rancher');
    topLevelMenu.openIfClosed();
    elemental.elementalIcon().should('exist');
    elemental.accessElementalMenu();
    // User without appropriate role will get a specific page
    cy.contains('Elemental is a software stack');
  });
});
