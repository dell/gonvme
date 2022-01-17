package gonvme

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	// ChrootDirectory allows the iscsiadm commands to be run within a chrooted path, helpful for containerized services
	ChrootDirectory = "chrootDirectory"

	// DefaultInitiatorNameFile is the default file which contains the initiator nqn
	DefaultInitiatorNameFile = "/etc/nvme/hostnqn"

	// NVMeCommand - nvme command
	NVMeCommand = "nvme"

	// NVMePort - port number
	NVMePort = "4420"
)

// NVMeTCP provides many iSCSI-specific functions
type NVMeTCP struct {
	NVMeType
}

// NewNVMeTCP - returns a new NVMeTCP client
func NewNVMeTCP(opts map[string]string) *NVMeTCP {
	nvme := NVMeTCP{
		NVMeType: NVMeType{
			mock:    false,
			options: opts,
		},
	}

	return &nvme
}

func (nvme *NVMeTCP) getChrootDirectory() string {
	s := nvme.options[ChrootDirectory]
	if s == "" {
		s = "/"
	}
	return s
}

func (nvme *NVMeTCP) buildNVMeCommand(cmd []string) []string {
	if nvme.getChrootDirectory() == "/" {
		return cmd
	}
	command := []string{"chroot", nvme.getChrootDirectory()}
	command = append(command, cmd...)
	return command
}

// DiscoverNVMeTCPTargets - runs nvme discovery and returns a list of targets.
func (nvme *NVMeTCP) DiscoverNVMeTCPTargets(address string, login bool) ([]NVMeTarget, error) {
	return nvme.discoverNVMeTCPTargets(address, login)
}

func (nvme *NVMeTCP) discoverNVMeTCPTargets(address string, login bool) ([]NVMeTarget, error) {
	// TODO: add injection check on address
	// nvme discovery is done via nvme cli
	// nvme discover -t tcp -a <NVMe interface IP> -s <port>
	exe := nvme.buildNVMeCommand([]string{NVMeCommand, "discover", "-t", "tcp", "-a", address, "-s", NVMePort})
	cmd := exec.Command(exe[0], exe[1:]...)

	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("\nError discovering %s: %v", address, err)
		return []NVMeTarget{}, err
	}

	targets := make([]NVMeTarget, 0)
	nvmeTarget := NVMeTarget{}
	entryCount := 0

	for _, line := range strings.Split(string(out), "\n") {
		// Output should look like:

		// Discovery Log Number of Records 2, Generation counter 2
		// =====Discovery Log Entry 0======
		// trtype:  fc
		// adrfam:  fibre-channel
		// subtype: nvme subsystem
		// treq:    not specified
		// portid:  0
		// trsvcid: none
		// subnqn:  nqn.1111-11.com.dell:powerstore:00:a1a1a1a111a1111a111a
		// traddr:  nn-0x11aaa111a1111a11:aa-0x11aaa11111111a11
		//
		// =====Discovery Log Entry 1======
		// trtype:  tcp
		// adrfam:  ipv4
		// subtype: nvme subsystem
		// treq:    not specified
		// portid:  2304
		// trsvcid: 4420
		// subnqn:  nqn.1111-11.com.dell:powerstore:00:a1a1a1a111a1111a111a
		// traddr:  1.1.1.1
		// sectype: none

		tokens := strings.Split(line, " ")
		key := tokens[0]
		value := strings.Join(tokens[1:], "")
		fmt.Println(value)
		switch key {

		case "=====Discovery":
			// add to array
			if entryCount != 0 {
				targets = append(targets, nvmeTarget)
				nvmeTarget = NVMeTarget{}
			}
			entryCount++
			continue

		case "trtype:":
			nvmeTarget.TargetType = value
			break

		case "traddr:":
			nvmeTarget.Portal = value
			break

		case "subnqn:":
			nvmeTarget.TargetNqn = value
			break

		case "adrfam:":
			nvmeTarget.AdrFam = tokens[1]
			break

		case "subtype:":
			nvmeTarget.SubType = tokens[1]
			break

		case "treq:":
			nvmeTarget.Treq = tokens[1]
			break

		case "portid:":
			nvmeTarget.PortID = tokens[1]
			break

		case "trsvcid:":
			nvmeTarget.TrsvcID = tokens[1]
			break

		case "sectype:":
			nvmeTarget.SecType = tokens[1]
			break

		default:
			fmt.Println("----------------------")
		}
	}
	targets = append(targets, nvmeTarget)

	nvmeTCPTargets := make([]NVMeTarget, 0)

	for _, target := range targets {
		if target.TargetType == NVMeNVMeTransportTypeTCP {
			nvmeTCPTargets = append(nvmeTCPTargets, target)
		}
	}

	// TODO: Add optional login
	// log into the target if asked
	/*if login {
		for _, t := range targets {
			iscsi.PerformLogin(t)
		}
	}*/

	return nvmeTCPTargets, nil
}

// GetInitiators returns a list of initiators on the local system.
func (nvme *NVMeTCP) GetInitiators(filename string) ([]string, error) {
	return nvme.getInitiators(filename)
}

func (nvme *NVMeTCP) getInitiators(filename string) ([]string, error) {

	// a slice of filename, which might exist and define the nvme initiators
	initiatorConfig := []string{}
	nqns := []string{}

	if filename == "" {
		// add default filename(s) here
		// /etc/iscsi/initiatorname.iscsi is the proper file for CentOS, RedHat, Debian, Ubuntu
		if nvme.getChrootDirectory() != "/" {
			initiatorConfig = append(initiatorConfig, nvme.getChrootDirectory()+"/"+DefaultInitiatorNameFile)
		} else {
			initiatorConfig = append(initiatorConfig, DefaultInitiatorNameFile)
		}
	} else {
		initiatorConfig = append(initiatorConfig, filename)
	}

	var err error
	// for each initiatior config file
	for _, init := range initiatorConfig {
		// make sure the file exists
		_, err = os.Stat(init)
		if err != nil {
			continue
		}

		// get the contents of the initiator config file
		// TODO: check if sys call is available for cat command
		cmd := exec.Command("cat", init)

		out, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error gathering initiator names: %v", err)
		}
		lines := strings.Split(string(out), "\n")

		for _, line := range lines {

			if line != "" {
				nqns = append(nqns, line)
			}
		}
	}

	if len(nqns) == 0 {
		return nqns, err
	}

	return nqns, nil
}
