package tracer

import (
	"bytes"
	"context"
	"testing"
	"fmt"
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