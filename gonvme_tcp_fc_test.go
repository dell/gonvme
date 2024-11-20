package gonvme

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock function to replace os.Stat in tests
type MockFileSystem struct{}

func (MockFileSystem) Stat(path string) (os.FileInfo, error) {
	if path == "/sbin/nvme" {
		return &mockFileInfo{isDir: false}, nil
	}
	return nil, os.ErrNotExist
}

type mockFileInfo struct {
	isDir bool
}

func (m *mockFileInfo) Name() string       { return "" }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() os.FileMode  { return 0 }
func (m *mockFileInfo) ModTime() time.Time { return time.Time{} }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }

func TestNewNVMe(t *testing.T) {
	// Inject the mock file system
	// fs := MockFileSystem{}
	opts := map[string]string{
		"ChrootDirectory": "/test",
	}

	nvme := &NVMe{
		NVMeType: NVMeType{
			mock:    false,
			options: opts,
		},
	}
	nvme.sessionParser = &sessionParser{}

	// nvme := &NVMe{
	// 	fs:          fs,
	// 	NVMeCommand: "/sbin/nvme",
	// 	options:     opts,
	// }

	assert.NotNil(t, nvme)
	// assert.Equal(t, "/sbin/nvme", nvme.NVMeCommand)
}

func TestGetChrootDirectory(t *testing.T) {
	// fs := MockFileSystem{}
	opts := map[string]string{
		"ChrootDirectory": "/test",
	}
	// nvme := &NVMe{
	// 	fs:          fs,
	// 	NVMeCommand: "/sbin/nvme",
	// 	options:     opts,
	// }

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
	// nvme = &NVMe{
	// 	fs:          fs,
	// 	NVMeCommand: "/sbin/nvme",
	// 	options:     opts,
	// }

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
					NodeName: "0x5005076400c7ec87",
					PortName: "0xc05076ffd6801e10",
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
			"testdata/nvme/hostnqn",
			"",
			nil,
			[]string{"nqn.2014-08.org.nvmexpress:uuid:4c4c4544-0042-5210-8053-b5c04f424433"},
			false,
		},
		{
			"successfully gets initiator specifying the file name",
			"testdata/nvme/hostnqn",
			"testdata/nvme/hostnqn",
			nil,
			[]string{"nqn.2014-08.org.nvmexpress:uuid:4c4c4544-0042-5210-8053-b5c04f424433"},
			false,
		},
		{
			"error path doesn't exist",
			"testdata/bad/nvme/hostnqn",
			"",
			nil,
			[]string{},
			true,
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

func TestListNVMeDeviceAndNamespace(t *testing.T) {
	tests := []struct {
		name         string
		getCommandFn func(name string, arg ...string) command
		want         []DevicePathAndNamespace
		wantErr      bool
	}{
		{
			"successfully lists devices",
			func(name string, arg ...string) command {
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
			"erorr listing devices",
			func(name string, arg ...string) command {
				return &mockCommand{
					err: errors.New("error listing devices"),
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
		getCommandFn func(name string, arg ...string) command
		devices      []DevicePathAndNamespace
		want         map[DevicePathAndNamespace][]string
		wantErr      bool
	}{
		{
			"successfully lists device IDs",
			func(name string, arg ...string) command {
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
			func(name string, arg ...string) command {
				return &mockCommand{
					err: errors.New("error listing devices"),
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
		getCommandFn func(name string, arg ...string) command
		want         []NVMESession
		wantErr      bool
	}{
		{
			"successfully gets sessions",
			func(name string, arg ...string) command {
				return &mockCommand{
					out: []byte(`[
		  {
		    "HostNQN":"nqn.2014-08.org.nvmexpress:uuid:6f08058a-af91-46bf-8311-a60da3a10348",
		    "HostID":"6f08058a-af91-46bf-8311-a60da3a10348",
		    "Subsystems":[
		      {
		        "Name":"nvme-subsys0",
		        "NQN":"nqn.1988-11.com.dell:powerstore:00:1b7322d7546dFD05675D",
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
					Target:            "nqn.1988-11.com.dell:powerstore:00:1b7322d7546dFD05675D",
					Portal:            "10.1.1.1:4420",
					Name:              "nvme3",
					NVMETransportName: "tcp",
					NVMESessionState:  "live",
				},
				{
					Target:            "nqn.1988-11.com.dell:powerstore:00:1b7322d7546dFD05675D",
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
			func(name string, arg ...string) command {
				return &mockCommand{
					err: errors.New("error"),
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
	err error
	out []byte
}

func (m mockCommand) Output() ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.out, nil
}

type MockCommand struct {
	mock.Mock
}

var tcpTestPortal string

func TestDiscoverNVMeTCPTargets(t *testing.T) {
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

tcpTestPortal = "1.1.1.1"

	cmdCommandFunc := func([]string) ([]byte, error) {
		return []byte(mockOutput), nil
	}
	c := NewNVMe(map[string]string{})
	_, err := c.discoverNVMeTCPTargets(tcpTestPortal, false, cmdCommandFunc)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestMockDiscoverNVMEFCTargets(t *testing.T) {
	var c NVMEinterface
	MockNumberOfFCTargets := "numberOfFCTargets"
	opts := map[string]string{}
	expected := 5
	opts[MockNumberOfFCTargets] = "5"
	c = NewNVMe(opts)
	// c = mock
	targets, err := c.DiscoverNVMeFCTargets("nn-0x11aaa111111a1a1a:pn-0x11aaa111111a1a1a", true)
	if err != nil {
		t.Error(err.Error())
	}
	if len(targets) != expected {
		t.Errorf("Expected to find %d targets, but got back %v", expected, targets)
	}
}

var testTarget string
func TestNVMeTCPLoginLogoutTargets(t *testing.T) {
	testTarget = "nqn.1988-11.com.mock:00:e6e2d5b871f1403E169D"
	c := NewNVMe(map[string]string{})
	tgt := NVMeTarget{
		Portal:     tcpTestPortal,
		TargetNqn:  testTarget,
		TrType:     "tcp",
		AdrFam:     "ipv4",
		SubType:    "nvme subsystem",
		Treq:       "not specified",
		PortID:     "0",
		TrsvcID:    "none",
		SecType:    "none",
		TargetType: "tcp",
	}
	err := c.NVMeTCPConnect(tgt, false)
	if err == nil {
		t.Error(err.Error())
		return
	}
	nvmeSessions, _ := c.GetSessions()
	if len(nvmeSessions) != 0 {
		err = c.NVMeDisconnect(tgt)
		if err != nil {
			t.Error(err.Error())
			return
		}
	}
}

var fcTestPortal string
var hostAddress string

func TestNVMeFCLoginLogoutTargets(t *testing.T) {
	fcTestPortal = "nn-0x11aaa111a1111a1a:pn-0x11aaa11111111a1a"
	testTarget = "nqn.1988-11.com.mock:00:e6e2d5b871f1403E169D"
	hostAddress = "nn-0x11aaa111a1111a1a:pn-0x11aaa11111111a1a"
	c := NewNVMe(map[string]string{})
	tgt := NVMeTarget{
		Portal:     fcTestPortal,
		TargetNqn:  testTarget,
		TrType:     "fc",
		AdrFam:     "fibre-channel",
		SubType:    "nvme subsystem",
		Treq:       "not specified",
		PortID:     "0",
		TrsvcID:    "none",
		SecType:    "none",
		TargetType: "fc",
		HostAdr:    hostAddress,
	}
	err := c.NVMeFCConnect(tgt, false)
	if err == nil {
		t.Error(err.Error())
		return
	}
	nvmeSessions, _ := c.GetSessions()
	if len(nvmeSessions) != 0 {
		err = c.NVMeDisconnect(tgt)
		if err != nil {
			t.Error(err.Error())
			return
		}
	}
}

func TestLoginLoginLogoutTargets(t *testing.T) {
	testTarget = "nqn.1988-11.com.mock:00:e6e2d5b871f1403E169D"
	tcpTestPortal = "1.1.1.1"
	c := NewNVMe(map[string]string{})
	tgt := NVMeTarget{
		Portal:     tcpTestPortal,
		TargetNqn:  testTarget,
		TrType:     "tcp",
		AdrFam:     "ipv4",
		SubType:    "nvme subsystem",
		Treq:       "not specified",
		PortID:     "0",
		TrsvcID:    "none",
		SecType:    "none",
		TargetType: "tcp",
	}
	err := c.NVMeTCPConnect(tgt, false)
	if err == nil {
		t.Error(err.Error())
		return
	}
	err = c.NVMeFCConnect(tgt, false)
	if err == nil {
		t.Error(err.Error())
		return
	}
	nvmeSessions, _ := c.GetSessions()
	if len(nvmeSessions) != 0 {
		err = c.NVMeDisconnect(tgt)
		if err != nil {
			t.Error(err.Error())
			return
		}
	}
}

func TestLogoutLogoutTargets(t *testing.T) {
	testTarget = "nqn.1988-11.com.mock:00:e6e2d5b871f1403E169D"
	tcpTestPortal = "1.1.1.1"
	c := NewNVMe(map[string]string{})
	tgt := NVMeTarget{
		Portal:     tcpTestPortal,
		TargetNqn:  testTarget,
		TrType:     "tcp",
		AdrFam:     "fibre-channel",
		SubType:    "nvme subsystem",
		Treq:       "not specified",
		PortID:     "0",
		TrsvcID:    "none",
		SecType:    "none",
		TargetType: "tcp",
	}
	// log out of the target, just in case we are logged in already
	_ = c.NVMeTCPConnect(tgt, false)
	nvmeSessions, _ := c.GetSessions()
	if len(nvmeSessions) != 0 {
		err := c.NVMeDisconnect(tgt)
		if err != nil {
			t.Error(err.Error())
			return
		}
	}
}