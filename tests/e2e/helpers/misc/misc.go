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

package misc

import (
	"math/rand"
	"time"
)

/*
Wait for random time
  - @param index Modulo for the seed
  - @returns Wait for the calculated time
*/
func RandomSleep(sequential bool, index int) {
	// Only useful in parallel mode
	if sequential {
		return
	}

	// Initialize the seed
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Get a pseudo-random value
	timeMax := 240000
	value := r.Intn(timeMax + (timeMax % index))

	// Wait until value is reached
	time.Sleep(time.Duration(value) * time.Millisecond)
}

/*
Wait for nodes to be booted
  - @param index Index of the current VM (usually used in a loop)
  - @param vmIndex Index of the first booted node
  - @param bootedNodes Number of already booted nodes
  - @param maxNodes Maximum number of nodes
  - @returns Returns (increment) the number of already booted nodes
*/
func WaitNodesBoot(index, vmIndex, bootedNodes, maxNodes int) int {
	if (index - vmIndex - bootedNodes) == maxNodes {
		// Wait a little
		time.Sleep(4 * time.Minute)
	}

	// Save the number of nodes already bootstrapped for the next round
	return (index - vmIndex)
}
