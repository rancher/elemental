import { Elemental } from '~/cypress/support/elemental';

Cypress.config();
describe('First login on Rancher', () => {
  const elemental = new Elemental();

  it('Log in and accept terms and conditions', () => {
    cy.visit('/auth/login');
    cy.get("span").then($text => {
      if ($text.text().includes('your first time visiting Rancher')) {
        elemental.firstLogin();
      }
      else {
        cy.log('Rancher already initialized, no need to handle first login.')
      }
    })
  });
});
