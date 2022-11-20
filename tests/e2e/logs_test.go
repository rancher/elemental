package e2e_test

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/rancher-sandbox/ele-testhelpers/vm"
	"os"
	"os/exec"
	"strings"
)

var _ = Describe("E2E - Getting logs node", Label("logs"), func() {
	var (
		sut *vm.SUT
	)
	BeforeEach(func() {
		sut = &vm.SUT{
			Host:     "192.168.122.2:22",
			Username: "root",
			Password: "root",
		}
	})
	It("gets the downstream logs", func() {
		// Get a more recent support binary
		sut.Command("curl -L https://github.com/Itxaka/elemental-operator/releases/download/v100.0.0/elemental-support_100.0.0_linux_amd64 -o /tmp/elemental-support")
		sut.Command("chmod +x /tmp/elemental-support")
		sut.Command("/tmp/elemental-support")
		out, _ := sut.Command("find /root -name `hostname`*.tar.gz -print")
		out = strings.TrimSpace(out)
		sut.GatherLog(out)
		// Get logs from local node?
		outcmd, err := exec.Command("kubectl", "get", "bundles", "--all-namespaces").CombinedOutput()
		if err != nil {
			fmt.Fprint(GinkgoWriter, err)
		}
		os.WriteFile("logs/bundles-resource.log", outcmd, os.ModePerm)
		outcmd, err = exec.Command("kubectl", "describe", "bundles", "--all-namespaces").CombinedOutput()
		if err != nil {
			fmt.Fprint(GinkgoWriter, err)
		}
		os.WriteFile("logs/bundles-describe.log", outcmd, os.ModePerm)
		outcmd, err = exec.Command("kubectl", "logs", "-n", "cattle-fleet-system", "-l", "app=fleet-controller").CombinedOutput()
		os.WriteFile("logs/fleet-controller.log", outcmd, os.ModePerm)
		outcmd, err = exec.Command("kubectl", "logs", "-n", "cattle-fleet-local-system", "-l", "app=fleet-agent").CombinedOutput()
		if err != nil {
			fmt.Fprint(GinkgoWriter, err)
		}
		os.WriteFile("logs/fleet-agent.log", outcmd, os.ModePerm)
		outcmd, err = exec.Command("kubectl", "describe", "nodes", "-o").CombinedOutput()
		if err != nil {
			fmt.Fprint(GinkgoWriter, err)
		}
		os.WriteFile("logs/node.log", outcmd, os.ModePerm)
	})
})
