package gonvme

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
