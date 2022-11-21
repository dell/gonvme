/*
 *
 * Copyright © 2022 Dell Inc. or its subsidiaries. All Rights Reserved.
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
	HostAdr    string //host_traddr
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

// FCHBAInfo holds information about host NVMe/FC ports
type FCHBAInfo struct {
	PortName string
	NodeName string
}
