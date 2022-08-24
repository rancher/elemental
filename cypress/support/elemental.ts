export class Elemental {

  firstLogin() {
    cy.get('input').type(Cypress.env('password'), {log: false});
    cy.clickButton('Log in with Local User');
    cy.contains('I agree').click('left');
    cy.clickButton('Continue');
    cy.contains("Getting Started", {timeout: 10000});
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
    expect($lis).to.have.length(8);
    expect($lis.eq(0)).to.contain('Dashboard');
    expect($lis.eq(1)).to.contain('Machine Registrations');
    expect($lis.eq(2)).to.contain('Machine Inventories');
    expect($lis.eq(3)).to.contain('Mach. Inv. Selectors');
    expect($lis.eq(4)).to.contain('Mach. Inv. Selec. Templates');
    expect($lis.eq(5)).to.contain('Managed OS Versions');
    expect($lis.eq(6)).to.contain('Managed OS Version Channels');
    expect($lis.eq(7)).to.contain('OS Image Upgrades');
    })      
  }
}
