/*
 *
 * Copyright © 2022-2024 Dell Inc. or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *      http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package gonvme

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

const (
	// ChrootDirectory allows the nvme commands to be run within a chrooted path, helpful for containerized services
	ChrootDirectory = "chrootDirectory"

	// NVMePort - port number
	NVMePort = "4420"

	// NVMEAlreadyConnected contains output holder for nvme connect
	NVMEAlreadyConnected = "already connected"

	// NVMeNoObjsFoundExitCode exit code indicates that no records/targets/sessions/portals
	// found to execute operation on
	NVMeNoObjsFoundExitCode = 21
)

var (
	fcHostPath = "/sys/class/fc_host/host*"

	// DefaultInitiatorNameFile is the default file which contains the initiator nqn
	DefaultInitiatorNameFile = "/etc/nvme/hostnqn"
)

type command interface {
	Output() ([]byte, error)
	Start() error
	Wait() error
	StderrPipe() (io.ReadCloser, error)
}

var getCommand = func(name string, arg ...string) command {
	return exec.Command(name, arg...)
}

// NVMe provides many nvme-specific functions
type NVMe struct {
	NVMeType
	sessionParser NVMeSessionParser
	NVMeCommand   string
}

// NewNVMe - returns a new NVMe client
func NewNVMe(opts map[string]string) *NVMe {
	nvme := NVMe{
		NVMeType: NVMeType{
			mock:    false,
			options: opts,
		},
	}
	nvme.sessionParser = &sessionParser{}

	paths := []string{"/sbin/nvme"}
	for _, path := range paths {
		pathCopy := path
		if nvme.getChrootDirectory() != "/" {
			path = nvme.getChrootDirectory() + path
		}

		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			log.Errorf("Error: Path %s does not exist\n", path)
		} else if err != nil {
			log.Errorf("Error: Unable to access path %s: %v\n", path, err)
		} else if info.IsDir() {
			log.Errorf("Error: Path %s is a directory, not an executable\n", path)
		} else {
			log.Infof("Success: Path %s exists and is an executable\n", path)
			nvme.NVMeCommand = pathCopy
			log.Infof("nvme.NVMeCommand: %s", nvme.NVMeCommand)
			break
		}
	}

	return &nvme
}

func (nvme *NVMe) getChrootDirectory() string {
	s := nvme.options[ChrootDirectory]
	if s == "" {
		s = "/"
	}
	return s
}

func (nvme *NVMe) buildNVMeCommand(cmd []string) []string {
	if nvme.getChrootDirectory() == "/" {
		return cmd
	}
	command := []string{"chroot", nvme.getChrootDirectory()}
	command = append(command, cmd...)
	return command
}

func (nvme *NVMe) getFCHostInfo() ([]FCHBAInfo, error) {
	match, err := filepath.Glob(fcHostPath)
	if err != nil {
		log.Errorf("Error gathering fc hosts: %v", err)
		return []FCHBAInfo{}, err
	}
	if len(match) == 0 {
		log.Errorf("The fc_host path doesn't exist")
		return []FCHBAInfo{}, err
	}

	var FCHostsInfo []FCHBAInfo
	for _, m := range match {

		var FCHostInfo FCHBAInfo
		portNamePath := path.Join(m, "port_name")
		data, err := os.ReadFile(filepath.Clean(portNamePath))
		if err != nil {
			log.Errorf("match: %s failed to read port_name file: %s", match, err.Error())
			continue
		}
		FCHostInfo.PortName = strings.TrimSpace(string(data))

		nodeNamePath := path.Join(m, "node_name")
		data, err = os.ReadFile(filepath.Clean(nodeNamePath))
		if err != nil {
			log.Errorf("match: %s failed to read node_name file: %s", match, err.Error())
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

// DiscoverNVMeTCPTargets - runs nvme discovery and returns a list of NVMeTCP targets.
func (nvme *NVMe) DiscoverNVMeTCPTargets(address string, login bool) ([]NVMeTarget, error) {
	return nvme.discoverNVMeTCPTargets(address, login)
}

func (nvme *NVMe) discoverNVMeTCPTargets(address string, login bool) ([]NVMeTarget, error) {
	// TODO: add injection check on address
	// nvme discovery is done via nvme cli
	// nvme discover -t tcp -a <NVMe interface IP> -s <port>
	exe := nvme.buildNVMeCommand([]string{nvme.NVMeCommand, "discover", "-t", "tcp", "-a", address, "-s", NVMePort})
	cmd := getCommand(exe[0], exe[1:]...) // #nosec G204

	out, err := cmd.Output()
	if err != nil {
		log.Errorf("\nError discovering %s: %v", address, err)
		return []NVMeTarget{}, err
	}

	targets := make([]NVMeTarget, 0)
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
		value := strings.Join(tokens[1:], " ")
		switch key {

		case "=====Discovery":
			// add to array
			if entryCount != 0 && !skipIteration {
				targets = append(targets, nvmeTarget)
			}
			nvmeTarget = NVMeTarget{}
			skipIteration = false
			entryCount++
			continue

		case "trtype:":
			nvmeTarget.TargetType = value
			if value != NVMeTransportTypeTCP {
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

		default:

		}
	}
	if !skipIteration && nvmeTarget.TargetNqn != "" {
		targets = append(targets, nvmeTarget)
	}

	// TODO: Add optional login
	// log into the target if asked
	if login {
		for _, t := range targets {
			err = nvme.NVMeTCPConnect(t, false)
			if err != nil {
				log.Errorf("Error during NVMeTCP connect")
			}
		}
	}

	return targets, nil
}

// DiscoverNVMeFCTargets - runs nvme discovery and returns a list of NVMeFC targets.
func (nvme *NVMe) DiscoverNVMeFCTargets(targetAddress string, login bool) ([]NVMeTarget, error) {
	return nvme.discoverNVMeFCTargets(targetAddress, login)
}

func (nvme *NVMe) discoverNVMeFCTargets(targetAddress string, login bool) ([]NVMeTarget, error) {
	// TODO: add injection check on address
	// nvme discovery is done via nvme cli
	// nvme discover -t fc -a traddr -w host_traddr
	// where traddr = nn-<Target_WWNN>:pn-<Target_WWPN> and host_traddr = nn-<Initiator_WWNN>:pn-<Initiator_WWPN>

	var out []byte
	FCHostsInfo, err := nvme.getFCHostInfo()
	if err != nil || len(FCHostsInfo) == 0 {
		log.Errorf("Error gathering NVMe/FC Hosts on the host side: %v", err)
		return []NVMeTarget{}, err
	}

	targets := make([]NVMeTarget, 0)
	for _, FCHostInfo := range FCHostsInfo {

		// host_traddr = nn-<Initiator_WWNN>:pn-<Initiator_WWPN>
		initiatorAddress := strings.Replace(fmt.Sprintf("nn-%s:pn-%s", FCHostInfo.NodeName, FCHostInfo.PortName), "\n", "", -1)
		exe := nvme.buildNVMeCommand([]string{nvme.NVMeCommand, "discover", "-t", "fc", "-a", targetAddress, "-w", initiatorAddress})
		cmd := getCommand(exe[0], exe[1:]...) // #nosec G204

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
				if entryCount != 0 && !skipIteration && nvmeTarget.Portal == targetAddress {
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
		if !skipIteration && nvmeTarget.TargetNqn != "" && nvmeTarget.Portal == targetAddress {
			targets = append(targets, nvmeTarget)
		}
	}

	if len(targets) == 0 {
		log.Errorf("Error discovering NVMe/FC targets: %v", err)
		return []NVMeTarget{}, err
	}

	// TODO: Add optional login
	// log into the target if asked
	if login {
		for _, t := range targets {
			err = nvme.NVMeFCConnect(t, false)
			if err != nil {
				log.Errorf("Error during NVMeFC connect")
			}
		}
	}

	return targets, nil
}

// GetInitiators returns a list of initiators on the local system.
func (nvme *NVMe) GetInitiators(filename string) ([]string, error) {
	return nvme.getInitiators(filename)
}

func (nvme *NVMe) getInitiators(filename string) ([]string, error) {
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
		out, err := os.ReadFile(filepath.Clean(init))
		if err != nil {
			log.Errorf("Error gathering initiator names: %v", err)
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

// NVMeTCPConnect will attempt to connect into a given NVMeTCP target
func (nvme *NVMe) NVMeTCPConnect(target NVMeTarget, duplicateConnect bool) error {
	return nvme.nvmeTCPConnect(target, duplicateConnect)
}

func (nvme *NVMe) nvmeTCPConnect(target NVMeTarget, duplicateConnect bool) error {
	// nvme connect is done via the nvme cli
	// nvme connect -t tcp -n <target NQN> -a <NVMe interface IP> -s 4420
	// D allows duplicate connections between same transport host and subsystem port
	var exe []string
	if duplicateConnect {
		exe = nvme.buildNVMeCommand([]string{nvme.NVMeCommand, "connect", "-t", "tcp", "-n", target.TargetNqn, "-a", target.Portal, "-s", NVMePort, "--ctrl-loss-tmo=-1", "-D"})
	} else {
		exe = nvme.buildNVMeCommand([]string{nvme.NVMeCommand, "connect", "-t", "tcp", "-n", target.TargetNqn, "-a", target.Portal, "-s", NVMePort, "--ctrl-loss-tmo=-1"})
	}
	cmd := getCommand(exe[0], exe[1:]...) // #nosec G204
	var Output string
	stderr, _ := cmd.StderrPipe()
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("starting nvme connect %s at %s: %v", target.TargetNqn, target.Portal, err)
	}

	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		Output = scanner.Text()
	}
	log.Debugf("connect output: %s", Output)
	err = cmd.Wait()

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
				// this is applicable if nvme cli version 1.16 or below
				if Output == "Failed to write to /dev/nvme-fabrics: Operation already in progress" || Output == "" {
					log.Infof("NVMe connection already exists\n")
					err = nil
				} else {
					log.Errorf("\nError during nvme connect %s at %s: %v", target.TargetNqn, target.Portal, err)
					return err
				}
			} else if nvmeConnectResult == 1 && strings.Contains(Output, NVMEAlreadyConnected) {
				// session already exists
				// this is applicable if nvme cli version is 2.0 and above
				log.Infof("NVMe connection already exists\n")
				err = nil
			} else {
				log.Errorf("\nnvme connect failure: %v, %s", err, err.Error())
			}
		} else {
			log.Errorf("\nError during nvme connect %s at %s: %v", target.TargetNqn, target.Portal, err)
		}

		if err != nil {
			log.Errorf("\nError during nvme connect %s at %s: %v", target.TargetNqn, target.Portal, err)
			return err
		}
	} else {
		log.Infof("\nnvme connect successful: %s", target.TargetNqn)
	}

	return nil
}

// NVMeFCConnect will attempt to connect into a given NVMeFC target
func (nvme *NVMe) NVMeFCConnect(target NVMeTarget, duplicateConnect bool) error {
	return nvme.nvmeFCConnect(target, duplicateConnect)
}

func (nvme *NVMe) nvmeFCConnect(target NVMeTarget, duplicateConnect bool) error {
	// nvme connect is done via the nvme cli
	// nvme connect -t fc -a traddr -w host_traddr -n target_nqn
	// where traddr = nn-<Target_WWNN>:pn-<Target_WWPN> and host_traddr = nn-<Initiator_WWNN>:pn-<Initiator_WWPN>
	// D allows duplicate connections between same transport host and subsystem port
	var exe []string
	if duplicateConnect {
		exe = nvme.buildNVMeCommand([]string{nvme.NVMeCommand, "connect", "-t", "fc", "-a", target.Portal, "-w", target.HostAdr, "-n", target.TargetNqn, "--ctrl-loss-tmo=-1", "-D"})
	} else {
		exe = nvme.buildNVMeCommand([]string{nvme.NVMeCommand, "connect", "-t", "fc", "-a", target.Portal, "-w", target.HostAdr, "-n", target.TargetNqn, "--ctrl-loss-tmo=-1"})
	}
	cmd := getCommand(exe[0], exe[1:]...) // #nosec G204
	var Output string
	stderr, _ := cmd.StderrPipe()
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("starting nvme connect %s at %s: %v", target.TargetNqn, target.Portal, err)
	}

	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		Output = scanner.Text()
	}
	err = cmd.Wait()

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
				// this is applicable if nvme cli version 1.16 or below
				if Output == "Failed to write to /dev/nvme-fabrics: Operation already in progress" || Output == "" {
					log.Infof("NVMe connection already exists\n")
					err = nil
				} else {
					log.Errorf("\nError during nvme connect %s at %s: %v", target.TargetNqn, target.Portal, err)
					return err
				}
			} else if nvmeConnectResult == 1 && strings.Contains(Output, NVMEAlreadyConnected) {
				// session already exists
				// this is applicable if nvme cli version is 2.0 and above
				log.Infof("NVMe connection already exists\n")
				err = nil
			} else {
				log.Errorf("NVMe/FC connect failure: %v", err)
			}
		} else {
			log.Errorf("Error during NVMe/FC connect %s at %s for %s host: %v", target.TargetNqn, target.Portal, target.HostAdr, err)
		}

		if err != nil {
			log.Errorf("Error during NVMe/FC connect %s at %s for %s host: %v", target.TargetNqn, target.Portal, target.HostAdr, err)
			return err
		}
	} else {
		log.Infof("NVMe/FC connect successful: %s", target.TargetNqn)
	}

	return nil
}

// NVMeDisconnect will attempt to disconnect from a given nvme target
func (nvme *NVMe) NVMeDisconnect(target NVMeTarget) error {
	return nvme.nvmeDisconnect(target)
}

func (nvme *NVMe) nvmeDisconnect(target NVMeTarget) error {
	// nvme disconnect is done via the nvme cli
	// nvme disconnect -n <target NQN>
	exe := nvme.buildNVMeCommand([]string{nvme.NVMeCommand, "disconnect", "-n", target.TargetNqn})
	cmd := getCommand(exe[0], exe[1:]...) // #nosec G204

	_, err := cmd.Output()

	if err != nil {
		log.Errorf("\nError during NVMe disconnect %s at %s: %v", target.TargetNqn, target.Portal, err)
	} else {
		log.Infof("\nnvme disconnect successful: %s", target.TargetNqn)
	}

	return err
}

// ListNVMeDeviceAndNamespace returns the NVME Device Paths and Namespace of each of the NVME device
func (nvme *NVMe) ListNVMeDeviceAndNamespace() ([]DevicePathAndNamespace, error) {
	/* ListNVMeDeviceAndNamespace Output
	{/dev/nvme0n1 54}
	{/dev/nvme0n2 55}
	{/dev/nvme1n1 54}
	{/dev/nvme1n2 55}
	*/
	exe := nvme.buildNVMeCommand([]string{"nvme", "list", "-o", "json"})

	/* nvme list -o json
	{
	  "Devices" : [
	    {
	      "NameSpace" : 9217,
	      "DevicePath" : "/dev/nvme0n1",
	      "Firmware" : "2.1.0.0",
	      "Index" : 0,
	      "ModelNumber" : "dellemc",
	      "SerialNumber" : "FP08RZ2",
	      "UsedBytes" : 0,
	      "MaximumLBA" : 10485760,
	      "PhysicalSize" : 5368709120,
	      "SectorSize" : 512
	    },
	    {
	      "NameSpace" : 9222,
	      "DevicePath" : "/dev/nvme0n2",
	      "Firmware" : "2.1.0.0",
	      "Index" : 0,
	      "ModelNumber" : "dellemc",
	      "SerialNumber" : "FP08RZ2",
	      "UsedBytes" : 0,
	      "MaximumLBA" : 10485760,
	      "PhysicalSize" : 5368709120,
	      "SectorSize" : 512
	    }
	  ]
	}
	*/
	cmd := getCommand(exe[0], exe[1:]...) // #nosec G204

	output, err := cmd.Output()
	if err != nil {
		return []DevicePathAndNamespace{}, err
	}

	str := string(output)
	lines := strings.Split(str, "\n")

	var result []DevicePathAndNamespace
	var currentPathAndNamespace *DevicePathAndNamespace
	var devicePath string
	var namespace string

	for _, line := range lines {

		line = strings.ReplaceAll(strings.TrimSpace(line), ",", "")

		switch {
		case strings.HasPrefix(line, "\"NameSpace\""):
			if len(strings.Split(line, ":")) >= 2 {
				namespace = strings.ReplaceAll(strings.TrimSpace(strings.Split(line, ":")[1]), "\"", "")
			}

		case strings.HasPrefix(line, "\"DevicePath\""):
			if len(strings.Split(line, ":")) >= 2 {
				devicePath = strings.ReplaceAll(strings.TrimSpace(strings.Split(line, ":")[1]), "\"", "")

				PathAndNamespace := DevicePathAndNamespace{}
				PathAndNamespace.Namespace = namespace
				PathAndNamespace.DevicePath = devicePath

				if currentPathAndNamespace != nil {
					result = append(result, *currentPathAndNamespace)
				}
				currentPathAndNamespace = &PathAndNamespace
			}
		}
	}
	if currentPathAndNamespace != nil {
		result = append(result, *currentPathAndNamespace)
	}

	return result, nil
}

