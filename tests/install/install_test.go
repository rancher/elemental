package smoke_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/os2/tests/sut"
	"os"
)

var _ = Describe("os2 installation", Label("setup"), func() {
	var s *sut.SUT
	BeforeEach(func() {
		s = sut.NewSUT()
		s.EventuallyConnects()
	})

	// This is used to setup the machine that will run other tests
	Context("First boot", func() {
		It("can install", func() {

			s.WriteInlineFile(`
rancheros:
 install:
   device: /dev/sda
   automatic: true
`, "/oem/userdata")

			out, err := s.Command("ros-installer --automatic --no-reboot-automatic && sync")
			Expect(out).To(And(
				ContainSubstring("Unmounting disk partitions"),
				ContainSubstring("Mounting disk partitions"),
				ContainSubstring("Finished copying COS_PASSIVE"),
				ContainSubstring("Grub install to device /dev/sda complete"),
			))
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

var _ = Describe("os2 setup tests", func() {
	var s *sut.SUT
	BeforeEach(func() {
		s = sut.NewSUT()
		s.EventuallyConnects()
	})

	Context("First boot", func() {
		It("can install from a container image", func() {

			containerImage := os.Getenv("CONTAINER_IMAGE")
			if containerImage == "" {
				Skip("No CONTAINER_IMAGE defined")
			}

			s.WriteInlineFile(`
rancheros:
 install:	
   device: /dev/sda
   containerImage: `+containerImage+`
   automatic: true
`, "/oem/userdata")

			out, err := s.Command("ros-installer --automatic --no-reboot-automatic && sync")
			Expect(out).To(And(
				ContainSubstring("Unmounting disk partitions"),
				ContainSubstring("Mounting disk partitions"),
				ContainSubstring("Finished copying COS_PASSIVE"),
				ContainSubstring("Unpacking docker image: "+containerImage),
				ContainSubstring("Grub install to device /dev/sda complete"),
			), out)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
