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
	"log/slog"
	"strconv"
	"strings"
	"sync"
)

var levelMux sync.Mutex

var levelSet = map[SLevel]Level{
	SLevelDebug: LevelDebug,
	SLevelInfo:  LevelInfo,
	SLevelWarn:  LevelWarn,
	SLevelError: LevelError,
}

func RegisterLevel(ls SLevel, ln Level) {
	if ls == "" {
		return
	}
	levelMux.Lock()
	levelSet[ls] = ln
	levelMux.Unlock()
}

func ParseLevel(ls SLevel) slog.Level {
	levelMux.Lock()
	defer levelMux.Unlock()
	// If it does not exist, a zero value will be returned,
	// which is equivalent to returning slog.LevelInfo
	return levelSet[ls]
}

const (
	SLevelDebug SLevel = "debug"
	SLevelInfo  SLevel = "info"
	SLevelWarn  SLevel = "warn"
	SLevelError SLevel = "error"
)

type SLevel string

func (l SLevel) String() string {
	return string(l)
}

func (l SLevel) Level() Level {
	parts := strings.SplitN(l.String(), "+", 2)
	kind := strings.ToLower(strings.TrimSpace(parts[0]))
	level := ParseLevel(SLevel(kind))
	if len(parts) != 2 {
		return level
	}
	offset, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
	return level + Level(offset)
}

func (l SLevel) getColorPrefix() string {
	parts := strings.SplitN(string(l), "+", 2)

	var level string
	switch len(parts) {
	case 1, 2:
		level = parts[0]
	default:
		return "\x1b[32m"
	}

	switch SLevel(strings.ToLower(level)) {
	case SLevelDebug:
		return "\x1b[37m" // gray
	case SLevelInfo:
		return "\x1b[36m" // blue
	case SLevelWarn:
		return "\x1b[33m" // yellow
	case SLevelError:
		return "\x1b[31m" // red
	default:
		return "\x1b[32m" // green
	}
}

func (l SLevel) getColorSuffix() string {
	return "\x1b[0m"
}

func (l SLevel) buildColorFormat(format string) string {
	return l.getColorPrefix() + format + l.getColorSuffix()
}
