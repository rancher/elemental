import * as cypressLib from '@rancher-ecp-qa/cypress-library';

// Check the Cypress tags
// Implemented but not used yet
export const isCypressTag = (tag: string) => {
  return (new RegExp(tag)).test(Cypress.env("cypress_tags"));
}

// Check the K8s version
export const isK8sVersion = (version: string) => {
  version = version.toLowerCase();
  return (new RegExp(version)).test(Cypress.env("k8s_downstream_version"));
}

// Check the Elemental operator version
export const isOperatorVersion = (version: string) => {
  return (new RegExp(version)).test(Cypress.env("operator_repo"));
}

// Check rancher manager version
export const isRancherManagerVersion = (version: string) => {
  return (new RegExp(version)).test(Cypress.env("rancher_version"));
}

// Check Elemental UI version
export const isUIVersion = (version: string) => {
  return (new RegExp(version)).test(Cypress.env("elemental_ui_version"));
}

// Check the upgrade target
export const isUpgradeOsChannel = (channel: string) => {
  return (new RegExp(channel)).test(Cypress.env("upgrade_os_channel"));
}

// Create Elemental cluster
export const createCluster = (clusterName: string, k8sVersion: string, proxy: string) => {
  cy.getBySel('button-create-elemental-cluster')
    .click();
  cy.getBySel('name-ns-description-name')
    .type(clusterName);
  cy.getBySel('name-ns-description-description')
    .type('My Elemental testing cluster');
  cy.contains('.labeled-input.create', 'Machine Count')
    .clear()
  if (isCypressTag('main')) {
    cy.contains('.labeled-input.create', 'Machine Count')
      .type('3');
  } else {
    cy.contains('.labeled-input.create', 'Machine Count')
      .type('1');
  }
  cy.contains('Show deprecated Kubernetes')
    .click();
  cy.contains('Kubernetes Version')
    .click();
  cy.contains(k8sVersion)
    .click();
  // Configure proxy if proxy is set to elemental
  if (Cypress.env('proxy') == "elemental") {
    cy.contains('Agent Environment Vars')
      .click();
    cy.get('#agentEnv > .key-value')
      .contains('Add')
      .click();
    cy.getBySel('input-kv-item-key-0')
      .type('HTTP_PROXY');
    cy.getBySel('kv-item-value-0')
      .type(proxy);
    cy.get('#agentEnv > .key-value')
      .contains('Add')
      .click();
    cy.getBySel('input-kv-item-key-1')
      .type('HTTPS_PROXY');
    cy.getBySel('kv-item-value-1').type(proxy);
    cy.get('#agentEnv > .key-value')
      .contains('Add')
      .click();
    cy.getBySel('input-kv-item-key-2')
      .type('NO_PROXY');
    cy.getBySel('kv-item-value-2')
      .type('localhost,127.0.0.0/8,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,.svc,.cluster.local');
    }
  cy.clickButton('Create');
  // This wait can be replaced by something cleaner
  // eslint-disable-next-line cypress/no-unnecessary-waiting
  cy.wait(3000);
  cypressLib.burgerMenuToggle();
  cypressLib.checkClusterStatus(clusterName, 'Updating', 300000);
  cypressLib.burgerMenuToggle();
  cypressLib.checkClusterStatus(clusterName, 'Active', 600000);
  // Ugly but needed unfortunately to make sure the cluster stops switching status
  // eslint-disable-next-line cypress/no-unnecessary-waiting
  cy.wait(240000);
}
