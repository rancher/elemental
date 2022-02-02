package smoke_test

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/os2/tests/sut"
	"os"
	"time"
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

			Eventually(func() string {
				out, _ := s.Command("k3s kubectl get nodes -o wide")
				return out
			}, 230*time.Second, 1*time.Second).Should(ContainSubstring("Ready"))
		})
	})

	Context("ros-operator", func() {
		It("installs and create a machine registration resource", func() {
			chart := os.Getenv("ROS_CHART")
			if chart == "" {
				Skip("No chart provided, skipping tests")
			}
			err := s.SendFile(chart, "/usr/local/ros.tgz", "0770")
			Expect(err).ToNot(HaveOccurred())

			err = s.SendFile("../assets/machineregistration.yaml", "/usr/local/machine.yaml", "0770")
			Expect(err).ToNot(HaveOccurred())

			err = s.SendFile("../assets/external_charts.yaml", "/usr/local/charts.yaml", "0770")
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() string {
				out, _ := s.Command("k3s kubectl get pods --all-namespaces")
				return out
			}, 960*time.Second, 1*time.Second).Should(ContainSubstring("Running"))

			s.Command(`k3s kubectl apply -f /usr/local/charts.yaml`)

			Eventually(func() string {
				out, _ := s.Command("k3s kubectl get pods --all-namespaces")
				return out
			}, 960*time.Second, 1*time.Second).Should(ContainSubstring("cattle-cluster-agent-"))

			out, err := s.Command("KUBECONFIG=/etc/rancher/k3s/k3s.yaml helm -n cattle-rancheros-operator-system install --create-namespace rancheros-operator /usr/local/ros.tgz")
			Expect(out).To(ContainSubstring("STATUS: DEPLOYED"))
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() string {
				out, _ := s.Command("k3s kubectl get pods --all-namespaces")
				return out
			}, 960*time.Second, 1*time.Second).Should(ContainSubstring("rancheros-operator-"))

			s.Command("k3s kubectl apply -f /usr/local/machine.yaml")

			Eventually(func() string {
				out, _ := s.Command("k3s kubectl get machineregistration -n fleet-default machine-registration -o json | jq '.status.registrationURL' -r")
				return out
			}, 960*time.Second, 1*time.Second).Should(ContainSubstring("http"))
		})
	})
})
