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

package main

import (
	"os"

	fleet "github.com/rancher/fleet/pkg/apis/fleet.cattle.io/v1alpha1"
	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	provv1 "github.com/rancher/rancher/pkg/apis/provisioning.cattle.io/v1"
	controllergen "github.com/rancher/wrangler/pkg/controller-gen"
	"github.com/rancher/wrangler/pkg/controller-gen/args"
)

func main() {
	os.Unsetenv("GOPATH")
	controllergen.Run(args.Options{
		OutputPackage: "github.com/rancher-sandbox/os2/pkg/generated",
		Boilerplate:   "scripts/boilerplate.go.txt",
		Groups: map[string]args.Group{
			"provisioning.cattle.io": {
				Types: []interface{}{
					provv1.Cluster{},
				},
			},
			"management.cattle.io": {
				Types: []interface{}{
					v3.Setting{},
					v3.ClusterRegistrationToken{},
				},
			},
			"fleet.cattle.io": {
				Types: []interface{}{
					fleet.Bundle{},
				},
			},
			"rancheros.cattle.io": {
				Types: []interface{}{
					"./pkg/apis/rancheros.cattle.io/v1",
				},
				GenerateTypes: true,
			},
		},
	})
}
