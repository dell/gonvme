package gonvme

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

const (
	// ChrootDirectory allows the nvme commands to be run within a chrooted path, helpful for containerized services
	ChrootDirectory = "chrootDirectory"

	// DefaultInitiatorNameFile is the default file which contains the initiator nqn
	DefaultInitiatorNameFile = "/etc/nvme/hostnqn"

	// NVMeCommand - nvme command
	NVMeCommand = "nvme"

	// NVMePort - port number
	NVMePort = "4420"

	// NVMeNoObjsFoundExitCode exit code indicates that no records/targets/sessions/portals
	// found to execute operation on
	NVMeNoObjsFoundExitCode = 21
)

// NVMeTCP provides many nvme-specific functions
type NVMeTCP struct {
	NVMeType
	sessionParser NVMeSessionParser
}

// NewNVMeTCP - returns a new NVMeTCP client
func NewNVMeTCP(opts map[string]string) *NVMeTCP {
	nvme := NVMeTCP{
		NVMeType: NVMeType{
			mock:    false,
			options: opts,
		},
	}
	nvme.sessionParser = &sessionParser{}
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
	/*if login {
		for _, t := range targets {
			gonvme.PerformLogin(t)
		}
	}*/

	return targets, nil
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

// NVMeConnect will attempt to connect into a given nvme target
func (nvme *NVMeTCP) NVMeConnect(target NVMeTarget) error {
	return nvme.nvmeConnect(target)
}

func (nvme *NVMeTCP) nvmeConnect(target NVMeTarget) error {
	// nvme connect is done via the nvme cli
	// nvme connect -t tcp -n <target NQN> -a <NVMe interface IP> -s 4420
	exe := nvme.buildNVMeCommand([]string{NVMeCommand, "connect", "-t", "tcp", "-n", target.TargetNqn, "-a", target.Portal, "-s", NVMePort})
	cmd := exec.Command(exe[0], exe[1:]...)

	var Output string
	stderr, _ := cmd.StderrPipe()
	err := cmd.Start()

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
				if Output == "Failed to write to /dev/nvme-fabrics: Operation already in progress" || Output == ""{
					fmt.Printf("NVMe connection already exists\n")
					err = nil
				} else {
					fmt.Printf("\nError during nvme connect %s at %s: %v", target.TargetNqn, target.Portal, err)
					return err
				}
			} else {
				fmt.Printf("\nnvme connect failure: %v", err)
			}
		} else {
			fmt.Printf("\nError during nvme connect %s at %s: %v", target.TargetNqn, target.Portal, err)
		}

		if err != nil {
			fmt.Printf("\nError during nvme connect %s at %s: %v", target.TargetNqn, target.Portal, err)
			return err
		}
	} else {
		fmt.Printf("\nnvme connect successful: %s", target.TargetNqn)
	}

	return nil
}

// NVMeDisconnect will attempt to disconnect from a given nvme target
func (nvme *NVMeTCP) NVMeDisconnect(target NVMeTarget) error {
	return nvme.nvmeDisconnect(target)
}

func (nvme *NVMeTCP) nvmeDisconnect(target NVMeTarget) error {
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

// ListNamespaceDevices returns the Device Paths and Namespace of each device and each output content
func (nvme *NVMeTCP) ListNamespaceDevices() map[DevicePathAndNamespace][]string {
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
	cmd := exec.Command(exe[0], exe[1:]...)

	output, _ := cmd.Output()
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

	namespaceDevices := make(map[DevicePathAndNamespace][]string)

	/* finding the namespaceDevices Output
	{devicePath namespace} [namespaceId1 namespaceId2]
	{/dev/nvme0n1 54} [0x36 0x37]
	{/dev/nvme0n2 55} [0x36 0x37]
	{/dev/nvme1n1 54} [0x36 0x37]
	{/dev/nvme1n2 55} [0x36 0x37]
	*/

	for _, devicePathAndNamespace := range result {

		devicePath = devicePathAndNamespace.DevicePath
		namespace = devicePathAndNamespace.Namespace

		exe := nvme.buildNVMeCommand([]string{"nvme", "list-ns", devicePath})
		/* nvme list-ns /dev/nvme0n1
		[   0]:0x2401
		[   1]:0x2406
		*/
		cmd := exec.Command(exe[0], exe[1:]...)
		output, _ := cmd.Output()

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
		namespaceDevices[devicePathAndNamespace] = namespaceDevice
	}
	return namespaceDevices
}

// GetNamespaceData returns the information of namespace specific to the namespace Id
func (nvme *NVMeTCP) GetNamespaceData(path string, namespaceID string) (string, string, error) {

	var nguid string
	var namespace string

	exe := nvme.buildNVMeCommand([]string{"nvme", "id-ns", path, "--namespace", namespaceID})
	cmd := exec.Command(exe[0], exe[1:]...)

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

	output, error := cmd.Output()
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
	return nguid, namespace, error
}

// GetSessions queries information about  NVMe sessions
func (nvme *NVMeTCP) GetSessions() ([]NVMESession, error) {
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

func isNoObjsExitCode(err error) bool {
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode() == NVMeNoObjsFoundExitCode
		}
	}
	return false
}
