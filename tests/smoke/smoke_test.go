package smoke_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/os2/tests/sut"
)

func systemdUnitIsStarted(s string, st *sut.SUT) {
	out, _ := st.Command(fmt.Sprintf("systemctl status %s", s))

	Expect(out).To(And(
		ContainSubstring(fmt.Sprintf("%s.service; enabled", s)),
		ContainSubstring("status=0/SUCCESS"),
	))
}

var _ = Describe("os2 Smoke tests", func() {
	var s *sut.SUT
	BeforeEach(func() {
		s = sut.NewSUT()
		s.EventuallyConnects()
	})

	AfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			s.Command("k3s kubectl get pods -A -o json > /tmp/pods.json")
			s.Command("k3s kubectl get events -A -o json > /tmp/events.json")
			s.Command("k3s kubectl get helmcharts -A -o json > /tmp/helm.json")
			s.Command("k3s kubectl get ingress -A -o json > /tmp/ingress.json")
			s.Command("df -h > /tmp/disk")
			s.Command("mount > /tmp/mounts")
			s.Command("blkid > /tmp/blkid")

			s.GatherAllLogs(
				[]string{
					"ros-installer",
					"cos-setup-boot",
					"cos-setup-network",
					"rancherd",
					"k3s",
				},
				[]string{
					"/tmp/pods.json",
					"/tmp/disk",
					"/tmp/mounts",
					"/tmp/blkid",
					"/tmp/events.json",
					"/tmp/helm.json",
					"/tmp/ingress.json",
				})
		}
	})

	Context("First boot", func() {
		for _, unit := range []string{"ros-installer", "rancherd", "cos-setup-rootfs"} {
			It(fmt.Sprintf("starts successfully %s on boot", unit), func() {
				systemdUnitIsStarted(unit, s)
			})
		}

		It("has default mounts", func() {
			out, _ := s.Command("mount")
			Expect(out).To(And(
				ContainSubstring("/var/lib/rancher"),
				ContainSubstring("/etc/ssh"),
				ContainSubstring("/etc/rancher"),
			))
		})

		It("has default cmdline", func() {
			out, _ := s.Command("cat /proc/cmdline")
			Expect(out).To(And(
				ContainSubstring("selinux=1"),
			))
		})

		// This test is flaky as it relies on dmesg output
		PIt("correctly starts cos services", func() {
			out, _ := s.Command("dmesg | grep cos")
			Expect(out).To(And(
				ContainSubstring("cos-immutable-rootfs.service: Succeeded"),
				ContainSubstring("cos-setup-rootfs.service: Succeeded"),
			))
		})

		// Added user via cloud-init is functional
		It("has the user added via cloud-init", func() {
			out, _ := s.Command(`su - vagrant -c 'id -un'`)
			Expect(out).To(Equal("vagrant\n"))

			out, _ = s.Command(`cat /run/vagrant/.ssh/authorized_keys`)
			Expect(out).To(ContainSubstring("ssh-rsa"))
			out, _ = s.Command(`sudo cat /root/.ssh/authorized_keys`)
			Expect(out).To(ContainSubstring("ssh-rsa"))
		})
	})

	Context("rancherd", func() {
		It("starts a single-node cluster", func() {
			err := s.SendFile("../assets/rancherd.yaml", "/oem/99_custom.yaml", "0770")
			Expect(err).ToNot(HaveOccurred())
			s.Command("systemctl restart --no-block rancherd")
			Eventually(func() string {
				out, _ := s.Command("ps aux")
				return out
			}, 230*time.Second, 1*time.Second).Should(ContainSubstring("k3s server"))

			systemdUnitIsStarted("k3s", s)

			Eventually(func() string {
				out, _ := s.Command("k3s kubectl get nodes -o wide")
				return out
			}, 230*time.Second, 1*time.Second).Should(ContainSubstring("Ready"))
		})
	})
})
