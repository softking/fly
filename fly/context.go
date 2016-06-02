package fly

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
		if !c.handlers[c.index](c) {
			c.index = s
			return
		}
	}
}

// Next 跳过中间件
func (c *Context) Next() {
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		if !c.handlers[c.index](c) {
			c.index = s
			return
		}
	}
}
