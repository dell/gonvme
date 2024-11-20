/*
 *
 * Copyright © 2022 Dell Inc. or its subsidiaries. All Rights Reserved.
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

package logger

import (
	"context"
	"testing"
)

func TestLogger(t *testing.T) {
	tests := []struct {
		name string
		fn   func(ctx context.Context, format string, args ...interface{})
	}{
		{
			name: "Info",
			fn:   Info,
		},
		{
			name: "Debug",
			fn:   Debug,
		},
		{
			name: "Error",
			fn:   Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldLogger := logger
			defer func() {
				logger = oldLogger
			}()
			logger = &DummyLogger{}
			tt.fn(context.Background(), "test")
		})
	}
}
