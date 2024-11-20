package gonvme

type NVMESession struct {
    Name              string
    Target            string
    NVMETransportName string
    Portal            string
    NVMESessionState  string
}

/*func TestSessionParser(t *testing.T) {
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
                    "Name": "nqn.2014-08.com.dell:shared-storage:fc:1234567890abcdef",
                    "NQN": "nqn.2014-08.com.dell:shared-storage:fc:1234567890abcdef",
                    "Paths": [{
                        "Transport": "tcp",
                        "Address": "10.230.1.1:4420",
                        "State": "live"
                    }]
                }]
            }`,
            expectedResult: []NVMESession{
                {
                    Name:              "nqn.2014-08.com.dell:shared-storage:fc:1234567890abcdef",
                    Target:            "nqn.2014-08.com.dell:shared-storage:fc:1234567890abcdef",
                    NVMETransportName: "tcp",
                    Portal:            "10.230.1.1:4420",
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
                    "Name": "nqn.2014-08.com.dell:shared-storage:fc:1234567890abcdef",
                    "NQN": "nqn.2014-08.com.dell:shared-storage:fc:1234567890abcdef",
                    "Paths": [{
                        "Transport": "fc",
                        "Address": "10.230.1.1:4420",
                        "State": "live"
                    }]
                }]
            }`,
            expectedResult: []NVMESession{
                {
                    Name:              "nqn.2014-08.com.dell:shared-storage:fc:1234567890abcdef",
                    Target:            "nqn.2014-08.com.dell:shared-storage:fc:1234567890abcdef",
                    NVMETransportName: "tcp",
                    Portal:            "10.230.1.1:4420",
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

	// NVMESession := Parse(tests)

    // for _, tt := range tests {
        // t.Run(tt.name, func(t *testing.T) {
        //     var result struct {
        //         Subsystems []struct {
        //             Name  string
        //             NQN   string
        //             Paths []struct {
        //                 Transport string
        //                 Address   string
        //                 State     string
        //             }
        //         }
        //     }
        //     err := json.Unmarshal([]byte(tt.input), &result)
        //     assert.NoError(t, err)

        //     var sessions []NVMESession
        //     for _, subsystem := range result.Subsystems {
        //         for _, path := range subsystem.Paths {
        //             sessions = append(sessions, NVMESession{
        //                 Name:              subsystem.Name,
        //                 Target:            subsystem.NQN,
        //                 NVMETransportName: path.Transport,
        //                 Portal:            path.Address,
        //                 NVMESessionState:  path.State,
        //             })
        //         }
        //     }

        //     assert.Equal(t, tt.expectedResult, sessions)
        // })
    // }
}
 */