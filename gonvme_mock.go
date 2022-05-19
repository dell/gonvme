package gonvme

import (
	"errors"
	"fmt"
	"strconv"
)

const (
	// MockNumberOfInitiators controls the number of initiators found in mock mode
	MockNumberOfInitiators = "numberOfInitiators"
	// MockNumberOfTCPTargets controls the number of NVMeTCP targets found in mock mode
	MockNumberOfTCPTargets = "numberOfTCPTargets"
	// MockNumberOfFCTargets controls the number of NVMeFC targets found in mock mode
	MockNumberOfFCTargets = "numberOfFCTargets"
	// MockNumberOfSessions controls the number of  NVMe sessions found in mock mode
	MockNumberOfSessions = "numberOfSession"
	// MockNumberOfNamespaceDevices controls the number of  NVMe Namespace Devices found in mock mode
	MockNumberOfNamespaceDevices = "numberOfNamespaceDevices"
)

var (
	// GONVMEMock is a struct controlling induced errors
	GONVMEMock struct {
		InduceDiscoveryError        bool
		InduceInitiatorError        bool
		InduceTCPLoginError         bool
		InduceFCLoginError          bool
		InduceLogoutError           bool
		InduceGetSessionsError      bool
		InducedNamespaceDeviceError bool
		InducedNamespaceDataError   bool
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

func (nvme *MockNVMe) discoverNVMeTCPTargets(address string, login bool) ([]NVMeTarget, error) {
	if GONVMEMock.InduceDiscoveryError {
		return []NVMeTarget{}, errors.New("discoverTargets induced error")
	}
	mockedTargets := make([]NVMeTarget, 0)
	count := getOptionAsInt(nvme.options, MockNumberOfTCPTargets)

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
				AdrFam:     "ipv4",
				SubType:    "nvme subsystem",
				Treq:       "not specified",
				PortID:     "0",
				TrsvcID:    "none",
				SecType:    "none",
				TargetType: "tcp",
			})
	}

	// send back a slice of targets
	return mockedTargets, nil
}

func (nvme *MockNVMe) discoverNVMeFCTargets(address string, login bool) ([]NVMeTarget, error) {
	if GONVMEMock.InduceDiscoveryError {
		return []NVMeTarget{}, errors.New("discoverTargets induced error")
	}
	mockedTargets := make([]NVMeTarget, 0)
	count := getOptionAsInt(nvme.options, MockNumberOfFCTargets)

	if count == 0 {
		count = 1
	}

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

func (nvme *MockNVMe) nvmeTCPConnect(target NVMeTarget, duplicateConnect bool) error {

	if GONVMEMock.InduceTCPLoginError {
		return errors.New("NVMeTCP Login induced error")
	}

	return nil
}

func (nvme *MockNVMe) nvmeFCConnect(target NVMeTarget, duplicateConnect bool) error {

	if GONVMEMock.InduceFCLoginError {
		return errors.New("NVMeFC Login induced error")
	}

	return nil
}

func (nvme *MockNVMe) nvmeDisconnect(target NVMeTarget) error {

	if GONVMEMock.InduceLogoutError {
		return errors.New("NVMe Logout induced error")
	}

	return nil
}

//GetNamespaceData returns the information of namespace specific to the namespace ID
func (nvme *MockNVMe) GetNamespaceData(path string, namespaceID string) (string, string, error) {
	if GONVMEMock.InducedNamespaceDataError {
		return "", "", errors.New("NVMe Namespace Data Induced Error")
	}

	nguid := "1a111a1111aa11111aaa1111111111a1"
	namespace := "11"

	return nguid, namespace, nil
}

//ListNamespaceDevices returns the Device Paths and Namespace of each NVMe device and each output content
func (nvme *MockNVMe) ListNamespaceDevices() (map[DevicePathAndNamespace][]string, error) {
	if GONVMEMock.InducedNamespaceDeviceError {
		return map[DevicePathAndNamespace][]string{}, errors.New("listNamespaceDevices induced error")
	}

	mockedNamespaceDevices := make(map[DevicePathAndNamespace][]string)
	count := getOptionAsInt(nvme.options, MockNumberOfNamespaceDevices)
	if count == 0 {
		count = 1
	}

	for idx := 0; idx < int(count); idx++ {
		init := fmt.Sprintf("%05d", idx)
		var currentPathAndNamespace DevicePathAndNamespace
		var namespaceDevice []string

		currentPathAndNamespace.DevicePath = "/dev/nvme0n" + init
		currentPathAndNamespace.Namespace = init

		namespaceDevice = append(namespaceDevice, "0x"+init)
		mockedNamespaceDevices[currentPathAndNamespace] = namespaceDevice
	}
	return mockedNamespaceDevices, nil
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

// DiscoverNVMeTCPTargets runs an NVMe discovery and returns a list of targets.
func (nvme *MockNVMe) DiscoverNVMeTCPTargets(address string, login bool) ([]NVMeTarget, error) {
	return nvme.discoverNVMeTCPTargets(address, login)
}

// DiscoverNVMeFCTargets runs an NVMe discovery and returns a list of targets.
func (nvme *MockNVMe) DiscoverNVMeFCTargets(address string, login bool) ([]NVMeTarget, error) {
	return nvme.discoverNVMeFCTargets(address, login)
}

// GetInitiators returns a list of NVMe initiators on the local system.
func (nvme *MockNVMe) GetInitiators(filename string) ([]string, error) {
	return nvme.getInitiators(filename)
}

// NVMeTCPConnect will attempt to log into an NVMe target
func (nvme *MockNVMe) NVMeTCPConnect(target NVMeTarget, duplicateConnect bool) error {
	return nvme.nvmeTCPConnect(target, duplicateConnect)
}

// NVMeFCConnect will attempt to log into an NVMe target
func (nvme *MockNVMe) NVMeFCConnect(target NVMeTarget, duplicateConnect bool) error {
	return nvme.nvmeFCConnect(target, duplicateConnect)
}

// NVMeDisconnect will attempt to log out of an NVMe target
func (nvme *MockNVMe) NVMeDisconnect(target NVMeTarget) error {
	return nvme.nvmeDisconnect(target)
}

// GetSessions Queries NVMe session info
func (nvme *MockNVMe) GetSessions() ([]NVMESession, error) {
	return nvme.getSessions()
}
