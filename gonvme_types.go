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

const (
	// NVMeTransportTypeTCP - Placeholder for NVMe Transport type TCP
	NVMeTransportTypeTCP = "tcp"

	// NVMeTransportTypeFC - Placeholder for NVMe Transport type FC
	NVMeTransportTypeFC = "fc"
)
