package tracer

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"
	datadog "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/opentracer"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// New is contructor for tracer
func New(opts ...StartOption) opentracing.Tracer {
	return datadog.New(opts...)
}

// Span represents a chunk of computation time. Spans have names, durations,
// timestamps and other metadata. A Tracer is used to create hierarchies of
// spans in a request, buffer and submit them to the server.
type Span interface {
	// SetTag sets a key/value pair as metadata on the span.
	SetTag(key string, value interface{})

	// SetOperationName sets the operation name for this span. An operation name should be
	// a representative name for a group of spans (e.g. "grpc.server" or "http.request").
	SetOperationName(operationName string)

	// BaggageItem returns the baggage item held by the given key.
	BaggageItem(key string) string

	// SetBaggageItem sets a new baggage item at the given key. The baggage
	// item should propagate to all descendant spans, both in- and cross-process.
	SetBaggageItem(key, val string)

	// Finish finishes the current span with the given options. Finish calls should be idempotent.
	Finish(opts ...FinishOption)

	// Context returns the SpanContext of this Span.
	Context() SpanContext
}

// SpanContext represents a span state that can propagate to descendant spans
// and across process boundaries. It contains all the information needed to
// spawn a direct descendant of the span that it belongs to. It can be used
// to create distributed tracing by propagating it using the provided interfaces.
type SpanContext = ddtrace.SpanContext

// StartSpanFromContext start a span from an input context.
func StartSpanFromContext(ctx context.Context, operationName string, opts ...StartSpanOption) (Span, context.Context) {
	return tracer.StartSpanFromContext(ctx, operationName, opts...)
}
