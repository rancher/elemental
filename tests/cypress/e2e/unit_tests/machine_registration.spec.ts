import { TopLevelMenu } from '~/cypress/support/toplevelmenu';
import '~/cypress/support/functions';
import { Elemental } from '~/cypress/support/elemental';

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
    cy.contains('Manage Registration Endpoints').click();
    cy.get('.outlet > header').contains('Registration Endpoints');
    cy.get('body').then(($body) => {
      if (!$body.text().includes('There are no rows to show.')) {
        cy.deleteAllResources();
      };
    });
  });
  

  it('Create machine registration with default options', () => {
    cy.createMachReg({machRegName: 'default-options-test'});
  });

  it('Create machine registration with labels and annotations', () => {
    cy.createMachReg({machRegName: 'labels-annotations-test', checkLabels: true, checkAnnotations: true});
  });

  it('Delete machine registration', () => {
    cy.createMachReg({machRegName: 'delete-test'});
    cy.deleteMachReg({machRegName: 'delete-test'});
  });

  it('Edit a machine registration with edit config button', () => {
    cy.createMachReg({machRegName: 'edit-config-test'});
    cy.editMachReg({machRegName: 'edit-config-test', addLabel: true, addAnnotation: true });
    cy.clickButton('Save');

    // Check that we can see our label and annotation in the YAML
    cy.checkMachRegLabel({machRegName: 'edit-config-test', labelName: 'myLabel1', labelValue: 'myLabelValue1'});
    cy.checkMachRegAnnotation({machRegName: 'edit-config-test', annotationName: 'myAnnotation1', annotationValue: 'myAnnotationValue1'});
  });

  it('Edit a machine registration with edit YAML button', () => {
    cy.createMachReg({machRegName: 'edit-yaml-test'});
    cy.editMachReg({machRegName: 'edit-yaml-test', addLabel: true, addAnnotation: true, withYAML: true });
    cy.clickButton('Save');

    // Check that we can see our label and annotation in the YAML
    cy.checkMachRegLabel({machRegName: 'edit-yaml-test', labelName: 'myLabel1', labelValue: 'myLabelValue1'});
    cy.checkMachRegAnnotation({machRegName: 'edit-yaml-test', annotationName: 'myAnnotation1', annotationValue: 'myAnnotationValue1'});
  });

  it('Clone a machine registration', () => {
    cy.createMachReg({machRegName: 'clone-test', checkLabels: true, checkAnnotations: true});
    cy.contains('clone-test').click();
    cy.get('div.actions > .role-multi-action').click()
    cy.contains('li', 'Clone').click();
    cy.typeValue({label: 'Name', value: 'cloned-machine-reg'});
    cy.clickButton('Create');
    cy.contains('.masthead', 'Registration Endpoint: cloned-machine-regActive').should('exist');
    
    // Check that we got the same label and annotation in both machine registration
    cy.checkMachRegLabel({machRegName: 'cloned-machine-reg', labelName: 'myLabel1', labelValue: 'myLabelValue1'});
    cy.contains('cloned-machine-reg').click();
    cy.checkMachRegAnnotation({machRegName: 'cloned-machine-reg', annotationName: 'myAnnotation1', annotationValue: 'myAnnotationValue1'});
    cy.contains('cloned-machine-reg').click();
  });

  it('Download Machine registration YAML', () => {
    cy.createMachReg({machRegName: 'download-yaml-test'});
    cy.contains('download-yaml-test').click();
    cy.get('div.actions > .role-multi-action').click()
    cy.contains('li', 'Download YAML').click();
    cy.verifyDownload('download-yaml-test.yaml');
  });

  // This test must stay the last one because we use this machine registration when we test adding a node.
  // It also tests using a custom cloud config by using read from file button.
  it('Create Machine registration we will use to test adding a node', () => {
    cy.createMachReg({machRegName: 'machine-registration', checkInventoryLabels: true, checkInventoryAnnotations: true, customCloudConfig: 'custom_cloud-config.yaml', checkDefaultCloudConfig: false});
    cy.checkMachInvLabel({machRegName: 'machine-registration', labelName: 'myInvLabel1', labelValue: 'myInvLabelValue1'});
  });
});
