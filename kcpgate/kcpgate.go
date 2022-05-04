package kcpgate

import (
	"strconv"
	"time"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/viakiba/gate-long-link/tools"
	"github.com/xtaci/kcp-go/v5"
)

// Start 启动服务
func Start() *kcp.Listener {
	addrListen := ":" + strconv.Itoa(tools.TomlConfig.Kcpgate.Port)
	listener, err := kcp.ListenWithOptions(addrListen, nil, 10, 3)
	if err != nil {
		tools.Error("kcpgate listen error:", err)
		return nil
	}
	tools.Info("Listening on " + addrListen)
	gopool.Go(func() {
		for {
			conn, err := listener.AcceptKCP()
			if err != nil {
				tools.Error(err)
				continue
			}
			tools.Info("获取一个新的客户端连接")
			gopool.Go(func() {
				handle(conn)
			})
		}
	})
	return listener
}

func handle(conn *kcp.UDPSession) {
	duration, _ := time.ParseDuration(strconv.Itoa(tools.TomlConfig.Kcpgate.DialTimeout) + "s")
	agent, done := tools.HandlerShake(conn, duration, tools.CONNECT_TYPE_KCP)
	if done != nil {
		conn.Close()
		return
	}
	tools.Handle(conn, agent, &tools.TomlConfig.Kcpgate)
}
