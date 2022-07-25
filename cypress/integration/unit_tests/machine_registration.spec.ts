import { TopLevelMenu } from '~/cypress/support/toplevelmenu';
import '~/cypress/support/functions';
import { Elemental } from '../../support/elemental';
import cypress from 'cypress';

Cypress.config();
describe('Machine registration testing', () => {
  const topLevelMenu = new TopLevelMenu();
  const elemental = new Elemental();

  beforeEach(() => {
    cy.login();
    cy.visit('/');

    // Open the navigation menu
    topLevelMenu.openIfClosed();

    // Click on the Elemental's icon
    elemental.accessElementalMenu(); 
    
    // Delete all files previously downloaded
    cy.exec('rm cypress/downloads/*', {failOnNonZeroExit: false});

    // Delete namespace
    cy.exec('kubectl --kubeconfig=/etc/rancher/k3s/k3s.yaml delete ns mynamespace', {failOnNonZeroExit: false});
    
    // Delete all existing machine registrations
    cy.get('body').then(($body) => {
      if ($body.text().includes('Manage Machine Registrations')) {
        cy.deleteAllMachReg();
      };
    });
  });

  it('Create machine registration with default options', () => {
    cy.createMachReg({machRegName: 'default-options-test'});
  });

  it('Create machine registration in custom namespace', () => {
    cy.createMachReg({machRegName: 'custom-namespace-test', namespace: 'mynamespace'});
  });

  it('Create machine registration with labels and annotations', () => {
    cy.createMachReg({machRegName: 'labels-annotations-test', checkLabels: true, checkAnnotations: true});
  });

  it.skip('Create machine registration with custom cloud-config', () => {
      // Cannot be tested yet due to https://github.com/rancher/dashboard/issues/6458
  });

  it('Delete machine registration', () => {
    cy.createMachReg({machRegName: 'delete-test'});
    cy.deleteMachReg({machRegName: 'delete-test'});
  });

});
