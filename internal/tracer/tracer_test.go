/*
 * Copyright © 2024-2025 Dell Inc. or its subsidiaries. All Rights Reserved.
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

package tracer

import (
	"bytes"
	"context"
	"fmt"
	"testing"
)

type mockTracer struct {
	output *bytes.Buffer
}

func (m *mockTracer) Trace(_ context.Context, format string, args ...interface{}) {
	m.output.WriteString(fmt.Sprintf(format+"\n", args...))
}

func TestSetTracer(t *testing.T) {
	var buf bytes.Buffer
	mock := &mockTracer{output: &buf}

	SetTracer(mock)

	tracer.Trace(context.Background(), "test message")
	expected := "test message\n"

	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestTraceFuncCall(t *testing.T) {
	var buf bytes.Buffer
	mock := &mockTracer{output: &buf}

	SetTracer(mock)

	done := TraceFuncCall(context.Background(), "TestFunc")
	done()

	expected := "START: TestFunc\nEND: TestFunc\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestDefaultTracer(t *testing.T) {
	tracer := &ConsoleTracer{}
	// Trace is a one liner that calls fmt.Printf.
	tracer.Trace(context.TODO(), "test message")
}
