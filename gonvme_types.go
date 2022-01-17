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
	// NVMeNVMeTransportTypeTCP - Placeholder for NVMe Transport type TCP
	NVMeNVMeTransportTypeTCP = "tcp"

	// NVMeNVMeTransportTypeFC - Placeholder for NVMe Transport type FC
	NVMeNVMeTransportTypeFC = "fc"
)
