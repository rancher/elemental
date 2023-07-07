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

// RANCHER FUNCTIONS
// /////////////////

export class Rancher {
  accesMenu(menu: string) {
    cy.contains(menu)
      .click();
  }

  addRepository(repositoryName: string, repositoryURL: string, repositoryType: string) {
    this.burgerMenuOpenIfClosed();
    cy.contains('local')
      .click();
    cy.addHelmRepo(repositoryName, repositoryURL, repositoryType);
  };

  burgerMenuOpenIfClosed() {
    cy.get('body').then((body) => {
      if (body.find('.menu.raised').length === 0) {
        this.burgerMenuToggle();
      };
    });
  };

  burgerMenuToggle() {
    cy.getBySel('top-level-menu', {timeout: 12000})
      .click();
  };

  checkClusterStatus(clusterName: string, clusterStatus: string, timeout: number) {
    this.burgerMenuOpenIfClosed();
      cy.contains('Home')
        .click();
      // The new cluster must be in active state
      cy.get('[data-node-id="fleet-default/'+clusterName+'"]')
        .contains(clusterStatus,  {timeout: timeout});
  };

  checkNavIcon(iconName: string) {
    return cy.get('.option .icon.group-icon.icon-'+iconName);
  } 
  
  enableExtensionSupport(withRancherRepo: boolean) {
    cy.contains('Extensions')
      .click();
    // Make sure we are on the Extensions page
    cy.contains('.message-icon', 'Extension support is not enabled');
    cy.clickButton('Enable');
    cy.contains('Enable Extension Support?')
    if (!withRancherRepo) {
      cy.contains('Add the Rancher Extension Repository')
        .click();
    }
  cy.clickButton('OK');
  cy.get('.tabs', {timeout: 40000})
    .contains('Installed Available Updates All');
  };
  
  // Handle first login in Rancher
  firstLogin() {
    cy.visit('/auth/login');
    cy.get("span").then($text => {
      if ($text.text().includes('your first time visiting Rancher')) {
        cy.get('input')
          .type(Cypress.env('password'), {log: false});
        cy.clickButton('Log in with Local User');
        cy.contains('By checking')
          .click('left');
        cy.clickButton('Continue');
        cy.getBySel('banner-title')
          .contains('Welcome to Rancher');
      } else {
        cy.log('Rancher already initialized, no need to handle first login.');
      };
    });
  };
};

// RANCHER CUSTOM CYPRESS COMMANDS
// ///////////////////////////////

declare global {
  namespace Cypress {
    interface Chainable {
      addHelmRepo(repoName: string, repoUrl: string, repoType?: string,): Chainable<Element>;
      byLabel(label: string,): Chainable<Element>;
      clickButton(label: string,): Chainable<Element>;
      clickNavMenu(listLabel: string[],): Chainable<Element>;
      confirmDelete(): Chainable<Element>;
      deleteAllResources():Chainable<Element>;
      login(username?: string, password?: string, cacheSession?: boolean,): Chainable<Element>;
      getBySel(dataTestAttribute: string, args?: any): Chainable<JQuery<HTMLElement>>;
      typeValue(label: string, value: string, noLabel?: boolean, log?: boolean): Chainable<Element>;
    }
}}

// Log into Rancher
Cypress.Commands.add('login', (
  username = Cypress.env('username'),
  password = Cypress.env('password'),
  cacheSession = Cypress.env('cache_session')) => {
    const login = () => {
      let loginPath ="/v3-public/localProviders/local*";
      cy.intercept('POST', loginPath).as('loginReq');
      
      cy.visit('/auth/login');

      cy.getBySel('local-login-username')
        .type(username, {log: false});

      cy.getBySel('local-login-password')
        .type(password, {log: false});

      cy.getBySel('login-submit')
        .click();
      cy.wait('@loginReq');
      cy.getBySel('banner-title').contains('Welcome to Rancher');
      } 

    if (cacheSession) {
      cy.session([username, password], login);
    } else {
      login();
    }
});

// Make sure we are in the desired menu inside a cluster (local by default)
// You can access submenu by giving submenu name in the array
// ex:  cy.clickNavMenu(['Menu', 'Submenu'])
Cypress.Commands.add('clickNavMenu', (listLabel: string[]) => {
  listLabel.forEach(label => cy.get('nav').contains(label).click());
});

// Search by data-testid selector
Cypress.Commands.add("getBySel", (selector, ...args) => {
  return cy.get(`[data-testid=${selector}]`, ...args);
});

// Search fields by label
Cypress.Commands.add('byLabel', (label) => {
  cy.get('.labeled-input')
    .contains(label)
    .siblings('input');
});

// Search button by label
Cypress.Commands.add('clickButton', (label) => {
  cy.get('.btn')
    .contains(label)
    .click();
});

// Confirm the delete operation
Cypress.Commands.add('confirmDelete', () => {
  cy.getBySel('prompt-remove-confirm-button')
    .click()
});

// Insert a value in a field *BUT* force a clear before!
Cypress.Commands.add('typeValue', (label, value, noLabel, log=true) => {
  if (noLabel === true) {
    cy.get(label)
      .focus()
      .clear()
      .type(value, {log: log});
  } else {
    cy.byLabel(label)
      .focus()
      .clear()
      .type(value, {log: log});
  }
});

// Delete all resources from a page
Cypress.Commands.add('deleteAllResources', () => {  
  cy.get('[width="30"] > .checkbox-outer-container')
    .click();
  cy.getBySel('sortable-table-promptRemove')
    .contains('Delete')
    .click()
  cy.confirmDelete();
  // Sometimes, UI is crashing when a resource is deleted
  // A reload should workaround the failure
  cy.get('body').then(($body) => {
    if (!$body.text().includes('There are no rows to show.')) {
        cy.reload();
        cy.log('RELOAD TRIGGERED');
        cy.screenshot('reload-triggered');
      };
    });
  cy.contains('There are no rows to show', {timeout: 15000});
});

// Add Helm repo
Cypress.Commands.add('addHelmRepo', (repoName, repoUrl, repoType) => {
  //cy.clickClusterMenu(['Apps', 'Repositories'])
  cy.clickNavMenu(['Apps', 'Repositories'])

  // Make sure we are in the 'Repositories' screen (test failed here before)
  cy.contains('header', 'Repositories')
    .should('be.visible');
  cy.contains('Create')
    .should('be.visible');

  cy.clickButton('Create');
  cy.contains('Repository: Create')
    .should('be.visible');
  cy.typeValue('Name', repoName);
  if (repoType === 'git') {
    cy.contains('Git repository')
      .click();
    cy.typeValue('Git Repo URL', repoUrl);
    cy.typeValue('Git Branch', 'main');
  } else {
    cy.typeValue('Index URL', repoUrl);
  }
  cy.clickButton('Create');
});
