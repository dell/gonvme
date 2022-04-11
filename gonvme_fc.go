package gonvme

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"
)

// NVMeFC provides many nvme/fc-specific functions
type NVMeFC struct {
	NVMeType
	sessionParser NVMeSessionParser
}

// NewNVMeFC - returns a new NVMeTCP client
func NewNVMeFC(opts map[string]string) *NVMeFC {
	nvme := NVMeFC{
		NVMeType: NVMeType{
			mock:    false,
			options: opts,
		},
	}
	nvme.sessionParser = &sessionParser{}
	return &nvme
}

func (nvme *NVMeFC) getChrootDirectory() string {
	s := nvme.options[ChrootDirectory]
	if s == "" {
		s = "/"
	}
	return s
}

func (nvme *NVMeFC) buildNVMeCommand(cmd []string) []string {
	if nvme.getChrootDirectory() == "/" {
		return cmd
	}
	command := []string{"chroot", nvme.getChrootDirectory()}
	command = append(command, cmd...)
	return command
}

func (nvme *NVMeFC) getFCHostInfo() ([]FCHBAInfo, error) {

	match, err := filepath.Glob("/sys/class/fc_host/host*")
	if err != nil {
		fmt.Printf("Error gathering fc hosts: %v", err)
		return []FCHBAInfo{}, err
	}

	var FCHostsInfo []FCHBAInfo
	for _, m := range match {

		var FCHostInfo FCHBAInfo
		data, err := ioutil.ReadFile(path.Join(m, "port_name"))
		if err != nil {
			fmt.Printf("match: %s failed to read port_name file: %s", match, err.Error())
			continue
		}
		FCHostInfo.PortName = strings.TrimSpace(string(data))

		data, err = ioutil.ReadFile(path.Join(m, "node_name"))
		if err != nil {
			fmt.Printf("match: %s failed to read node_name file: %s", match, err.Error())
			continue
		}
		FCHostInfo.NodeName = strings.TrimSpace(string(data))
		FCHostsInfo = append(FCHostsInfo, FCHostInfo)
	}

	if len(FCHostsInfo) == 0 {
		return []FCHBAInfo{}, err
	}
	return FCHostsInfo, nil
}

// GetInitiators returns a list of initiators on the local system.
func (nvme *NVMeFC) GetInitiators(filename string) ([]string, error) {
	return nvme.getInitiators(filename)
}

