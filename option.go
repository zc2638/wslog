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
	"io"
	"log/slog"
)

type ReplaceAttrFunc func(groups []string, a slog.Attr) slog.Attr

type Params struct {
	Writer      io.Writer
	Level       slog.Leveler
	ReplaceAttr ReplaceAttrFunc
}

type Option func(params *Params)

func WriterOption(w io.Writer) Option {
	return func(params *Params) {
		if w == nil {
			return
		}
		params.Writer = w
	}
}

func LevelOption(level slog.Leveler) Option {
	return func(params *Params) {
		if level == nil {
			return
		}
		params.Level = level
	}
}

func ReplaceAttrOption(fn ReplaceAttrFunc) Option {
	return func(params *Params) {
		params.ReplaceAttr = fn
	}
}
