package gonvme

import (
	// "encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionParser(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedResult []NVMESession
	}{
		{
			name: "TCP",
			input: `{
                "HostNQN": "something",
                "HostID": "something",
                "Subsystems": [{
                    "NQN": "nqn.2014-08.com.dell:shared-storage:fc:1234567890abcdef",
                    "Paths": [{
                        "Name": "nqn.2014-08.com.dell:shared-storage:fc:1234567890abcdef",
                        "Transport": "tcp",
                        "Address": "traddr=10.0.0.1,trsvcid=4420,src=00",
                        "State": "live"
                    }]
                }]
            }`,
			expectedResult: []NVMESession{
				{
					Name:              "nqn.2014-08.com.dell:shared-storage:fc:1234567890abcdef",
					Target:            "nqn.2014-08.com.dell:shared-storage:fc:1234567890abcdef",
					NVMETransportName: "tcp",
					Portal:            "10.0.0.1:4420",
					NVMESessionState:  "live",
				},
			},
		},
		{
			name: "FC",
			input: `{
                "HostNQN": "something",
                "HostID": "something",
                "Subsystems": [{
                    "NQN": "nqn.2014-08.com.dell:shared-storage:fc:1234567890abcdef",
                    "Paths": [{
                        "Name": "nqn.2014-08.com.dell:shared-storage:fc:1234567890abcdef",
                        "Transport": "fc",
                        "Address": "traddr=10.0.0.1:4420 trsvcid=4420 src=00",
                        "State": "live"
                    }]
                }]
            }`,
			expectedResult: []NVMESession{
				{
					Name:              "nqn.2014-08.com.dell:shared-storage:fc:1234567890abcdef",
					Target:            "nqn.2014-08.com.dell:shared-storage:fc:1234567890abcdef",
					NVMETransportName: "fc",
					Portal:            "10.0.0.1:4420",
					NVMESessionState:  "live",
				},
			},
		},
	}

	sp := &sessionParser{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sp.Parse([]byte(tt.input))
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
