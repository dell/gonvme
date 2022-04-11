/*
 *
 * Copyright © 2022 Dell Inc. or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package gonvme

import (
	"time"

	"github.com/dell/gonvme/internal/logger"
	"github.com/dell/gonvme/internal/tracer"
)

// Logger - Placeholder for logger
type Logger = logger.Logger

// Tracer - Placeholder for tracer
type Tracer = tracer.Tracer

// NVMEinterface is the interface that provides the NVMe client functionality
type NVMEinterface interface {

	// DiscoverNVMeTargets discovers the targets exposed via a given portal
	// returns an array of NVMe Target instances
	DiscoverNVMeTargets(address string, login bool) ([]NVMeTarget, error)

	// GetInitiators get a list of iSCSI initiators defined in a specified file
	// To use the system default file of "/etc/nvme/hostnqn", provide a filename of ""
	GetInitiators(filename string) ([]string, error)

	//NVMeConnect connects into a specified NVMe target
	NVMeConnect(target NVMeTarget) error

	// NVMeDisconnect disconnect from the specified NVMe target
	NVMeDisconnect(target NVMeTarget) error

	// GetSessions queries information about sessions
	GetSessions() ([]NVMESession, error)

	// generic implementations
	isMock() bool
	getOptions() map[string]string
}

// NVMeType is the base structure for each platform implementation
type NVMeType struct {
	mock    bool
	options map[string]string
}

// SetLogger set custom logger for gonvme
func SetLogger(customLogger Logger) {
	logger.SetLogger(customLogger)
}

// SetTracer set custom tracer for gonvme
func SetTracer(customTracer Tracer) {
	tracer.SetTracer(customTracer)
}

func setTimeouts(prop *time.Duration, value time.Duration, defaultVal time.Duration) {
	if value == 0 {
		*prop = defaultVal
	} else {
		*prop = value
	}
}

func (i *NVMeType) isMock() bool {
	return i.mock
}

func (i *NVMeType) getOptions() map[string]string {
	return i.options
}
