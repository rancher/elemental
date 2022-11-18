import { TopLevelMenu } from '~/cypress/support/toplevelmenu';
import '~/cypress/support/functions';
import { Elemental } from '../../support/elemental';

Cypress.config();
describe('Machine registration testing', () => {
  const topLevelMenu   = new TopLevelMenu();
  const elemental      = new Elemental();
  const ui_account     = Cypress.env('ui_account');
  const elemental_user = "elemental-user"
  const ui_password    = "rancherpassword"

  beforeEach(() => {
    (ui_account == "user") ? cy.login(elemental_user, ui_password) : cy.login();
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
    cy.contains('Manage Machine Registrations').click();
    cy.get('.outlet > header').contains('Machine Registrations');
    cy.get('body').then(($body) => {
      if (!$body.text().includes('There are no rows to show.')) {
        cy.deleteAllMachReg();
      };
    });
  });


  // This test must stay the last one because we use this machine registration when we test adding a node.
  // It also tests using a custom cloud config by using read from file button.
  it('Create Machine registration we will use to test adding a node', () => {
    cy.createMachReg({machRegName: 'machine-registration', checkInventoryLabels: true, checkInventoryAnnotations: true, customCloudConfig: 'custom_cloud-config.yaml', checkDefaultCloudConfig: false});
    cy.checkMachInvLabel({machRegName: 'machine-registration', labelName: 'myInvLabel1', labelValue: 'myInvLabelValue1'});
  });
});
