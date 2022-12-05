package e2e_test

import (
	"fmt"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("E2E - Getting logs node", Label("logs"), func() {
	It("Get the upstream cluster logs", func() {
		By("Downloading tools to generate logs", func() {
			_ = os.Mkdir("logs", 0755)
			_ = os.Chdir("logs")
			err := exec.Command("curl", "-L", "https://github.com/rancher/elemental-operator/releases/download/v1.0.0/elemental-support_1.0.0_linux_amd64", "-o", "elemental-support").Run()
			Expect(err).To(Not(HaveOccurred()))
			err = exec.Command("chmod", "+x", "elemental-support").Run()
			Expect(err).To(Not(HaveOccurred()))
			err = exec.Command("curl", "-L", "https://raw.githubusercontent.com/rancherlabs/support-tools/master/collection/rancher/v2.x/logs-collector/rancher2_logs_collector.sh", "-o", "rancher2_logs_collector.sh").Run()
			Expect(err).To(Not(HaveOccurred()))
			err = exec.Command("chmod", "+x", "rancher2_logs_collector.sh").Run()
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Executing binaries to collect logs", func() {
			err := exec.Command("./elemental-support").Run()
			Expect(err).To(Not(HaveOccurred()))
			err = exec.Command("sudo", "./rancher2_logs_collector.sh", "-d", "../logs").Run()
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Collecting additionals logs with kubectl commands", func() {
			outcmd, err := exec.Command("kubectl", "get", "bundles", "--all-namespaces").CombinedOutput()
			if err != nil {
				fmt.Fprint(GinkgoWriter, err)
			}
			os.WriteFile("bundles-resource.log", outcmd, os.ModePerm)
			outcmd, err = exec.Command("kubectl", "describe", "bundles", "--all-namespaces").CombinedOutput()
			if err != nil {
				fmt.Fprint(GinkgoWriter, err)
			}
			os.WriteFile("bundles-describe.log", outcmd, os.ModePerm)
		})
	})
})
