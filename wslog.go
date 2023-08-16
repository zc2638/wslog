// Copyright © 2023 zc2638 <zc2638@qq.com>.
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

package wslog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync/atomic"
)

type Config struct {
	Level  SLevel `json:"level,omitempty" yaml:"level,omitempty"`
	Format string `json:"format,omitempty" yaml:"format,omitempty"`
	Source bool   `json:"source,omitempty" yaml:"source,omitempty"`

	// only use for default log handler
	Colorful bool `json:"colorful,omitempty" yaml:"colorful,omitempty"`

	Filename   string `json:"filename,omitempty" yaml:"filename,omitempty"`
	MaxSize    int    `json:"maxSize,omitempty" yaml:"maxSize,omitempty"`
	MaxAge     int    `json:"maxAge,omitempty" yaml:"maxAge,omitempty"`
	MaxBackups int    `json:"maxBackups,omitempty" yaml:"maxBackups,omitempty"`
	LocalTime  bool   `json:"localTime,omitempty" yaml:"localTime,omitempty"`
	Compress   bool   `json:"compress,omitempty" yaml:"compress,omitempty"`
}

func (c *Config) HandlerOptions() *HandlerOptions {
	level := new(LevelVar)
	level.Set(c.Level.Level())
	return &HandlerOptions{
		AddSource: c.Source,
		Level:     level,
	}
}

func (c *Config) Writer() io.Writer {
	return NewWriter(*c)
}

func New(cfg Config, opts ...any) *Logger {
	handlerOpts := cfg.HandlerOptions()

	var (
		handler Handler
		writer  io.Writer
	)
	for _, opt := range opts {
		switch v := opt.(type) {
		case io.Writer:
			writer = v
		case *HandlerOptions:
			if v != nil {
				handlerOpts = v
			}
		case func(groups []string, a Attr) Attr:
			handlerOpts.ReplaceAttr = v
		case Leveler:
			handlerOpts.Level = v
		case Handler:
			handler = v
		}
	}

	if handler == nil {
		writer = cfg.Writer()
		switch strings.ToLower(cfg.Format) {
		case "json":
			handler = slog.NewJSONHandler(writer, handlerOpts)
		case "text":
			handler = slog.NewTextHandler(writer, handlerOpts)
		default:
			handler = NewLogHandler(writer, handlerOpts, cfg.Colorful)
		}
	}
	return NewLogger(handler)
}

var defaultLogger atomic.Value

func init() {
	defaultLogger.Store(NewLogger(NewLogHandler(os.Stdout, nil, true)))
}

// Default returns the default Logger.
func Default() *Logger { return defaultLogger.Load().(*Logger) }

// SetDefault makes l the default Logger.
// After this call, output from the log package's default Logger
// (as with [log.Print], etc.) will be logged at LevelInfo using l's Handler.
func SetDefault(l *Logger) {
	defaultLogger.Store(l)
}

// With calls Logger.With on the default logger.
func With(args ...any) *Logger {
	return Default().With(args...)
}

// Debug calls Logger.Debug on the default logger.
func Debug(msg string, args ...any) {
	Default().log(emptyCtx, LevelDebug, msg, args...)
}

// Debugf calls Logger.Debugf on the default logger.
func Debugf(format string, args ...any) {
	Default().log(emptyCtx, LevelDebug, fmt.Sprintf(format, args...))
}

// DebugCtx calls Logger.DebugCtx on the default logger.
func DebugCtx(ctx context.Context, msg string, args ...any) {
	Default().log(ctx, LevelDebug, msg, args...)
}

// Info calls Logger.Info on the default logger.
func Info(msg string, args ...any) {
	Default().log(emptyCtx, LevelInfo, msg, args...)
}

// Infof calls Logger.Infof on the default logger.
func Infof(format string, args ...any) {
	Default().log(emptyCtx, LevelInfo, fmt.Sprintf(format, args...))
}

// InfoCtx calls Logger.InfoCtx on the default logger.
func InfoCtx(ctx context.Context, msg string, args ...any) {
	Default().log(ctx, LevelInfo, msg, args...)
}

// Warn calls Logger.Warn on the default logger.
func Warn(msg string, args ...any) {
	Default().log(emptyCtx, LevelWarn, msg, args...)
}

// Warnf calls Logger.Warnf on the default logger.
func Warnf(format string, args ...any) {
	Default().log(emptyCtx, LevelWarn, fmt.Sprintf(format, args...))
}

// WarnCtx calls Logger.WarnCtx on the default logger.
func WarnCtx(ctx context.Context, msg string, args ...any) {
	Default().log(ctx, LevelWarn, msg, args...)
}

// Error calls Logger.Error on the default logger.
func Error(msg string, args ...any) {
	Default().log(emptyCtx, LevelError, msg, args...)
}

// Errorf calls Logger.Errorf on the default logger.
func Errorf(format string, args ...any) {
	Default().log(emptyCtx, LevelError, fmt.Sprintf(format, args...))
}

// ErrorCtx calls Logger.ErrorCtx on the default logger.
func ErrorCtx(ctx context.Context, msg string, args ...any) {
	Default().log(ctx, LevelError, msg, args...)
}

// Log calls Logger.Log on the default logger.
func Log(level Level, msg string, args ...any) {
	Default().log(emptyCtx, level, msg, args...)
}

// LogCtx calls Logger.LogCtx on the default logger.
func LogCtx(ctx context.Context, level Level, msg string, args ...any) {
	Default().log(ctx, level, msg, args...)
}

// LogAttrs calls Logger.LogAttrs on the default logger.
func LogAttrs(level Level, msg string, attrs ...Attr) {
	Default().logAttrs(emptyCtx, level, msg, attrs...)
}

// LogAttrsCtx calls Logger.LogAttrsCtx on the default logger.
func LogAttrsCtx(ctx context.Context, level Level, msg string, attrs ...Attr) {
	Default().logAttrs(ctx, level, msg, attrs...)
}