// ListNVMeNamespaceID returns the namespace IDs for each NVME device path
func (nvme *NVMe) ListNVMeNamespaceID(NVMeDeviceAndNamespace []DevicePathAndNamespace) (map[DevicePathAndNamespace][]string, error) {
	/* ListNVMeNamespaceID Output
	{devicePath namespace} [namespaceId1 namespaceId2]
	{/dev/nvme0n1 54} [0x36 0x37]
	{/dev/nvme0n2 55} [0x36 0x37]
	{/dev/nvme1n1 54} [0x36 0x37]
	{/dev/nvme1n2 55} [0x36 0x37]
	*/
	namespaceIDs := make(map[DevicePathAndNamespace][]string)

	var err error
	for _, devicePathAndNamespace := range NVMeDeviceAndNamespace {

		devicePath := devicePathAndNamespace.DevicePath

		exe := nvme.buildNVMeCommand([]string{"nvme", "list-ns", devicePath})
		/* nvme list-ns /dev/nvme0n1
		[   0]:0x2401
		[   1]:0x2406
		*/
		cmd := getCommand(exe[0], exe[1:]...) // #nosec G204
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		str := string(output)
		lines := strings.Split(str, "\n")

		var namespaceDevice []string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				if len(strings.Split(line, ":")) >= 2 {
					nsDevice := strings.Split(line, ":")[1]
					namespaceDevice = append(namespaceDevice, nsDevice)
				}
			}
		}
		namespaceIDs[devicePathAndNamespace] = namespaceDevice
	}

	if len(namespaceIDs) == 0 {
		return map[DevicePathAndNamespace][]string{}, err
	}

	return namespaceIDs, nil
}

