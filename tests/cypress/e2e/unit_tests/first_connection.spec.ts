/*
Copyright © 2022 - 2023 SUSE LLC

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

import { Elemental } from '~/cypress/support/elemental';
import filterTests from '~/cypress/support/filterTests.js';

filterTests(['main', 'upgrade'], () => {
  Cypress.config();
  describe('First login on Rancher', () => {
    const elemental = new Elemental();

    it('Log in and accept terms and conditions', () => {
      cy.visit('/auth/login');
      cy.get("span").then($text => {
        if ($text.text().includes('your first time visiting Rancher')) {
          elemental.firstLogin();
        }
        else {
          cy.log('Rancher already initialized, no need to handle first login.')
        }
      })
    });
  });
})
