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

// InventoryInput represents a request to create an inventory
type InventoryInput struct {
	Op   string // "create"
	Name string
}

// InventoryOutput represents the result of the creation
type InventoryOutput struct {
	Success bool
}

var inventoryModel = porcupine.Model{
	Init: func() interface{} { return make(map[string]bool) },
	Step: func(state, input, output interface{}) (bool, interface{}) {
		inp := input.(InventoryInput)
		out := output.(InventoryOutput)
		st := state.(map[string]bool)

		// Copy state to be safe (though porcupine might copy for us, it's safer to not mutate input state)
		newState := make(map[string]bool)
		for k, v := range st {
			newState[k] = v
		}

		if inp.Op == "create" {
			if out.Success {
				newState[inp.Name] = true
				return true, newState
			}
			// If failure, state doesn't change
			return true, newState
		}
		return false, newState
	},
}

// InventoryStore Interface
type InventoryStore interface {
	Setup() error
	Teardown()
	CreateInventory(name string) error
}

// MemoryInventoryStore (Simulation)
type MemoryInventoryStore struct {
	mu    sync.Mutex
	items map[string]bool
}

func (m *MemoryInventoryStore) Setup() error {
	m.items = make(map[string]bool)
	return nil
}
func (m *MemoryInventoryStore) Teardown() {}
func (m *MemoryInventoryStore) CreateInventory(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Simulate success
	m.items[name] = true
	return nil
}

// KubectlInventoryStore (Real)
type KubectlInventoryStore struct {
	Namespace string
}

func (k *KubectlInventoryStore) Setup() error {
	return nil
}

func (k *KubectlInventoryStore) Teardown() {
	_, _ = kubectl.Run("delete", "machineinventory", "-n", k.Namespace, "-l", "test=inventory-race")
}

func (k *KubectlInventoryStore) CreateInventory(name string) error {
	manifest := fmt.Sprintf(`
apiVersion: elemental.cattle.io/v1beta1
kind: MachineInventory
metadata:
  name: %s
  namespace: %s
  labels:
    test: inventory-race
spec:
  tpmHash: "test-hash-%s"
`, name, k.Namespace, name)

	f, _ := os.CreateTemp("", "inv-*.yaml")
	defer os.Remove(f.Name())
	f.WriteString(manifest)
	f.Close()

	return kubectl.Apply(k.Namespace, f.Name())
}

var _ = Describe("Inventory Race Tests", Label("inventory_race"), func() {
	var store InventoryStore
	const namespace = "default"

	BeforeEach(func() {
		useCluster := false
		if os.Getenv("KUBECONFIG") != "" || os.Getenv("API_URL") != "" {
			useCluster = true
		}

		if useCluster {
			GinkgoWriter.Println("Running in REAL mode (Cluster)")
			store = &KubectlInventoryStore{Namespace: namespace}
		} else {
			GinkgoWriter.Println("Running in SIMULATION mode (Memory)")
			store = &MemoryInventoryStore{}
		}

		err := store.Setup()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if store != nil {
			store.Teardown()
		}
	})

	It("should be linearizable when creating multiple inventories", func() {
		opsCount := 100
		concurrency := 20
		history := make([]porcupine.Operation, 0, opsCount)
		var historyMu sync.Mutex
		var wg sync.WaitGroup

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < opsCount/concurrency; j++ {
					invName := fmt.Sprintf("race-inv-%d-%d-%s", id, j, rand.String(5))

					start := time.Now()
					err := store.CreateInventory(invName)
					end := time.Now()

					output := InventoryOutput{Success: err == nil}

					historyMu.Lock()
					history = append(history, porcupine.Operation{
						ClientId: id,
						Input:    InventoryInput{Op: "create", Name: invName},
						Call:     start.UnixNano(),
						Output:   output,
						Return:   end.UnixNano(),
					})
					historyMu.Unlock()
				}
			}(i)
		}
		wg.Wait()

		res, info := porcupine.CheckOperationsVerbose(inventoryModel, history, 0)
		if res == porcupine.Illegal {
			f, _ := os.Create("inventory_race_failure.html")
			porcupine.Visualize(inventoryModel, info, f)
			f.Close()
			Fail("Operations were NOT linearizable! See inventory_race_failure.html")
		} else if res == porcupine.Unknown {
			Fail("Operations linearizability could not be determined (timed out).")
		} else {
			GinkgoWriter.Println("Operations were linearizable.")
		}
	})
})
