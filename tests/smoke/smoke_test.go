/*
Copyright © 2022 - 2023 SUSE LLC

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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	sut "github.com/rancher-sandbox/ele-testhelpers/vm"
)

func systemdUnitIsStarted(s string, st *sut.SUT) {
	out, err := st.Command(fmt.Sprintf("systemctl status %s", s))
	Expect(err).To(Not(HaveOccurred()))

	Expect(out).To(And(
		ContainSubstring(fmt.Sprintf("%s.service; enabled", s)),
		ContainSubstring("status=0/SUCCESS"),
	))
}

var _ = Describe("Elemental Smoke tests", func() {
	var s *sut.SUT
	BeforeEach(func() {
		s = sut.NewSUT()
		s.EventuallyConnects()
	})

	AfterEach(func() {
		if CurrentSpecReport().Failed() {
			cmds := []string{"pods", "events", "helmcharts", "ingress"}
			for _, c := range cmds {
				_, err := s.Command("k3s kubectl get " + c + " -A -o json > /tmp/" + c + ".json")
				Expect(err).To(Not(HaveOccurred()))
			}
			_, err := s.Command("df -h > /tmp/disk")
			Expect(err).To(Not(HaveOccurred()))
			_, err = s.Command("mount > /tmp/mounts")
			Expect(err).To(Not(HaveOccurred()))
			_, err = s.Command("blkid > /tmp/blkid")
			Expect(err).To(Not(HaveOccurred()))

			s.GatherAllLogs()
		}
	})

	Context("First boot", func() {
		for _, unit := range []string{"elemental-setup-initramfs", "elemental-setup-network", "elemental-setup-rootfs", "elemental-setup-boot", "elemental-setup-fs"} {
			It(fmt.Sprintf("starts successfully %s on boot", unit), func() {
				systemdUnitIsStarted(unit, s)
			})
		}

		It("has default mounts", func() {
			out, err := s.Command("mount")
			Expect(err).To(Not(HaveOccurred()))
			Expect(out).To(And(
				ContainSubstring("/var/lib/rancher"),
				ContainSubstring("/etc/ssh"),
				ContainSubstring("/etc/rancher"),
			))
		})

		It("has default cmdline", func() {
			out, err := s.Command("cat /proc/cmdline")
			Expect(err).To(Not(HaveOccurred()))
			Expect(out).To(And(
				ContainSubstring("rd.neednet=0"),
			))
		})

		// Added user via cloud-init is functional
		It("has the user added via cloud-init", func() {
			out, err := s.Command("su - vagrant -c 'id -un'")
			Expect(err).To(Not(HaveOccurred()))
			Expect(out).To(Equal("vagrant\n"))

			out, err = s.Command("cat /run/vagrant/.ssh/authorized_keys")
			Expect(err).To(Not(HaveOccurred()))
			Expect(out).To(ContainSubstring("ssh-rsa"))
			out, err = s.Command("sudo cat /root/.ssh/authorized_keys")
			Expect(err).To(Not(HaveOccurred()))
			Expect(out).To(ContainSubstring("ssh-rsa"))
		})
	})
})
