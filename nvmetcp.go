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

	// nvmeNoObjsFoundExitCode exit code indicates that no records/targets/sessions/portals
	// found to execute operation on
	NVMeNoObjsFoundExitCode = 21

	// NVMeCommand
	NVMeCommand = "nvme"

	// NVMePort
	NVMePort = "4420"
)

// NVMeTCP provides many iSCSI-specific functions
type NVMeTCP struct {
	NVMeType
}

// NewLinuxNVMe returns an LinuxNVMe client
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

// DiscoverTargets runs nvme discovery and returns a list of targets.
func (nvme *NVMeTCP) DiscoverNVMeTCPTargets(address string, login bool) ([]NVMeTarget, error) {
	return nvme.discoverNVMeTCPTargets(address, login)
}

func (iscsi *NVMeTCP) discoverNVMeTCPTargets(address string, login bool) ([]NVMeTarget, error) {
	// TODO: add injection check on address
	// iSCSI discovery is done via the iscsiadm cli
	// iscsiadm -m discovery -t st --portal <target>
	exe := iscsi.buildNVMeCommand([]string{NVMeCommand, "discover", "-t", "tcp", "-a", address, "-s", NVMePort})
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

		tokens := strings.Split(line, " (.+)")

		switch tokens[0] {

		case "=====Discovery":
			// add to array
			if entryCount != 0 {
				targets = append(targets, nvmeTarget)
				nvmeTarget = NVMeTarget{}
			}
			entryCount++
			continue

		case "trtype:":
			if tokens[1] != "tcp" {
				continue
			} else {
				break
			}

		case "traddr:":
			nvmeTarget.Portal = tokens[1]
			break

		case "subnqn:":
			nvmeTarget.TargetNqn = tokens[1]
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
		}
	}
	targets = append(targets, nvmeTarget)

	// TODO: Add optional login
	// log into the target if asked
	/*if login {
		for _, t := range targets {
			iscsi.PerformLogin(t)
		}
	}*/

	return targets, nil
}

// GetInitiators returns a list of initiators on the local system.
func (iscsi *LinuxNVMe) GetInitiators(filename string) ([]string, error) {
	return iscsi.getInitiators(filename)
}

func (iscsi *LinuxNVMe) getInitiators(filename string) ([]string, error) {

	// a slice of filename, which might exist and define the nvme initiators
	initiatorConfig := []string{}
	nqns := []string{}

	if filename == "" {
		// add default filename(s) here
		// /etc/iscsi/initiatorname.iscsi is the proper file for CentOS, RedHat, Debian, Ubuntu
		if iscsi.getChrootDirectory() != "/" {
			initiatorConfig = append(initiatorConfig, iscsi.getChrootDirectory()+"/"+DefaultInitiatorNameFile)
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
		nqns = append(nqns, lines...)
	}

	if len(nqns) == 0 {
		return nqns, err
	}

	return nqns, nil
}
