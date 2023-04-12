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

export class Elemental {
  // Go into the cluster creation menu
  accessClusterMenu() {
    cy.contains('Dashboard')
      .click();
    cy.getBySel('elemental-main-title')
      .should('exist');
    cy.getBySel('card-clusters')
      .contains('Create Elemental Cluster')
      .should('exist');
    cy.getBySel('button-create-elemental-cluster')
      .click();
  }

  // Go into the Elemental menu
  accessElementalMenu() {
    cy.contains('OS Management')
      .click();
  }

  // Make sure we get all menus
  checkElementalNav() {
    // Open advanced accordion
    cy.get('div.header > i')
      .eq(0)
      .click()
    cy.get('div.header')
      .contains('Advanced')
      .should('be.visible')
    // Check all listed options once accordion is opened
    cy.get('li.child.nav-type')
      .should(($lis) => {
    expect($lis).to.have.length(6);
    expect($lis.eq(0)).to.contain('Dashboard');
    expect($lis.eq(1)).to.contain('Registration Endpoints');
    expect($lis.eq(2)).to.contain('Inventory of Machines');
    expect($lis.eq(3)).to.contain('Update Groups');
    expect($lis.eq(4)).to.contain('OS Versions');
    expect($lis.eq(5)).to.contain('OS Version Channels');
    })      
  }

  // Make sure Elemental logo appears
  elementalIcon() {
    return cy.get('.option .icon.group-icon.icon-elemental');
  } 
  
  // Handle first login in Rancher
  firstLogin() {
    cy.get('input')
      .type(Cypress.env('password'), {log: false});
    cy.clickButton('Log in with Local User');
    cy.contains('By checking')
      .click('left');
    cy.clickButton('Continue');
    cy.getBySel('banner-title')
      .contains('Welcome to Rancher');
  }
}
