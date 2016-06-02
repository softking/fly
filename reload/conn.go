package reload

import (
	"errors"
	"net"
)

type flyConn struct {
	net.Conn
	server *Server
}

// Close 关闭
func (c flyConn) Close() (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()
	c.server.wg.Done()
	return c.Conn.Close()
}
