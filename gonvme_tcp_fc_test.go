package gonvme

import (
    "os"
    "testing"

    "github.com/stretchr/testify/assert"
)

// Mock function to replace os.Stat in tests
func mockOsStat(path string) (os.FileInfo, error) {
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
    // Save the original function to restore later
    originalOsStat := osStat
    // Replace the osStat function with the mock
    osStat = mockOsStat
    // Restore the original function after the test
    defer func() { osStat = originalOsStat }()

    opts := map[string]string{
        "ChrootDirectory": "/test",
    }
    nvme := NewNVMe(opts)

    assert.NotNil(t, nvme)
    assert.Equal(t, "/sbin/nvme", nvme.NVMeCommand)
}

func TestGetChrootDirectory(t *testing.T) {
    opts := map[string]string{
        "ChrootDirectory": "/test",
    }
    nvme := NewNVMe(opts)

    chrootDir := nvme.getChrootDirectory()
    assert.Equal(t, "/test", chrootDir)

    opts = map[string]string{}
    nvme = NewNVMe(opts)

    chrootDir = nvme.getChrootDirectory()
    assert.Equal(t, "/", chrootDir)
}