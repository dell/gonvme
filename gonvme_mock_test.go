/*
 * Copyright © 2025 Dell Inc. or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *      http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package gonvme

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMockNVMe(t *testing.T) {
	opts := map[string]string{
		"numberOfInitiators":       "1",
		"numberOfTCPTargets":       "2",
		"numberOfFCTargets":        "3",
		"numberOfSession":          "10",
		"numberOfNamespaceDevices": "5",
	}

	nvme := NewMockNVMe(opts)
	assert.NotNil(t, nvme)
	assert.Equal(t, opts, nvme.options)
}

func TestMockedDiscoverNVMeTCPTargets(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{
		MockNumberOfTCPTargets: "2",
	})
	GONVMEMock.InduceDiscoveryError = false
	targets, err := nvme.discoverNVMeTCPTargets("1.1.1.1", false)

	if len(targets) != 2 {
		t.Errorf("Expected 2 targets, got %d", len(targets))
	}

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestMockedDiscoverNVMeTCPTargetsError(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{
		MockNumberOfTCPTargets: "2",
	})
	GONVMEMock.InduceDiscoveryError = true
	_, err := nvme.discoverNVMeTCPTargets("1.1.1.1", false)
	assert.NotNil(t, err)
}

func TestMockedDiscoverNVMeFCTargets(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{
		MockNumberOfFCTargets: "2",
	})
	GONVMEMock.InduceDiscoveryError = false

	targets, err := nvme.discoverNVMeFCTargets("nn-0x11aaa11111111a11:pn-0x11aaa11111111a11", false)

	if len(targets) != 2 {
		t.Errorf("Expected 2 targets, got %d", len(targets))
	}

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestMockedDiscoverNVMeFCTargetsError(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{
		MockNumberOfFCTargets: "2",
	})
	GONVMEMock.InduceDiscoveryError = true
	_, err := nvme.discoverNVMeFCTargets("nn-0x11aaa11111111a11:pn-0x11aaa11111111a11", false)
	assert.NotNil(t, err)
}

func TestMockedGetInitiators(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{
		MockNumberOfInitiators: "2",
	})
	initiators, err := nvme.getInitiators("")
	assert.Nil(t, err)
	assert.Len(t, initiators, 2)
}

func TestMockedGetInitiatorsZero(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{
		MockNumberOfInitiators: "0",
	})
	initiators, err := nvme.getInitiators("")
	assert.Nil(t, err)
	assert.Len(t, initiators, 1)
}

func TestMockedGetInitiatorsError(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{
		MockNumberOfInitiators: "2",
	})
	GONVMEMock.InduceInitiatorError = true
	_, err := nvme.getInitiators("")
	assert.NotNil(t, err)
}

func TestMockedNvmeTCPConnect(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{})
	err := nvme.nvmeTCPConnect(NVMeTarget{}, false)
	assert.Nil(t, err)
}

func TestMockedNvmeTCPConnectError(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{})
	GONVMEMock.InduceTCPLoginError = true
	err := nvme.nvmeTCPConnect(NVMeTarget{}, false)
	assert.NotNil(t, err)
}

func TestMockedNvmeFCConnect(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{})
	err := nvme.nvmeFCConnect(NVMeTarget{}, false)
	assert.Nil(t, err)
}

func TestMockedNvmeFCConnectError(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{})
	GONVMEMock.InduceFCLoginError = true
	err := nvme.nvmeFCConnect(NVMeTarget{}, false)
	assert.NotNil(t, err)
}

func TestMockedNvmeDisconnect(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{})
	err := nvme.nvmeDisconnect(NVMeTarget{})
	assert.Nil(t, err)
}

func TestMockedNvmeDisconnectError(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{})
	GONVMEMock.InduceLogoutError = true
	err := nvme.nvmeDisconnect(NVMeTarget{})
	assert.NotNil(t, err)
}

func TestMockedGetNVMeDeviceData(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{})
	_, _, err := nvme.GetNVMeDeviceData("")
	assert.Nil(t, err)
}

func TestMockedGetNVMeDeviceDataError(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{})
	GONVMEMock.InducedNVMeDeviceDataError = true
	_, _, err := nvme.GetNVMeDeviceData("")
	assert.NotNil(t, err)
}

func TestMockedListNVMeNamespaceID(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{})
	_, err := nvme.ListNVMeNamespaceID(nil)
	assert.Nil(t, err)
}

func TestMockedListNVMeNamespaceIDError(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{})
	GONVMEMock.InducedNVMeNamespaceIDError = true
	_, err := nvme.ListNVMeNamespaceID(nil)
	assert.NotNil(t, err)
}

func TestMockedListNVMeDeviceAndNamespace(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{})
	_, err := nvme.ListNVMeDeviceAndNamespace()
	assert.Nil(t, err)
}

func TestMockedListNVMeDeviceAndNamespaceError(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{})
	GONVMEMock.InducedNVMeDeviceAndNamespaceError = true
	_, err := nvme.ListNVMeDeviceAndNamespace()
	assert.NotNil(t, err)
}

func TestMockedGetSessions(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{})
	_, err := nvme.GetSessions()
	assert.Nil(t, err)
}

func TestMockedGetSessionsError(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{})
	GONVMEMock.InduceGetSessionsError = true
	_, err := nvme.GetSessions()
	assert.NotNil(t, err)
}
