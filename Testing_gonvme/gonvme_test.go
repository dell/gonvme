package gonvme

import (
	"testing"
	"time"
	"context"
	// "github.com/dell/gonvme/internal/logger"
	// "github.com/dell/gonvme/internal/tracer"
	"github.com/stretchr/testify/assert"
)

type NVMESession struct {
    Name              string
    Target            string
    NVMETransportName string
    Portal            string
    NVMESessionState  string
}

type NVMeTarget struct {
	Portal     string
	TargetNqn  string
	TrType     string
	AdrFam     string
	SubType    string
	Treq       string
	PortID     string
	TrsvcID    string
	SecType    string
	TargetType string
	HostAdr    string
}
type DevicePathAndNamespace struct {
	DevicePath string
	Namespace  string
}

// type Logger interface {
// 	Info(ctx context.Context, format string, args ...interface{})
// 	Debug(ctx context.Context, format string, args ...interface{})
// 	Error(ctx context.Context, format string, args ...interface{})
// }

type MockLogger struct{}
func (m *MockLogger) Info(ctx context.Context, format string, args ...interface{})  {}
func (m *MockLogger) Infof(ctx context.Context, format string, args ...interface{}) {}
func (m *MockLogger) Error(ctx context.Context, format string, args ...interface{}) {}
func (m *MockLogger) Errorf(ctx context.Context, format string, args ...interface{}) {}
func (m *MockLogger) Debug(ctx context.Context, format string, args ...interface{}) {}
func (m *MockLogger) Debugf(ctx context.Context, format string, args ...interface{}) {}


type MockTracer struct{}
type MockSpan struct{}

// func (m *MockTracer) Start(ctx context.Context, name string) (context.Context, tracer.Span) {
// 	return ctx, nil
// }

// func (m *MockTracer) End(ctx context.Context, span tracer.Span) {}

// func TestDiscoverNVMeFCTargets(t *testing.T) {
// 	nvme := &NVMeType{}
// 	address := "10.0.0.1"

// 	expectedTargets := []NVMeTarget{
// 		{
// 			Portal:     address,
// 			TargetNqn:  "nqn.1988-11.com.dell.mock:e6e2d5b871f1403E169D00000",
// 			TrType:     "fc",
// 			AdrFam:     "fibre-channel",
// 			SubType:    "nvme subsystem",
// 			Treq:       "not specified",
// 			PortID:     "0",
// 			TrsvcID:    "none",
// 			SecType:    "none",
// 			TargetType: "fc",
// 			HostAdr:    "nn-0x58aaa11111111a11:pn-0x58aaa11111111a11",
// 		},
// 	}

// 	targets, err := nvme.DiscoverNVMeTCPTargets(address, true)
// 	assert.NoError(t, err)
// 	assert.Equal(t, expectedTargets, targets)
// }

func TestSetLogger(t *testing.T) {
	// Mock logger
	// mockLogger := &logger.Logger{}
	mockLogger := &MockLogger{}
	SetLogger(mockLogger)

	// Check if the logger is set correctly
	// assert.Equal(t, mockLogger, mockLogger)
}

// func TestSetTracer(t *testing.T) {
	// Mock tracer
	// mockTracer := &tracer.Tracer{}
	// SetTracer(mockTracer)

	// Check if the tracer is set correctly
	// assert.Equal(t, mockTracer, tracer.GetTracer())
// }

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