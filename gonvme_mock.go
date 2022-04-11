package gonvme

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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

// MockNVMe provides a mock implementation of an NVMe client
type MockNVMe struct {
	NVMeType
}

// NewMockNVMe - returns a mock NVMe client
func NewMockNVMe(opts map[string]string) *MockNVMe {
	nvme := MockNVMe{
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

func (nvme *MockNVMe) discoverNVMeTargets(address string, login bool) ([]NVMeTarget, error) {
	if GONVMEMock.InduceDiscoveryError {
		return []NVMeTarget{}, errors.New("discoverTargets induced error")
	}
	mockedTargets := make([]NVMeTarget, 0)
	count := getOptionAsInt(nvme.options, MockNumberOfTargets)

	if count == 0 {
		count = 1
	}

	if strings.HasPrefix(address, "nn-") {
		for idx := 0; idx < int(count); idx++ {
			tgt := fmt.Sprintf("%05d", idx)
			mockedTargets = append(mockedTargets,
				NVMeTarget{
					Portal:     address,
					TargetNqn:  "nqn.1988-11.com.dell.mock:e6e2d5b871f1403E169D" + tgt,
					TrType:     "fc",
					AdrFam:     "fibre-channel",
					SubType:    "nvme subsystem",
					Treq:       "not specified",
					PortID:     "0",
					TrsvcID:    "none",
					SecType:    "none",
					TargetType: "fc",
					HostAdr:    "nn-0x58aaa11111111a11:pn-0x58aaa11111111a11",
				})
		}
	} else {
		for idx := 0; idx < int(count); idx++ {
			tgt := fmt.Sprintf("%05d", idx)
			mockedTargets = append(mockedTargets,
				NVMeTarget{
					Portal:     address,
					TargetNqn:  "nqn.1988-11.com.dell.mock:e6e2d5b871f1403E169D" + tgt,
					TrType:     "tcp",
					AdrFam:     "ipv4",
					SubType:    "nvme subsystem",
					Treq:       "not specified",
					PortID:     "0",
					TrsvcID:    "none",
					SecType:    "none",
					TargetType: "tcp",
				})
		}
	}

	// send back a slice of targets
	return mockedTargets, nil
}

func (nvme *MockNVMe) getInitiators(filename string) ([]string, error) {

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

func (nvme *MockNVMe) nvmeConnect(target NVMeTarget) error {

	if GONVMEMock.InduceLoginError {
		return errors.New("NVMe Login induced error")
	}

	return nil
}

func (nvme *MockNVMe) nvmeDisconnect(target NVMeTarget) error {

	if GONVMEMock.InduceLogoutError {
		return errors.New("NVMe Logout induced error")
	}

	return nil
}

func (nvme *MockNVMe) getSessions() ([]NVMESession, error) {

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

// DiscoverNVMeTargets runs an NVMe discovery and returns a list of targets.
func (nvme *MockNVMe) DiscoverNVMeTargets(address string, login bool) ([]NVMeTarget, error) {
	return nvme.discoverNVMeTargets(address, login)
}

// GetInitiators returns a list of NVMe initiators on the local system.
func (nvme *MockNVMe) GetInitiators(filename string) ([]string, error) {
	return nvme.getInitiators(filename)
}

// NVMeConnect will attempt to log into an NVMe target
func (nvme *MockNVMe) NVMeConnect(target NVMeTarget) error {
	return nvme.nvmeConnect(target)
}

// NVMeDisconnect will attempt to log out of an NVMe target
func (nvme *MockNVMe) NVMeDisconnect(target NVMeTarget) error {
	return nvme.nvmeDisconnect(target)
}

// GetSessions Queries NVMe session info
func (nvme *MockNVMe) GetSessions() ([]NVMESession, error) {
	return nvme.getSessions()
}
