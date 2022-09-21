import { TopLevelMenu } from '~/cypress/support/toplevelmenu';
import '~/cypress/support/functions';
import { Elemental } from '../../support/elemental';

Cypress.config();
describe('Upgrade tests', () => {
  const topLevelMenu = new TopLevelMenu();
  const elemental    = new Elemental();

  beforeEach(() => {
    cy.login();
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
    cy.contains('myelementalcluster').click();
    cy.typeValue({label: 'OS Image', value: 'quay.io/costoolkit/elemental-ci:latest'});
    cy.clickButton('Create');
    cy.get('[data-testid="sortable-cell-0-0"]').contains('Active');

    // Check if the node reboots to apply the upgrade
    cy.clickNavMenu(["Dashboard"]);
    cy.clickButton('Manage Elemental Clusters');
    cy.get('.title').contains('Clusters');
    cy.contains('myelementalcluster').click();
    cy.get('.primaryheader').contains('Active');
    cy.get('.primaryheader').contains('Updating', {timeout: 240000});
    cy.get('.primaryheader').contains('Active', {timeout: 240000});
  });
});
