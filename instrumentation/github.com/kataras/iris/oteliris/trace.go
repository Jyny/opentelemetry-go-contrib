// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package oteliris // import "go.opentelemetry.io/contrib/instrumentation/github.com/kataras/iris/oteliris"

import (
	iris "github.com/kataras/iris/v12"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/semconv/v1.13.0/httpconv"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	tracerName = "go.opentelemetry.io/contrib/instrumentation/github.com/kataras/iris/oteliris"
)

// NewMiddleware returns middleware that will trace incoming requests.
// opts parameter can be used to configure the middleware.
func NewMiddleware(opts ...Option) iris.Handler {
	cfg := config{}
	for _, opt := range opts {
		opt.apply(&cfg)
	}

	if cfg.TracerProvider == nil {
		cfg.TracerProvider = otel.GetTracerProvider()
	}
	tracer := cfg.TracerProvider.Tracer(
		tracerName,
		trace.WithInstrumentationVersion(SemVersion()),
	)
	if cfg.Propagators == nil {
		cfg.Propagators = otel.GetTextMapPropagator()
	}
	if cfg.SpanNameFormatter == nil {
		cfg.SpanNameFormatter = func(ctx iris.Context) string {
			return ctx.RouteName()
		}
	}

	return func(ctx iris.Context) {
		// trace all requests if no filters are specified
		for _, f := range cfg.Filters {
			// trace the request if the filter returns true
			if !f(ctx) {
				ctx.Next()
				return
			}
		}

		ctxTrace := cfg.Propagators.Extract(ctx.Request().Context(), propagation.HeaderCarrier(ctx.Request().Header))
		spanName := cfg.SpanNameFormatter(ctx)
		opts := []trace.SpanStartOption{
			trace.WithAttributes(
				append(
					httpconv.ServerRequest(cfg.ServerName, ctx.Request()),
					semconv.HTTPRoute(ctx.Path()),
				)...,
			),
			trace.WithSpanKind(trace.SpanKindServer),
		}

		ctxTrace, span := tracer.Start(ctxTrace, spanName, opts...)
		defer span.End()

		// pass the span through the request context
		*ctx.Request() = *ctx.Request().WithContext(ctxTrace)

		ctx.Next()

		span.SetStatus(httpconv.ServerStatus(ctx.GetStatusCode()))
		span.SetAttributes(semconv.HTTPStatusCode(ctx.GetStatusCode()))
	}
}
