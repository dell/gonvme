package gonvme

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetTimeouts(t *testing.T) {
	var prop time.Duration

	// Test with value 0, should set to defaultVal
	setTimeouts(&prop, 0, 10*time.Second)
	assert.Equal(t, 10*time.Second, prop)

	// Test with non-zero value, should set to value
	setTimeouts(&prop, 5*time.Second, 10*time.Second)
	assert.Equal(t, 5*time.Second, prop)
}

func TestNVMeType_isMock(t *testing.T) {
	nvme := &NVMeType{mock: true}
	assert.True(t, nvme.isMock())

	nvme.mock = false
	assert.False(t, nvme.isMock())
}

func TestNVMeType_getOptions(t *testing.T) {
	options := map[string]string{"key": "value"}
	nvme := &NVMeType{options: options}
	assert.Equal(t, options, nvme.getOptions())
}
