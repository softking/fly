package fly

import (
	"net/http"
	"strings"
	"net"
)

// Handler Handler
type Handler func(*Context)

// Context Context
type Context struct {
	setState bool
	state int
	index    int
	Writer   http.ResponseWriter
	Request  *http.Request
	Param  map[string]string
	Data     map[string]interface{}
	handlers []Handler
}

// WriteString 输出字符串
func (c *Context) WriteString(code int, context string) {
	if !c.setState{
		c.state = code
		c.Writer.WriteHeader(code)
		c.setState = true
		return
	}

	c.Writer.Write([]byte(context))
}

// State 获取State
func (c *Context)State()int{
	return c.state
}

// GetParam 获取参数
func (c *Context) GetParam(key string) (string, bool) {
	data, ok := c.Request.Form[key]
	if !ok || len(data) < 1 {
		return "", false
	}
	return data[0], true
}

// ClientIP 获取客户端ip
func (c *Context) ClientIP() string {
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}

func (c *Context) dispatch() {
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

// Next 跳过中间件
func (c *Context) Next() {
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

// Abort 中断中间件
func (c *Context)Abort(){
	c.index = len(c.handlers)
}
