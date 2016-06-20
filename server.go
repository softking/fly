package fly

import (
	"github.com/softking/fly/reload"
	"net/http"
)

// Server ServeHttp
type Server interface {
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

// Run 正常run
func Run(mod Server, addr ...string) error {
	address := resolveAddress(addr)
	err := http.ListenAndServe(address, mod)
	return err
}

// ReloadRun 热更新方式run
func ReloadRun(mod Server, addr ...string) error {
	address := resolveAddress(addr)
	err := reload.ListenAndServe(address, mod)
	return err
}