// GetNVMeDeviceData returns the information (nguid and namespace) of an NVME device path
func (nvme *NVMe) GetNVMeDeviceData(path string) (string, string, error) {
	var nguid string
	var namespace string

	exe := nvme.buildNVMeCommand([]string{"nvme", "id-ns", path})
	cmd := getCommand(exe[0], exe[1:]...) // #nosec G204

	/*
		nvme id-ns /dev/nvme3n1 0x95
		NVME Identify Namespace 149:
		nsze    : 0x1000000
		ncap    : 0x1000000
		nuse    : 0x223b8
		nsfeat  : 0xb
		nlbaf   : 0
		flbas   : 0
		mc      : 0
		dpc     : 0
		dps     : 0
		nmic    : 0x1
		rescap  : 0xff
		fpi     : 0
		dlfeat  : 9
		nawun   : 2047
		nawupf  : 2047
		nacwu   : 0
		nabsn   : 2047
		nabo    : 0
		nabspf  : 2047
		noiob   : 0
		nvmcap  : 0
		mssrl   : 0
		mcl     : 0
		msrc    : 0
		anagrpid: 2
		nsattr  : 0
		nvmsetid: 0
		endgid  : 0
		nguid   : 507911ecda65a2498ccf0968009a5d07
		eui64   : 0000000000000000
		lbaf  0 : ms:0   lbads:9  rp:0 (in use)
	*/

	output, err := cmd.Output()
	if err != nil {
		return "", "", err
	}
	str := string(output)
	lines := strings.Split(str, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "nguid") {
			if len(strings.Split(line, ":")) >= 2 {
				nguid = strings.TrimSpace(strings.Split(line, ":")[1])
			}
		}
		if strings.HasPrefix(line, "NVME Identify Namespace") {
			if len(strings.Fields(line)) >= 4 {
				namespace = strings.ReplaceAll(strings.TrimSpace(strings.Fields(line)[3]), ":", "")
			}
		}

		if nguid != "" && namespace != "" {
			return nguid, namespace, nil
		}
	}
	return nguid, namespace, err
}

