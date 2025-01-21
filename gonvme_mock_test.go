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
	targets, err := nvme.DiscoverNVMeTCPTargets("1.1.1.1", false)

	if len(targets) != 2 {
		t.Errorf("Expected 2 targets, got %d", len(targets))
	}

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestMockedDiscoverNVMeTCPTargetsZero(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{
		MockNumberOfTCPTargets: "0",
	})
	GONVMEMock.InduceDiscoveryError = false
	targets, _ := nvme.DiscoverNVMeTCPTargets("1.1.1.1", false)

	assert.Len(t, targets, 1)
}

func TestMockedDiscoverNVMeTCPTargetsError(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{
		MockNumberOfTCPTargets: "2",
	})
	GONVMEMock.InduceDiscoveryError = true
	_, err := nvme.DiscoverNVMeTCPTargets("1.1.1.1", false)
	assert.NotNil(t, err)
}

func TestMockedDiscoverNVMeFCTargets(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{
		MockNumberOfFCTargets: "2",
	})
	GONVMEMock.InduceDiscoveryError = false

	targets, err := nvme.DiscoverNVMeFCTargets("nn-0x11aaa11111111a11:pn-0x11aaa11111111a11", false)

	if len(targets) != 2 {
		t.Errorf("Expected 2 targets, got %d", len(targets))
	}

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestMockedDiscoverNVMeFCTargetsZero(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{
		MockNumberOfFCTargets: "0",
	})
	GONVMEMock.InduceDiscoveryError = false
	targets, _ := nvme.DiscoverNVMeFCTargets("nn-0x11aaa11111111a11:pn-0x11aaa11111111a11", false)
	assert.Len(t, targets, 1)
}

func TestMockedDiscoverNVMeFCTargetsError(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{
		MockNumberOfFCTargets: "2",
	})
	GONVMEMock.InduceDiscoveryError = true
	_, err := nvme.DiscoverNVMeFCTargets("nn-0x11aaa11111111a11:pn-0x11aaa11111111a11", false)
	assert.NotNil(t, err)
}

func TestMockedGetInitiators(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{
		MockNumberOfInitiators: "2",
	})
	initiators, err := nvme.GetInitiators("")
	assert.Nil(t, err)
	assert.Len(t, initiators, 2)
}

func TestMockedGetInitiatorsZero(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{
		MockNumberOfInitiators: "0",
	})
	initiators, err := nvme.GetInitiators("")
	assert.Nil(t, err)
	assert.Len(t, initiators, 1)
}

func TestMockedGetInitiatorsError(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{
		MockNumberOfInitiators: "2",
	})
	GONVMEMock.InduceInitiatorError = true
	_, err := nvme.GetInitiators("")
	assert.NotNil(t, err)
}

func TestMockedNVMeTCPConnect(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{})
	err := nvme.NVMeTCPConnect(NVMeTarget{}, false)
	assert.Nil(t, err)
}

func TestMockedNVMeTCPConnectError(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{})
	GONVMEMock.InduceTCPLoginError = true
	err := nvme.NVMeTCPConnect(NVMeTarget{}, false)
	assert.NotNil(t, err)
}

func TestMockedNVMeFCConnect(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{})
	err := nvme.NVMeFCConnect(NVMeTarget{}, false)
	assert.Nil(t, err)
}

func TestMockedNVMeFCConnectError(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{})
	GONVMEMock.InduceFCLoginError = true
	err := nvme.NVMeFCConnect(NVMeTarget{}, false)
	assert.NotNil(t, err)
}

func TestMockedNVMeDisconnect(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{})
	err := nvme.NVMeDisconnect(NVMeTarget{})
	assert.Nil(t, err)
}

func TestMockedNVMeDisconnectError(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{})
	GONVMEMock.InduceLogoutError = true
	err := nvme.NVMeDisconnect(NVMeTarget{})
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

func TestMockedDeviceRescan(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{})
	GONVMEMock.InduceGetSessionsError = false
	err := nvme.DeviceRescan("")
	assert.Nil(t, err)
}

func TestMockedDeviceRescanError(t *testing.T) {
	nvme := NewMockNVMe(map[string]string{})
	GONVMEMock.InduceGetSessionsError = true
	err := nvme.DeviceRescan("")
	assert.NotNil(t, err)
}
