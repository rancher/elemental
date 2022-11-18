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
		sut.Command("elemental-support")
		out, _ := sut.Command("find /root -name `hostname`*.tar.gz -print")
		sut.GatherLog(out)
		hostname, _ := sut.Command("hostname")
		filename, _ := exec.Command("find", ".", "--name", fmt.Sprintf("%s*.tar.gz", hostname), "-print").CombinedOutput()
		os.Rename(string(filename), strings.Replace(string(filename), ":", "-", -1))
	})
})
