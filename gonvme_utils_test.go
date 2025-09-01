/*
 *
 * Copyright © 2021-2024 Dell Inc. or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

/*
 *
 * Copyright © 2024 Dell Inc. or its subsidiaries. All Rights Reserved.
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
		{
			name: "Skip invalid transport",
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
                    },
					{
                        "Name": "nqn.2014-08.com.dell:shared-storage:fc:1234567890abcdef",
                        "Transport": "invalid",
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
		{
			name: "Fail to parse",
			input: `{
                "HostNQN": 1,
                "HostID": "something",
                "Subsystems": [{
                    "NQN": "nqn.2014-08.com.dell:shared-storage:fc:1234567890abcdef",
                    "Paths": [{
                        "Name": "nqn.2014-08.com.dell:shared-storage:fc:1234567890abcdef",
                        "Transport": "fc",
                        "Address": "traddr=10.0.0.1:4420 trsvcid=4420 src=00",
                        "State": "live"
                    },
					{
                        "Name": "nqn.2014-08.com.dell:shared-storage:fc:1234567890abcdef",
                        "Transport": "invalid",
                        "Address": "traddr=10.0.0.1:4420 trsvcid=4420 src=00",
                        "State": "live"
                    }]
                }]
            }`,
			expectedResult: []NVMESession(nil),
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
