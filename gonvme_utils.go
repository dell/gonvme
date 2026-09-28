/*
 *
 * Copyright © 2022-2026 Dell Inc. or its subsidiaries. All Rights Reserved.
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
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/dell/csmlog"
)

type sessionParser struct{}

// Single subsystem object
type subsystems struct {
	Name  string
	NQN   string
	Paths []map[string]string `json:"Paths"`
}

// SubSysResponse of subsystems.
type SubSysResponse struct {
	HostNQN    string       `json:"HostNQN"`
	HostID     string       `json:"HostID"`
	Subsystems []subsystems `json:"Subsystems"`
}

func (sp *sessionParser) Parse(data []byte) []NVMESession {
	str := string(data)
	if str[0] == '{' {
		str = fmt.Sprintf("[%s]", str)
	}
	var result []NVMESession
	var response []SubSysResponse
	err := json.Unmarshal([]byte(str), &response)
	if err != nil {
		csmlog.WithFields(csmlog.Fields{
			"error": err.Error(),
		}).Error("JSON-encoded parsing error")
		return result
	}
	for _, resp := range response {
		for _, system := range resp.Subsystems {
			session := NVMESession{}
			session.Target = system.NQN
			for _, path := range system.Paths {
				session.Name = path["Name"]
				session.NVMETransportName = NVMETransportName(path["Transport"])
				// session is reused across the paths of one subsystem, so clear the
				// per-path source address before parsing this path. FC paths carry no
				// src_addr and must not inherit one from a preceding TCP path.
				session.SourceAddr = ""
				if path["Transport"] == NVMeTransportTypeFC {
					fields := strings.Fields(path["Address"])
					if len(fields) > 0 {
						parts := strings.Split(fields[0], "=")
						if len(parts) > 1 {
							session.Portal = parts[1]
						}
					}
				} else if path["Transport"] == NVMeTransportTypeTCP {
					// Parse the comma-separated key=value address field.
					// Format: "traddr=<IP>,trsvcid=<port>[,<key>=<value>...]"
					// Both IPv4 and IPv6 bare addresses are supported; net.ParseIP
					// is the authoritative validator, replacing the former IPv4-only
					// regex which silently ignored all IPv6 NVMe/TCP sessions.
					var trAddr, trsvcid, srcAddr string
					for _, item := range strings.Split(path["Address"], ",") {
						item = strings.TrimSpace(item)
						switch {
						case strings.HasPrefix(item, "traddr="):
							trAddr = strings.ReplaceAll(strings.TrimPrefix(item, "traddr="), "\"", "")
						case strings.HasPrefix(item, "trsvcid="):
							trsvcid = strings.ReplaceAll(strings.TrimPrefix(item, "trsvcid="), "\"", "")
						case strings.HasPrefix(item, "src_addr="):
							// Sessions established with "nvme connect --host-traddr" report
							// the bound source address here. It is recorded as-is, with no
							// validation: consumers only log it, and an absent or empty
							// value leaves SourceAddr empty rather than failing the parse.
							srcAddr = strings.ReplaceAll(strings.TrimPrefix(item, "src_addr="), "\"", "")
						}
					}
					// srcAddr is scoped to this path, so a session without src_addr never
					// inherits the value from the previously parsed path of the same subsystem.
					session.SourceAddr = srcAddr
					if net.ParseIP(trAddr) != nil {
						// Use strings.Contains(trAddr, ":") — not net.ParseIP().To4() — to
						// decide whether to append the port. This intentionally mirrors the
						// predicate used by gobrick.addDefaultNVMePortToVolumeInfoPortals:
						//
						//   if !strings.Contains(t.Portal, ":") { t.Portal += ":4420" }
						//
						// session.Portal must equal target.Portal for gobrick session-matching
						// to succeed, so both sides must apply the same gate. Using To4() would
						// diverge for IPv4-mapped IPv6 addresses (e.g. ::ffff:10.0.0.1): To4()
						// returns non-nil (IPv4 path → appends port), while gobrick sees ":"
						// and skips the port-append, causing a mismatch.
						if strings.Contains(trAddr, ":") {
							// IPv6 or IPv4-mapped IPv6: gobrick skips port-append; emit bare
							// address so session.Portal == target.Portal.
							session.Portal = trAddr
						} else if trsvcid != "" {
							// Pure IPv4: gobrick appends ":port"; mirror that format here.
							session.Portal = trAddr + ":" + trsvcid
						} else {
							session.Portal = trAddr
						}
					}
				} else {
					continue
				}
				session.NVMESessionState = NVMESessionState(path["State"])
				result = append(result, session)
			}
		}
	}
	return result
}
