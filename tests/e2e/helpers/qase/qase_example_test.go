/*
Copyright © 2022 - 2024 SUSE LLC

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

/*
Example code to use the Qase reporter helper in Ginkgo tests.

This test can be run with this command:
$ QASE_LOG_LEVEL=debug                                          \
  QASE_REPORT=1                                                 \
  QASE_RUN_COMPLETE=1                                           \
  QASE_RUN_NAME="Qase/Gingo Integration Run"                    \
  QASE_RUN_DESCRIPTION="Automated test for Qase/Ginkgo library" \
  QASE_PROJECT_CODE=ELEMENTAL                                   \
  QASE_API_TOKEN=<TOKEN>                                        \
  ginkgo --label-filter qase -r .
*/

package qase_test

import (
	"fmt"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func testQaseFunc() ([]byte, error) {
	fmt.Println("Hello World from function!")

	return (exec.Command("pwd").Output())
}

var _ = Describe("Qase Ginkgo Integration - Basic tests", Label("qase"), func() {
	// Nothing should be logged in Qase for this test
	It("Test a sleep function", func() {
		fmt.Println("Sleep for 2s...")
		time.Sleep(2 * time.Second)
	})

	// The whole test with ID=30 should be marked as failed
	It("Test the Qase function with ID=30", func() {
		// Report to Qase
		testCaseID = 30

		By("testing that output is not empty (will pass)", func() {
			fmt.Println("Hello World from test 1!")
			Ω(testQaseFunc()).Should(Not(BeEmpty()))
		})

		// Short delay
		time.Sleep(5 * time.Second)

		By("testing that output is empty (will fail)", func() {
			fmt.Println("Hello World from test 2!")
			Ω(testQaseFunc()).Should(BeEmpty())
		})
	})

	// The whole test with ID=31 should be marked as passed
	It("Test the Qase function with ID=31", func() {
		// Report to Qase
		testCaseID = 31

		// Short delay
		time.Sleep(3 * time.Second)

		By("testing that output is not empty (will pass)", func() {
			fmt.Println("Hello World from test 1!")
			Ω(testQaseFunc()).Should(Not(BeEmpty()))
		})
	})
})
