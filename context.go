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
	"net/http"
)

type loggerKey struct{}

// WithContext returns a new context with the provided logger.
// Use in combination with logger.With(key, value) for great effect.
func WithContext(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// FromContext retrieves the current logger from the context.
// If no logger is available, the default logger is returned.
func FromContext(ctx context.Context) *Logger {
	logger := ctx.Value(loggerKey{})
	if logger == nil {
		return Default()
	}
	return logger.(*Logger)
}

// FromRequest retrieves the current logger from the request.
// If no logger is available, the default logger is returned.
func FromRequest(r *http.Request) *Logger {
	return FromContext(r.Context())
}
