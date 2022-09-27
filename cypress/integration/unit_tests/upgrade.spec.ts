import { TopLevelMenu } from '~/cypress/support/toplevelmenu';
import '~/cypress/support/functions';
import { Elemental } from '../../support/elemental';
import 'cypress-file-upload';

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

  it('Create an OS Version Channels', () => {
    // Create ManagedOSVersionChannel resource
    cy.exec(`sed -i 's/# namespace: fleet-default/namespace: fleet-default/g' tests/assets/managedOSVersionChannel.yaml`)
    cy.get('.nav').contains('Advanced').click();
    cy.get('.nav').contains('Managed OS Version Channels').click();
    cy.clickButton('Create from YAML');
    // Wait needed to avoid crash with the upload
    cy.wait(2000);
    cy.get('input[type="file"]').attachFile({filePath: '../../tests/assets/managedOSVersionChannel.yaml'});
    // Wait needed to avoid crash with the upload
    cy.wait(2000);
    // Wait for os-versions to be printed, that means the upload is done
    cy.contains('os-versions');
    cy.clickButton('Create');
    // The new resource must be active
    cy.contains('Active');
  });

  it('Check OS Versions', () => {
    cy.get('.nav').contains('Advanced').click();
    cy.get('.nav').contains('Managed OS Versions').click();
    cy.contains('Active teal-5.2', {timeout: 120000});
    cy.contains('Active teal-5.3');
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
