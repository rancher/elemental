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
			s.GatherAllLogs()
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
		})
	})
})
