// Copyright © 2022 Attestant Limited.
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

package loggers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// NewGinLogger creates a gin logger using the supplied zerolog.
//
//nolint:zerologlint
func NewGinLogger(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()

		c.Next()

		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		if raw != "" {
			path = path + "?" + raw
		}

		var e *zerolog.Event

		switch {
		case c.Writer.Status() >= 400 && c.Writer.Status() < 500:
			e = log.Warn()
		case c.Writer.Status() >= 500 && c.Writer.Status() < 600:
			e = log.Error()
		default:
			e = log.Trace()
		}

		e.
			Int("status_code", c.Writer.Status()).
			Dur("latency_ms", time.Since(started)).
			Str("client", c.ClientIP()).
			Str("method", c.Request.Method).
			Str("path", path).
			Msg(c.Errors.ByType(gin.ErrorTypePrivate).String())
	}
}
