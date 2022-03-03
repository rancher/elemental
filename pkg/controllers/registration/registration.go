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

package registration

import (
	"context"
	"fmt"

	v1 "github.com/rancher-sandbox/os2/pkg/apis/rancheros.cattle.io/v1"
	"github.com/rancher-sandbox/os2/pkg/clients"
	ranchercontrollers "github.com/rancher-sandbox/os2/pkg/generated/controllers/management.cattle.io/v3"
	roscontrollers "github.com/rancher-sandbox/os2/pkg/generated/controllers/rancheros.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/randomtoken"
)

type handler struct {
	settingsCache ranchercontrollers.SettingCache
}

func Register(ctx context.Context, clients *clients.Clients) {
	h := handler{
		settingsCache: clients.Rancher.Setting().Cache(),
	}
	roscontrollers.RegisterMachineRegistrationStatusHandler(ctx, clients.OS.MachineRegistration(), "Ready", "machine-registration",
		h.OnChange)
}

func (h *handler) OnChange(obj *v1.MachineRegistration, status v1.MachineRegistrationStatus) (v1.MachineRegistrationStatus, error) {
	serverURL, err := h.serverURL()
	if err != nil {
		return status, err
	}

	if status.RegistrationToken == "" {
		status.RegistrationToken, err = randomtoken.Generate()
		if err != nil {
			return status, err
		}
	}

	status.RegistrationURL = serverURL + "/v1-rancheros/registration/" + status.RegistrationToken
	return status, nil
}

func (h *handler) serverURL() (string, error) {
	setting, err := h.settingsCache.Get("server-url")
	if err != nil {
		return "", err
	}
	if setting.Value == "" {
		return "", fmt.Errorf("server-url is not set")
	}
	return setting.Value, nil
}
