export class Elemental {
  firstLogin() {
    cy.get('input').type(Cypress.env('password'), {log: false});
    cy.clickButton('Log in with Local User');
    cy.contains('By checking').click('left');
    cy.clickButton('Continue');
    cy.get('[data-testid="banner-title"]').contains('Welcome to Rancher');
  }

  elementalIcon() {
    return cy.get('.option .icon.group-icon.icon-os-management');
  } 

  // Go into the Elemental Menu
  accessElementalMenu() {
    cy.contains('OS Management').click();
  }

  checkElementalNav() {
    // Open Advanced accordion
    cy.get('div.header > i').eq(0).click()
    cy.get('div.header').contains('Advanced').should('be.visible')

    // Check all listed options once accordion is opened
    cy.get('li.child.nav-type').should(($lis) => {
    expect($lis).to.have.length(6);
    expect($lis.eq(0)).to.contain('Dashboard');
    expect($lis.eq(1)).to.contain('Registration Endpoints');
    expect($lis.eq(2)).to.contain('Inventory of Machines');
    expect($lis.eq(3)).to.contain('Update Groups');
    expect($lis.eq(4)).to.contain('OS Versions');
    expect($lis.eq(5)).to.contain('OS Version Channels');
    })      
  }

  // Go into the cluster creation menu
  accessClusterMenu() {
    cy.contains('.title', 'Dashboard').should('exist');
    cy.contains('Dashboard').click();
    cy.contains('Create Elemental Cluster').should('exist');
    cy.contains('Create Elemental Cluster').click();
  }
}
