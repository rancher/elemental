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

import { sleep } from 'k6';
import { testMachineRegistration } from '/scripts/tests/machineregistration.js';

export const options = {
    vus: 10,
    duration: '30s',
};

export default function () {
    const url = __ENV.API_URL;
    const token = __ENV.SA_TOKEN;
    const ns = 'default';

    if (!url) {
        console.log("No API_URL provided, running smoke test only");
        sleep(1);
        return;
    }

    testMachineRegistration(url, token, ns);
}
