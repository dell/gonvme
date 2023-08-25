/*
 *
 * Copyright © 2022-2023 Dell Inc. or its subsidiaries. All Rights Reserved.
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

// Response of subsystems.
type Response struct {
	Subsys []subsystems `json:"Subsystems"`
}

func (sp *sessionParser) Parse(data []byte) []NVMESession {
	str := string(data)
	var result []NVMESession
	var response Response
	err := json.Unmarshal([]byte(str), &response)
	if err != nil {
		log.Error("JSON-encoded parsing error: ", err.Error())
		return result
	}
	for _, system := range response.Subsys {
		session := NVMESession{}
		session.Target = system.NQN
		reAdd := `(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`
		re := regexp.MustCompilePOSIX(reAdd)
		for _, path := range system.Paths {
			session.Name = path["Name"]
			session.NVMETransportName = NVMETransportName(path["Transport"])
			if path["Transport"] == NVMeTransportTypeFC {
				session.Portal = strings.Split(strings.Fields(path["Address"])[0], "=")[1]
			} else if path["Transport"] == NVMeTransportTypeTCP {
				if re.MatchString(path["Address"]) {
					session.Portal = re.FindString(path["Address"]) + ":" + strings.ReplaceAll(strings.Split(strings.Fields(path["Address"])[1], "=")[1], "\"", "")
				}
			} else {
				continue
			}
			session.NVMESessionState = NVMESessionState(path["State"])
			result = append(result, session)

		}
	}
	return result
}