// GetSessions queries information about  NVMe sessions
func (nvme *NVMe) GetSessions() ([]NVMESession, error) {
	exe := nvme.buildNVMeCommand([]string{"nvme", "list-subsys", "-o", "json"})
	cmd := getCommand(exe[0], exe[1:]...) // #nosec G204

	/*
		[
		  {
		    "HostNQN":"nqn.2014-08.org.nvmexpress:uuid:1a11111a-aa11-11aa-1111-a10aa1a11111",
		    "HostID":"1a11111a-aa11-11aa-1111-a11aa1a1111",
		    "Subsystems":[
		      {
		        "Name":"nvme-subsys0",
		        "NQN":"nqn.1988-11.com.dell:mock:00:1a1111a1111aA11111A",
		        "IOPolicy":"numa",
		        "Paths":[
		          {
		            "Name":"nvme3",
		            "Transport":"tcp",
		            "Address":"traddr=10.1.1.1,trsvcid=4420,src_addr=10.1.1.2",
		            "State":"live"
		          },
		          {
		            "Name":"nvme2",
		            "Transport":"tcp",
		            "Address":"traddr=10.1.1.2,trsvcid=4420,src_addr=10.1.1.2",
		            "State":"live"
		          },
		        ]
		      }
		    ]
		  }
		]
	*/

	output, err := cmd.Output()
	if err != nil {
		if isNoObjsExitCode(err) {
			return []NVMESession{}, nil
		}
		return []NVMESession{}, err
	}
	return nvme.sessionParser.Parse(output), nil
}

func isNoObjsExitCode(err error) bool {
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode() == NVMeNoObjsFoundExitCode
		}
	}
	return false
}

// DeviceRescan rescan the NVMe controller device
func (nvme *NVMe) DeviceRescan(device string) error {
	exe := nvme.buildNVMeCommand([]string{"nvme", "ns-rescan", device})
	cmd := getCommand(exe[0], exe[1:]...) // #nosec G204
	_, err := cmd.Output()
	if err != nil {
		return err
	}
	return nil
}
