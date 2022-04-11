package gonvme

import (
	"regexp"
	"strings"
)

type sessionParser struct{}

func (sp *sessionParser) Parse(data []byte) []NVMESession {
	str := string(data)
	lines := strings.Split(str, "\n")

	var result []NVMESession
	var curSession *NVMESession
	var Target string
	var Name string

	for _, line := range lines {
		line = strings.ReplaceAll(strings.TrimSpace(line), ",", "")

		switch {
		case regexp.MustCompile("^\"NQN\"").Match([]byte(line)):
			Target = strings.ReplaceAll(strings.Fields(line)[2], "\"", "")

		case regexp.MustCompile("^\"Name\"").Match([]byte(line)):
			Name = strings.ReplaceAll(strings.Fields(line)[2], "\"", "")

		case regexp.MustCompile("^\"Transport\"").Match([]byte(line)):
			session := NVMESession{}
			session.Target = Target
			session.Name = Name
			session.NVMETransportName = NVMETransportName(strings.ReplaceAll(strings.Fields(line)[2], "\"", ""))
			if curSession != nil {
				result = append(result, *curSession)
			}
			curSession = &session

		case regexp.MustCompile("^\"Address\"").Match([]byte(line)):
			targetIP := strings.Split(strings.Fields(line)[2], "=")[1]
			if strings.Split(strings.Fields(line)[3], "=")[0] == "host_traddr" {
				curSession.Portal = targetIP
			} else {
				targetPortal := strings.ReplaceAll(strings.Split(strings.Fields(line)[3], "=")[1], "\"", "")
				curSession.Portal = targetIP + ":" + targetPortal
			}

		case regexp.MustCompile("^\"State\"").Match([]byte(line)):
			curSession.NVMESessionState = NVMESessionState(strings.ReplaceAll(strings.Fields(line)[2], "\"", ""))
		}
	}
	if curSession != nil {
		result = append(result, *curSession)
	}
	return result
}