func (nvme *NVMeFC) getInitiators(filename string) ([]string, error) {

	// a slice of filename, which might exist and define the nvme initiators
	initiatorConfig := []string{}
	nqns := []string{}

	if filename == "" {
		// add default filename(s) here
		// /etc/nvme/hostnqn is the proper file for CentOS, RedHat, Sles, Ubuntu
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

// DiscoverNVMeTargets - runs NVMe/FC discovery and returns a list of targets.
func (nvme *NVMeFC) DiscoverNVMeTargets(targetAddress string, login bool) ([]NVMeTarget, error) {
	return nvme.discoverNVMeTargets(targetAddress, login)
}

func (nvme *NVMeFC) discoverNVMeTargets(targetAddress string, login bool) ([]NVMeTarget, error) {
	// TODO: add injection check on address
	// nvme discovery is done via nvme cli
	// nvme discover -t fc -a traddr -w host_traddr
	// where traddr = nn-<Target_WWNN>:pn-<Target_WWPN> and host_traddr = nn-<Initiator_WWNN>:pn-<Initiator_WWPN>

	var out []byte
	FCHostsInfo, err := nvme.getFCHostInfo()
	if err != nil {
		fmt.Printf("Error gathering NVMe/FC Hosts on the host side: %v", err)
		return []NVMeTarget{}, nil
	}

	targets := make([]NVMeTarget, 0)
	for _, FCHostInfo := range FCHostsInfo {

		initiatorAddress := strings.Replace(fmt.Sprintf("nn-%s:pn-%s", FCHostInfo.NodeName, FCHostInfo.PortName), "\n", "", -1)
		exe := nvme.buildNVMeCommand([]string{NVMeCommand, "discover", "-t", "fc", "-a", targetAddress, "-w", initiatorAddress})
		cmd := exec.Command(exe[0], exe[1:]...)

		out, err = cmd.Output()
		if err != nil {
			continue
		}

		nvmeTarget := NVMeTarget{}
		entryCount := 0
		skipIteration := false

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

			tokens := strings.Fields(line)
			if len(tokens) < 2 {
				continue
			}
			key := tokens[0]
			value := strings.Join(tokens[1:], "")
			switch key {

			case "=====Discovery":
				// add to array
				if entryCount != 0 && !skipIteration {
					targets = append(targets, nvmeTarget)
				}
				nvmeTarget = NVMeTarget{}
				nvmeTarget.HostAdr = initiatorAddress
				skipIteration = false
				entryCount++
				continue

			case "trtype:":
				nvmeTarget.TargetType = value
				if value != NVMeTransportTypeFC {
					skipIteration = true
				}
				break

			case "traddr:":
				nvmeTarget.Portal = value
				break

			case "subnqn:":
				nvmeTarget.TargetNqn = value
				break

			case "adrfam:":
				nvmeTarget.AdrFam = value
				break

			case "subtype:":
				nvmeTarget.SubType = value
				break

			case "treq:":
				nvmeTarget.Treq = value
				break

			case "portid:":
				nvmeTarget.PortID = value
				break

			case "trsvcid:":
				nvmeTarget.TrsvcID = value
				break

			case "sectype:":
				nvmeTarget.SecType = value
				break

			}
		}
		if !skipIteration && nvmeTarget.TargetNqn != "" {
			targets = append(targets, nvmeTarget)
		}
	}

	if len(targets) == 0 {
		fmt.Printf("\n Error discovering NVMe/FC targets: %s", err)
		return []NVMeTarget{}, err
	}

	// TODO: Add optional login
	// log into the target if asked
	/*if login {
		for _, t := range targets {
			gonvme.PerformLogin(t)
		}
	}*/

	return targets, nil
}

// NVMeConnect will attempt to connect into a given nvme target
func (nvme *NVMeFC) NVMeConnect(target NVMeTarget) error {
	return nvme.nvmeConnect(target)
}

func (nvme *NVMeFC) nvmeConnect(target NVMeTarget) error {
	// nvme connect is done via the nvme cli
	// nvme connect -t tcp -n <target NQN> -a <NVMe interface IP> -s 4420
	exe := nvme.buildNVMeCommand([]string{NVMeCommand, "connect", "-t", "fc", "-a", target.Portal, "-w", target.HostAdr, "-n", target.TargetNqn})
	cmd := exec.Command(exe[0], exe[1:]...)
	_, err := cmd.Output()

	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// nvme connect exited with an exit code != 0
			nvmeConnectResult := -1
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				nvmeConnectResult = status.ExitStatus()
			}
			if nvmeConnectResult == 114 || nvmeConnectResult == 70 {
				// session already exists
				// do not treat this as a failure
				fmt.Printf("NVMe/FC connection already exists\n")
				err = nil
			} else {
				fmt.Printf("\nnvme connect failure: %v", err)
			}
		} else {
			fmt.Printf("\nError during NVMe/FC connect %s at %s for %s host: %v", target.TargetNqn, target.Portal, target.HostAdr, err)
		}

		if err != nil {
			fmt.Printf("\nError during NVMe/FC connect %s at %s for %s host: %v", target.TargetNqn, target.Portal, target.HostAdr, err)
			return err
		}
	} else {
		fmt.Printf("\nNVMe/FC connect successful: %s", target.TargetNqn)
	}

	return nil
}

// NVMeDisconnect will attempt to disconnect from a given nvme target
func (nvme *NVMeFC) NVMeDisconnect(target NVMeTarget) error {
	return nvme.nvmeDisconnect(target)
}

func (nvme *NVMeFC) nvmeDisconnect(target NVMeTarget) error {
	// nvme disconnect is done via the nvme cli
	// nvme disconnect -n <target NQN>
	exe := nvme.buildNVMeCommand([]string{NVMeCommand, "disconnect", "-n", target.TargetNqn})
	cmd := exec.Command(exe[0], exe[1:]...)

	_, err := cmd.Output()

	if err != nil {
		fmt.Printf("\nError logging %s at %s: %v", target.TargetNqn, target.Portal, err)
	} else {
		fmt.Printf("\nnvme disconnect successful: %s", target.TargetNqn)
	}

	return err
}

// GetSessions queries information about  NVMe sessions
func (nvme *NVMeFC) GetSessions() ([]NVMESession, error) {
	exe := nvme.buildNVMeCommand([]string{"nvme", "list-subsys", "-o", "json"})
	cmd := exec.Command(exe[0], exe[1:]...)
	output, err := cmd.Output()
	if err != nil {
		if isNoObjsExitCode(err) {
			return []NVMESession{}, nil
		}
		return []NVMESession{}, err
	}
	return nvme.sessionParser.Parse(output), nil
}

// ifNVVMeFCDiscover checks if the NVMe/FC discover and connect is possible with the given initiator and target address
func (nvme *NVMeFC) ifNVVMeFCDiscover(initiatorAddress string, targetAddress string) bool {

	exe := nvme.buildNVMeCommand([]string{NVMeCommand, "discover", "-t", "fc", "-a", targetAddress, "-w", initiatorAddress})
	cmd := exec.Command(exe[0], exe[1:]...)

	_, err := cmd.Output()
	if err != nil {
		return false
	}
	return true
}
