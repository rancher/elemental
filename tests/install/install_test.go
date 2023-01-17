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
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	sut "github.com/rancher-sandbox/ele-testhelpers/vm"
)

var _ = Describe("Elemental Installation tests", func() {
	var s *sut.SUT
	BeforeEach(func() {
		s = sut.NewSUT()
		s.EventuallyConnects()
	})

	// This is used to setup the machine that will run other tests
	Context("From ISO", Label("iso"), func() {
		It("can install", func() {
			err := s.SendFile("../assets/cloud_init.yaml", "/tmp/cloud_init.yaml", "0640")
			Expect(err).ToNot(HaveOccurred())

			out, err := s.Command(s.ElementalCmd("install", "/dev/sda", "--cloud-init", "/tmp/cloud_init.yaml"))
			fmt.Printf(out)
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(And(
				ContainSubstring("Unmounting disk partitions"),
				ContainSubstring("Mounting disk partitions"),
				ContainSubstring("Finished copying /run/rootfsbase into /run/cos/workingtree"),
				ContainSubstring("Setting default grub entry to Elemental"),
			), out)
		})
	})

	Context("From container", Label("container"), func() {
		It("can install", func() {
			containerImage := os.Getenv("CONTAINER_IMAGE")
			if containerImage == "" {
				Skip("No CONTAINER_IMAGE defined")
			}

			err := s.SendFile("../assets/cloud_init.yaml", "/tmp/cloud_init.yaml", "0640")
			Expect(err).ToNot(HaveOccurred())

			out, err := s.Command(s.ElementalCmd("install", "/dev/sda", "--cloud-init", "/tmp/cloud_init.yaml", "--system.uri", "docker:"+containerImage))
			fmt.Printf(out)
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(And(
				ContainSubstring("Unmounting disk partitions"),
				ContainSubstring("Mounting disk partitions"),
				ContainSubstring("Finished copying "+containerImage+" into /run/cos/workingtree"),
				ContainSubstring("Unpacking a container image: "+containerImage),
				ContainSubstring("Setting default grub entry to Elemental"),
			), out)
		})
	})

	It("has customization applied", Label("iso", "container"), func() {
		// Reboot to check the installed OS
		s.Reboot()

		By("Checking we booted from the installed OS")
		s.AssertBootedFrom(sut.Active)

		By("Checking config file was run")
		_, err := s.Command("stat /oem/90_custom.yaml")
		Expect(err).ToNot(HaveOccurred())

		out, err := s.Command("hostname")
		Expect(err).ToNot(HaveOccurred())
		Expect(out).To(ContainSubstring("my-own-name"))
	})
})
