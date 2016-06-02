package fly

import (
	"fmt"
	"net/http"
	"wpweb/fly/reload"
)

// IWillFly 我要飞得更高,非得更更更更, 咳咳,卡带了
func IWillFly() *server {
	return &server{
		Mid:    []Handler{},
		Router: make(map[string]map[string][]Handler),
	}
}

func (mod *server) Midware(handler ...Handler) {
	mod.Mid = handler
}

func (mod *server) AddMidware(handler ...Handler) {
	mod.Mid = append(mod.Mid, handler...)
}

func (mod *server) Get(path string, handler ...Handler) {
	hand, ok := mod.Router[path]
	if !ok {
		hand = make(map[string][]Handler)
	}
	hand[path] = handler
	mod.Router["GET"] = hand
}

func (mod *server) POST(path string, handler ...Handler) {
	hand, ok := mod.Router[path]
	if !ok {
		hand = make(map[string][]Handler)
	}
	hand[path] = handler
	mod.Router["POST"] = hand
}

func (mod *server) Run(addr ...string) error {
	address := resolveAddress(addr)
	err := http.ListenAndServe(address, mod)
	return err
}

func (mod *server) RunReload(addr ...string) error {
	address := resolveAddress(addr)
	err := reload.ListenAndServe(address, mod)
	return err
}

func (mod *server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	c := &Context{}
	c.Request = request
	c.Writer = writer
	mod.handleHTTPRequest(c)
}

func (mod *server) handleHTTPRequest(c *Context) {
	hand, ok := mod.Router[c.Request.Method]
	if !ok {
		fmt.Println("No Method", c.Request.Method)
		return
	}

	handler, ok := hand[c.Request.URL.Path]
	if !ok {
		fmt.Println("No Handler", c.Request.URL.Path)
		return
	}

	c.Request.ParseForm()

	c.handlers = []Handler{}
	c.handlers = append(c.handlers, mod.Mid...)
	c.handlers = append(c.handlers, handler...)
	c.dispatch()
}
