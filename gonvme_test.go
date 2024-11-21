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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetTimeouts(t *testing.T) {
	var prop time.Duration

	// Test with value 0, should set to defaultVal
	setTimeouts(&prop, 0, 10*time.Second)
	assert.Equal(t, 10*time.Second, prop)

	// Test with non-zero value, should set to value
	setTimeouts(&prop, 5*time.Second, 10*time.Second)
	assert.Equal(t, 5*time.Second, prop)
}

func TestNVMeType_isMock(t *testing.T) {
	nvme := &NVMeType{mock: true}
	assert.True(t, nvme.isMock())

	nvme.mock = false
	assert.False(t, nvme.isMock())
}

func TestNVMeType_getOptions(t *testing.T) {
	options := map[string]string{"key": "value"}
	nvme := &NVMeType{options: options}
	assert.Equal(t, options, nvme.getOptions())
}
