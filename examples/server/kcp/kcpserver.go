package main

import (
	"strconv"
	"time"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/viakiba/gate-long-link/tools"
	"github.com/xtaci/kcp-go/v5"
)

func main() {
	tools.InitLogSimple()
	//key := pbkdf2.Key([]byte("demo pass"), []byte("demo salt"), 1024, 32, sha1.New)
	//block, _ := kcp.NewAESBlockCrypt(key)
	if listener, err := kcp.ListenWithOptions("127.0.0.1:8083", nil, 10, 3); err == nil {
		for {
			s, err := listener.AcceptKCP()
			if err != nil {
				tools.Error(err)
			}
			tools.Info("new client: %s", s.RemoteAddr().String())
			gopool.Go(func() { accept(s) })
		}
	} else {
		tools.Error(err)
	}
}

func accept(conn *kcp.UDPSession) {
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
