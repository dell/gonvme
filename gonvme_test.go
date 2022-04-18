package gonvme

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"strings"
	"testing"
)

type testData struct {
	TCPPortal     string
	FCPortal      string
	Target        string
	FCHostAddress string
}

var (
	tcpTestPortal string
	fcTestPortal  string
	testTarget    string
	hostAddress   string
)

func reset() {

	testValuesFile, err := ioutil.ReadFile("testdata/unittest_values.json")
	if err != nil {
		log.Infof("Error Reading the file: %s ", err)
	}
	var testValues testData
	err = json.Unmarshal(testValuesFile, &testValues)
	if err != nil {
		log.Infof("Error during unmarshal: %s", err)
	}
	tcpTestPortal = testValues.TCPPortal
	fcTestPortal = testValues.FCPortal
	testTarget = testValues.Target
	hostAddress = testValues.FCHostAddress

	GONVMEMock.InduceDiscoveryError = false
	GONVMEMock.InduceInitiatorError = false
	GONVMEMock.InduceLoginError = false
	GONVMEMock.InduceLogoutError = false
	GONVMEMock.InduceGetSessionsError = false
}

func TestTCPPolymorphichCapability(t *testing.T) {
	reset()
	var c NVMEinterface
	// start off with a real implementation
	c = NewNVMeTCP(map[string]string{})
	if c.isMock() {
		// this should not be a mock implementation
		t.Error("Expected a real implementation but got a mock one")
		return
	}
	// switch it to mock
	c = NewMockNVMe(map[string]string{})
	if !c.isMock() {
		// this should not be a real implementation
		t.Error("Expected a mock implementation but got a real one")
		return
	}
	// switch back to a real implementation
	c = NewNVMeTCP(map[string]string{})
	if c.isMock() {
		// this should not be a mock implementation
		t.Error("Expected a real implementation but got a mock one")
		return
	}

}

func TestFCPolymorphichCapability(t *testing.T) {
	reset()
	var c NVMEinterface
	// start off with a real implementation
	c = NewNVMeFC(map[string]string{})
	if c.isMock() {
		// this should not be a mock implementation
		t.Error("Expected a real implementation but got a mock one")
		return
	}
	// switch it to mock
	c = NewMockNVMe(map[string]string{})
	if !c.isMock() {
		// this should not be a real implementation
		t.Error("Expected a mock implementation but got a real one")
		return
	}
	// switch back to a real implementation
	c = NewNVMeFC(map[string]string{})
	if c.isMock() {
		// this should not be a mock implementation
		t.Error("Expected a real implementation but got a mock one")
		return
	}

}

func TestNVMeTCPDiscoverTargets(t *testing.T) {
	reset()
	c := NewNVMeTCP(map[string]string{})
	_, err := c.DiscoverNVMeTargets(tcpTestPortal, false)
	if err == nil {
		t.Error(err.Error())
	}
}

func TestNVMeFCDiscoverTargets(t *testing.T) {
	reset()
	c := NewNVMeFC(map[string]string{})
	_, err := c.DiscoverNVMeTargets(tcpTestPortal, false)
	if err == nil {
		t.Error(err.Error())
	}
}

func TestNVMeTCPLoginLogoutTargets(t *testing.T) {
	reset()
	c := NewNVMeTCP(map[string]string{})
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
	err := c.NVMeConnect(tgt)
	if err == nil {
		t.Error(err.Error())
		return
	}
	err = c.NVMeDisconnect(tgt)
	if err != nil {
		t.Error(err.Error())
		return
	}
}

func TestNVMeFCLoginLogoutTargets(t *testing.T) {
	reset()
	c := NewNVMeFC(map[string]string{})
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

	err := c.NVMeConnect(tgt)
	if err == nil {
		t.Error(err.Error())
		return
	}
	err = c.NVMeDisconnect(tgt)
	if err != nil {
		t.Error(err.Error())
		return
	}
}

func TestNVMeTCPLoginLoginLogoutTargets(t *testing.T) {
	reset()
	c := NewNVMeTCP(map[string]string{})
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
	err := c.NVMeConnect(tgt)
	if err == nil {
		t.Error(err.Error())
		return
	}
	err = c.NVMeConnect(tgt)
	if err == nil {
		t.Error(err.Error())
		return
	}
	err = c.NVMeDisconnect(tgt)
	if err != nil {
		t.Error(err.Error())
		return
	}
}

func TestNVMeFCLoginLoginLogoutTargets(t *testing.T) {
	reset()
	c := NewNVMeFC(map[string]string{})
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
	}
	err := c.NVMeConnect(tgt)
	if err == nil {
		t.Error(err.Error())
		return
	}
	err = c.NVMeConnect(tgt)
	if err == nil {
		t.Error(err.Error())
		return
	}
	err = c.NVMeDisconnect(tgt)
	if err != nil {
		t.Error(err.Error())
		return
	}
}

