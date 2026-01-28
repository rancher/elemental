/*
Copyright © 2022 - 2025 SUSE LLC

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

package e2e_test

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/anishathalye/porcupine"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"k8s.io/apimachinery/pkg/util/rand"
)

type KvInput struct {
	Op    string // "read", "write"
	Value string
}

type KvOutput struct {
	Value string
	Found bool
}

var mrModel = porcupine.Model{
	Init: func() interface{} { return "" },
	Step: func(state, input, output interface{}) (bool, interface{}) {
		inp := input.(KvInput)
		out := output.(KvOutput)
		st := state.(string)

		if inp.Op == "read" {
			if !out.Found {
				return st == "", st
			}
			return out.Value == st, st
		} else if inp.Op == "write" {
			return true, inp.Value
		}
		return false, st
	},
}

type Store interface {
	Setup() error
	Teardown()
	Read() (string, bool)
	Write(val string) error
}

type MemoryStore struct {
	mu  sync.Mutex
	val string
	set bool
}

func (m *MemoryStore) Setup() error { return nil }
func (m *MemoryStore) Teardown()    {}
func (m *MemoryStore) Read() (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Simulate failing to find if not set
	return m.val, m.set
}
func (m *MemoryStore) Write(val string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.val = val
	m.set = true
	return nil
}

type KubectlStore struct {
	Namespace string
	Name      string
}

func (k *KubectlStore) Setup() error {
	manifest := fmt.Sprintf(`
apiVersion: elemental.cattle.io/v1beta1
kind: MachineRegistration
metadata:
  name: %s
  namespace: %s
spec:
  machineName: test-node
  machineInventoryLabels:
    test: "true"
`, k.Name, k.Namespace)

	f, _ := os.CreateTemp("", "mr-*.yaml")
	defer os.Remove(f.Name())
	f.WriteString(manifest)
	f.Close()

	return kubectl.Apply(k.Namespace, f.Name())
}

func (k *KubectlStore) Teardown() {
	_, _ = kubectl.Run("delete", "machineregistration", k.Name, "-n", k.Namespace)
}

func (k *KubectlStore) Read() (string, bool) {
	out, err := kubectl.RunWithoutErr("get", "machineregistration", k.Name, "-n", k.Namespace, "-o", "jsonpath={.metadata.annotations.consistency-check}")
	if err == nil {
		return out, true
	}
	return "", false
}

func (k *KubectlStore) Write(val string) error {
	patch := fmt.Sprintf(`{"metadata":{"annotations":{"consistency-check":"%s"}}}`, val)
	_, err := kubectl.RunWithoutErr("patch", "machineregistration", k.Name, "-n", k.Namespace, "--type=merge", "-p", patch)
	return err
}

var _ = Describe("Consistency Tests", Label("consistency"), func() {
	var store Store
	var mrName string
	const namespace = "default"

	BeforeEach(func() {
		mrName = "porcupine-mr-" + rand.String(5)

		// Check for Cluster
		useCluster := false
		if os.Getenv("KUBECONFIG") != "" || os.Getenv("API_URL") != "" {
			// Naive check, can be improved
			useCluster = true
		}

		if useCluster {
			GinkgoWriter.Println("Running in REAL mode (Cluster)")
			store = &KubectlStore{Namespace: namespace, Name: mrName}
		} else {
			GinkgoWriter.Println("Running in SIMULATION mode (Memory)")
			store = &MemoryStore{}
		}

		err := store.Setup()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if store != nil {
			store.Teardown()
		}
	})

	It("should be linearizable with concurrent annotation updates", func() {
		opsCount := 20
		concurrency := 4
		history := make([]porcupine.Operation, 0, opsCount)
		var historyMu sync.Mutex
		var wg sync.WaitGroup

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < opsCount/concurrency; j++ {
					opType := "read"
					val := ""
					if rand.Intn(2) == 0 {
						opType = "write"
						val = fmt.Sprintf("val-%d-%d", id, j)
					}

					start := time.Now()
					var output KvOutput

					if opType == "write" {
						_ = store.Write(val)
					} else {
						v, found := store.Read()
						output.Value = v
						output.Found = found
					}
					end := time.Now()

					historyMu.Lock()
					history = append(history, porcupine.Operation{
						ClientId: id,
						Input:    KvInput{Op: opType, Value: val},
						Call:     start.UnixNano(),
						Output:   output,
						Return:   end.UnixNano(),
					})
					historyMu.Unlock()
				}
			}(i)
		}
		wg.Wait()

		res, info := porcupine.CheckOperationsVerbose(mrModel, history, 0)
		if res == porcupine.Illegal {
			f, _ := os.Create("consistency_failure.html")
			porcupine.Visualize(mrModel, info, f)
			f.Close()
			Fail("Operations were NOT linearizable! See consistency_failure.html")
		} else if res == porcupine.Unknown {
			Fail("Operations linearizability could not be determined (timed out).")
		} else {
			GinkgoWriter.Println("Operations were linearizable.")
		}
	})
})
