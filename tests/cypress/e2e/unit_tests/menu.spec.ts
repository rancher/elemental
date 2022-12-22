import { TopLevelMenu } from '~/cypress/support/toplevelmenu';
import '~/cypress/support/functions';
import { Elemental } from '~/cypress/support/elemental';

Cypress.config();
describe('Menu testing', () => {
  const topLevelMenu   = new TopLevelMenu();
  const elemental      = new Elemental();
  const ui_account     = Cypress.env('ui_account');
  const elemental_user = "elemental-user"
  const ui_password    = "rancherpassword"

  beforeEach(() => {
    (ui_account == "user") ? cy.login(elemental_user, ui_password) : cy.login();
    cy.visit('/');
  });

  it('Check Elemental logo', () => {
    topLevelMenu.openIfClosed();

    // Elemental's icon should appear in the side menu
      elemental.elementalIcon().should('exist');
  });
  
  it('Check Elemental menu', () => {
    topLevelMenu.openIfClosed();

    // Elemental's icon should appear in the side menu
      elemental.elementalIcon().should('exist');

      // Click on the Elemental's icon
      elemental.accessElementalMenu(); 

    // Check Elemental's side menu
    elemental.checkElementalNav();
  });
});
