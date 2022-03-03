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

package clients

import (
	"context"

	"github.com/rancher-sandbox/os2/pkg/generated/controllers/fleet.cattle.io"
	fleetcontrollers "github.com/rancher-sandbox/os2/pkg/generated/controllers/fleet.cattle.io/v1alpha1"
	"github.com/rancher-sandbox/os2/pkg/generated/controllers/management.cattle.io"
	ranchercontrollers "github.com/rancher-sandbox/os2/pkg/generated/controllers/management.cattle.io/v3"
	"github.com/rancher-sandbox/os2/pkg/generated/controllers/provisioning.cattle.io"
	provcontrollers "github.com/rancher-sandbox/os2/pkg/generated/controllers/provisioning.cattle.io/v1"
	"github.com/rancher-sandbox/os2/pkg/generated/controllers/rancheros.cattle.io"
	oscontrollers "github.com/rancher-sandbox/os2/pkg/generated/controllers/rancheros.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/clients"
	"github.com/rancher/wrangler/pkg/generic"
	"k8s.io/client-go/rest"
)

const (
	// SystemNamespace Default namespace for rancher system objects
	SystemNamespace = "cattle-system"
)

type Clients struct {
	*clients.Clients
	Fleet        fleetcontrollers.Interface
	OS           oscontrollers.Interface
	Rancher      ranchercontrollers.Interface
	Provisioning provcontrollers.Interface
}

func NewFromConfig(restConfig *rest.Config) (*Clients, error) {
	c, err := clients.NewFromConfig(restConfig, nil)
	if err != nil {
		return nil, err
	}

	opts := &generic.FactoryOptions{
		SharedControllerFactory: c.SharedControllerFactory,
	}
	return &Clients{
		Clients:      c,
		Fleet:        fleet.NewFactoryFromConfigWithOptionsOrDie(restConfig, opts).Fleet().V1alpha1(),
		OS:           rancheros.NewFactoryFromConfigWithOptionsOrDie(restConfig, opts).Rancheros().V1(),
		Rancher:      management.NewFactoryFromConfigWithOptionsOrDie(restConfig, opts).Management().V3(),
		Provisioning: provisioning.NewFactoryFromConfigWithOptionsOrDie(restConfig, opts).Provisioning().V1(),
	}, nil
}

func (c *Clients) Start(ctx context.Context) error {
	return c.SharedControllerFactory.Start(ctx, 5)
}
