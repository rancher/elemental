/*
Copyright © 2022 SUSE LLC

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

package smoke_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/os2/tests/sut"
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
				ContainSubstring("Unpacking a container image: "+containerImage),
				ContainSubstring("Grub install to device /dev/sda complete"),
			), out)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
