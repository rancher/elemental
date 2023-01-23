import { TopLevelMenu } from '~/cypress/support/toplevelmenu';
import '~/cypress/support/functions';
import { Elemental } from '~/cypress/support/elemental';
import 'cypress-file-upload';

Cypress.config();
describe('Upgrade tests', () => {
  const topLevelMenu     = new TopLevelMenu();
  const elemental        = new Elemental();
  const ui_account       = Cypress.env('ui_account');
  const operator_version = Cypress.env('operator_version');
  const elemental_user   = "elemental-user"
  const ui_password      = "rancherpassword"

  beforeEach(() => {
    (ui_account == "user") ? cy.login(elemental_user, ui_password) : cy.login();
    cy.visit('/');

    // Open the navigation menu
    topLevelMenu.openIfClosed();

    // Click on the Elemental's icon
    elemental.accessElementalMenu(); 
  });

  it('Create an OS Version Channels', () => {
    (operator_version == "1.0") ? cy.exec(`sed -i '/syncInterval/d' assets/managedOSVersionChannel.yaml`) : "" ;
    // Create ManagedOSVersionChannel resource
    cy.exec(`sed -i 's/# namespace: fleet-default/namespace: fleet-default/g' assets/managedOSVersionChannel.yaml`);
    cy.get('.nav').contains('Advanced').click();
    cy.get('.nav').contains('OS Version Channels').click();
    cy.clickButton('Create from YAML');
    // Wait needed to avoid crash with the upload
    cy.wait(2000);
    cy.get('input[type="file"]').attachFile({filePath: '../../assets/managedOSVersionChannel.yaml'});
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
    cy.get('.nav').contains('OS Versions').click();
    cy.contains('Active fake-image', {timeout: 120000});
    cy.contains('Active teal-5.3');
  });

  it('Delete OS Versions', () => {
    cy.get('.nav').contains('Advanced').click();
    cy.get('.nav').contains('OS Versions').click();
    cy.contains('fake-image').parent().parent().click();
    cy.clickButton('Delete');
    cy.confirmDelete();
    cy.contains('fake-image').should('not.exist');
  });

  it('Delete OS Versions Channels', () => {
    cy.get('.nav').contains('Advanced').click();
    cy.get('.nav').contains('OS Version Channels').click();
    cy.deleteAllResources();
    cy.get('.nav').contains('Advanced').click();
    cy.get('.nav').contains('OS Versions').click();
    cy.contains('There are no rows to show');
  });

  it('Upgrade one node with OS Image Upgrades', () => {
    // Create ManagedOSImage resource
    cy.get('.nav').contains('Advanced').click();
    cy.get('.nav').contains('Update Groups').click();
    cy.clickButton('Create');
    cy.get('.primaryheader').contains('Update Group: Create');
    cy.typeValue({label: 'Name', value: 'upgrade'});
    cy.contains('Target Cluster').click();
    cy.contains('myelementalcluster').click();
    cy.typeValue({label: 'OS Image', value: 'quay.io/costoolkit/elemental-ci:latest'});
    cy.clickButton('Create');
    // Status changes a lot right after the creation so let's wait 10 secondes
    // before checking
    cy.wait(10000);
    cy.get('[data-testid="sortable-cell-0-0"]').contains('Active');

    // Workaround to avoid sporadic issue with Upgrade
    // https://github.com/rancher/elemental/issues/410
    // Restart fleet agent inside downstream cluster
    topLevelMenu.openIfClosed();
    cy.contains('myelementalcluster').click();
    cy.contains('Workload').click();
    cy.contains('Pods').click();
    cy.get('.header-buttons > :nth-child(2)').click();
    cy.wait(20000);
    cy.get('.shell-body').type('kubectl scale deployment/fleet-agent -n cattle-fleet-system --replicas=0{enter}');
    cy.get('.shell-body').type('kubectl scale deployment/fleet-agent -n cattle-fleet-system --replicas=1{enter}');

    // Check if the node reboots to apply the upgrade
    topLevelMenu.openIfClosed();
    elemental.accessElementalMenu();
    cy.clickNavMenu(["Dashboard"]);
    cy.clickButton('Manage Elemental Clusters');
    cy.get('.title').contains('Clusters');
    cy.contains('myelementalcluster').click();
    cy.get('.primaryheader').contains('Active');
    cy.get('.primaryheader').contains('Updating', {timeout: 240000});
    cy.get('.primaryheader').contains('Active', {timeout: 240000});
  });
});
