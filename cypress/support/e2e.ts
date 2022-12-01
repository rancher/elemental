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
      clickClusterMenu(listLabel: string[],): Chainable<Element>;
      typeValue(label: string, value: string, noLabel?: boolean, log?: boolean): Chainable<Element>;
      typeKeyValue(key: string, value: string,): Chainable<Element>;
      getDetail(name: string, type: string, namespace?: string): Chainable<Element>;
      createMachReg(machRegName: string, namespace?: string, checkLabels?: boolean, checkAnnotations?: boolean, customCloudConfig?: string, checkDefaultCloudConfig?: boolean): Chainable<Element>;
      deleteMachReg(machRegName: string): Chainable<Element>;
      deleteAllResources():Chainable<Element>;
      addMachRegLabel(labelName: string, labelValue: string):Chainable<Element>;
      checkMachRegLabel(machRegName: string, labelName: string, labelValue: string):Chainable<Element>;
      checkMachRegAnnotation(machRegName: string, annotationName: string, annotationValue: string):Chainable<Element>;
      addMachRegAnnotation(annotationName: string, annotationValue: string):Chainable<Element>;
      addMachInvLabel(labelName: string, labelValue: string):Chainable<Element>;
      checkMachInvLabel(machRegName: string, labelName: string, labelValue: string):Chainable<Element>;
      checkMachInvAnnotation(machRegName: string, annotationName: string, annotationValue: string):Chainable<Element>;
      addMachInvAnnotation(annotationName: string, annotationValue: string):Chainable<Element>;
      editMachReg(machRegName: string, addLabel?: boolean, addAnnotation?: boolean, withYAML?: boolean): Chainable<Element>;
      addHelmRepo(repoName: string, repoUrl: string, repoType?: string,): Chainable<Element>;
      importMachineInventory(machineInventoryFile: string, machineInventoryName: string): Chainable<Element>;
      checkFilter(filterName: string, testFilterOne: boolean, testFilterTwo: boolean, shouldNotMatch: boolean): Chainable<Element>;
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

