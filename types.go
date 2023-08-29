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
	"log/slog"
)

type (
	Attr           = slog.Attr
	Record         = slog.Record
	Handler        = slog.Handler
	HandlerOptions = slog.HandlerOptions
)

type (
	Level    = slog.Level
	Leveler  = slog.Leveler
	LevelVar = slog.LevelVar
)

const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

type Kind = slog.Kind

const (
	KindAny       = slog.KindAny
	KindBool      = slog.KindBool
	KindDuration  = slog.KindDuration
	KindFloat64   = slog.KindFloat64
	KindInt64     = slog.KindInt64
	KindString    = slog.KindString
	KindTime      = slog.KindTime
	KindUint64    = slog.KindUint64
	KindGroup     = slog.KindGroup
	KindLogValuer = slog.KindLogValuer
)

const (
	// TimeKey is the key used by the built-in handlers for the time
	// when the log method is called. The associated Value is a [time.Time].
	TimeKey = slog.TimeKey
	// LevelKey is the key used by the built-in handlers for the level
	// of the log call. The associated value is a [Level].
	LevelKey = slog.LevelKey
	// MessageKey is the key used by the built-in handlers for the
	// message of the log call. The associated value is a string.
	MessageKey = slog.MessageKey
	// SourceKey is the key used by the built-in handlers for the source file
	// and line of the log call. The associated value is a string.
	SourceKey = slog.SourceKey
)

const BadKey = "!BADKEY"

func argsToAttrSlice(args []any) []Attr {
	var (
		attr  Attr
		attrs []Attr
	)
	for len(args) > 0 {
		attr, args = argsToAttr(args)
		attrs = append(attrs, attr)
	}
	return attrs
}

func argsToAttr(args []any) (Attr, []any) {
	switch x := args[0].(type) {
	case string:
		if len(args) == 1 {
			return slog.String(BadKey, x), nil
		}
		return slog.Any(x, args[1]), args[2:]

	case Attr:
		return x, args[1:]

	default:
		return slog.Any(BadKey, x), args[1:]
	}
}

func needsQuoting(s string) bool {
	if len(s) == 0 {
		return true
	}
	for _, b := range s {
		if !((b >= 'a' && b <= 'z') ||
			(b >= 'A' && b <= 'Z') ||
			(b >= '0' && b <= '9') ||
			b == '-' || b == '.' || b == '_' || b == '/' || b == '@' || b == '^' || b == '+') {
			return true
		}
	}
	return false
}

const (
	quoteChar  = 34
	splitChar  = 61
	sepChar    = 32
	escapeChar = 92
)

var quoteSuffix = []byte{quoteChar, sepChar}

func convertToColorKey(b []byte, colorPrefix, colorSuffix []byte) []byte {
	bl := len(b)
	if bl == 0 {
		return b
	}

	var buf bytes.Buffer
	for {
		index := bytes.IndexByte(b, splitChar)
		if index == -1 {
			buf.Write(b)
			break
		}

		key := b[:index]
		buf.Write(colorPrefix)
		buf.Write(key)
		buf.Write(colorSuffix)
		buf.WriteByte(splitChar)

		val := b[index+1:]
		index = bytes.IndexByte(val, quoteChar)
		// match quote prefix
		if index == 0 {
			buf.WriteByte(quoteChar)
			val = val[1:]

			// 循环查找 结束符，如果找到转义的结束符，继续查找
			var eof bool
			for {
				index = bytes.Index(val, quoteSuffix)
				// break when the quote suffix is not matched
				if index == -1 {
					buf.Write(val)
					eof = true
					break
				}

				buf.Write(val[:index])
				buf.Write(quoteSuffix)
				if index > 0 && val[index-1] != escapeChar {
					b = val[index+2:]
					break
				}
				val = val[index+2:]
			}
			if eof {
				break
			}
			continue
		}

		index = bytes.IndexByte(val, sepChar)
		// break when the sep is not matched
		if index == -1 {
			buf.Write(val)
			break
		}
		buf.Write(val[:index])
		buf.WriteByte(sepChar)
		b = val[index+1:]
	}
	return buf.Bytes()
}
