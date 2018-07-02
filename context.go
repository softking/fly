package fly

import (
	"net"
	"net/http"
	"net/url"
	"strings"
	"encoding/json"
	"log"
)

// Handler Handler
type Handler func(*Context)

// Context Context
type Context struct {
	setState bool
	state    int
	index    int
	Writer   http.ResponseWriter
	Request  *http.Request
	Params   map[string]string
	Data     map[string]interface{}
	handlers []Handler
}

// SetCookie SetCookie
func (c *Context) SetCookie(
	name string,
	value string,
	maxAge int,
	path string,
	domain string,
	secure bool,
	httpOnly bool,
) {
	if path == "" {
		path = "/"
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
}

// NewContext NewContext
func NewContext(state int) *Context {
	return &Context{
		state: state,
	}
}

// Redirect Redirect
func (c *Context) Redirect(code int, location string) {
	http.Redirect(c.Writer, c.Request, location, code)
}

// Header is a intelligent shortcut for c.Writer.Header().Set(key, value)
func (c *Context) Header(key, value string) {
	if len(value) == 0 {
		c.Writer.Header().Del(key)
	} else {
		c.Writer.Header().Set(key, value)
	}
}

func (c *Context) Cookie(name string) (string, error) {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return "", err
	}
	val, _ := url.QueryUnescape(cookie.Value)
	return val, nil
}

// Put Put
func (c *Context) Put(key string, value interface{}) {
	c.Data[key] = value
}

// Get Get
func (c *Context) Get(key string) (interface{}, bool) {
	data, has := c.Data[key]
	return data, has
}

// SetCode http code
func (c *Context) SetCode(code int) {
	if !c.setState {
		c.state = code
		c.Writer.WriteHeader(code)
		c.setState = true
	}
}

// WriteString 输出字符串
func (c *Context) WriteString(code int, context string) {
	if !c.setState {
		c.state = code
		c.Writer.WriteHeader(code)
		c.setState = true
	}

	c.Writer.Write([]byte(context))
}

// WriteJSON 输出JSON
func (c *Context) WriteJSON(code int, context interface{}) {
	if !c.setState {
		c.state = code
		c.Writer.WriteHeader(code)
		c.setState = true
	}

	data, err := json.Marshal(context)
	if err != nil{
		log.Fatal("json error")
	}
	c.Writer.Write(data)
}

// Write 输出
func (c *Context) Write(code int, context []byte) {
	if !c.setState {
		c.state = code
		c.Writer.WriteHeader(code)
		c.setState = true
	}

	c.Writer.Write(context)
}

// State 获取State
func (c *Context) State() int {
	return c.state
}

func (c *Context) Param(key string) string {
	return c.Params[key]
}

// Query 获取参数
func (c *Context) Query(key string) string {
	data, ok := c.Request.Form[key]
	if !ok || len(data) < 1 {
		return ""
	}
	return data[0]
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
func (c *Context) Abort() {
	c.index = len(c.handlers)
}
