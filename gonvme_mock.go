package gonvme

import (
	"errors"
	"fmt"
	"strconv"
)

const (
	// MockNumberOfInitiators controls the number of initiators found in mock mode
	MockNumberOfInitiators = "numberOfInitiators"
	// MockNumberOfTargets controls the number of targets found in mock mode
	MockNumberOfTargets = "numberOfTargets"
	// MockNumberOfSessions controls the number of  NVMe sessions found in mock mode
	MockNumberOfSessions = "numberOfSession"
)

var (
	// GONVMEMock is a struct controlling induced errors
	GONVMEMock struct {
		InduceDiscoveryError   bool
		InduceInitiatorError   bool
		InduceLoginError       bool
		InduceLogoutError      bool
		InduceGetSessionsError bool
	}
)

// MockNVMeTCP provides a mock implementation of an NVMe client
type MockNVMeTCP struct {
	NVMeType
}

// NewMockNVMeTCP - returns a mock NVMeTCP client
func NewMockNVMeTCP(opts map[string]string) *MockNVMeTCP {
	nvme := MockNVMeTCP{
		NVMeType: NVMeType{
			mock:    true,
			options: opts,
		},
	}

	return &nvme
}

func getOptionAsInt(opts map[string]string, key string) int64 {
	v, _ := strconv.ParseInt(opts[key], 10, 64)
	return v
}

func (nvme *MockNVMeTCP) discoverNVMeTCPTargets(address string, login bool) ([]NVMeTarget, error) {
	if GONVMEMock.InduceDiscoveryError {
		return []NVMeTarget{}, errors.New("discoverTargets induced error")
	}
	mockedTargets := make([]NVMeTarget, 0)
	count := getOptionAsInt(nvme.options, MockNumberOfTargets)

	if count == 0 {
		count = 1
	}

	for idx := 0; idx < int(count); idx++ {
		tgt := fmt.Sprintf("%05d", idx)
		mockedTargets = append(mockedTargets,
			NVMeTarget{
				Portal:     address,
				TargetNqn:  "nqn.1988-11.com.dell.mock:e6e2d5b871f1403E169D" + tgt,
				TrType:     "tcp",
				AdrFam:     "fibre-channel",
				SubType:    "nvme subsystem",
				Treq:       "not specified",
				PortID:     "2368",
				TrsvcID:    "none",
				SecType:    "none",
				TargetType: "tcp",
			})
	}

	// send back a slice of targets
	return mockedTargets, nil
}

func (nvme *MockNVMeTCP) getInitiators(filename string) ([]string, error) {

	if GONVMEMock.InduceInitiatorError {
		return []string{}, errors.New("getInitiators induced error")
	}

	mockedInitiators := make([]string, 0)
	count := getOptionAsInt(nvme.options, MockNumberOfInitiators)
	if count == 0 {
		count = 1
	}

	for idx := 0; idx < int(count); idx++ {
		init := fmt.Sprintf("%05d", idx)
		mockedInitiators = append(mockedInitiators,
			"nqn.1988-11.com.dell.mock:01:00000000"+init)
	}
	return mockedInitiators, nil
}

func (nvme *MockNVMeTCP) nvmeConnect(target NVMeTarget) error {

	if GONVMEMock.InduceLoginError {
		return errors.New("NVMe Login induced error")
	}

	return nil
}

func (nvme *MockNVMeTCP) nvmeDisconnect(target NVMeTarget) error {

	if GONVMEMock.InduceLogoutError {
		return errors.New("NVMe Logout induced error")
	}

	return nil
}

func (nvme *MockNVMeTCP) getSessions() ([]NVMESession, error) {

	if GONVMEMock.InduceGetSessionsError {
		return []NVMESession{}, errors.New("getSessions induced error")
	}

	var sessions []NVMESession
	count := getOptionAsInt(nvme.options, MockNumberOfSessions)
	if count == 0 {
		count = 1
	}
	for idx := 0; idx < int(count); idx++ {
		init := fmt.Sprintf("%0d", idx)
		session := NVMESession{}
		session.Target = fmt.Sprintf("nqn.1988-11.com.dell.mock:00:e6e2d5b871f1403E169D%d", idx)
		session.Portal = fmt.Sprintf("192.168.1.%d", idx)
		session.Name = "nvme" + init
		session.NVMESessionState = NVMESessionStateLive
		session.NVMETransportName = NVMETransportNameTCP
		sessions = append(sessions, session)
	}
	return sessions, nil
}

// ====================================================================
// Architecture agnostic code for the mock implementation

// DiscoverNVMeTCPTargets runs an NVMe discovery and returns a list of targets.
func (nvme *MockNVMeTCP) DiscoverNVMeTCPTargets(address string, login bool) ([]NVMeTarget, error) {
	return nvme.discoverNVMeTCPTargets(address, login)
}

// GetInitiators returns a list of NVMe initiators on the local system.
func (nvme *MockNVMeTCP) GetInitiators(filename string) ([]string, error) {
	return nvme.getInitiators(filename)
}

// NVMeConnect will attempt to log into an NVMe target
func (nvme *MockNVMeTCP) NVMeConnect(target NVMeTarget) error {
	return nvme.nvmeConnect(target)
}

// NVMeDisconnect will attempt to log out of an NVMe target
func (nvme *MockNVMeTCP) NVMeDisconnect(target NVMeTarget) error {
	return nvme.nvmeDisconnect(target)
}

// GetSessions Queries NVMe session info
func (nvme *MockNVMeTCP) GetSessions() ([]NVMESession, error) {
	return nvme.getSessions()
}
