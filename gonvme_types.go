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

// NVMeTransportName - Placeholder for NVMe Transport name
type NVMeTransportName string

const (
	NVMeTCP NVMeTransportName = "tcp"
	NVMeFC  NVMeTransportName = "fc"
)
