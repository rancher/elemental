package e2e_test

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher/elemental/tests/e2e/helpers/misc"
)

func checkRC(err error) {
	if err != nil {
		GinkgoWriter.Printf("%s\n", err)
	}
}

var _ = Describe("E2E - Getting logs node", Label("logs"), func() {
	type Binary struct {
		Url  string
		Name string
	}

	type getResourceLog struct {
		Name string
		Verb []string
	}
	var elementalBinary string = fmt.Sprintf("https://github.com/rancher/elemental-operator/releases/download/v%s/elemental-support_%s_linux_amd64", elementalSupportVersion, elementalSupportVersion)

	It("Get the upstream cluster logs", func() {
		By("Downloading and executing tools to generate logs", func() {
			elemental := Binary{
				elementalBinary,
				"elemental-support",
			}

			logCollector := Binary{
				"https://raw.githubusercontent.com/rancherlabs/support-tools/master/collection/rancher/v2.x/logs-collector/rancher2_logs_collector.sh",
				"rancher2_logs_collector.sh",
			}

			_ = os.Mkdir("logs", 0755)
			_ = os.Chdir("logs")
			myDir, _ := os.Getwd()

			var binaries []Binary = []Binary{elemental, logCollector}
			for _, b := range binaries {
				Eventually(func() error {
					err := exec.Command("curl", "-L", b.Url, "-o", b.Name).Run()
					return err
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
					outcmd, err := exec.Command("kubectl", v, r.Name, "--all-namespaces").CombinedOutput()
					checkRC(err)
					err = os.WriteFile(r.Name+"-"+v+".log", outcmd, os.ModePerm)
					checkRC(err)
				}
			}
		})
	})
})
