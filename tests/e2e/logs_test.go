package e2e_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
	"github.com/rancher-sandbox/ele-testhelpers/vm"
)

var _ = Describe("E2E - Getting logs node", Label("logs"), func() {
	var (
		sut *vm.SUT
	)
	BeforeEach(func() {
		hostData, err := tools.GetHostNetConfig(".*name=\""+vmName+"\".*", netDefaultFileName)
		Expect(err).To(Not(HaveOccurred()))

		sut = &vm.SUT{
			Host:      string(hostData.IP) + ":22",
			Username:  userName,
			Password:  userPassword,
			MachineID: "test",
		}
	})
	It("gets the downstream logs", func() {
		sut.Command("elemental-support")
		out, _ := sut.Command("find /root -name `hostname`*.tar.gz -print")
		sut.GatherLog(out)
	})
})
