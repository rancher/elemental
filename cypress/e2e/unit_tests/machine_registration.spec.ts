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
    cy.contains('.masthead', 'Machine Registration: cloned-machine-reg Active').should('exist');
    
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

});
