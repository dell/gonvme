/*
 *
 * Copyright © 2024-2026 Dell Inc. or its subsidiaries. All Rights Reserved.
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
			// IPv6 NVMe/TCP session — the former IPv4-only regex would leave
			// session.Portal empty, causing gobrick session-matching to fail
			// even when the NVMe connection was successfully established.
			name: "TCP IPv6 global unicast",
			input: `{
                "HostNQN": "something",
                "HostID": "something",
                "Subsystems": [{
                    "NQN": "nqn.2014-08.com.dell:shared-storage:nvme:ipv6test",
                    "Paths": [{
                        "Name": "nvme3",
                        "Transport": "tcp",
                        "Address": "traddr=2001:db8::1,trsvcid=4420,src_addr=fe80::2",
                        "State": "live"
                    }]
                }]
            }`,
			// Portal must be the bare IPv6 address (no port) to match
			// gobrick.addDefaultNVMePortToVolumeInfoPortals which skips
			// port-append when the portal string already contains ":".
			expectedResult: []NVMESession{
				{
					Name:              "nvme3",
					Target:            "nqn.2014-08.com.dell:shared-storage:nvme:ipv6test",
					NVMETransportName: "tcp",
					Portal:            "2001:db8::1",
					NVMESessionState:  "live",
					SourceAddr:        "fe80::2",
				},
			},
		},
		{
			name: "TCP IPv6 no src_addr field",
			input: `{
                "HostNQN": "something",
                "HostID": "something",
                "Subsystems": [{
                    "NQN": "nqn.2014-08.com.dell:shared-storage:nvme:ipv6test2",
                    "Paths": [{
                        "Name": "nvme4",
                        "Transport": "tcp",
                        "Address": "traddr=fd12:3456:789a:1::1,trsvcid=4420",
                        "State": "live"
                    }]
                }]
            }`,
			// No src_addr in the address field: SourceAddr stays empty, never an error.
			expectedResult: []NVMESession{
				{
					Name:              "nvme4",
					Target:            "nqn.2014-08.com.dell:shared-storage:nvme:ipv6test2",
					NVMETransportName: "tcp",
					Portal:            "fd12:3456:789a:1::1",
					NVMESessionState:  "live",
					SourceAddr:        "",
				},
			},
		},
		{
			// Host-managed NVMe/TCP sessions are established with --host-traddr, so
			// the kernel reports src_addr. The driver logs it to help operators spot
			// a session bound to the wrong interface; the value is never validated.
			name: "TCP IPv4 with src_addr",
			input: `{
                "HostNQN": "something",
                "HostID": "something",
                "Subsystems": [{
                    "NQN": "nqn.2014-08.com.dell:shared-storage:nvme:srcaddr",
                    "Paths": [{
                        "Name": "nvme8",
                        "Transport": "tcp",
                        "Address": "traddr=10.11.12.13,trsvcid=4420,src_addr=10.10.10.21",
                        "State": "live"
                    }]
                }]
            }`,
			expectedResult: []NVMESession{
				{
					Name:              "nvme8",
					Target:            "nqn.2014-08.com.dell:shared-storage:nvme:srcaddr",
					NVMETransportName: "tcp",
					Portal:            "10.11.12.13:4420",
					NVMESessionState:  "live",
					SourceAddr:        "10.10.10.21",
				},
			},
		},
		{
			// A src_addr key with no value must yield an empty SourceAddr rather than
			// an error or a panic — the session is still usable without it.
			name: "TCP src_addr present but empty",
			input: `{
                "HostNQN": "something",
                "HostID": "something",
                "Subsystems": [{
                    "NQN": "nqn.2014-08.com.dell:shared-storage:nvme:emptysrc",
                    "Paths": [{
                        "Name": "nvme9",
                        "Transport": "tcp",
                        "Address": "traddr=10.11.12.14,trsvcid=4420,src_addr=",
                        "State": "live"
                    }]
                }]
            }`,
			expectedResult: []NVMESession{
				{
					Name:              "nvme9",
					Target:            "nqn.2014-08.com.dell:shared-storage:nvme:emptysrc",
					NVMETransportName: "tcp",
					Portal:            "10.11.12.14:4420",
					NVMESessionState:  "live",
					SourceAddr:        "",
				},
			},
		},
		{
			// Same defensive unquoting as traddr and trsvcid.
			name: "TCP quoted src_addr value",
			input: `{
                "HostNQN": "something",
                "HostID": "something",
                "Subsystems": [{
                    "NQN": "nqn.2014-08.com.dell:shared-storage:nvme:quotedsrc",
                    "Paths": [{
                        "Name": "nvme10",
                        "Transport": "tcp",
                        "Address": "traddr=\"10.0.0.3\",trsvcid=\"4420\",src_addr=\"10.10.10.22\"",
                        "State": "live"
                    }]
                }]
            }`,
			expectedResult: []NVMESession{
				{
					Name:              "nvme10",
					Target:            "nqn.2014-08.com.dell:shared-storage:nvme:quotedsrc",
					NVMETransportName: "tcp",
					Portal:            "10.0.0.3:4420",
					NVMESessionState:  "live",
					SourceAddr:        "10.10.10.22",
				},
			},
		},
		{
			// An unparsable traddr leaves Portal empty, but src_addr is independent
			// and must still be captured for diagnostics.
			name: "TCP invalid traddr with valid src_addr",
			input: `{
                "HostNQN": "something",
                "HostID": "something",
                "Subsystems": [{
                    "NQN": "nqn.2014-08.com.dell:shared-storage:nvme:badaddrsrc",
                    "Paths": [{
                        "Name": "nvme11",
                        "Transport": "tcp",
                        "Address": "traddr=not-an-ip,trsvcid=4420,src_addr=10.10.10.23",
                        "State": "live"
                    }]
                }]
            }`,
			expectedResult: []NVMESession{
				{
					Name:              "nvme11",
					Target:            "nqn.2014-08.com.dell:shared-storage:nvme:badaddrsrc",
					NVMETransportName: "tcp",
					Portal:            "",
					NVMESessionState:  "live",
					SourceAddr:        "10.10.10.23",
				},
			},
		},
		{
			// FC sessions carry no src_addr; the FC branch must leave it empty.
			name: "FC leaves SourceAddr empty",
			input: `{
                "HostNQN": "something",
                "HostID": "something",
                "Subsystems": [{
                    "NQN": "nqn.2014-08.com.dell:shared-storage:nvme:fcsrc",
                    "Paths": [{
                        "Name": "nvme12",
                        "Transport": "fc",
                        "Address": "traddr=nn-0x1:pn-0x2 host_traddr=nn-0x3:pn-0x4",
                        "State": "live"
                    }]
                }]
            }`,
			expectedResult: []NVMESession{
				{
					Name:              "nvme12",
					Target:            "nqn.2014-08.com.dell:shared-storage:nvme:fcsrc",
					NVMETransportName: "fc",
					Portal:            "nn-0x1:pn-0x2",
					NVMESessionState:  "live",
					SourceAddr:        "",
				},
			},
		},
		{
			name: "TCP invalid address (not a valid IP)",
			input: `{
                "HostNQN": "something",
                "HostID": "something",
                "Subsystems": [{
                    "NQN": "nqn.2014-08.com.dell:shared-storage:nvme:badaddr",
                    "Paths": [{
                        "Name": "nvme5",
                        "Transport": "tcp",
                        "Address": "traddr=not-an-ip,trsvcid=4420",
                        "State": "live"
                    }]
                }]
            }`,
			// net.ParseIP returns nil for invalid addresses; session.Portal stays "".
			expectedResult: []NVMESession{
				{
					Name:              "nvme5",
					Target:            "nqn.2014-08.com.dell:shared-storage:nvme:badaddr",
					NVMETransportName: "tcp",
					Portal:            "",
					NVMESessionState:  "live",
				},
			},
		},
		{
			// Defensive: both traddr and trsvcid values are unquoted before use,
			// so a hypothetical kernel output with quoted values still parses correctly.
			name: "TCP quoted traddr and trsvcid values",
			input: `{
                "HostNQN": "something",
                "HostID": "something",
                "Subsystems": [{
                    "NQN": "nqn.2014-08.com.dell:shared-storage:nvme:quoted",
                    "Paths": [{
                        "Name": "nvme6",
                        "Transport": "tcp",
                        "Address": "traddr=\"10.0.0.2\",trsvcid=\"4420\"",
                        "State": "live"
                    }]
                }]
            }`,
			expectedResult: []NVMESession{
				{
					Name:              "nvme6",
					Target:            "nqn.2014-08.com.dell:shared-storage:nvme:quoted",
					NVMETransportName: "tcp",
					Portal:            "10.0.0.2:4420",
					NVMESessionState:  "live",
				},
			},
		},
		{
			// IPv4-mapped IPv6 addresses (::ffff:10.0.0.1) contain ":" so gobrick's
			// addDefaultNVMePortToVolumeInfoPortals skips port-append, producing bare
			// target.Portal. The parser must use strings.Contains(trAddr, ":") — NOT
			// net.ParseIP().To4() — so that session.Portal matches target.Portal.
			// To4() returns non-nil for ::ffff:10.0.0.1 and would append the port,
			// causing a mismatch and breaking session-matching for this address family.
			name: "TCP IPv4-mapped IPv6 address",
			input: `{
                "HostNQN": "something",
                "HostID": "something",
                "Subsystems": [{
                    "NQN": "nqn.2014-08.com.dell:shared-storage:nvme:v4mapped",
                    "Paths": [{
                        "Name": "nvme7",
                        "Transport": "tcp",
                        "Address": "traddr=::ffff:10.0.0.1,trsvcid=4420",
                        "State": "live"
                    }]
                }]
            }`,
			expectedResult: []NVMESession{
				{
					Name:              "nvme7",
					Target:            "nqn.2014-08.com.dell:shared-storage:nvme:v4mapped",
					NVMETransportName: "tcp",
					Portal:            "::ffff:10.0.0.1",
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
