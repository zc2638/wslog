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

var levelSet = map[Level]slog.Level{
	LevelDebug: slog.LevelDebug,
	LevelInfo:  slog.LevelInfo,
	LevelWarn:  slog.LevelWarn,
	LevelError: slog.LevelError,
}

func RegisterLevel(ls Level, ln slog.Level) {
	levelMux.Lock()
	levelSet[ls] = ln
	levelMux.Unlock()
}

func ParseLevel(ls Level) slog.Level {
	levelMux.Lock()
	defer levelMux.Unlock()
	// If it does not exist, a zero value will be returned,
	// which is equivalent to returning slog.LevelInfo
	return levelSet[ls]
}

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

type Level string

func (l Level) String() string {
	return string(l)
}

func (l Level) Level() slog.Level {
	parts := strings.SplitN(l.String(), "+", 2)
	kind := strings.ToLower(strings.TrimSpace(parts[0]))
	level := ParseLevel(Level(kind))
	if len(parts) != 2 {
		return level
	}
	offset, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
	return level + slog.Level(offset)
}