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

package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	qase "github.com/rancher-sandbox/qase-ginkgo"
	"github.com/sirupsen/logrus"
)

var (
	runID    int32
	runStrID = os.Getenv("QASE_RUN_ID")
)

func init() {
	if runStrID != "" {
		i, err := strconv.ParseInt(runStrID, 10, 32)
		if err != nil {
			logrus.Fatalf("Error on converting string to int64: %v", err)
		}
		runID = int32(i)
	}
}

func main() {
	// Define the allowed options
	createRun := flag.Bool("create", false, "create a new Qase run")
	deleteRun := flag.Bool("delete", false, "delete a Qase run, QASE_RUN_ID should be set")
	publishRun := flag.Bool("publish", false, "publish a Qase report, QASE_RUN_ID should be set, it also depends on QASE_REPORT and QASE_RUN_COMPLETE")

	// Parse the arguments
	flag.Parse()

	// Only one option at a time is allowed
	if *createRun {
		id := qase.CreateRun()
		if id <= 0 {
			logrus.Fatalln("Error on creating Qase run")
		}
		logrus.Debugf("Qase run id %d created", id)
		fmt.Printf("%d", id)
	} else if *deleteRun {
		qase.DeleteRun()
		logrus.Debugf("Qase run id %d deleted", runID)
	} else if *publishRun {
		qase.FinalizeResults()
		logrus.Debugf("Qase finalization for run id %d has been done", runID)
	} else {
		logrus.Debugln("Nothing to do!")
	}
}
