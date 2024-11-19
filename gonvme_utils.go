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
	 "encoding/json"
	 "fmt"
	 "regexp"
	 "strings"
 
	 log "github.com/sirupsen/logrus"
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
        log.Error("JSON-encoded parsing error: ", err.Error())
        return result
    }
    for _, resp := range response {
        for _, system := range resp.Subsystems {
            session := NVMESession{}
            session.Target = system.NQN
            reAdd := `(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`
            re := regexp.MustCompilePOSIX(reAdd)
            for _, path := range system.Paths {
                session.Name = path["Name"]
                session.NVMETransportName = NVMETransportName(path["Transport"]) 
                if path["Transport"] == NVMeTransportTypeFC {
                    fields := strings.Fields(path["Address"])
                    if len(fields) > 0 {
                        parts := strings.Split(fields[0], "=")
                        if len(parts) > 1 {
                            session.Portal = parts[1]
                        }
                    }
                } else if path["Transport"] == NVMeTransportTypeTCP {
                    if re.MatchString(path["Address"]) {
                        ip := re.FindString(path["Address"])
                        portHolder := ""
                        for _, item := range strings.Split(path["Address"], ",") { // fmt: [traddr=10.230.1.1,trsvcid=4420,src=00]
                            if strings.Contains(item, "trsvcid") {
                                portHolder = item
                                break
                            }
                        }
                        if portHolder != "" {
                            portParts := strings.Split(portHolder, "=")
                            if len(portParts) > 1 {
                                port := strings.ReplaceAll(portParts[1], "\"", "")
                                session.Portal = ip + ":" + port
                            }
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
 