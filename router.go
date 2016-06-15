
package fly

import (
	"net/http"
)


// Param node
type Param struct {
	Key   string
	Value string
}

// Params Param-slice
type Params []Param

// Router struct
type Router struct {
	trees map[string]*node
	RedirectTrailingSlash bool
	RedirectFixedPath bool
	HandleMethodNotAllowed bool
	HandleOPTIONS bool
	NotFoundUseMidWare bool
	MethodNotAllowedUseMidware bool

	NotFound  Handler
	MethodNotAllowed Handler
	PanicHandler func(http.ResponseWriter, *http.Request, interface{})
	Mid []Handler
}


// IWillFly 飞翔吧,骚年
func IWillFly() *Router {
	return &Router{
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
		HandleOPTIONS:          true,
	}
}

// MidWare MidWare
func (r *Router) MidWare(handler ...Handler) {
	r.Mid = handler
}

// AddMidware AddMidware
func (r *Router) AddMidware(handler ...Handler) {
	r.Mid = append(r.Mid, handler...)
}


// GET is a shortcut for router.Handle("GET", path, handle)
func (r *Router) GET(path string, handle ...Handler) {
	r.Handle("GET", path, handle...)
}

// HEAD is a shortcut for router.Handle("HEAD", path, handle)
func (r *Router) HEAD(path string, handle ...Handler) {
	r.Handle("HEAD", path, handle...)
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handle)
func (r *Router) OPTIONS(path string, handle ...Handler) {
	r.Handle("OPTIONS", path, handle...)
}

// POST is a shortcut for router.Handle("POST", path, handle)
func (r *Router) POST(path string, handle ...Handler) {
	r.Handle("POST", path, handle...)
}

// PUT is a shortcut for router.Handle("PUT", path, handle)
func (r *Router) PUT(path string, handle ...Handler) {
	r.Handle("PUT", path, handle...)
}

// PATCH is a shortcut for router.Handle("PATCH", path, handle)
func (r *Router) PATCH(path string, handle ...Handler) {
	r.Handle("PATCH", path, handle...)
}

// DELETE is a shortcut for router.Handle("DELETE", path, handle)
func (r *Router) DELETE(path string, handle ...Handler) {
	r.Handle("DELETE", path, handle...)
}

// Handle Handle
func (r *Router) Handle(method, path string, handle ...Handler) {
	if path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}

	if r.trees == nil {
		r.trees = make(map[string]*node)
	}

	root := r.trees[method]
	if root == nil {
		root = new(node)
		r.trees[method] = root
	}

	root.addRoute(path, handle...)
}

func (r *Router) recv(w http.ResponseWriter, req *http.Request) {
	if rcv := recover(); rcv != nil {
		r.PanicHandler(w, req, rcv)
	}
}

// Lookup method + path
func (r *Router) Lookup(method, path string) ([]Handler, Params, bool) {
	if root := r.trees[method]; root != nil {
		return root.getValue(path)
	}
	return nil, nil, false
}

func (r *Router) allowed(path, reqMethod string) (allow string) {
	if path == "*" { // server-wide
		for method := range r.trees {
			if method == "OPTIONS" {
				continue
			}

			// add request method to list of allowed methods
			if len(allow) == 0 {
				allow = method
			} else {
				allow += ", " + method
			}
		}
	} else { // specific path
		for method := range r.trees {
			// Skip the requested method - we already tried this one
			if method == reqMethod || method == "OPTIONS" {
				continue
			}

			handle, _, _ := r.trees[method].getValue(path)
			if handle != nil {
				// add request method to list of allowed methods
				if len(allow) == 0 {
					allow = method
				} else {
					allow += ", " + method
				}
			}
		}
	}
	if len(allow) > 0 {
		allow += ", OPTIONS"
	}
	return
}

// ServeHTTP 总的请求入口 由此分发
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if r.PanicHandler != nil {
		defer r.recv(w, req)
	}

	path := req.URL.Path

	if root := r.trees[req.Method]; root != nil {
		if handle, ps, tsr := root.getValue(path); handle != nil {
			r.handleHTTPRequest(w,req,ps, handle)
			return
		} else if req.Method != "CONNECT" && path != "/" {
			code := 301 // Permanent redirect, request with GET method
			if req.Method != "GET" {
				// Temporary redirect, request with same method
				// As of Go 1.3, Go does not support status code 308.
				code = 307
			}

			if tsr && r.RedirectTrailingSlash {
				if len(path) > 1 && path[len(path)-1] == '/' {
					req.URL.Path = path[:len(path)-1]
				} else {
					req.URL.Path = path + "/"
				}
				http.Redirect(w, req, req.URL.String(), code)
				return
			}

			// Try to fix the request path
			if r.RedirectFixedPath {
				fixedPath, found := root.findCaseInsensitivePath(
					CleanPath(path),
					r.RedirectTrailingSlash,
				)
				if found {
					req.URL.Path = string(fixedPath)
					http.Redirect(w, req, req.URL.String(), code)
					return
				}
			}
		}
	}

	if req.Method == "OPTIONS" {
		// Handle OPTIONS requests
		if r.HandleOPTIONS {
			if allow := r.allowed(path, req.Method); len(allow) > 0 {
				w.Header().Set("Allow", allow)
				return
			}
		}
	} else {
		// Handle 405
		if r.HandleMethodNotAllowed {
			if allow := r.allowed(path, req.Method); len(allow) > 0 {
				w.Header().Set("Allow", allow)
				if r.MethodNotAllowed != nil {
					c := &Context{}
					c.Request = req
					c.Writer = w
					c.Data = make(map[string]interface{})
					c.handlers = []Handler{}
					if r.MethodNotAllowedUseMidware {
						c.handlers = append(c.handlers, r.Mid...)
					}
					c.handlers = append(c.handlers, r.MethodNotAllowed)
					c.dispatch()
				} else {
					http.Error(w,
						http.StatusText(http.StatusMethodNotAllowed),
						http.StatusMethodNotAllowed,
					)
				}
				return
			}
		}
	}

	// Handle 404
	if r.NotFound != nil {
		c := &Context{}
		c.Request = req
		c.Writer = w
		c.Data = make(map[string]interface{})
		c.handlers = []Handler{}
		if r.NotFoundUseMidWare{
			c.handlers = append(c.handlers, r.Mid...)
		}
		c.handlers = append(c.handlers, r.NotFound)
		c.dispatch()
	} else {
		http.NotFound(w, req)
	}
}

func (r *Router) handleHTTPRequest(writer http.ResponseWriter, request *http.Request, params Params, handle []Handler) {

	c := &Context{}
	c.Request = request
	c.Writer = writer
	c.Data = make(map[string]interface{})
	c.Params = make(map[string]string)
	for _,p:=range params{
		c.Params[p.Key] = p.Value
	}

	c.Request.ParseForm()

	c.handlers = []Handler{}
	c.handlers = append(c.handlers, r.Mid...)
	c.handlers = append(c.handlers, handle...)
	c.dispatch()
}