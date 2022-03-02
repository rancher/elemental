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

package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"regexp"
	"strings"

	v1 "github.com/rancher-sandbox/os2/pkg/apis/rancheros.cattle.io/v1"
	"github.com/rancher/fleet/pkg/apis/fleet.cattle.io/v1alpha1"
	values "github.com/rancher/wrangler/pkg/data"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const defaultName = "m-${System Information/Manufacturer}-${System Information/Product Name}-${System Information/Serial Number}-"

var (
	sanitize   = regexp.MustCompile("[^0-9a-zA-Z]")
	doubleDash = regexp.MustCompile("--+")
	start      = regexp.MustCompile("^[a-zA-Z]")
)

func (i *InventoryServer) register(resp http.ResponseWriter, req *http.Request) {
	machineInventory, machineRegister, data, err := i.buildResponse(req)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusUnauthorized)
		return
	}

	machine, writer, err := i.authMachine(resp, req, machineInventory.Namespace)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusUnauthorized)
		return
	}
	if writer == nil {
		if err := i.sampleConfig(machineRegister, resp); err != nil {
			http.Error(resp, "authorization required", http.StatusUnauthorized)
		}
		return
	}
	defer writer.Close()

	if err := i.saveMachine(machine.Spec.TPMHash, machineInventory); err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	_, _ = writer.Write(data)
}

func (i *InventoryServer) sampleConfig(machineRegistration *v1.MachineRegistration, writer io.Writer) error {
	_, err := writer.Write([]byte("#cloud-config\n"))
	if err != nil {
		return err
	}

	installSection := map[string]interface{}{
		"registrationURL": machineRegistration.Status.RegistrationURL,
	}
	certs := i.cacert()
	if certs != "" {
		installSection["registrationCaCert"] = certs
	}

	return yaml.NewEncoder(writer).Encode(map[string]interface{}{
		"rancheros": map[string]interface{}{
			"install": installSection,
		},
	})
}

func (i *InventoryServer) saveMachine(tpmHash string, machineInventory *v1.MachineInventory) error {
	machines, err := i.machineCache.GetByIndex(tpmHashIndex, tpmHash)
	if err != nil || len(machines) > 0 {
		return err
	}

	machineInventory.Spec.TPMHash = tpmHash
	_, err = i.machineClient.Create(machineInventory)
	return err
}

func buildName(data map[string]interface{}, name string) string {
	str := name
	result := &strings.Builder{}
	for {
		i := strings.Index(str, "${")
		if i == -1 {
			result.WriteString(str)
			break
		}
		j := strings.Index(str[i:], "}")
		if j == -1 {
			result.WriteString(str)
			break
		}

		result.WriteString(str[:i])
		obj := values.GetValueN(data, strings.Split(str[i+2:j+i], "/")...)
		if str, ok := obj.(string); ok {
			result.WriteString(str)
		}
		str = str[j+i+1:]
	}

	resultStr := sanitize.ReplaceAllString(result.String(), "-")
	resultStr = doubleDash.ReplaceAllString(resultStr, "-")
	if !start.MatchString(resultStr) {
		resultStr = "m" + resultStr
	}
	if len(resultStr) > 58 {
		resultStr = resultStr[:58]
	}
	return strings.ToLower(resultStr)
}

func (i *InventoryServer) buildResponse(req *http.Request) (*v1.MachineInventory, *v1.MachineRegistration, []byte, error) {
	token := path.Base(req.URL.Path)

	smbios, err := getSMBios(req)
	if err != nil {
		return nil, nil, nil, err
	}

	machineRegisters, err := i.machineRegistrationCache.GetByIndex(registrationTokenIndex, token)
	if apierrors.IsNotFound(err) || len(machineRegisters) != 1 {
		if len(machineRegisters) > 1 {
			logrus.Errorf("Multiple MachineRegistrations have the same token %s: %v", token, machineRegisters)
		}
		if err == nil && len(machineRegisters) == 0 {
			err = fmt.Errorf("MachineRegistration does not exist")
		}
		return nil, nil, nil, err
	}
	machineRegister := machineRegisters[0]

	name := machineRegister.Spec.MachineName
	if name == "" {
		name = defaultName
	}

	installConfig := map[string]interface{}{}
	if machineRegister.Spec.CloudConfig != nil && len(machineRegister.Spec.CloudConfig.Data) > 0 {
		installConfig = values.MergeMapsConcatSlice(installConfig, machineRegister.Spec.CloudConfig.Data)
	}

	serverURL, err := i.serverURL()
	if err != nil {
		return nil, nil, nil, err
	}
	values.PutValue(installConfig, serverURL, "rancherd", "server")
	values.PutValue(installConfig, "tpm://", "rancherd", "token")
	values.PutValue(installConfig, true, "rancheros", "install", "automatic")

	data, err := json.Marshal(installConfig)
	if err != nil {
		return nil, nil, nil, err
	}

	return &v1.MachineInventory{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: buildName(smbios, name),
			Namespace:    machineRegister.Namespace,
			Labels:       machineRegisters[0].Spec.MachineInventoryLabels,
			Annotations:  machineRegisters[0].Spec.MachineInventoryAnnotations,
		},
		Spec: v1.MachineInventorySpec{
			SMBIOS: &v1alpha1.GenericMap{
				Data: smbios,
			},
		},
	}, machineRegister, data, nil
}

func getSMBios(req *http.Request) (map[string]interface{}, error) {
	smbios := req.Header.Get("X-Cattle-Smbios")
	if smbios == "" {
		return nil, nil
	}
	smbiosData, err := base64.StdEncoding.DecodeString(smbios)
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{}
	return data, json.Unmarshal(smbiosData, &data)
}
