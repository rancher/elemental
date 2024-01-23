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

// qase-ginkgo integration library

package qase

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	ginkgo "github.com/onsi/ginkgo/v2"
	ginkgoTypes "github.com/onsi/ginkgo/v2/types"
	"github.com/sirupsen/logrus"
	qase "go.qase.io/client"
)

// Defined ResultStatus Enum
type ResultStatusEnum uint

const (
	// NOTE: we can't update from all status to another one!
	// From IN_PROGRESS and UNTESTED we can update to: PASSED, FAILED, BLOCKED, SKIPPED and INVALID
	// From these above status we can't go back nor set another status, a new case needs
	// to be added (with the same id), so this will be seen as Retries in UI
	invalid ResultStatusEnum = iota
	inProgress
	passed
	failed
	blocked
	skipped
	untested
)

/*
This method returns the value of ResultStatus enum in a readable (string) format
  - @return the value in string format
*/
func (r ResultStatusEnum) String() string {
	return [...]string{"invalid", "in_progress", "passed", "failed", "blocked", "skipped", "untested"}[r]
}

var (
	apiToken       = os.Getenv("QASE_API_TOKEN")
	environmentID  int64
	envStrID       = os.Getenv("QASE_ENVIRONMENT_ID")
	logLevel       = strings.ToUpper(os.Getenv("QASE_LOG_LEVEL"))
	projectCode    = os.Getenv("QASE_PROJECT_CODE")
	report         = os.Getenv("QASE_REPORT")
	runComplete    = os.Getenv("QASE_RUN_COMPLETE")
	runDescription = os.Getenv("QASE_RUN_DESCRIPTION")
	runID          int32
	runName        = os.Getenv("QASE_RUN_NAME")
	runStrID       = os.Getenv("QASE_RUN_ID")
)

/*
This function checks the availability of a specific project.
  - @param client Client configuration for Qase instance access
  - @param prj Project name
  - @return true or false, depending of project availability
*/
func checkProject(client *qase.APIClient, prj string) bool {
	prjChk, _, err := client.ProjectsApi.GetProject(context.TODO(), prj)
	if err != nil {
		logrus.Fatalf("Error on checking project: %v", err)
	}

	return prjChk.Status
}

/*
This function creates a run in a specific project.
  - @param client Client configuration for Qase instance access
  - @param name Name of the run
  - @param description Short description of the run
  - @param ids List of case IDs to add in the run
  - @return IdResponse struct
*/
func createRun(client *qase.APIClient, name, description string, ids []int64) *qase.IdResponse {
	if name == "" {
		name = "Automated run " + time.Now().Format(time.RFC3339)
	}

	if description == "" {
		description = "Ginkgo automated run"
	}

	runObject := qase.RunCreate{
		Title:         name,
		Description:   description,
		Cases:         ids,
		IsAutotest:    true,
		EnvironmentId: environmentID,
	}
	logrus.Debugf("runObject: %v", runObject)

	idResponse, _, err := client.RunsApi.CreateRun(context.TODO(), runObject, projectCode)
	if err != nil {
		logrus.Fatalf("Error on creating run: %v", err)
	}

	return &idResponse
}

/*
This function creates a run in a specific project.
  - @param name Name of the run
  - @param description Short description of the run
  - @return ID of the created run
*/
func CreateRun() int32 {
	var createdID int32

	cfg := qase.NewConfiguration()
	cfg.AddDefaultHeader("Token", apiToken)
	client := qase.NewAPIClient(cfg)

	if checkProject(client, projectCode) {
		logrus.Debugf("Project %s is validated", projectCode)

		// Create test run
		idReponse := createRun(client, runName, runDescription, []int64{})
		createdID = int32(idReponse.Result.Id)
		logrus.Debugf("Run named '%s' with description '%s' created with id %d", runName, runDescription, createdID)

		// Check that we can access the run
		checkRun(client, createdID)

		// Export runID for all functions to use it
		os.Setenv("QASE_RUN_ID", fmt.Sprint(runID))
	}

	return createdID
}

/*
This function checks the availability of a specific run.
  - @param client Client configuration for Qase instance access
  - @param id ID of the run to check
  - @return Fatal on error
*/
func checkRun(client *qase.APIClient, id int32) {
	runResponse, _, err := client.RunsApi.GetRun(context.TODO(), projectCode, id)
	if err != nil || !runResponse.Status {
		logrus.Fatalf("Error on checking run: %v", err)
	}
}

/*
This function sets a specific run as complete.
  - @param client Client configuration for Qase instance access
  - @param id ID of the run to check
  - @return Fatal on error
*/
func completeRun(client *qase.APIClient, id int32) {
	completeResponse, _, err := client.RunsApi.CompleteRun(context.TODO(), projectCode, id)
	if err != nil || !completeResponse.Status {
		logrus.Fatalf("Error on completing run: %v", err)
	}
}

/*
This function completely deletes a specific run.
  - @param client Client configuration for Qase instance access
  - @param id ID of the run to check
  - @return Fatal on error
*/
func deleteRun(client *qase.APIClient, id int32) {
	idResponse, _, err := client.RunsApi.DeleteRun(context.TODO(), projectCode, id)
	if err != nil || !idResponse.Status {
		logrus.Fatalf("Error on deleting run with id %d: %v", id, err)
	}
}

