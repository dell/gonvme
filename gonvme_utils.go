package gonvme

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strings"
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
