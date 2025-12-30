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

import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

export function testMachineRegistration(url, token, ns) {
    const resourceName = `k6-mr-${uuidv4()}`;

    const params = {
        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
        },
        insecureSkipTLSVerify: true,
    };

    const payload = JSON.stringify({
        apiVersion: "elemental.cattle.io/v1beta1",
        kind: "MachineRegistration",
        metadata: {
            name: resourceName,
            namespace: ns,
        },
        spec: {
            machineName: "k6-node-${System Information/UUID}",
            machineInventoryLabels: {
                "testing.k6.io/managed": "true"
            }
        }
    });

    const createRes = http.post(`${url}/apis/elemental.cattle.io/v1beta1/namespaces/${ns}/machineregistrations`, payload, params);

    check(createRes, {
        'create status is 201': (r) => r.status === 201,
    });

    if (createRes.status !== 201) {
        console.log(`Failed to create ${resourceName}: ${createRes.status} ${createRes.body}`);
        return;
    }

    for (let i = 0; i < 5; i++) {
        const getRes = http.get(`${url}/apis/elemental.cattle.io/v1beta1/namespaces/${ns}/machineregistrations/${resourceName}`, params);
        if (getRes.status !== 200) break;

        const obj = getRes.json();

        if (!obj.metadata.annotations) obj.metadata.annotations = {};
        obj.metadata.annotations[`k6-update-${i}`] = String(Date.now());

        const putRes = http.put(`${url}/apis/elemental.cattle.io/v1beta1/namespaces/${ns}/machineregistrations/${resourceName}`, JSON.stringify(obj), params);

        check(putRes, {
            'update status is 200 or 409': (r) => r.status === 200 || r.status === 409,
        });

        sleep(0.1);
    }

    const delRes = http.del(`${url}/apis/elemental.cattle.io/v1beta1/namespaces/${ns}/machineregistrations/${resourceName}`, null, params);
    check(delRes, {
        'delete status is 200 or 202 or 204': (r) => [200, 202, 204].includes(r.status),
    });
}
