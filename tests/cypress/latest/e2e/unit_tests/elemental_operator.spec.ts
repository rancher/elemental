/*
Copyright © 2022 - 2025 SUSE LLC

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
import { isCypressTag, isGitRepo, isOperatorInstallType, isOperatorVersion, isRancherManagerVersion } from '~/support/utils';
import { Elemental } from '~/support/elemental';

filterTests(['main', 'upgrade'], () => {
  describe('Install Elemental Operator', () => {
    const elemental = new Elemental();
    const upgradeFromVersion = Cypress.env('upgrade_from_version');
    const chartmuseumRepo = Cypress.env('chartmuseum_repo') + ':8080';

    beforeEach(() => {
      cy.viewport(1920, 1080);
      cy.login();
      cy.visit('/');
      cypressLib.burgerMenuToggle();
    });

    if (isOperatorVersion('marketplace') && isGitRepo('github')) {
      it('Configure from which git repo charts are pulled from', () => {
        cypressLib.addRepository('rancher-dev', 'https://github.com/rancher/charts.git', 'git', isRancherManagerVersion('2.9') ? 'dev-v2.9' : 'dev-v2.8');
      });
    }
    // Add dev repo for main test or if the test runs on Rancher 2.7 (because operator is not in the 2.7 marketplace)
    if ((!isOperatorVersion('marketplace') && isCypressTag('main')) || isRancherManagerVersion('2.7')) {
      it('Add local chartmuseum repo', () => {
        cypressLib.addRepository('elemental-operator', chartmuseumRepo, 'helm', 'none');
      });

      qase(10,
        it('Install latest dev Elemental operator', () => {
          elemental.installElementalOperator(upgradeFromVersion);
      }));
    } else if (!isRancherManagerVersion('2.7') && !isOperatorInstallType('cli')) {
      qase(57,
        it('Install latest stable Elemental operator', () => {
          elemental.installElementalOperator(upgradeFromVersion);
      }));
    }
  });
});
