package tracer

import (
	"context"
	"fmt"
)

var tracer Tracer

// Tracer  tracing interface for gonvme
type Tracer interface {
	Trace(ctx context.Context, format string, args ...interface{})
}

// SetTracer - set custom tracer
func SetTracer(customTracer Tracer) {
	tracer = customTracer
}

func init() {
	tracer = &DummyTracer{}
}

// TraceFuncCall - trace definitions
func TraceFuncCall(ctx context.Context, funcName string) func() {
	tracer.Trace(ctx, "START: %s", funcName)
	return func() {
		Trace(ctx, "END: %s", funcName)
	}
}

// DummyTracer - default tracer
type DummyTracer struct{}

// Trace - default trace
func (dl *DummyTracer) Trace(ctx context.Context, format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// Trace - custom trace
func Trace(ctx context.Context, format string, args ...interface{}) {
	tracer.Trace(ctx, format, args...)
}