func TestNVMeTCPLogoutLogoutTargets(t *testing.T) {
	reset()
	c := NewNVMeTCP(map[string]string{})
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
	// log out of the target, just in case we are logged in already
	_ = c.NVMeConnect(tgt)
	err := c.NVMeDisconnect(tgt)
	if err != nil {
		t.Error(err.Error())
		return
	}
}

func TestNVMeFCLogoutLogoutTargets(t *testing.T) {
	reset()
	c := NewNVMeFC(map[string]string{})
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
	}
	// log out of the target, just in case we are logged in already
	_ = c.NVMeConnect(tgt)
	err := c.NVMeDisconnect(tgt)
	if err != nil {
		t.Error(err.Error())
		return
	}
}

func TestNVMeTCPGetInitiators(t *testing.T) {
	reset()
	var testdata = []struct {
		filename string
		count    int
	}{
		{"testdata/initiatorname.nvme", 1},
		{"testdata/multiple_nqn.nvme", 2},
		{"testdata/no_nqn.nvme", 0},
		{"testdata/valid.nvme", 1},
	}

	c := NewNVMeTCP(map[string]string{})
	for _, tt := range testdata {
		initiators, err := c.GetInitiators(tt.filename)
		if err != nil {
			t.Errorf("Error getting %d initiators from %s: %s", tt.count, tt.filename, err.Error())
		}
		if len(initiators) != tt.count {
			t.Errorf("Expected %d initiators in %s, but got %d", tt.count, tt.filename, len(initiators))
		}
	}

}

func TestNVMeFCGetInitiators(t *testing.T) {
	reset()
	var testdata = []struct {
		filename string
		count    int
	}{
		{"testdata/initiatorname.nvme", 1},
		{"testdata/multiple_nqn.nvme", 2},
		{"testdata/no_nqn.nvme", 0},
		{"testdata/valid.nvme", 1},
	}

	c := NewNVMeFC(map[string]string{})
	for _, tt := range testdata {
		initiators, err := c.GetInitiators(tt.filename)
		if err != nil {
			t.Errorf("Error getting %d initiators from %s: %s", tt.count, tt.filename, err.Error())
		}
		if len(initiators) != tt.count {
			t.Errorf("Expected %d initiators in %s, but got %d", tt.count, tt.filename, len(initiators))
		}
	}

}

func TestBuildNVMeTCPCommand(t *testing.T) {
	reset()
	opts := map[string]string{}
	initial := []string{"/bin/ls"}
	opts[ChrootDirectory] = "/test"
	c := NewNVMeTCP(opts)
	command := c.buildNVMeCommand(initial)
	// the length of the resulting command should the length of the initial command +2
	if len(command) != (len(initial) + 2) {
		t.Errorf("Expected to %d items in the command slice but received %v", len(initial)+2, command)
	}
	if command[0] != "chroot" {
		t.Error("Expected the command to be run with chroot")
	}
	if command[1] != opts[ChrootDirectory] {
		t.Errorf("Expected the command to chroot to %s but got %s", opts[ChrootDirectory], command[1])
	}
}

func TestBuildNVMeFCCommand(t *testing.T) {
	reset()
	opts := map[string]string{}
	initial := []string{"/bin/ls"}
	opts[ChrootDirectory] = "/test"
	c := NewNVMeFC(opts)
	command := c.buildNVMeCommand(initial)
	// the length of the resulting command should the length of the initial command +2
	if len(command) != (len(initial) + 2) {
		t.Errorf("Expected to %d items in the command slice but received %v", len(initial)+2, command)
	}
	if command[0] != "chroot" {
		t.Error("Expected the command to be run with chroot")
	}
	if command[1] != opts[ChrootDirectory] {
		t.Errorf("Expected the command to chroot to %s but got %s", opts[ChrootDirectory], command[1])
	}
}

