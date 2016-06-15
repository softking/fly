package reload

import (
	"net"
	"os"
	"syscall"
	"time"
)

type flyListener struct {
	net.Listener
	stop    chan error
	stopped bool
	server  *Server
}

func newFlyListener(l net.Listener, srv *Server) (el *flyListener) {
	el = &flyListener{
		Listener: l,
		stop:     make(chan error),
		server:   srv,
	}
	go func() {
		_ = <-el.stop
		el.stopped = true
		el.stop <- el.Listener.Close()
	}()
	return
}

// Accept 接受连接
func (gl *flyListener) Accept() (c net.Conn, err error) {
	tc, err := gl.Listener.(*net.TCPListener).AcceptTCP()
	if err != nil {
		return
	}

	err = tc.SetKeepAlive(true)
	if err != nil{
		return
	}
	err = tc.SetKeepAlivePeriod(30 * time.Second)
	if err != nil{
		return
	}

	c = &flyConn{
		Conn:   tc,
		server: gl.server,
	}
	gl.server.wg.Add(1)
	return
}

func (gl *flyListener) Close() error {
	if gl.stopped {
		return syscall.EINVAL
	}
	gl.stop <- nil
	return <-gl.stop
}

func (gl *flyListener) File() *os.File {
	tl := gl.Listener.(*net.TCPListener)
	fl, _ := tl.File()
	return fl
}
