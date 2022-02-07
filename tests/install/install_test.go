package smoke_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/os2/tests/sut"
)

var _ = Describe("os2 install tests", func() {
	var s *sut.SUT
	BeforeEach(func() {
		s = sut.NewSUT()
		s.EventuallyConnects()
	})

	Context("First boot", func() {
		It("can install", func() {
			// err := s.SendFile("../assets/cloud_init.yaml", "/tmp/install.yaml", "0770")
			// Expect(err).ToNot(HaveOccurred())

			out, err := s.Command("cos-installer /dev/sda && sync")
			Expect(out).To(And(
				ContainSubstring("Deployment done, now you might want to reboot"),
			))
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
