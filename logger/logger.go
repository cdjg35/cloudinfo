// Copyright © 2018 Banzai Cloud
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

package logger

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ctxMarker struct{}

var (
	ctxKey = &ctxMarker{}
	logger = logrus.New() // default logger
)

const (
	correlationIdKey = "correlation-id"
	scrapeIdFullKey  = "scrape-id-full"
	scrapeIdShortKey = "scrape-id-short"

	providerKey = "provider"
	serviceKey  = "service"
	regionKey   = "region"
)

// InitLogger sets level and format for Logger
func InitLogger(level, format string) {

	logger = newLogger(Config{
		Level:  level,
		Format: format,
	})

}

// Config holds information necessary for customizing the logger.
type Config struct {
	Level  string
	Format string
}

func newLogger(config Config) *logrus.Logger {
	logger := logrus.New()

	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}

	logger.Level = level

	switch config.Format {
	case "json":
		logger.Formatter = new(logrus.JSONFormatter)

	default:
		textFormatter := new(logrus.TextFormatter)
		textFormatter.FullTimestamp = true

		logger.Formatter = textFormatter
	}

	return logger
}

// Extract assembles the entry with the fields extracted from the context
func Extract(ctx context.Context) ContextLogger {

	fds, ok := ctx.Value(ctxKey).(map[string]interface{})
	if !ok || fds == nil {
		return logrus.NewEntry(logger)
	}

	fields := logrus.Fields{}
	for k, v := range fds {
		fields[k] = v
	}

	return logger.WithFields(fields)
}

// ToContext adds
func ToContext(ctx context.Context, fields map[string]interface{}) context.Context {

	// retrieving the "parent" context
	parentVals, ok := ctx.Value(ctxKey).(map[string]interface{})

	if parentVals == nil && ok {
		// there is no logger context set in the parent
		return context.WithValue(ctx, ctxKey, fields)
	}

	if ok { // the parent context is successfully retrieved
		for k, v := range parentVals { // copy parent context values into the current context
			fields[k] = v
		}
	}

	return context.WithValue(ctx, ctxKey, fields)
}

// GetCorrelationId get correlation id from gin context
func GetCorrelationId(c *gin.Context) string {
	id := c.GetString(ContextKey)
	return id
}

// LogEntryWrapper wraps the logger entry implementation
// By embedding the library specific entry (logrus here), we have the default implementation "out of the box"
type LogEntryWrapper struct {
	// the default logging library is logrus
	*logrus.Entry
}

// ContextLogger gathers all the log operations used in the application, mainly operations implemented by "conventional" loggers
// The interface is meant to decouple application dependency on logger libraries
type ContextLogger interface {
	WithError(err error) *logrus.Entry
	WithField(key string, value interface{}) *logrus.Entry
	WithFields(fields logrus.Fields) *logrus.Entry
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Fatal(args ...interface{})
}

// logCtxBuilder helper struct to build the context for logging purposes
type logCtxBuilder struct {
	ctx map[string]interface{}
}

// NewLogCtxBuilder creates a new struct initializes it and returns the pointer to it
func NewLogCtxBuilder() *logCtxBuilder {
	lCtx := logCtxBuilder{}
	lCtx.init()
	return &lCtx
}

func (cb *logCtxBuilder) init() {
	if cb.ctx == nil {
		cb.ctx = make(map[string]interface{})
	}
}

// WithProvider sets the provider into the logger context
func (cb *logCtxBuilder) WithProvider(provider string) *logCtxBuilder {
	return cb.WithField(providerKey, provider)
}

// WithService sets the service into the logger context
func (cb *logCtxBuilder) WithService(service string) *logCtxBuilder {
	return cb.WithField(serviceKey, service)
}

// WithRegion sets the region into the logger context
func (cb *logCtxBuilder) WithRegion(region string) *logCtxBuilder {
	return cb.WithField(regionKey, region)
}

// // WithCorrelationId sets the correlation id into the logger context
func (cb *logCtxBuilder) WithCorrelationId(cid string) *logCtxBuilder {
	return cb.WithField(correlationIdKey, cid)
}

// WithScrapeIdShort sets the short lived scraping identifier into the logger context
func (cb *logCtxBuilder) WithScrapeIdShort(id interface{}) *logCtxBuilder {
	return cb.WithField(scrapeIdShortKey, id)
}

// WithScrapeIdFull sets the scraping identifier into the logger context
func (cb *logCtxBuilder) WithScrapeIdFull(id interface{}) *logCtxBuilder {
	return cb.WithField(scrapeIdFullKey, id)
}

// WithField adds an arbitrary value to the logger context with the provided keys
func (cb *logCtxBuilder) WithField(field string, value interface{}) *logCtxBuilder {
	cb.init()
	cb.ctx[field] = value
	return cb
}

// Build gets the map representing logger context
func (cb *logCtxBuilder) Build() map[string]interface{} {
	cb.init()
	return cb.ctx
}
