/*
Copyright © 2022 - 2024 SUSE LLC

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

package e2e_test

import (
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
)

func checkRC(err error) {
	if err != nil {
		GinkgoWriter.Printf("%s\n", err)
	}
}

var _ = Describe("E2E - Getting logs node", Label("logs"), func() {
	type binary struct {
		Url  string
		Name string
	}

	type getResourceLog struct {
		Name string
		Verb []string
	}

	It("Get the upstream cluster logs", func() {
		// Report to Qase
		testCaseID = 69

		By("Downloading and executing tools to generate logs", func() {
			elemental := binary{
				elementalSupport,
				"elemental-support",
			}

			logCollector := binary{
				rancherLogCollector,
				"rancher2_logs_collector.sh",
			}

			_ = os.Mkdir("logs", 0755)
			_ = os.Chdir("logs")
			myDir, _ := os.Getwd()

			for _, b := range []binary{elemental, logCollector} {
				Eventually(func() error {
					return exec.Command("curl", "-L", b.Url, "-o", b.Name).Run()
				}, tools.SetTimeout(1*time.Minute), 5*time.Second).Should(BeNil())

				err := exec.Command("chmod", "+x", b.Name).Run()
				checkRC(err)
				if b.Name == "elemental-support" {
					err := exec.Command(myDir + "/" + b.Name).Run()
					checkRC(err)
				} else {
					err := exec.Command("sudo", myDir+"/"+b.Name, "-d", "../logs").Run()
					checkRC(err)
				}
			}
		})

		By("Collecting additionals logs with kubectl commands", func() {
			Bundles := getResourceLog{
				"bundles",
				[]string{"get", "describe"},
			}

			var getResources []getResourceLog = []getResourceLog{Bundles}
			for _, r := range getResources {
				for _, v := range r.Verb {
					outcmd, err := kubectl.RunWithoutErr(v, r.Name, "--all-namespaces")
					checkRC(err)
					err = os.WriteFile(r.Name+"-"+v+".log", []byte(outcmd), os.ModePerm)
					checkRC(err)
				}
			}
		})

		if proxy == "elemental" || proxy == "rancher" {
			By("Collecting proxy log and make sure traffic went through it", func() {
				out, err := exec.Command("docker", "exec", "squid_proxy", "cat", "/var/log/squid/access.log").CombinedOutput()
				checkRC(err)
				err = os.WriteFile("squid.log", []byte(out), os.ModePerm)
				checkRC(err)
				Expect(out).Should(MatchRegexp("TCP_TUNNEL/200.*CONNECT.*rancher.io"))
			})
		}
	})
})
