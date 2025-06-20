/*
 *
 * Copyright © 2024 Dell Inc. or its subsidiaries. All Rights Reserved.
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
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type testData struct {
	TCPPortal     string
	FCPortal      string
	Target        string
	FCHostAddress string
}

var (
	tcpTestPortal string
	testTarget    string
)

func reset() {
	testValuesFile, err := os.ReadFile("testdata/unittest_values.json")
	if err != nil {
		log.Infof("Error Reading the file: %s ", err)
	}
	var testValues testData
	err = json.Unmarshal(testValuesFile, &testValues)
	if err != nil {
		log.Infof("Error during unmarshal: %s", err)
	}
	tcpTestPortal = testValues.TCPPortal
	testTarget = testValues.Target
}

func TestNewNVMe(t *testing.T) {
	opts := map[string]string{
		"chrootDirectory": "/test",
	}

	nvme := &NVMe{
		NVMeType: NVMeType{
			mock:    false,
			options: opts,
		},
	}
	nvme.sessionParser = &sessionParser{}

	assert.NotNil(t, nvme)

	originalGetPaths := getPaths
	defer func() {
		getPaths = originalGetPaths
	}()

	// this path is used for hostnqn but can serve as the nvme command for testing
	getPaths = func() []string {
		return []string{"testdata/hostnqn"}
	}

	nvme = NewNVMe(nil)
	assert.NotNil(t, nvme)
}

func TestNewNVMeBadPaths(t *testing.T) {
	originalGetPaths := getPaths
	defer func() {
		getPaths = originalGetPaths
	}()

	getPaths = func() []string {
		return []string{"/bad/path/does/not/exist"}
	}
	nvme := NewNVMe(nil)
	assert.NotNil(t, nvme)

	getPaths = func() []string {
		return []string{"testdata/fc_host"}
	}
	nvme = NewNVMe(nil)
	assert.NotNil(t, nvme)
}

func TestGetChrootDirectory(t *testing.T) {
	opts := map[string]string{
		"chrootDirectory": "/test",
	}

	nvme := &NVMe{
		NVMeType: NVMeType{
			mock:    false,
			options: opts,
		},
	}
	nvme.sessionParser = &sessionParser{}

	chrootDir := nvme.getChrootDirectory()
	assert.Equal(t, "/test", chrootDir)

	opts = map[string]string{}

	nvme = &NVMe{
		NVMeType: NVMeType{
			mock:    false,
			options: opts,
		},
	}
	nvme.sessionParser = &sessionParser{}

	chrootDir = nvme.getChrootDirectory()
	assert.Equal(t, "/", chrootDir)
}

func TestBuildNVMeCommand(t *testing.T) {
	opts := map[string]string{
		"chrootDirectory": "/test",
	}
	nvme := NewNVMe(opts)

	cmd := []string{"nvme", "list"}
	builtCmd := nvme.buildNVMeCommand(cmd)
	expectedCmd := []string{"chroot", "/test", "nvme", "list"}
	assert.Equal(t, expectedCmd, builtCmd)

	opts = map[string]string{}
	nvme = NewNVMe(opts)

	builtCmd = nvme.buildNVMeCommand(cmd)
	assert.Equal(t, cmd, builtCmd)
}

func TestGetFCHostInfo(t *testing.T) {
	tests := []struct {
		name          string
		fcHostPattern string
		want          []FCHBAInfo
		wantErr       bool
	}{
		{
			"successfully gets fibre channel hosts",
			"testdata/fc_host/host*",
			[]FCHBAInfo{
				{
					NodeName: "00:00:00:00:00:00:00:01",
					PortName: "00:00:00:00:00:00:00:01",
				},
			},
			false,
		},
		{
			"no fibre channel hosts due to path doesn't exist",
			"testdata/bad/fc_host/host*",
			[]FCHBAInfo{},
			false,
		},
		{
			"no fibre channel hosts due to malformed hosts",
			"testdata/fc_host_bad/host*",
			[]FCHBAInfo{},
			false,
		},
		{
			"error reading path due to malformed path",
			"**/[invalid",
			[]FCHBAInfo{},
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			originalFCHostPattern := fcHostPath
			fcHostPath = tc.fcHostPattern
			defer func() { fcHostPath = originalFCHostPattern }()

			nvme := NewNVMe(nil)
			got, err := nvme.getFCHostInfo()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestGetInitiators(t *testing.T) {
	tests := []struct {
		name          string
		initiatorFile string
		fileName      string
		options       map[string]string
		want          []string
		wantErr       bool
	}{
		{
			"successfully gets initiator",
			"testdata/hostnqn",
			"",
			nil,
			[]string{"nqn.2014-08.org.mock:uuid:00a00000-0000-0000-0000-aa0a0000000a"},
			false,
		},
		{
			"successfully gets initiator specifying the file name",
			"testdata/hostnqn",
			"testdata/hostnqn",
			nil,
			[]string{"nqn.2014-08.org.mock:uuid:00a00000-0000-0000-0000-aa0a0000000a"},
			false,
		},
		{
			"error path doesn't exist",
			"testdata/empty_hostnqn",
			"",
			nil,
			[]string{},
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			originalInitiatorFile := DefaultInitiatorNameFile
			DefaultInitiatorNameFile = tc.initiatorFile
			defer func() { DefaultInitiatorNameFile = originalInitiatorFile }()

			nvme := NewNVMe(nil)
			got, err := nvme.GetInitiators(tc.fileName)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

// There are two known formats for the output of the nvme list -o json command.
// Newer versions of nvme, around 2.11 (RHEL 9.6) introduced a version which is
// incompatable with the older versions.
func TestListNVMeDeviceAndNamespace(t *testing.T) {
	tests := []struct {
		name         string
		getCommandFn func(_ string, _ ...string) command
		want         []DevicePathAndNamespace
		wantErr      bool
	}{
		{
			"nvme-cli pre 2_11 format",
			func(_ string, _ ...string) command {
				return &mockCommand{
					out: []byte(`{
						"Devices" : [
						  {
							"NameSpace" : 9217,
							"DevicePath" : "/dev/nvme0n1",
							"Firmware" : "2.1.0.0",
							"Index" : 0,
							"ModelNumber" : "dellemc",
							"SerialNumber" : "FP08RZ2",
							"UsedBytes" : 0,
							"MaximumLBA" : 10485760,
							"PhysicalSize" : 5368709120,
							"SectorSize" : 512
						  },
						  {
							"NameSpace" : 9222,
							"DevicePath" : "/dev/nvme0n2",
							"Firmware" : "2.1.0.0",
							"Index" : 0,
							"ModelNumber" : "dellemc",
							"SerialNumber" : "FP08RZ2",
							"UsedBytes" : 0,
							"MaximumLBA" : 10485760,
							"PhysicalSize" : 5368709120,
							"SectorSize" : 512
						  }
						]
					  }`),
				}
			},
			[]DevicePathAndNamespace{
				{
					DevicePath: "/dev/nvme0n1",
					Namespace:  "9217",
				},
				{
					DevicePath: "/dev/nvme0n2",
					Namespace:  "9222",
				},
			},
			false,
		},
		{
			"nvme-cli 2_11 format",
			func(_ string, _ ...string) command {
				return &mockCommand{
					out: []byte(`{
						"Devices":[
							{
							"HostNQN":"nqn.2014-08.org.nvmexpress:uuid:a66f1c42-4bce-a619-9c59-9ae6ac2ccb8a",
							"HostID":"a2d57d74-a198-4e6b-aa78-97af9cd00f31",
							"Subsystems":[
								{
								"Subsystem":"nvme-subsys0",
								"SubsystemNQN":"nqn.1988-11.com.dell:powerstore:00:42c92aa830b1FF113003",
								"Controllers":[
									{
									"Controller":"nvme0",
									"Cntlid":"4102",
									"SerialNumber":"883YCJ3",
									"ModelNumber":"dellemc-powerstore",
									"Firmware":"4.1.0.0",
									"Transport":"tcp",
									"Address":"traddr=10.11.12.13,trsvcid=4420,src_addr=10.10.10.21",
									"Slot":"",
									"Namespaces":[
									],
									"Paths":[
										{
										"Path":"nvme0c0n1",
										"ANAState":"optimized"
										}
									]
									},
									{
									"Controller":"nvme1",
									"Cntlid":"8",
									"SerialNumber":"883YCJ3",
									"ModelNumber":"dellemc-powerstore",
									"Firmware":"4.1.0.0",
									"Transport":"tcp",
									"Address":"traddr=10.11.12.14,trsvcid=4420,src_addr=10.10.10.21",
									"Slot":"",
									"Namespaces":[
									],
									"Paths":[
										{
										"Path":"nvme0c1n1",
										"ANAState":"non-optimized"
										}
									]
									}
								],
								"Namespaces":[
									{
									"NameSpace":"nvme0n1",
									"Generic":"ng0n1",
									"NSID":293,
									"UsedBytes":620130304,
									"MaximumLBA":6291456,
									"PhysicalSize":3221225472,
									"SectorSize":512
									}
								]
								}
							]
							}
						]
					}`),
				}
			},
			[]DevicePathAndNamespace{
				{
					DevicePath: "/dev/nvme0n1",
					Namespace:  "293",
				},
			},
			false,
		},
		{
			"error listing devices",
			func(_ string, _ ...string) command {
				return &mockCommand{
					outErr: errors.New("error listing devices"),
				}
			},
			nil,
			true,
		},
		{
			"error on unmarshalling json",
			func(_ string, _ ...string) command {
				return &mockCommand{
					out: []byte(`{
						"Devices" : [
						  {
						]
					  }`),
				}
			},
			nil,
			true,
		},
		{
			"unknown data format",
			func(_ string, _ ...string) command {
				return &mockCommand{
					out: []byte(`{
						"Devices" : [
						  {
							"ValidButNotWhatWeExpect" : "value"
						  }
						]
					}`),
				}
			},
			nil,
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			originalGetCommand := getCommand
			getCommand = tc.getCommandFn
			defer func() { getCommand = originalGetCommand }()

			nvme := NewNVMe(nil)
			got, err := nvme.ListNVMeDeviceAndNamespace()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestListNVMeNamespaceID(t *testing.T) {
	tests := []struct {
		name         string
		getCommandFn func(_ string, _ ...string) command
		devices      []DevicePathAndNamespace
		want         map[DevicePathAndNamespace][]string
		wantErr      bool
	}{
		{
			"successfully lists device IDs",
			func(_ string, _ ...string) command {
				return &mockCommand{
					out: []byte(`
		[   0]:0x2401
		[   1]:0x2406`),
				}
			},
			[]DevicePathAndNamespace{
				{
					DevicePath: "/dev/nvme0n1",
					Namespace:  "9217",
				},
			},
			map[DevicePathAndNamespace][]string{
				{
					DevicePath: "/dev/nvme0n1",
					Namespace:  "9217",
				}: {"0x2401", "0x2406"},
			},
			false,
		},
		{
			"empty resposne from error listing",
			func(_ string, _ ...string) command {
				return &mockCommand{
					outErr: errors.New("error listing devices"),
				}
			},
			[]DevicePathAndNamespace{
				{
					DevicePath: "/dev/nvme0n1",
					Namespace:  "9217",
				},
			},
			map[DevicePathAndNamespace][]string{},
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			originalGetCommand := getCommand
			getCommand = tc.getCommandFn
			defer func() { getCommand = originalGetCommand }()

			nvme := NewNVMe(nil)
			got, err := nvme.ListNVMeNamespaceID(tc.devices)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestGetSessions(t *testing.T) {
	tests := []struct {
		name         string
		getCommandFn func(_ string, _ ...string) command
		want         []NVMESession
		wantErr      bool
	}{
		{
			"successfully gets sessions",
			func(_ string, _ ...string) command {
				return &mockCommand{
					out: []byte(`[
		  {
		    "HostNQN":"nqn.2014-08.org.nvmexpress:uuid:1a11111a-aa11-11aa-1111-a11aa1a11111",
		    "HostID":"6f08058a-af91-46bf-8311-a60da3a10348",
		    "Subsystems":[
		      {
		        "Name":"nvme-subsys0",
		        "NQN":"nqn.1988-11.com.dell:mock:00:1a1111a1111aAA11111A",
		        "IOPolicy":"numa",
		        "Paths":[
		          {
		            "Name":"nvme3",
		            "Transport":"tcp",
		            "Address":"traddr=10.1.1.1,trsvcid=4420,src_addr=10.1.1.2",
		            "State":"live"
		          },
		          {
		            "Name":"nvme2",
		            "Transport":"tcp",
		            "Address":"traddr=10.1.1.2,trsvcid=4420,src_addr=10.1.1.2",
		            "State":"live"
		          }
		        ]
		      }
		    ]
		  }
		]`),
				}
			},
			[]NVMESession{
				{
					Target:            "nqn.1988-11.com.dell:mock:00:1a1111a1111aAA11111A",
					Portal:            "10.1.1.1:4420",
					Name:              "nvme3",
					NVMETransportName: "tcp",
					NVMESessionState:  "live",
				},
				{
					Target:            "nqn.1988-11.com.dell:mock:00:1a1111a1111aAA11111A",
					Portal:            "10.1.1.2:4420",
					Name:              "nvme2",
					NVMETransportName: "tcp",
					NVMESessionState:  "live",
				},
			},
			false,
		},
		{
			"error listing sessions",
			func(_ string, _ ...string) command {
				return &mockCommand{
					outErr: errors.New("error"),
				}
			},
			nil,
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			originalGetCommand := getCommand
			getCommand = tc.getCommandFn
			defer func() { getCommand = originalGetCommand }()

			nvme := NewNVMe(nil)
			got, err := nvme.GetSessions()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

type mockCommand struct {
	outErr   error
	startErr error
	waitErr  error
	out      []byte
	stdErr   []byte
}

func (m mockCommand) Output() ([]byte, error) {
	if m.outErr != nil {
		return nil, m.outErr
	}
	return m.out, nil
}

func (m mockCommand) Start() error {
	return m.startErr
}

func (m mockCommand) Wait() error {
	return m.waitErr
}

func (m mockCommand) StderrPipe() (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(m.stdErr)), nil
}

func TestDiscoverNVMeTCPTargets(t *testing.T) {
	nvme := NewNVMe(map[string]string{})
	mockOutput := `=====Discovery Log Entry 0======
trtype:  tcp
adrfam:  ipv4
subtype: nvme subsystem
treq:    not specified
portid:  4420
trsvcid: 4420
subnqn:  nqn.1988-11.com.dell:powerstore:00:1a1111a1111aAA11111A
traddr:  10.0.0.1
sectype: none
`
	reset()
	originalGetCommand := getCommand
	getCommandFunc := func(_ string, _ ...string) command {
		return &mockCommand{
			out:    []byte(mockOutput),
			outErr: nil,
		}
	}
	getCommand = getCommandFunc
	defer func() { getCommand = originalGetCommand }()
	_, err := nvme.discoverNVMeTCPTargets(tcpTestPortal, false)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestDiscoverNVMeFCTargets(t *testing.T) {
	opts := map[string]string{}
	nvme := NewNVMe(opts)

	mockOutput := `Discovery Log Number of Records 2, Generation counter 2
=====Discovery Log Entry 0======
trtype:  fc
adrfam:  fibre-channel
subtype: nvme subsystem
treq:    not specified
portid:  0
trsvcid: none
subnqn:  nqn.1111-11.com.dell:powerstore:00:a1a1a1a111a1111a111a
traddr:  nn-0x11aaa111a1111a11:aa-0x11aaa11111111a11
`
	originalGetCommand := getCommand
	getCommandFunc := func(_ string, _ ...string) command {
		return &mockCommand{
			out:    []byte(mockOutput),
			outErr: nil,
		}
	}
	getCommand = getCommandFunc
	defer func() { getCommand = originalGetCommand }()

	originalFCHostPattern := fcHostPath
	fcHostPath = "testdata/fc_host/host*"
	defer func() { fcHostPath = originalFCHostPattern }()

	_, err := nvme.discoverNVMeFCTargets("nn-0x11aaa111111a1a1a:pn-0x11aaa111111a1a1a", true)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestGetNVMeDeviceDatatest(t *testing.T) {
	var c NVMEinterface
	opts := map[string]string{}
	c = NewNVMe(opts)

	mockOutput := `
nvme id-ns /dev/nvme3n1 0x95
NVME Identify Namespace 149:
nsze    : 0x1000000
ncap    : 0x1000000
nuse    : 0x223b8
nsfeat  : 0xb
nlbaf   : 0
flbas   : 0
mc      : 0
dpc     : 0
dps     : 0
nmic    : 0x1
rescap  : 0xff
fpi     : 0
dlfeat  : 9
nawun   : 2047
nawupf  : 2047
nacwu   : 0
nabsn   : 2047
nabo    : 0
nabspf  : 2047
noiob   : 0
nvmcap  : 0
mssrl   : 0
mcl     : 0
msrc    : 0
anagrpid: 2
nsattr  : 0
nvmsetid: 0
endgid  : 0
nguid   : 111111aaaa11a1111aaa1111111a1a11
eui64   : 0000000000000000
lbaf  0 : ms:0   lbads:9  rp:0 (in use)
	`
	originalGetCommand := getCommand
	getCommandFunc := func(_ string, _ ...string) command {
		return &mockCommand{
			out:    []byte(mockOutput),
			outErr: nil,
		}
	}
	getCommand = getCommandFunc
	defer func() { getCommand = originalGetCommand }()

	guid, namespace, err := c.GetNVMeDeviceData("testdata/device_data")
	if err != nil {
		t.Error(err.Error())
	}

	if guid != "111111aaaa11a1111aaa1111111a1a11" {
		t.Errorf("want %s, got %s", "111111aaaa11a1111aaa1111111a1a11", guid)
	}

	if namespace != "149" {
		t.Errorf("want %s, got %s", "149", namespace)
	}
}

func TestGetNVMeDeviceDataError(t *testing.T) {
	var c NVMEinterface
	opts := map[string]string{}
	c = NewNVMe(opts)

	originalGetCommand := getCommand
	getCommandFunc := func(_ string, _ ...string) command {
		return &mockCommand{
			outErr: errors.New("error"),
		}
	}
	getCommand = getCommandFunc
	defer func() { getCommand = originalGetCommand }()

	_, _, err := c.GetNVMeDeviceData("/nvmeMock/0n1")
	if err == nil {
		t.Error("Expected an induced error")
		return
	}
}

func TestNVMeTCPConnect(t *testing.T) {
	tests := []struct {
		name             string
		nvmeTarget       NVMeTarget
		duplicateConnect bool
		getCommandFn     func(_ string, _ ...string) command
		wantErr          bool
		errContains      string
	}{
		{
			"successfully connects",
			NVMeTarget{
				Portal:    "1.1.1.1",
				TargetNqn: "nqn.1988-11.com.mock:00:a1a1a1a111a1111A111A",
			},
			false,
			func(_ string, _ ...string) command {
				return &mockCommand{
					startErr: nil,
					waitErr:  nil,
				}
			},
			false,
			"",
		},
		{
			"successfully connects duplicate",
			NVMeTarget{
				Portal:    "1.1.1.1",
				TargetNqn: "nqn.1988-11.com.mock:00:a1a1a1a111a1111A111A",
			},
			true,
			func(_ string, _ ...string) command {
				return &mockCommand{
					startErr: nil,
					waitErr:  nil,
				}
			},
			false,
			"",
		},
		{
			"error connecting",
			NVMeTarget{
				Portal:    "1.1.1.1",
				TargetNqn: "nqn.1988-11.com.mock:00:a1a1a1a111a1111A111A",
			},
			false,
			func(_ string, _ ...string) command {
				return &mockCommand{
					startErr: nil,
					waitErr:  errors.New("error should be in output"),
				}
			},
			true,
			"error should be in output",
		},
		{
			"error connecting with code 114",
			NVMeTarget{
				Portal:    "1.1.1.1",
				TargetNqn: "nqn.1988-11.com.mock:00:a1a1a1a111a1111A111A",
			},
			false,
			func(_ string, _ ...string) command {
				return &mockCommand{
					startErr: nil,
					waitErr:  &exec.ExitError{ProcessState: &os.ProcessState{}},
				}
			},
			true,
			"error connecting to nvme target",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			originalGetCommand := getCommand
			getCommand = tc.getCommandFn
			defer func() { getCommand = originalGetCommand }()

			c := NewNVMe(map[string]string{})
			err := c.NVMeTCPConnect(tc.nvmeTarget, tc.duplicateConnect)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNVMeFCConnect(t *testing.T) {
	tests := []struct {
		name             string
		nvmeTarget       NVMeTarget
		duplicateConnect bool
		getCommandFn     func(_ string, _ ...string) command
		wantErr          bool
	}{
		{
			"successfully connects",
			NVMeTarget{
				Portal:    "1.1.1.1",
				TargetNqn: "nqn.1988-11.com.mock:00:a1a1a1a111a1111A111A",
			},
			false,
			func(_ string, _ ...string) command {
				return &mockCommand{
					startErr: nil,
					waitErr:  nil,
				}
			},
			false,
		},
		{
			"successfully connects duplicate",
			NVMeTarget{
				Portal:    "1.1.1.1",
				TargetNqn: "nqn.1988-11.com.mock:00:a1a1a1a111a1111A111A",
			},
			true,
			func(_ string, _ ...string) command {
				return &mockCommand{
					startErr: nil,
					waitErr:  nil,
				}
			},
			false,
		},
		{
			"error connecting",
			NVMeTarget{
				Portal:    "1.1.1.1",
				TargetNqn: "nqn.1988-11.com.mock:00:a1a1a1a111a1111A111A",
			},
			false,
			func(_ string, _ ...string) command {
				return &mockCommand{
					startErr: nil,
					waitErr:  errors.New("error"),
				}
			},
			true,
		},
		{
			"error connecting with code 114",
			NVMeTarget{
				Portal:    "1.1.1.1",
				TargetNqn: "nqn.1988-11.com.mock:00:a1a1a1a111a1111A111A",
			},
			false,
			func(_ string, _ ...string) command {
				return &mockCommand{
					startErr: nil,
					waitErr:  &exec.ExitError{ProcessState: &os.ProcessState{}},
				}
			},
			true,
		},
	}

	for _, tc := range tests {
		originalGetCommand := getCommand
		getCommand = tc.getCommandFn
		defer func() { getCommand = originalGetCommand }()

		c := NewNVMe(map[string]string{})
		err := c.NVMeFCConnect(tc.nvmeTarget, tc.duplicateConnect)
		if tc.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestNVMeDisconnect(t *testing.T) {
	tests := []struct {
		name         string
		nvmeTarget   NVMeTarget
		getCommandFn func(_ string, _ ...string) command
		wantErr      bool
	}{
		{
			"successfully disconnects",
			NVMeTarget{
				Portal:    "1.1.1.1",
				TargetNqn: "nqn.1988-11.com.mock:00:a1a1a1a111a1111A111A",
			},
			func(_ string, _ ...string) command {
				return &mockCommand{
					outErr: nil,
				}
			},
			false,
		},
		{
			"error disconnecting",
			NVMeTarget{
				Portal:    "1.1.1.1",
				TargetNqn: "nqn.1988-11.com.mock:00:a1a1a1a111a1111A111A",
			},
			func(_ string, _ ...string) command {
				return &mockCommand{
					outErr: errors.New("error"),
				}
			},
			true,
		},
	}

	for _, tc := range tests {
		originalGetCommand := getCommand
		getCommand = tc.getCommandFn
		defer func() { getCommand = originalGetCommand }()

		c := NewNVMe(map[string]string{})
		err := c.NVMeDisconnect(tc.nvmeTarget)
		if tc.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestDeviceRescan(t *testing.T) {
	tests := []struct {
		name         string
		getCommandFn func(_ string, _ ...string) command
		wantErr      bool
	}{
		{
			"successfully rescans",
			func(_ string, _ ...string) command {
				return &mockCommand{
					outErr: nil,
				}
			},
			false,
		},
		{
			"error rescanning",
			func(_ string, _ ...string) command {
				return &mockCommand{
					outErr: errors.New("error"),
				}
			},
			true,
		},
	}

	for _, tc := range tests {
		originalGetCommand := getCommand
		getCommand = tc.getCommandFn
		defer func() { getCommand = originalGetCommand }()

		c := NewNVMe(map[string]string{})
		err := c.DeviceRescan("device")
		if tc.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestIsNoObjsExitCode(t *testing.T) {
	r := isNoObjsExitCode(nil)
	assert.False(t, r)
}
