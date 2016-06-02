package fly

import "net/http"

// Handler Handler
type Handler func(*Context) bool

// Context Context
type Context struct {
	index    int
	handlers []Handler
	Writer   http.ResponseWriter
	Request  *http.Request
	Data     map[string]interface{}
}

type server struct {
	Router map[string]map[string][]Handler // map["get"]map["/hello"] func
	Mid    []Handler
}
