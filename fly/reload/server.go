package reload

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Server Server
type Server struct {
	*http.Server
	FlyListener      net.Listener
	SignalHooks      map[int]map[os.Signal][]func()
	tlsInnerListener *flyListener
	wg               sync.WaitGroup
	sigChan          chan os.Signal
	isChild          bool
	state            uint8
	Network          string
}

// Serve 开启服务
func (srv *Server) Serve() (err error) {
	srv.state = StateRunning
	err = srv.Server.Serve(srv.FlyListener)
	log.Println(syscall.Getpid(), "Waiting for connections to finish...")
	srv.wg.Wait()
	srv.state = StateTerminate
	return
}

// ListenAndServe 开启服务
func (srv *Server) ListenAndServe() (err error) {
	addr := srv.Addr
	if addr == "" {
		addr = ":http"
	}

	go srv.handleSignals()

	l, err := srv.getListener(addr)
	if err != nil {
		log.Println(err)
		return err
	}

	srv.FlyListener = newFlyListener(l, srv)

	if srv.isChild {
		process, err := os.FindProcess(os.Getppid())
		if err != nil {
			log.Println(err)
			return err
		}
		err = process.Signal(syscall.SIGTERM)
		if err != nil {
			return err
		}
	}

	log.Println(os.Getpid(), srv.Addr)
	return srv.Serve()
}

// ListenAndServeTLS https
func (srv *Server) ListenAndServeTLS(certFile, keyFile string) (err error) {
	addr := srv.Addr
	if addr == "" {
		addr = ":https"
	}

	config := &tls.Config{}
	if srv.TLSConfig != nil {
		*config = *srv.TLSConfig
	}
	if config.NextProtos == nil {
		config.NextProtos = []string{"http/1.1"}
	}

	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return
	}

	go srv.handleSignals()

	l, err := srv.getListener(addr)
	if err != nil {
		log.Println(err)
		return err
	}

	srv.tlsInnerListener = newFlyListener(l, srv)
	srv.FlyListener = tls.NewListener(srv.tlsInnerListener, config)

	if srv.isChild {
		process, err := os.FindProcess(os.Getppid())
		if err != nil {
			log.Println(err)
			return err
		}
		err = process.Signal(syscall.SIGTERM)
		if err != nil {
			return err
		}
	}
	log.Println(os.Getpid(), srv.Addr)
	return srv.Serve()
}

func (srv *Server) getListener(laddr string) (l net.Listener, err error) {
	if srv.isChild {
		var ptrOffset uint
		if len(socketPtrOffsetMap) > 0 {
			ptrOffset = socketPtrOffsetMap[laddr]
			log.Println("laddr", laddr, "ptr offset", socketPtrOffsetMap[laddr])
		}

		f := os.NewFile(uintptr(3+ptrOffset), "")
		l, err = net.FileListener(f)
		if err != nil {
			err = fmt.Errorf("net.FileListener error: %v", err)
			return
		}
	} else {
		l, err = net.Listen(srv.Network, laddr)
		if err != nil {
			err = fmt.Errorf("net.Listen error: %v", err)
			return
		}
	}
	return
}

func (srv *Server) handleSignals() {
	var sig os.Signal

	signal.Notify(
		srv.sigChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	pid := syscall.Getpid()
	for {
		sig = <-srv.sigChan
		srv.signalHooks(PreSignal, sig)
		switch sig {
		case syscall.SIGHUP:
			log.Println(pid, "Received SIGHUP. forking.")
			err := srv.fork()
			if err != nil {
				log.Println("Fork err:", err)
			}
		case syscall.SIGINT:
			log.Println(pid, "Received SIGINT.")
			srv.shutdown()
		case syscall.SIGTERM:
			log.Println(pid, "Received SIGTERM.")
			srv.shutdown()
		default:
			log.Printf("Received %v: nothing i care about...\n", sig)
		}
		srv.signalHooks(PostSignal, sig)
	}
}

func (srv *Server) signalHooks(ppFlag int, sig os.Signal) {
	if _, notSet := srv.SignalHooks[ppFlag][sig]; !notSet {
		return
	}
	for _, f := range srv.SignalHooks[ppFlag][sig] {
		f()
	}
	return
}

func (srv *Server) shutdown() {
	if srv.state != StateRunning {
		return
	}

	srv.state = StateShuttingDown
	if DefaultTimeout >= 0 {
		go srv.serverTimeout(DefaultTimeout)
	}
	err := srv.FlyListener.Close()
	if err != nil {
		log.Println(syscall.Getpid(), "Listener.Close() error:", err)
	} else {
		log.Println(syscall.Getpid(), srv.FlyListener.Addr(), "Listener closed.")
	}
}

func (srv *Server) serverTimeout(d time.Duration) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("WaitGroup at 0", r)
		}
	}()
	if srv.state != StateShuttingDown {
		return
	}
	time.Sleep(d)
	log.Println("[STOP - Hammer Time] Forcefully shutting down parent")
	for {
		if srv.state == StateTerminate {
			break
		}
		srv.wg.Done()
	}
}

func (srv *Server) fork() (err error) {
	regLock.Lock()
	defer regLock.Unlock()
	if runningServersForked {
		return
	}
	runningServersForked = true

	var files = make([]*os.File, len(runningServers))
	var orderArgs = make([]string, len(runningServers))
	for _, srvPtr := range runningServers {
		switch srvPtr.FlyListener.(type) {
		case *flyListener:
			files[socketPtrOffsetMap[srvPtr.Server.Addr]] = srvPtr.FlyListener.(*flyListener).File()
		default:
			files[socketPtrOffsetMap[srvPtr.Server.Addr]] = srvPtr.tlsInnerListener.File()
		}
		orderArgs[socketPtrOffsetMap[srvPtr.Server.Addr]] = srvPtr.Server.Addr
	}

	log.Println(files)
	path := os.Args[0]
	var args []string
	if len(os.Args) > 1 {
		for _, arg := range os.Args[1:] {
			if arg == "-reload" {
				break
			}
			args = append(args, arg)
		}
	}
	args = append(args, "-reload")
	if len(runningServers) > 1 {
		args = append(args, fmt.Sprintf(`-socketorder=%s`, strings.Join(orderArgs, ",")))
		log.Println(args)
	}
	cmd := exec.Command(path, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = files
	err = cmd.Start()
	if err != nil {
		log.Fatalf("Restart: Failed to launch, error: %v", err)
	}

	return
}
