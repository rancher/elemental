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
	"context"
	"fmt"
	"io"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var _ = Describe("Performance Benchmarks", Label("performance", "k6"), func() {
	It("should run k6 race condition test", func() {
		ctx := context.Background()

		absPath, err := filepath.Abs("k6")
		Expect(err).NotTo(HaveOccurred())

		req := testcontainers.ContainerRequest{
			Image: "grafana/k6:latest",
			Cmd:   []string{"run", "/scripts/main.js"},
			Files: []testcontainers.ContainerFile{
				{
					HostFilePath:      filepath.Join(absPath, "main.js"),
					ContainerFilePath: "/scripts/main.js",
					FileMode:          0755,
				},
				{
					HostFilePath:      filepath.Join(absPath, "tests", "machineregistration.js"),
					ContainerFilePath: "/scripts/tests/machineregistration.js",
					FileMode:          0755,
				},
			},
			WaitingFor: wait.ForExit(),
		}

		k6Container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		Expect(err).NotTo(HaveOccurred())

		defer func() {
			k6Container.Terminate(ctx)
		}()

		state, err := k6Container.State(ctx)
		Expect(err).NotTo(HaveOccurred())

		logs, err := k6Container.Logs(ctx)
		if err != nil {
			fmt.Printf("Failed to get logs: %v\n", err)
		} else {
			logBytes, _ := io.ReadAll(logs)
			fmt.Printf("K6 Container Logs:\n%s\n", string(logBytes))
		}

		Expect(state.ExitCode).To(Equal(0), "K6 test failed")
	})
})
