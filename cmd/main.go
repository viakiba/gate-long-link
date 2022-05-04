package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/viakiba/gate-long-link/kcpgate"
	"github.com/viakiba/gate-long-link/tcpgate"
	"github.com/viakiba/gate-long-link/tools"
	"github.com/viakiba/gate-long-link/wsgate"
)

func main() {
	tools.CommonInit()
	if tools.TomlConfig.Kcpgate.Enabled {
		l := kcpgate.Start()
		if l != nil {
			defer l.Close()
		}
	}
	if tools.TomlConfig.Tcpgate.Enabled {
		l := tcpgate.Start()
		if l != nil {
			defer l.Close()
		}
	}
	if tools.TomlConfig.Wsgate.Enabled {
		l := wsgate.Start()
		if l != nil {
			defer l.Close()
		}
	}
	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, syscall.SIGTERM)
	signal.Notify(exitChan, syscall.SIGINT)
	<-exitChan
}
