# Integration of Ginkgo tests results into Qase

## Introduction

This `qase-ginkgo` integration library is a a basic library that can be used to report results from Ginkgo Go tests to Qase reporting tool.

## Quick test

You can simply copy/paste the `qase_example_test.go` file and adapt it to your needs.

## Step-by-step explanation

### API token

To use this simple library you first need an API token from your Qase instance (`QASE_API_TOKEN`). Some permissions are also needed like creation/deletion, please refer to Qase document for this.

### Environment variables

Some variables can/have to be set for the integration to run. Three are mandatory: `QASE_API_TOKEN`, `QASE_PROJECT_CODE` and `QASE_RUN_ID`.  
The first one is descripted above, the second is the code of your Qase projectand the third is the id on an already created run.  
`QASE_RUN_ID` can also be set to `auto` to automatically creates a new run.

Here the variables you can use/define:
| Variable | Description | Needed? |
|:---:|:---:|:---:|
| `QASE_API_TOKEN` | API token to access your Qase instance | Mandatory |
| `QASE_DELETE_RUN` | Delete the specified run (mainly useful for debugging purposes) | Optional |
| `QASE_ENVIRONNMENT_ID` | Use a specific environnment in a project | Optional |
| `QASE_LOG_DEVEL` | Define log level use, see [logrus](https://pkg.go.dev/github.com/sirupsen/logrus#readme-level-logging) for more infos | Optional |
| `QASE_PROJECT_CODE` | Code of your Qase project | Mandatory |
| `QASE_REPORT` | Create a public Qase report | Optional |
| `QASE_RUN_COMPLETE` | Set the run as complete | Optional |
| `QASE_RUN_DESCRIPTION` | Description of your run | Optional |
| `QASE_RUN_ID` | Id of the run to use, can be set to `auto` to create one | Mandatory |
| `QASE_RUN_NAME` | Name of your run | Optional |

### Ginkgo integration

Using the library is not too complicated: specific calls are mainly done in the `suite_test.go` file.

First, you have to import the `qase-ginkgo` library:
```go
package mycode_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/rancher/elemental/tests/e2e/helpers/qase"
)
```

Then add these lines:
```go
// To be able to log test results in Qase
var testCaseID int64

var _ = ReportAfterSuite("Qase Report", func(report Report) {
	// Finalize Qase report
	FinalizeResults()
})

var _ = ReportBeforeEach(func(report SpecReport) {
	// Reset case ID
	testCaseID = -1
})

var _ = ReportAfterEach(func(report SpecReport) {
	// Add result in Qase if asked
	Qase(testCaseID, report)
})
```

Please note that you can change the `testCaseID` with your own variable, just pay attention to be consistent!
You can also add more code in the `Report*` Ginkgo functions, this configuration is just the minimal stuff you need to add for the `qse-ginkgo` libray to work!

Then, in your test code you have to add the case id value (`testCaseID` in our example) you want to report in Qase, for example like this:
```go
func testFunc() ([]byte, error) {
	fmt.Println("Hello World from function!")

	return (exec.Command("pwd").Output())
}

var _ = Describe("Qase Ginkgo Integration", func() {
	// The case with ID=10 should be marked as passed
	It("Test the Qase integration with case ID=10", func() {
		// Report to Qase
		testCaseID = 10    // <== Line to add!

		// Short delay
		time.Sleep(3 * time.Second)

		By("testing that output is not empty", func() {
			fmt.Println("Hello World from It!")
			Ω(testQaseFunc()).Should(Not(BeEmpty()))
		})
	})
})
```

And that's all!

**NOTE:** be aware that `testCaseID` cannot be set multiple times in the same `It` call, only the last set value will be used.  
This is pure Ginkgo limitation. Please check Ginkgo [documentation](https://onsi.github.io/ginkgo/#spec-subjects-it) for more information.
