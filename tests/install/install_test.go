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
			out, err := s.Command("ELEMENTAL_TARGET=/dev/sda elemental install && sync")
			Expect(out).To(And(
				ContainSubstring("Unmounting disk partitions"),
				ContainSubstring("Mounting disk partitions"),
				ContainSubstring("Copying Passive image..."),
			))
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
