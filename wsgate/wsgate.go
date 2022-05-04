package wsgate

import (
	"net"
	"strconv"
	"time"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/viakiba/gate-long-link/tools"
)

// Start 启动服务
func Start() net.Listener {
	addrListen := ":" + strconv.Itoa(tools.TomlConfig.Wsgate.Port)
	listen, error := net.Listen("tcp", addrListen)
	if error != nil {
		tools.Error("Error listening:", error)
		return nil
	}
	tools.Info("Listening on " + addrListen)
	gopool.Go(func() {
		for {
			conn, err := listen.Accept()
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
	return listen
}

func handle(conn net.Conn) {
	duration, _ := time.ParseDuration(strconv.Itoa(tools.TomlConfig.Wsgate.DialTimeout) + "s")
	agent, err := tools.WsHandshake(conn, duration)
	if err != nil {
		conn.Close()
		return
	}
	tools.WsHandle(conn, agent, &tools.TomlConfig.Wsgate)
}
