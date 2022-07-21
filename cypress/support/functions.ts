// Generic functions

// Log into Rancher
Cypress.Commands.add('login', (username = Cypress.env('username'), password = Cypress.env('password'), cacheSession = Cypress.env('cache_session')) => {
  const login = () => {
    let loginPath
    loginPath="/v3-public/localProviders/local*";
    cy.intercept('POST', loginPath).as('loginReq');
    
    cy.visit('/auth/login');

    cy.byLabel('Username')
      .focus()
      .type(username, {log: false});

    cy.byLabel('Password')
      .focus()
      .type(password, {log: false});

    cy.get('button').click();
    cy.wait('@loginReq');
    cy.contains("Getting Started", {timeout: 10000}).should('be.visible');
    } 

  if (cacheSession) {
    cy.session([username, password], login);
  } else {
    login();
  }
});

// Search fields by label
Cypress.Commands.add('byLabel', (label) => {
  cy.get('.labeled-input').contains(label).siblings('input');
});

// Search button by label
Cypress.Commands.add('clickButton', (label) => {
  cy.get('.btn').contains(label).click();
});

// Insert a value in a field *BUT* force a clear before!
Cypress.Commands.add('typeValue', ({label, value, noLabel, log=true}) => {
  if (noLabel === true) {
    cy.get(label).focus().clear().type(value, {log: log});
  } else {
    cy.byLabel(label).focus().clear().type(value, {log: log});
  }
});

// Insert a key/value pair
Cypress.Commands.add('typeKeyValue', ({key, value}) => {
  cy.get(key).clear().type(value);
});

Cypress.Commands.overwrite('type', (originalFn, subject, text, options = {}) => {
  options.delay = 100;

  return originalFn(subject, text, options);
});

// Add a delay between command without using cy.wait()
// https://github.com/cypress-io/cypress/issues/249#issuecomment-443021084
const COMMAND_DELAY = 1000;

for (const command of ['visit', 'click', 'trigger', 'type', 'clear', 'reload', 'contains']) {
    Cypress.Commands.overwrite(command, (originalFn, ...args) => {
        const origVal = originalFn(...args);

        return new Promise((resolve) => {
            setTimeout(() => {
                resolve(origVal);
            }, COMMAND_DELAY);
        });
    });
}; 
