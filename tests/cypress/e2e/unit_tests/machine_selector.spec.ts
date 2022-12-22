import { TopLevelMenu } from '~/cypress/support/toplevelmenu';
import '~/cypress/support/functions';
import { Elemental } from '~/cypress/support/elemental';

Cypress.config();
describe('Machine selector testing', () => {
  const topLevelMenu   = new TopLevelMenu();
  const elemental      = new Elemental();
  const k8s_version    = Cypress.env('k8s_version');
  const ui_account     = Cypress.env('ui_account');
  const elemental_user = "elemental-user"
  const ui_password    = "rancherpassword"

  beforeEach(() => {
    (ui_account == "user") ? cy.login(elemental_user, ui_password) : cy.login();
    cy.visit('/');

    // Open the navigation menu
    topLevelMenu.openIfClosed();

    // Click on the Elemental's icon
    elemental.accessElementalMenu(); 

    // Go to the cluster creation page
    elemental.accessClusterMenu(); 
  });

  it('Testing selector without any rule', () => {
    cy.contains('.banner', 'Matches all 1 existing Inventory of Machines').should('exist');
  });

  it('Testing selector with unmatching rule', () => {
    //cy.clickButton('Add Rule');
    // TODO: Cannot use the clickButton here, I do not know why yet 
    cy.get('.mt-20 > .btn').contains('Add Rule').click();
    cy.get('[data-testid="input-match-expression-values-0"] > input').click().type('wrong');
    cy.contains('.banner', 'Matches no existing Inventory of Machines').should('exist');
  });

  it('Testing selector with matching rule', () => {
    //cy.clickButton('Add Rule');
    // TODO: Cannot use the clickButton here, I do not know why yet 
    cy.get('.mt-20 > .btn').contains('Add Rule').click();
    cy.get('#vs6__combobox').click()
    cy.contains('myInvLabel1').click();
    cy.get('[data-testid="input-match-expression-values-0"] > input').click().type('myInvLabelValue1');
    cy.contains('.banner', 'Matches all 1 existing Inventory of Machines').should('exist');
  });
});
