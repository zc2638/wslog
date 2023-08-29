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
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

func NewLogHandler(w io.Writer, opts *HandlerOptions, disableColor bool) Handler {
	if opts == nil {
		opts = new(HandlerOptions)
	}
	return &logHandler{
		w:            w,
		opts:         *opts,
		mu:           new(sync.Mutex),
		sep:          ".",
		disableColor: disableColor,
	}
}

type logHandler struct {
	w    io.Writer
	opts HandlerOptions
	mu   *sync.Mutex

	sep          string
	groups       []string
	attrBuffer   bytes.Buffer
	disableColor bool
}

func (h *logHandler) clone() *logHandler {
	return &logHandler{
		mu:           h.mu, // mutex shared among all clones of this handler
		w:            h.w,
		opts:         h.opts,
		sep:          h.sep,
		groups:       slices.Clip(h.groups),
		attrBuffer:   h.attrBuffer,
		disableColor: h.disableColor,
	}
}

func (h *logHandler) Enabled(_ context.Context, level Level) bool {
	minLevel := LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

func (h *logHandler) Handle(_ context.Context, record Record) error {
	var (
		defBuf  bytes.Buffer
		attrBuf bytes.Buffer
	)

	logTime := record.Time.Round(0)
	defAttrs := []Attr{
		slog.Any(LevelKey, record.Level),        // level
		slog.Time(TimeKey, logTime),             // time: strip monotonic to match Attr behavior
		slog.String(MessageKey, record.Message), // message
	}
	h.addAttrs(&defBuf, nil, defAttrs)
	defBuf.WriteString(" ")

	// source
	if h.opts.AddSource {
		fs := runtime.CallersFrames([]uintptr{record.PC})
		f, _ := fs.Next()
		source := &slog.Source{
			Function: f.Function,
			File:     f.File,
			Line:     f.Line,
		}
		sourceAttr := slog.Any(SourceKey, source)
		h.addAttrs(&attrBuf, nil, []Attr{sourceAttr})
	}

	attrBuf.Write(h.attrBuffer.Bytes())
	extraAttrs := make([]Attr, 0, record.NumAttrs())
	record.Attrs(func(attr slog.Attr) bool {
		extraAttrs = append(extraAttrs, attr)
		return true
	})
	h.addAttrs(&attrBuf, nil, extraAttrs)

	attrBytes := attrBuf.Bytes()
	if !h.disableColor {
		slevel := SLevel(record.Level.String())
		colorPrefix, colorSuffix := slevel.getColorPrefix(), slevel.getColorSuffix()
		attrBytes = convertToColorKey(attrBytes, []byte(colorPrefix), []byte(colorSuffix))
	}

	defBuf.Write(attrBytes)
	// TODO write record attr
	defBuf.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.w.Write(defBuf.Bytes())
	return err
}

func (h *logHandler) WithGroup(name string) Handler {
	cp := h.clone()
	cp.groups = append(cp.groups, name)
	return cp
}

func (h *logHandler) WithAttrs(attrs []Attr) Handler {
	cp := h.clone()
	groups := make([]string, len(cp.groups))
	copy(groups[:], cp.groups[:])
	cp.addAttrs(&cp.attrBuffer, groups, attrs)
	return cp
}

func (h *logHandler) addAttrs(buf *bytes.Buffer, groups []string, attrs []Attr) {
	groupPrefix := strings.Join(groups, ".")
	for _, a := range attrs {
		if raFn := h.opts.ReplaceAttr; raFn != nil && a.Value.Kind() != KindGroup {
			a.Value = a.Value.Resolve()
			a = raFn(groups, a)
		}
		a.Value = a.Value.Resolve()

		// Elide empty Attrs.
		if a.Key == "" {
			continue
		}

		kind := a.Value.Kind()
		switch kind {
		case KindAny:
			// Special case: Source.
			if src, ok := a.Value.Any().(*slog.Source); ok {
				a.Value = slog.StringValue(fmt.Sprintf("%s:%d", src.File, src.Line))
			}
		case KindGroup:
			as := a.Value.Group()
			// Output only non-empty groups.
			if len(as) > 0 {
				// Inline a group with an empty key.
				g2 := make([]string, 0, len(groups)+1)
				g2 = append(g2, groups...)
				if a.Key != "" {
					g2 = append(g2, a.Key)
				}
				h.addAttrs(buf, g2, as)
			}
			continue
		}

		switch a.Key {
		case LevelKey:
			levelStr := a.Value.String()
			if !h.disableColor {
				slevel := SLevel(levelStr)
				format := slevel.buildColorFormat("%s")
				levelStr = fmt.Sprintf(format, levelStr)
			}
			buf.WriteString(levelStr)
		case TimeKey:
			buf.WriteString("[")
			if kind == KindTime {
				buf.WriteString(a.Value.Time().Format(time.RFC3339))
			} else {
				buf.WriteString(a.Value.String())
			}
			buf.WriteString("]")
		case MessageKey:
			buf.WriteString(" ")
			buf.WriteString(a.Value.String())
		default:
			buf.WriteString(" ")
			if groupPrefix != "" {
				buf.WriteString(groupPrefix)
				buf.WriteString(h.sep)
			}
			str := a.Value.String()
			if needsQuoting(str) {
				str = strconv.Quote(str)
			}
			buf.WriteString(a.Key)
			buf.WriteString("=")
			buf.WriteString(str)
		}
	}
}

func NewMultiHandler(handlers ...Handler) Handler {
	return &multiHandler{handlers: handlers}
}

type multiHandler struct {
	handlers []Handler
}

func (h *multiHandler) Enabled(ctx context.Context, level Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *multiHandler) Handle(ctx context.Context, record Record) error {
	var errs []error
	for _, handler := range h.handlers {
		if err := handler.Handle(ctx, record); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (h *multiHandler) WithAttrs(attrs []Attr) Handler {
	cp := &multiHandler{handlers: make([]Handler, len(h.handlers))}
	for index, handler := range h.handlers {
		cp.handlers[index] = handler.WithAttrs(attrs)
	}
	return cp
}

func (h *multiHandler) WithGroup(name string) Handler {
	cp := &multiHandler{handlers: make([]Handler, len(h.handlers))}
	for index, handler := range h.handlers {
		cp.handlers[index] = handler.WithGroup(name)
	}
	return cp
}
