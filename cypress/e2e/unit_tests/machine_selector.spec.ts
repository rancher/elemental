import { TopLevelMenu } from '~/cypress/support/toplevelmenu';
import '~/cypress/support/functions';
import { Elemental } from '../../support/elemental';

Cypress.config();
describe('Machine selector testing', () => {
  const topLevelMenu = new TopLevelMenu();
  const elemental    = new Elemental();
  const k8s_version  = Cypress.env('k8s_version');

  beforeEach(() => {
    cy.login();
    cy.visit('/');

    // Open the navigation menu
    topLevelMenu.openIfClosed();

    // Click on the Elemental's icon
    elemental.accessElementalMenu(); 

    // Go to the cluster creation page
    elemental.accessClusterMenu(); 
  });

  it('Testing selector without any rule', () => {
    cy.contains('.banner', 'Matches all 1 existing Machine Inventories').should('exist');
  });

  it('Testing selector with unmatching rule', () => {
    //cy.clickButton('Add Rule');
    // TODO: Cannot use the clickButton here, I do not know why yet 
    cy.get('.mt-20 > .btn').contains('Add Rule').click();
    cy.get('[data-testid="input-match-expression-values-0"] > input').click().type('wrong');
    cy.contains('.banner', 'Matches no existing Machine Inventories').should('exist');
  });

  it('Testing selector with matching rule', () => {
    //cy.clickButton('Add Rule');
    // TODO: Cannot use the clickButton here, I do not know why yet 
    cy.get('.mt-20 > .btn').contains('Add Rule').click();
    cy.get('[data-testid="input-match-expression-key-0"] > input').type('cypress');
    cy.get('[data-testid="input-match-expression-values-0"] > input').click().type('uitesting');
    cy.contains('.banner', 'Matches all 1 existing Machine Inventories').should('exist');
  });
});
