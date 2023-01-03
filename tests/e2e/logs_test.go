package e2e_test

import (
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher/elemental/tests/e2e/helpers/misc"
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
				}, misc.SetTimeout(1*time.Minute), 5*time.Second).Should(BeNil())

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
					outcmd, err := kubectl.Run(v, r.Name, "--all-namespaces")
					checkRC(err)
					err = os.WriteFile(r.Name+"-"+v+".log", []byte(outcmd), os.ModePerm)
					checkRC(err)
				}
			}
		})

		if proxy != "" {
			By("Collecting proxy log and make sure traffic went through it", func() {
				out, err := exec.Command("docker", "exec", "squid_proxy", "cat", "/var/log/squid/access.log").CombinedOutput()
				checkRC(err)
				err = os.WriteFile("squid.log", []byte(out), os.ModePerm)
				checkRC(err)
				Expect(out).Should(MatchRegexp("TCP_TUNNEL/200.*CONNECT git.rancher.io"))
			})
		}
	})
})
