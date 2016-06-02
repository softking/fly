package reload

import (
	"flag"
	"net/http"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"
)

// 各种信号
const (
	PreSignal = iota
	PostSignal
	StateInit
	StateRunning
	StateShuttingDown
	StateTerminate
)

var (
	regLock              *sync.Mutex
	runningServers       map[string]*Server
	runningServersOrder  []string
	socketPtrOffsetMap   map[string]uint
	runningServersForked bool

	// DefaultReadTimeOut 你猜
	DefaultReadTimeOut time.Duration
	// DefaultWriteTimeOut 你猜
	DefaultWriteTimeOut time.Duration
	// DefaultMaxHeaderBytes 你猜
	DefaultMaxHeaderBytes int
	// DefaultTimeout 你猜
	DefaultTimeout = 60 * time.Second

	isChild     bool
	socketOrder string
	once        sync.Once
)

func onceInit() {
	regLock = &sync.Mutex{}
	flag.BoolVar(&isChild, "reload", false, "listen on open fd (after forking)")
	flag.StringVar(&socketOrder, "socketorder", "", "previous initialization order - used when more than one listener was started")
	runningServers = make(map[string]*Server)
	runningServersOrder = []string{}
	socketPtrOffsetMap = make(map[string]uint)
}

// NewServer http server
func NewServer(addr string, handler http.Handler) (srv *Server) {
	once.Do(onceInit)
	regLock.Lock()
	defer regLock.Unlock()
	if !flag.Parsed() {
		flag.Parse()
	}
	if len(socketOrder) > 0 {
		for i, addr := range strings.Split(socketOrder, ",") {
			socketPtrOffsetMap[addr] = uint(i)
		}
	} else {
		socketPtrOffsetMap[addr] = uint(len(runningServersOrder))
	}

	srv = &Server{
		wg:      sync.WaitGroup{},
		sigChan: make(chan os.Signal),
		isChild: isChild,
		SignalHooks: map[int]map[os.Signal][]func(){
			PreSignal: {
				syscall.SIGHUP:  {},
				syscall.SIGINT:  {},
				syscall.SIGTERM: {},
			},
			PostSignal: {
				syscall.SIGHUP:  {},
				syscall.SIGINT:  {},
				syscall.SIGTERM: {},
			},
		},
		state:   StateInit,
		Network: "tcp",
	}
	srv.Server = &http.Server{}
	srv.Server.Addr = addr
	srv.Server.ReadTimeout = DefaultReadTimeOut
	srv.Server.WriteTimeout = DefaultWriteTimeOut
	srv.Server.MaxHeaderBytes = DefaultMaxHeaderBytes
	srv.Server.Handler = handler

	runningServersOrder = append(runningServersOrder, addr)
	runningServers[addr] = srv

	return
}

// ListenAndServe start
func ListenAndServe(addr string, handler http.Handler) error {
	server := NewServer(addr, handler)
	return server.ListenAndServe()
}

// ListenAndServeTLS https
func ListenAndServeTLS(addr string, certFile string, keyFile string, handler http.Handler) error {
	server := NewServer(addr, handler)
	return server.ListenAndServeTLS(certFile, keyFile)
}
