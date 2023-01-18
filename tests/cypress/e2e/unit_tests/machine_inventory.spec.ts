import { TopLevelMenu } from '~/cypress/support/toplevelmenu';
import '~/cypress/support/functions';
import { Elemental } from '~/cypress/support/elemental';

Cypress.config();
describe('Machine inventory testing', () => {
  const topLevelMenu   = new TopLevelMenu();
  const elemental      = new Elemental();
  const k8s_version    = Cypress.env('k8s_version');
  const ui_account     = Cypress.env('ui_account');
  const elemental_user = 'elemental-user'
  const ui_password    = 'rancherpassword'
  const proxy          = 'http://172.17.0.1:3128'

  beforeEach(() => {
    (ui_account == "user") ? cy.login(elemental_user, ui_password) : cy.login();
    cy.visit('/');

    // Open the navigation menu
    topLevelMenu.openIfClosed();

    // Click on the Elemental's icon
    elemental.accessElementalMenu(); 
  });

  it('Check that machine inventory has been created', () => {
    cy.clickNavMenu(["Inventory of Machines"]);
    cy.contains('.badge-state', 'Active').should('exist');
    cy.contains('Namespace: fleet-default').should('exist');
  });

  it('Create Elemental cluster', () => {
    cy.contains('Create Elemental Cluster').click();
    cy.typeValue({label: 'Cluster Name', value: 'myelementalcluster'});
    cy.typeValue({label: 'Cluster Description', value: 'My Elemental testing cluster'});
    cy.contains('Show deprecated Kubernetes').click();
    cy.contains('Kubernetes Version').click();
    cy.contains(k8s_version).click();
    // Configure proxy if proxy is set to elemental
    if ( Cypress.env('proxy') == "elemental") {
      cy.contains('Agent Environment Vars').click();
      cy.get('#agentEnv > .key-value').contains('Add').click();
      cy.get('.key > input').type('HTTP_PROXY');
      cy.get('.no-resize').type(proxy);
      cy.get('#agentEnv > .key-value').contains('Add').click();
      cy.get(':nth-child(7) > input').type('HTTPS_PROXY');
      cy.get(':nth-child(8) > .no-resize').type(proxy);
      cy.get('#agentEnv > .key-value').contains('Add').click();
      cy.get(':nth-child(10) > input').type('NO_PROXY');
      cy.get(':nth-child(11) > .no-resize').type('localhost,127.0.0.0/8,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,.svc,.cluster.local');
    }
    cy.clickButton('Create');
    cy.contains('Updating myelementalcluster', {timeout: 20000});
    cy.contains('Active myelementalcluster', {timeout: 360000});
  });

  it('Check Elemental cluster status', () => {
    topLevelMenu.openIfClosed();
    cy.contains('Home').click();
    // The new cluster must be in active state
    cy.get('[data-node-id="fleet-default/myelementalcluster"]').contains('Active');
    // Go into the dedicated cluster page
    topLevelMenu.openIfClosed();
    cy.contains('myelementalcluster').click();
  })
});
