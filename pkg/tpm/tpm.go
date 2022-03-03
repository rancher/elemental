/*
Copyright © 2022 SUSE LLC

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

package tpm

import (
	v1 "github.com/rancher-sandbox/os2/pkg/apis/rancheros.cattle.io/v1"
	"github.com/rancher-sandbox/os2/pkg/clients"
	roscontrollers "github.com/rancher-sandbox/os2/pkg/generated/controllers/rancheros.cattle.io/v1"
	corecontrollers "github.com/rancher/wrangler/pkg/generated/controllers/core/v1"
)

const (
	machineByHash = "machineByHash"
	tpmCACert     = "tpm-ca"
)

type AuthServer struct {
	machineCache  roscontrollers.MachineInventoryCache
	machineClient roscontrollers.MachineInventoryClient
	secretCache   corecontrollers.SecretCache
}

func New(clients *clients.Clients) *AuthServer {
	a := &AuthServer{
		machineCache:  clients.OS.MachineInventory().Cache(),
		machineClient: clients.OS.MachineInventory(),
		secretCache:   clients.Core.Secret().Cache(),
	}

	a.machineCache.AddIndexer(machineByHash, func(obj *v1.MachineInventory) ([]string, error) {
		if obj.Spec.TPMHash == "" {
			return nil, nil
		}
		return []string{obj.Spec.TPMHash}, nil
	})

	return a
}
