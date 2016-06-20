// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package midware

import (
	"fmt"
	"io"
	"time"

	"github.com/softking/fly"
	"os"
)

// DefaultWriter 默认的输出
var DefaultWriter io.Writer = os.Stdout

// Logger 日志中间件
func Logger(c *fly.Context) {
	out := DefaultWriter
	// Start timer
	start := time.Now()
	path := c.Request.URL.Path

	// Process request
	c.Next()

	end := time.Now()
	latency := end.Sub(start)
	clientIP := c.ClientIP()
	method := c.Request.Method
	state := c.State()

	fmt.Fprintf(out, "[GIN] %v | %13v | %d |%s  %s %s\n",
		end.Format("2006/01/02 - 15:04:05"),
		latency,
		state,
		clientIP,
		method,
		path,
	)

}
