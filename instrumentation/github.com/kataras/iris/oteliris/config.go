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

// base on https://github.com/open-telemetry/opentelemetry-go-contrib/blob/main/instrumentation/github.com/gin-gonic/gin/otelgin/option.go

package oteliris // import "go.opentelemetry.io/contrib/instrumentation/github.com/kataras/iris/oteliris"

import (
	"github.com/kataras/iris/v12"

	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type config struct {
	ServerName        string
	TracerProvider    oteltrace.TracerProvider
	Propagators       propagation.TextMapPropagator
	Filters           []Filter
	SpanNameFormatter SpanNameFormatter
}

// Filter is a predicate used to determine whether a given iris.Context should be traced.
// A Filter must return true if the request should be traced.
// If no filters are provided then **all requests are traced.**.
type Filter func(iris.Context) bool

// SpanNameFormatter is used to set span name from iris.Context.
type SpanNameFormatter func(iris.Context) string

// Option specifies instrumentation configuration options.
type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

// WithPropagators specifies propagators to use for extracting
// information from the HTTP requests. If none are specified, global
// ones will be used.
func WithPropagators(propagators propagation.TextMapPropagator) Option {
	return optionFunc(func(cfg *config) {
		if propagators != nil {
			cfg.Propagators = propagators
		}
	})
}

// WithTracerProvider specifies a tracer provider to use for creating a tracer.
// If none is specified, the global provider is used.
func WithTracerProvider(provider oteltrace.TracerProvider) Option {
	return optionFunc(func(cfg *config) {
		if provider != nil {
			cfg.TracerProvider = provider
		}
	})
}

// WithFilter adds a filter to the list of filters used by the handler.
// If any filter indicates to exclude a request then the request will not be
// traced. All filters must allow a request to be traced for a Span to be created.
// If no filters are provided then all requests are traced.
// Filters will be invoked for each processed request, it is advised to make them
// simple and fast.
func WithFilter(f ...Filter) Option {
	return optionFunc(func(c *config) {
		c.Filters = append(c.Filters, f...)
	})
}

// WithSpanNameFormatter takes a function that will be called on every
// request and the returned string will become the Span Name.
func WithSpanNameFormatter(f func(ctx iris.Context) string) Option {
	return optionFunc(func(c *config) {
		c.SpanNameFormatter = f
	})
}

// WithServiceName sets the service name for the tracer.
func WithServiceName(serviceName string) Option {
	return optionFunc(func(cfg *config) {
		cfg.ServerName = serviceName
	})
}