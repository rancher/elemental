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
export class Rancher {
  burgerMenuToggle() {
    cy.getBySel('top-level-menu', {timeout: 12000})
      .click();
  }

  burgerMenuOpenIfClosed() {
    cy.get('body').then((body) => {
      if (body.find('.menu.raised').length === 0) {
        this.burgerMenuToggle();
      }
    });
  }

  addRepository(repositoryName: string, repositoryURL: string, repositoryType: string) {
    this.burgerMenuOpenIfClosed();
    cy.contains('local')
      .click();
    cy.addHelmRepo(repositoryName, repositoryURL, repositoryType);
  }
}

// RANCHER CUSTOM CYPRESS COMMANDS

declare global {
  namespace Cypress {
    interface Chainable {
      addHelmRepo(repoName: string, repoUrl: string, repoType?: string,): Chainable<Element>;
      byLabel(label: string,): Chainable<Element>;
      clickButton(label: string,): Chainable<Element>;
      confirmDelete(): Chainable<Element>;
      login(username?: string, password?: string, cacheSession?: boolean,): Chainable<Element>;
      getBySel(dataTestAttribute: string, args?: any): Chainable<JQuery<HTMLElement>>;
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

// Add Helm repo
Cypress.Commands.add('addHelmRepo', (repoName, repoUrl, repoType) => {
  cy.clickClusterMenu(['Apps', 'Repositories'])

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
});1