/*
Copyright © 2023 SUSE LLC

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

package install

import (
	"strings"

	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
)

// Install or upgrade Rancher Manager
func DeployRancherManager(hostname, channel, version, ca, proxy string) {
	channelName := "rancher-" + channel

	// Add Helm repository
	err := kubectl.RunHelmBinaryWithCustomErr("repo", "add", channelName,
		"https://releases.rancher.com/server-charts/"+channel,
	)
	Expect(err).To(Not(HaveOccurred()))

	err = kubectl.RunHelmBinaryWithCustomErr("repo", "update")
	Expect(err).To(Not(HaveOccurred()))

	// Set flags for Rancher Manager installation
	flags := []string{
		"upgrade", "--install", "rancher", channelName + "/rancher",
		"--namespace", "cattle-system",
		"--create-namespace",
		"--set", "hostname=" + hostname,
		"--set", "extraEnv[0].name=CATTLE_SERVER_URL",
		"--set", "extraEnv[0].value=https://" + hostname,
		"--set", "extraEnv[1].name=CATTLE_BOOTSTRAP_PASSWORD",
		"--set", "extraEnv[1].value=rancherpassword",
		"--set", "replicas=1",
		"--set", "global.cattle.psp.enabled=false",
	}

	// Set specified version if needed
	if version != "" && version != "latest" {
		if version == "devel" {
			flags = append(flags,
				"--devel",
				"--set", "rancherImageTag=v2.7-head",
			)
		} else if strings.Contains(version, "-rc") {
			flags = append(flags,
				"--devel",
				"--version", version,
			)
		} else {
			flags = append(flags, "--version", version)
		}
	}

	// For Private CA
	if ca == "private" {
		flags = append(flags,
			"--set", "ingress.tls.source=secret",
			"--set", "privateCA=true",
		)
	}

	// Use Rancher Manager behind proxy
	if proxy == "rancher" {
		flags = append(flags,
			"--set", "proxy=http://172.17.0.1:3128",
			"--set", "noProxy=127.0.0.0/8\\,10.0.0.0/8\\,cattle-system.svc\\,172.16.0.0/12\\,192.168.0.0/16\\,.svc\\,.cluster.local",
		)
	}

	err = kubectl.RunHelmBinaryWithCustomErr(flags...)
	Expect(err).To(Not(HaveOccurred()))
}
