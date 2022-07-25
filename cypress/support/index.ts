import './functions';

declare global {
  // eslint-disable-next-line no-unused-vars
  namespace Cypress {
    interface Chainable {
      // Functions declared in functions.ts
      login(username?: string, password?: string, cacheSession?: boolean,): Chainable<Element>;
      byLabel(label: string,): Chainable<Element>;
      clickButton(label: string,): Chainable<Element>;
      confirmDelete(): Chainable<Element>;
      clickNavMenu(listLabel: string[],): Chainable<Element>;
      clickElementalMenu(label: string,): Chainable<Element>;
      typeValue(label: string, value: string, noLabel?: boolean, log?: boolean): Chainable<Element>;
      typeKeyValue(key: string, value: string,): Chainable<Element>;
      getDetail(name: string, type: string, namespace?: string): Chainable<Element>;
      createMachReg(machRegName: string, namespace?: string, checkLabels?: boolean, checkAnnotations?: boolean): Chainable<Element>;
      deleteMachReg(machRegName: string): Chainable<Element>;
      deleteAllMachReg():Chainable<Element>;
    }
}}

// TODO handle redirection errors better?
// we see a lot of 'error navigation cancelled' uncaught exceptions that don't actually break anything; ignore them here
Cypress.on('uncaught:exception', (err, runnable) => {
  // returning false here prevents Cypress from failing the test
  if (err.message.includes('navigation guard')) {
    return false;
  }
});

require('cypress-dark');
require('cy-verify-downloads').addCustomCommand();
require('cypress-plugin-tab');

