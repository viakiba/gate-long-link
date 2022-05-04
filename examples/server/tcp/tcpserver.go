package main

import (
	"net"
	"strconv"
	"time"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/viakiba/gate-long-link/tools"
)

func main() {
	tools.InitLogSimple()
	addrListen := "0.0.0.0:7083"
	listen, error := net.Listen("tcp", addrListen)
	if error != nil {
		tools.Error(error)
		return
	}
	tools.Info("Listening on " + addrListen)
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			conn.Close()
			tools.Error(err)
			continue
		}
		tools.Info("获取一个新的客户端连接")
		gopool.Go(func() {
			accept(conn)
		})
	}
}

func accept(conn net.Conn) {
	gopool.Go(func() {
		ticker := time.NewTicker(3 * time.Second)
		for {
			select {
			case <-ticker.C:
				data := tools.CalSendData("hello world\n")
				_, err := conn.Write(data)
				if err != nil {
					tools.Error("客户端断线：" + err.Error())
					conn.Close()
					return
				}
			}
		}
	})
	gopool.Go(func() {
		buf := make([]byte, 0)
		for {
			temp := make([]byte, 256)
			readLen, err := conn.Read(temp)
			if err != nil {
				conn.Close()
				break
			}
			length, payload, buf1 := tools.CalReadData(readLen, buf, temp)
			if length == -1 {
				buf = buf1
				continue
			}
			tools.Info("数据长度: " + strconv.FormatInt(int64(length), 10) + " 数据内容: " + string(payload))
			buf = buf1[4+length:]
		}
	})

}
