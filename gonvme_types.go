package gonvme

// NVMeTarget defines an NVMe target
type NVMeTarget struct {
	Portal    string //traddr
	TargetNqn string //subnqn
	TrType    string //trtype
	AdrFam    string //adrfam
	SubType   string //subtype
	Treq      string //treq
	PortID    string //portid
	TrsvcID   string //trsvcid
	SecType   string //sectype
}

type NVMeSessionState string
type NVMeConnectionState string
type NVMeTransportName string

const (
	NVMeSessionState_LOGGED_IN NVMeSessionState = "LOGGED_IN"
	NVMeSessionState_FAILED    NVMeSessionState = "FAILED"
	NVMeSessionState_FREE      NVMeSessionState = "FREE"

	NVMeConnectionState_FREE             NVMeConnectionState = "FREE"
	NVMeConnectionState_TRANSPORT_WAIT   NVMeConnectionState = "TRANSPORT WAIT"
	NVMeConnectionState_IN_LOGIN         NVMeConnectionState = "IN LOGIN"
	NVMeConnectionState_LOGGED_IN        NVMeConnectionState = "LOGGED IN"
	NVMeConnectionState_IN_LOGOUT        NVMeConnectionState = "IN LOGOUT"
	NVMeConnectionState_LOGOUT_REQUESTED NVMeConnectionState = "LOGOUT REQUESTED"
	NVMeConnectionState_CLEANUP_WAIT     NVMeConnectionState = "CLEANUP WAIT"

	NVMeTransportName_TCP  NVMeTransportName = "tcp"
	NVMeTransportName_ISER NVMeTransportName = "iser"
)

// NVMeSession defines an NVMe session info
type NVMeSession struct {
	Target              string
	Portal              string
	SID                 string
	IfaceTransport      NVMeTransportName
	IfaceInitiatorname  string
	IfaceIPaddress      string
	NVMeSessionState    NVMeSessionState
	NVMeConnectionState NVMeConnectionState
	Username            string
	Password            string
	UsernameIn          string
	PasswordIn          string
}

// NVMeSession defines an NVMe node info
type NVMeNode struct {
	Target string
	Portal string
	Fields map[string]string
}

type NVMeSessionParser interface {
	Parse([]byte) []NVMeSession
}

type NVMeNodeParser interface {
	Parse([]byte) []NVMeNode
}
