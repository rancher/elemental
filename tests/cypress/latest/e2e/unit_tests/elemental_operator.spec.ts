/*
Copyright © 2022 - 2024 SUSE LLC

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

import '~/support/commands';
import filterTests from '~/support/filterTests.js';
import * as cypressLib from '@rancher-ecp-qa/cypress-library';
import { qase } from 'cypress-qase-reporter/dist/mocha';
import { isCypressTag, isRancherManagerVersion } from '~/support/utils';
import { Elemental } from '~/support/elemental';
import { slowCypressDown } from 'cypress-slow-down'

// slow down each command by 500ms
slowCypressDown(500)

filterTests(['main', 'upgrade'], () => {
  Cypress.config();
  describe('Install Elemental Operator', () => {
    const elemental = new Elemental();
  
    beforeEach(() => {
      cy.login();
      cy.visit('/');
      cypressLib.burgerMenuToggle();
    });
    // Add dev repo for main test or if the test runs on rancher 2.7 (because operator is not in the 2.7 marketplace)
    if (isCypressTag('main') || isRancherManagerVersion('2.7')) {
      it('Add local chartmuseum repo', () => {
        cypressLib.addRepository('elemental-operator', Cypress.env('chartmuseum_repo')+':8080', 'helm', 'none');
      });
      qase(10,
        it('Install latest dev Elemental operator', () => {
          elemental.installElementalOperator();
        })
      );
    } else if (isCypressTag('upgrade') && !isRancherManagerVersion('2.7')) {
      qase(57,
        it('Install latest stable Elemental operator', () => {
          elemental.installElementalOperator();
        })
      );
    };
  });
});
