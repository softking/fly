package fly

import "net/http"

// Handler Handler
type Handler func(*Context)

// Context Context
type Context struct {
	index    int
	Writer   http.ResponseWriter
	Request  *http.Request
	Param  map[string]string
	Data     map[string]interface{}
	handlers []Handler
}

// WriteString 输出字符串
func (c *Context) WriteString(context string) {
	c.Writer.Write([]byte(context))
}

// GetParam 获取参数
func (c *Context) GetParam(key string) (string, bool) {
	data, ok := c.Request.Form[key]
	if !ok || len(data) < 1 {
		return "", false
	}
	return data[0], true
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

func (c *Context)Abort(){
	c.index = len(c.handlers)
}