/*
This function finalises the results for a specific run.
  - @return Fatal on error
*/
func FinalizeResults() {
	cfg := qase.NewConfiguration()
	cfg.AddDefaultHeader("Token", apiToken)
	client := qase.NewAPIClient(cfg)

	// Do something only if runID is valid
	if runID > 0 {
		// Complete run if needed
		if runComplete != "" {
			completeRun(client, runID)

			// Log in Ginkgo
			ginkgo.GinkgoWriter.Printf("Report for run ID %d has been complete\n", runID)
		}

		// Make the run publicly available
		if report != "" {
			runPublicResponse, _, err := client.RunsApi.UpdateRunPublicity(context.TODO(), qase.RunPublic{Status: true}, projectCode, runID)
			if err != nil {
				logrus.Fatalf("Error on publishing run: %v", err)
			}
			logrus.Debugf("Published run available here: %s", runPublicResponse.Result.Url)

			// Log in Ginkgo
			ginkgo.GinkgoWriter.Printf("Report for run ID %d available: %s\n", runID, runPublicResponse.Result.Url)
		}
	} else {
		logrus.Debug("Nothing to finalize!")
	}
}

/*
This function updates a specific run.
  - @param client Client configuration for Qase instance access
  - @param r ResultUpdate struct with needed informations
  - @param id ID of the run to check
  - @param hash Hash of the case to update in the specified run
  - @return Fatal on error
*/
//lint:ignore U1000 Ignore unused function because this one is not used yet
func updateRun(client *qase.APIClient, r qase.ResultUpdate, id int32, hash string) {
	hashResponse, _, err := client.ResultsApi.UpdateResult(context.TODO(), r, projectCode, id, hash)
	if err != nil {
		logrus.Fatalf("Error on updating run result with id %d: %v", id, err)
	}

	newHash := hashResponse.Result.Hash
	if newHash != hash {
		logrus.Fatalf("Error: new hash (%s) is different from original hash (%s)", newHash, hash)
	}

	logrus.Debugf("Run %d updated for case hash %s", id, hash)
}

/*
This function creates/updates run/cases for specific Ginkgo tests.
  - @param testReport Ginkgo SpecReport struct used to retrieve the test status
  - @param id ID of the case to create/update
  - @return Fatal on error
*/
func Qase(id int64, testReport ginkgo.SpecReport) {
	// A negative or zero run/case ID means that we won't log anything
	if runID == 0 || id <= 0 {
		return
	}

	cfg := qase.NewConfiguration()
	cfg.AddDefaultHeader("Token", apiToken)
	client := qase.NewAPIClient(cfg)

	// Check that project exists
	if checkProject(client, projectCode) {
		logrus.Debugf("Project %s is validated", projectCode)

		// Check that we can access the run
		checkRun(client, runID)
		logrus.Debugf("Using run with id %d", runID)

		// Check that the case exists
		testCaseResponse, _, err := client.CasesApi.GetCase(context.TODO(), projectCode, int32(id))
		if err != nil {
			logrus.Fatalf("Error on getting case: %v", err)
		}
		logrus.Debugf("Case id %d found with title '%s'", id, testCaseResponse.Result.Title)

		// Check run status
		var status ResultStatusEnum
		if testReport.State.Is(ginkgoTypes.SpecStateFailureStates) {
			status = failed
		} else {
			switch testReport.State.String() {
			case "passed":
				status = passed
			case "pending":
				status = blocked
			case "skipped":
				status = skipped
			default:
				// Code unknown, set status to INVALID!
				status = invalid
			}
		}
		logrus.Debugf("Case id %d status from Ginkgo is %v, so Qase status set to %v", id, testReport.State.String(), status)

		// ResultCreate struct, set status to the one provided by Ginkgo
		resultCreate := qase.ResultCreate{
			CaseId: id,
			Status: status.String(),
			TimeMs: int64(testReport.RunTime.Milliseconds()),
		}
		logrus.Debugf("resultCreate: %v", resultCreate)

		// Create result
		inlineResponse, _, err := client.ResultsApi.CreateResult(context.TODO(), resultCreate, projectCode, int64(runID))
		if err != nil {
			logrus.Fatalf("Error on creating result: %v", err)
		}
		hashID := inlineResponse.Result.Hash
		logrus.Debugf("Hash %s created for case id %d", hashID, id)

		// Log in Ginkgo
		ginkgo.GinkgoWriter.Printf("Qase ID %d created for run ID %d on project %s\n", id, runID, projectCode)
	}
}

/*
This function extracts log level based on string input.
  - @param lvl Log level usually set by env var
  - @return logrus.Level
*/
func getLogLvl(lvl string) logrus.Level {
	switch lvl {
	case "PANIC":
		return logrus.PanicLevel
	case "FATAL":
		return logrus.FatalLevel
	case "ERROR":
		return logrus.ErrorLevel
	case "WARN":
		return logrus.WarnLevel
	case "INFO":
		return logrus.InfoLevel
	case "DEBUG":
		return logrus.DebugLevel
	case "TRACE":
		return logrus.TraceLevel
	default:
		// Default level
		return logrus.InfoLevel
	}
}

/*
This function initialises basics things for Qase integration.
*/
func init() {
	var err error

	// Set Debug loglevel
	logrus.SetLevel(getLogLvl(logLevel))

	// OS environment variables are all strings, some need to be modified
	if envStrID != "" {
		environmentID, err = strconv.ParseInt(envStrID, 10, 64)
		if err != nil {
			logrus.Fatalf("Error on converting string to int64: %v", err)
		}
	}

	if runStrID != "" {
		i, err := strconv.ParseInt(runStrID, 10, 32)
		if err != nil {
			logrus.Fatalf("Error on converting string to int64: %v", err)
		}
		runID = int32(i)
	}
}
