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
	"runtime"
	"time"
)

var emptyCtx = context.Background()

// NewLogger creates a new Logger with the given non-nil Handler.
func NewLogger(h Handler) *Logger {
	return NewLoggerSkip(h, 3)
}

func NewLoggerSkip(h Handler, skip int) *Logger {
	if h == nil {
		panic("nil Handler")
	}
	l := &Logger{handler: h, skip: skip}
	return l
}

type Logger struct {
	handler Handler
	skip    int
}

func (l *Logger) clone() *Logger {
	c := *l
	return &c
}

func (l *Logger) Handler() Handler { return l.handler }

func (l *Logger) With(args ...any) *Logger {
	if len(args) == 0 {
		return l
	}
	c := l.clone()
	c.handler = l.handler.WithAttrs(argsToAttrSlice(args))
	return c
}

// WithGroup returns a Logger that starts a group if the name is non-empty.
// The keys of all attributes added to the Logger will be qualified by the given
// name. (How that qualification happens depends on the [Handler.WithGroup]
// method of the Logger's Handler.)
//
// If the name is empty, WithGroup returns the receiver.
func (l *Logger) WithGroup(name string) *Logger {
	if name == "" {
		return l
	}
	c := l.clone()
	c.handler = l.handler.WithGroup(name)
	return c

}

// EnabledCtx reports whether l emits log records at the given context and level.
func (l *Logger) EnabledCtx(ctx context.Context, level Level) bool {
	if ctx == nil {
		ctx = emptyCtx
	}
	return l.Handler().Enabled(ctx, level)
}

// Enabled reports whether l emits log records at the given level.
func (l *Logger) Enabled(level Level) bool {
	return l.Handler().Enabled(emptyCtx, level)
}

// LogCtx emitting a log record with the current time and the given level and message.
// The Record's Attrs consist of the Logger's attributes followed by
// the Attrs specified by args.
//
// The attribute arguments are processed as follows:
//   - If an argument is an Attr, it is used as is.
//   - If an argument is a string and this is not the last argument,
//     the following argument is treated as the value and the two are combined
//     into an Attr.
//   - Otherwise, the argument is treated as a value with key `!BADKEY`.
func (l *Logger) LogCtx(ctx context.Context, level Level, msg string, args ...any) {
	l.log(ctx, level, msg, args...)
}

func (l *Logger) Log(level Level, msg string, args ...any) {
	l.log(emptyCtx, level, msg, args...)
}

// LogAttrsCtx is a more efficient version of [Logger.Log] that accepts only Attrs.
func (l *Logger) LogAttrsCtx(ctx context.Context, level Level, msg string, attrs ...Attr) {
	l.logAttrs(ctx, level, msg, attrs...)
}

func (l *Logger) LogAttrs(level Level, msg string, attrs ...Attr) {
	l.logAttrs(emptyCtx, level, msg, attrs...)
}

// Debug logs at LevelDebug.
func (l *Logger) Debug(msg string, args ...any) {
	l.log(emptyCtx, LevelDebug, msg, args...)
}

// Debugf logs at LevelDebug with the given format.
func (l *Logger) Debugf(format string, args ...any) {
	l.log(emptyCtx, LevelDebug, fmt.Sprintf(format, args...))
}

// DebugCtx logs at LevelDebug with the given context.
func (l *Logger) DebugCtx(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelDebug, msg, args...)
}

// Info logs at LevelInfo.
func (l *Logger) Info(msg string, args ...any) {
	l.log(context.Background(), LevelInfo, msg, args...)
}

// Infof logs at LevelInfo with the given format.
func (l *Logger) Infof(format string, args ...any) {
	l.log(emptyCtx, LevelInfo, fmt.Sprintf(format, args...))
}

// InfoCtx logs at LevelInfo with the given context.
func (l *Logger) InfoCtx(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelInfo, msg, args...)
}

// Warn logs at LevelWarn.
func (l *Logger) Warn(msg string, args ...any) {
	l.log(context.Background(), LevelWarn, msg, args...)
}

// Warnf logs at LevelWarn with the given format.
func (l *Logger) Warnf(format string, args ...any) {
	l.log(emptyCtx, LevelWarn, fmt.Sprintf(format, args...))
}

// WarnCtx logs at LevelWarn with the given context.
func (l *Logger) WarnCtx(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelWarn, msg, args...)
}

// Error logs at LevelError.
func (l *Logger) Error(msg string, args ...any) {
	l.log(emptyCtx, LevelError, msg, args...)
}

// Errorf logs at LevelError with the given format.
func (l *Logger) Errorf(format string, args ...any) {
	l.log(emptyCtx, LevelError, fmt.Sprintf(format, args...))
}

// ErrorCtx logs at LevelError with the given context.
func (l *Logger) ErrorCtx(ctx context.Context, msg string, args ...any) {
	l.log(ctx, LevelError, msg, args...)
}

// log is the low-level logging method for methods that take ...any.
// It must always be called directly by an exported logging method
// or function, because it uses a fixed call depth to obtain the pc.
func (l *Logger) log(ctx context.Context, level Level, msg string, args ...any) {
	if !l.EnabledCtx(ctx, level) {
		return
	}

	var pcs [1]uintptr
	// skip [runtime.Callers, this function, this function's caller]
	runtime.Callers(l.skip, pcs[:])
	pc := pcs[0]

	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.Add(args...)
	if ctx == nil {
		ctx = emptyCtx
	}
	_ = l.Handler().Handle(ctx, r)
}

// logAttrs is like [Logger.log], but for methods that take ...Attr.
func (l *Logger) logAttrs(ctx context.Context, level Level, msg string, attrs ...Attr) {
	if !l.EnabledCtx(ctx, level) {
		return
	}

	var pcs [1]uintptr
	// skip [runtime.Callers, this function, this function's caller]
	runtime.Callers(l.skip, pcs[:])
	pc := pcs[0]

	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.AddAttrs(attrs...)
	if ctx == nil {
		ctx = emptyCtx
	}
	_ = l.Handler().Handle(ctx, r)
}
