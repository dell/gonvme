package gonvme

// NVMeTarget defines an NVMe target
type NVMeTarget struct {
	Portal     string //traddr
	TargetNqn  string //subnqn
	TrType     string //trtype
	AdrFam     string //adrfam
	SubType    string //subtype
	Treq       string //treq
	PortID     string //portid
	TrsvcID    string //trsvcid
	SecType    string //sectype
	TargetType string //trtype
}

//NVMESessionState defines the NVMe connection state
type NVMESessionState string

//NVMETransportName defines the NMVe protocol
type NVMETransportName string

const (
	// NVMeTransportTypeTCP - Placeholder for NVMe Transport type TCP
	NVMeTransportTypeTCP = "tcp"

	// NVMeTransportTypeFC - Placeholder for NVMe Transport type FC
	NVMeTransportTypeFC = "fc"

	//NVMESessionStateLive indicates the NVMe connection state as live
	NVMESessionStateLive NVMESessionState = "live"
	//NVMESessionStateDeleting indicates the NVMe connection state as deleting
	NVMESessionStateDeleting NVMESessionState = "deleting"
	//NVMESessionStateConnecting indicates the NVMe connection state as connecting
	NVMESessionStateConnecting NVMESessionState = "connecting"

	//NVMETransportNameTCP indicates the NVMe protocol as tcp
	NVMETransportNameTCP NVMETransportName = "tcp"
	//NVMETransportNameFC indicates the NVMe protocol as fc
	NVMETransportNameFC NVMETransportName = "fc"
	//NVMETransportNameRDMA indicates the NVMe protocol as rdma
	NVMETransportNameRDMA NVMETransportName = "rdma"
)

// DevicePathAndNamespace  defines the device path and namespace of a device
type DevicePathAndNamespace struct {
	DevicePath string
	Namespace  string
}

// NVMESession defines an iSCSI session info
type NVMESession struct {
	Target            string
	Portal            string
	Name              string
	NVMESessionState  NVMESessionState
	NVMETransportName NVMETransportName
}

// NVMeSessionParser defines an NVMe session parser
type NVMeSessionParser interface {
	Parse([]byte) []NVMESession
}
