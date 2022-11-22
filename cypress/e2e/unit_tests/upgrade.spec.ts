import { TopLevelMenu } from '~/cypress/support/toplevelmenu';
import '~/cypress/support/functions';
import { Elemental } from '../../support/elemental';
import 'cypress-file-upload';

Cypress.config();
describe('Upgrade tests', () => {
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

  });

  it('Upgrade one node with OS Image Upgrades', () => {
    // Create ManagedOSImage resource
    cy.get('.nav').contains('Advanced').click();
    cy.get('.nav').contains('OS Image Upgrades').click();
    cy.clickButton('Create');
    cy.get('.primaryheader').contains('OS Image Upgrade: Create');
    cy.typeValue({label: 'Name', value: 'upgrade'});
    cy.contains('Target Cluster').click();
    cy.contains('1').click();
    cy.typeValue({label: 'OS Image', value: 'quay.io/costoolkit/elemental-ci:latest'});
    cy.clickButton('Create');
    cy.wait(5000);
    cy.get('[data-testid="sortable-cell-0-0"]').contains('Active');

    // Check if the node reboots to apply the upgrade
    topLevelMenu.openIfClosed();
    cy.contains('Cluster Management').click();
    cy.get('.title').contains('Clusters');
    cy.contains('1').click();
    cy.get('.primaryheader').contains('Active');
    cy.reload()
    // wait half an hour to get logs of a working example
    //cy.wait(1800000)
    cy.get('.primaryheader').contains('Updating', {timeout: 120000});
    cy.get('.primaryheader').contains('Active', {timeout: 240000});
  });
});