func TestNVMeTCPGetSessions(t *testing.T) {
	reset()
	c := NewNVMeTCP(map[string]string{})
	_, err := c.GetSessions()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestNVMeFCGetSessions(t *testing.T) {
	reset()
	c := NewNVMeFC(map[string]string{})
	_, err := c.GetSessions()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestMockDiscoverTargets(t *testing.T) {
	reset()
	var c NVMEinterface
	opts := map[string]string{}
	expected := 5
	opts[MockNumberOfTargets] = fmt.Sprintf("%d", expected)
	c = NewMockNVMe(opts)
	//c = mock
	targets, err := c.DiscoverNVMeTargets("1.1.1.1", false)
	if err != nil {
		t.Error(err.Error())
	}
	if len(targets) != expected {
		t.Errorf("Expected to find %d targets, but got back %v", expected, targets)
	}
}

func TestMockDiscoverTargetsError(t *testing.T) {
	reset()
	opts := map[string]string{}
	expected := 5
	opts[MockNumberOfTargets] = fmt.Sprintf("%d", expected)
	c := NewMockNVMe(opts)
	GONVMEMock.InduceDiscoveryError = true
	targets, err := c.DiscoverNVMeTargets("1.1.1.1", false)
	if err == nil {
		t.Error("Expected an induced error")
		return
	}
	if !strings.Contains(err.Error(), "induced") {
		t.Error("Expected an induced error")
		return
	}
	if len(targets) != 0 {
		t.Errorf("Expected to receive 0 targets when inducing an error. Received %v", targets)
		return
	}
}

func TestMockGetInitiators(t *testing.T) {
	reset()
	opts := map[string]string{}
	expected := 3
	opts[MockNumberOfInitiators] = fmt.Sprintf("%d", expected)
	c := NewMockNVMe(opts)
	initiators, err := c.GetInitiators("")
	if err != nil {
		t.Error(err.Error())
	}
	if len(initiators) != expected {
		t.Errorf("Expected to find %d initiators, but got back %v", expected, initiators)
	}
}

func TestMockGetInitiatorsError(t *testing.T) {
	reset()
	opts := map[string]string{}
	expected := 3
	opts[MockNumberOfInitiators] = fmt.Sprintf("%d", expected)
	c := NewMockNVMe(opts)
	GONVMEMock.InduceInitiatorError = true
	initiators, err := c.GetInitiators("")
	if err == nil {
		t.Error("Expected an induced error")
		return
	}
	if !strings.Contains(err.Error(), "induced") {
		t.Error("Expected an induced error")
		return
	}
	if len(initiators) != 0 {
		t.Errorf("Expected to receive 0 initiators when inducing an error. Received %v", initiators)
		return
	}
}

func TestNVMeTCPMockLoginLogoutTargets(t *testing.T) {
	reset()
	c := NewMockNVMe(map[string]string{})
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
	err := c.NVMeConnect(tgt)
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = c.NVMeDisconnect(tgt)
	if err != nil {
		t.Error(err.Error())
		return
	}
}

func TestMockLogoutTargetsError(t *testing.T) {
	reset()
	c := NewMockNVMe(map[string]string{})
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
	}
	GONVMEMock.InduceLogoutError = true
	err := c.NVMeConnect(tgt)
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = c.NVMeDisconnect(tgt)
	if err == nil {
		t.Error("Expected an induced error")
		return
	}
	if !strings.Contains(err.Error(), "induced") {
		t.Error("Expected an induced error")
		return
	}
}

func TestMockLoginTargetsError(t *testing.T) {
	reset()
	c := NewMockNVMe(map[string]string{})
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
	GONVMEMock.InduceLoginError = true
	err := c.NVMeConnect(tgt)
	if err == nil {
		t.Error("Expected an induced error")
		return
	}
	if !strings.Contains(err.Error(), "induced") {
		t.Error("Expected an induced error")
		return
	}
}

func TestMockGetSessions(t *testing.T) {
	reset()
	c := NewMockNVMe(map[string]string{})
	// check without induced error
	data, err := c.GetSessions()
	if len(data) == 0 || len(data[0].Target) == 0 {
		t.Error("invalid response from mock")
	}
	if err != nil {
		t.Error(err.Error())
		return
	}
}

func TestMockGetSessionsError(t *testing.T) {
	reset()
	c := NewMockNVMe(map[string]string{})
	// check with induced error
	GONVMEMock.InduceGetSessionsError = true
	_, err := c.GetSessions()
	if err == nil {
		t.Error("Expected an induced error")
		return
	}
	if !strings.Contains(err.Error(), "induced") {
		t.Error("Expected an induced error")
		return
	}
}

func TestSessionParserParse(t *testing.T) {
	sp := &sessionParser{}
	fileErrMsg := "can't read file with test data"

	// test valid data
	data, err := ioutil.ReadFile("testdata/session_info_valid")
	if err != nil {
		t.Error(fileErrMsg)
	}
	sessions := sp.Parse(data)
	if len(sessions) != 2 {
		t.Error("unexpected results count")
	}
	for i, session := range sessions {
		if i == 0 {
			compareStr(t, session.Target, "nqn.1988-11.com.dell.mock:00:e6e2d5b871f1403E169D")
			compareStr(t, session.Portal, "10.230.1.1:4420")
			compareStr(t, string(session.NVMESessionState), string(NVMESessionStateLive))
			compareStr(t, string(session.NVMETransportName), string(NVMETransportNameTCP))
		} else {
			compareStr(t, session.Target, "nqn.1988-11.com.dell.mock:00:e6e2d5b871f1403E169D")
			compareStr(t, session.Portal, "10.230.1.2:4420")
			compareStr(t, string(session.NVMESessionState), string(NVMESessionStateDeleting))
			compareStr(t, string(session.NVMETransportName), string(NVMETransportNameTCP))
		}
	}

	// test invalid data parsing
	data, err = ioutil.ReadFile("testdata/session_info_invalid")
	if err != nil {
		t.Error(fileErrMsg)
	}
	r := sp.Parse(data)
	if len(r) != 0 {
		t.Error("non empty result while parsing invalid data")
	}
}

func compareStr(t *testing.T, str1 string, str2 string) {
	if str1 != str2 {
		t.Errorf("strings are not equal: %s != %s", str1, str2)
	}
}
