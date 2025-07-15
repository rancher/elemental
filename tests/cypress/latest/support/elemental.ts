/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import { isCypressTag, isGitRepo, isOperatorVersion, isRancherManagerVersion, isRancherPrime } from '~/support/utils';

export class Elemental {
  // Go into the cluster creation menu
  accessClusterMenu(): void {
    cy.contains('Dashboard').click();
    cy.getBySel('elemental-main-title').should('exist');
    cy.getBySel('card-clusters').contains('Create Elemental Cluster').should('exist');
    cy.getBySel('button-create-elemental-cluster').click();
  }

  // Make sure we get all menus
  checkElementalNav(): void {
    // Open advanced accordion
    if (isRancherManagerVersion('2.12') || isRancherManagerVersion('rancher:head')) {
      cy.get('.accordion-item > .icon').eq(0).click();
    } else {
      cy.get('div.header > i').eq(0).click();
    }
    cy.get('div.header').contains('Advanced').should('be.visible');
    // Check all listed options once accordion is opened
    cy.get('li.child.nav-type').should(($lis) => {
      expect($lis).to.have.length(7);
      // There is a bug with Dashboard entry in rancher 2.10
      // https://github.com/rancher/elemental-ui/issues/230
      if (isRancherManagerVersion('2.9')) {
        expect($lis.eq(0)).to.contain('Dashboard');
      }
      expect($lis.eq(1)).to.contain('Registration Endpoints');
      expect($lis.eq(2)).to.contain('Inventory of Machines');
      expect($lis.eq(3)).to.contain('Update Groups');
      expect($lis.eq(4)).to.contain('OS Versions');
      expect($lis.eq(5)).to.contain('OS Version Channels');
      expect($lis.eq(6)).to.contain('Seed Images');
    });
  }

  // Get elemental operator version
  getOperatorVersion(): void {
    cy.contains('local').click();
    cy.contains('Workloads').click();
    cy.contains('Deployments').click();
    cy.contains('elemental-operator').click()
    cy.get('[data-testid="sortable-cell-0-2"] > .formatter-pod-images > span').invoke('text').then((version:string) => {
      Cypress.env('elemental_operator_version', version);
      cy.log(`Elemental operator version: ${version}`);
    });
  }

  installElementalOperator(upgrade_from_version: string): void {
    cy.contains('local').click();
    cy.get('.nav').contains('Apps').click();

    if (isCypressTag('main') && !isOperatorVersion('marketplace')) {
      isRancherManagerVersion('rancher:head') ? cy.get('[data-testid="item-card-cluster/elemental-operator/elemental-operator"]').click() : cy.contains('.item.has-description.color1', 'Elemental', { timeout: 30000 }).click();
    } else {
        // Uncheck Rancher (rancher.io) repo if it's checked
        if (isGitRepo('github')) {
          cy.get('#vs1__combobox > .vs__selected-options').click();
          cy.get('#vs1__option-1 > .checkbox-outer-container > .checkbox-container').click();
        }
      cy.contains('Elemental', { timeout: 30000 }).click();
    }

    cy.contains('Charts: Elemental', { timeout: 30000 });

    if (isCypressTag('upgrade') && isOperatorVersion('marketplace')) {
      cy.contains(upgrade_from_version, { timeout: 30000 }).click();
    }

    cy.clickButton('Install');
    if (isRancherManagerVersion('2.11') || isRancherManagerVersion('rancher:head')) {
      cy.contains('.top > .title', 'Elemental') 
    } else {
      cy.contains('.outer-container > .header', 'Elemental');
    }
    if (isRancherPrime() && isCypressTag('main') && !isOperatorVersion('marketplace')) {
      const registryLabel = 'Container Registry';
      cy.byLabel(registryLabel).clear();
      if (isOperatorVersion('staging')) {
        cy.byLabel(registryLabel).type('registry.opensuse.org/isv/rancher/elemental/staging/containers');
      } else if (isOperatorVersion('dev')) {
        cy.byLabel(registryLabel).type('registry.opensuse.org/isv/rancher/elemental/dev/containers');
      }
    }

    cy.clickButton('Next');
    cy.clickButton('Install');
    cy.contains('SUCCESS: helm', { timeout: 120000 });
    cy.reload();
    if (isRancherManagerVersion('2.9') || isRancherManagerVersion('2.8')) {
      // eslint-disable-next-line cypress/unsafe-to-chain-command
      cy.contains('Only User Namespaces').click().type('cattle-elemental-system{enter}{esc}');
      cy.get('.outlet').contains('Deployed elemental-operator cattle-elemental-system', { timeout: 120000 });
    } else {
      cy.contains('Only User Namespaces').click()
      // Select All Namespaces entry
      cy.getBySel('namespaces-option-0').click();
      cy.get('.outlet').contains(new RegExp('Deployed.*elemental-operator'), { timeout: 120000 });
    }
  }
}
