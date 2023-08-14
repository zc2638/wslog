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
	"log/slog"
	"os"
	"strings"
	"sync/atomic"
)

type Config struct {
	Level  Level  `json:"level,omitempty" yaml:"level,omitempty"`
	Format string `json:"format,omitempty" yaml:"format,omitempty"`
	Source bool   `json:"source,omitempty" yaml:"source,omitempty"`

	Filename   string `json:"filename,omitempty" yaml:"filename,omitempty"`
	MaxSize    int    `json:"maxSize,omitempty" yaml:"maxSize,omitempty"`
	MaxAge     int    `json:"maxAge,omitempty" yaml:"maxAge,omitempty"`
	MaxBackups int    `json:"maxBackups,omitempty" yaml:"maxBackups,omitempty"`
	LocalTime  bool   `json:"localTime,omitempty" yaml:"localTime,omitempty"`
	Compress   bool   `json:"compress,omitempty" yaml:"compress,omitempty"`
}

func New(cfg Config, opts ...Option) *Logger {
	writer := NewWriter(cfg)
	level := new(slog.LevelVar)
	level.Set(cfg.Level.Level())

	params := &Params{Writer: writer, Level: level}
	for _, v := range opts {
		v(params)
	}

	opt := &slog.HandlerOptions{
		AddSource:   cfg.Source,
		Level:       params.Level,
		ReplaceAttr: params.ReplaceAttr,
	}

	if params.Handler == nil {
		if strings.ToLower(cfg.Format) == "json" {
			params.Handler = slog.NewJSONHandler(writer, opt)
		} else {
			params.Handler = slog.NewTextHandler(writer, opt)
		}
	}
	return NewLogger(params.Handler, params.Level, params.Writer)
}

var defaultLogger atomic.Value

func init() {
	defaultLogger.Store(NewLogger(slog.NewTextHandler(os.Stdout, nil)))
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
	Default().log(emptyCtx, slog.LevelDebug, msg, args...)
}

// Debugf calls Logger.Debugf on the default logger.
func Debugf(format string, args ...any) {
	Default().log(emptyCtx, slog.LevelDebug, fmt.Sprintf(format, args...))
}

// DebugCtx calls Logger.DebugCtx on the default logger.
func DebugCtx(ctx context.Context, msg string, args ...any) {
	Default().log(ctx, slog.LevelDebug, msg, args...)
}

// Info calls Logger.Info on the default logger.
func Info(msg string, args ...any) {
	Default().log(emptyCtx, slog.LevelInfo, msg, args...)
}

// Infof calls Logger.Infof on the default logger.
func Infof(format string, args ...any) {
	Default().log(emptyCtx, slog.LevelInfo, fmt.Sprintf(format, args...))
}

// InfoCtx calls Logger.InfoCtx on the default logger.
func InfoCtx(ctx context.Context, msg string, args ...any) {
	Default().log(ctx, slog.LevelInfo, msg, args...)
}

// Warn calls Logger.Warn on the default logger.
func Warn(msg string, args ...any) {
	Default().log(emptyCtx, slog.LevelWarn, msg, args...)
}

// Warnf calls Logger.Warnf on the default logger.
func Warnf(format string, args ...any) {
	Default().log(emptyCtx, slog.LevelWarn, fmt.Sprintf(format, args...))
}

// WarnCtx calls Logger.WarnCtx on the default logger.
func WarnCtx(ctx context.Context, msg string, args ...any) {
	Default().log(ctx, slog.LevelWarn, msg, args...)
}

// Error calls Logger.Error on the default logger.
func Error(msg string, args ...any) {
	Default().log(emptyCtx, slog.LevelError, msg, args...)
}

// Errorf calls Logger.Errorf on the default logger.
func Errorf(format string, args ...any) {
	Default().log(emptyCtx, slog.LevelError, fmt.Sprintf(format, args...))
}

// ErrorCtx calls Logger.ErrorCtx on the default logger.
func ErrorCtx(ctx context.Context, msg string, args ...any) {
	Default().log(ctx, slog.LevelError, msg, args...)
}

// Log calls Logger.Log on the default logger.
func Log(level slog.Level, msg string, args ...any) {
	Default().log(emptyCtx, level, msg, args...)
}

// LogCtx calls Logger.LogCtx on the default logger.
func LogCtx(ctx context.Context, level slog.Level, msg string, args ...any) {
	Default().log(ctx, level, msg, args...)
}

// LogAttrs calls Logger.LogAttrs on the default logger.
func LogAttrs(level slog.Level, msg string, attrs ...slog.Attr) {
	Default().logAttrs(emptyCtx, level, msg, attrs...)
}

// LogAttrsCtx calls Logger.LogAttrsCtx on the default logger.
func LogAttrsCtx(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	Default().logAttrs(ctx, level, msg, attrs...)
}
