package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/viakiba/gate-long-link/tools"
	"github.com/xtaci/kcp-go/v5"
)

/**
https://github.com/skywind3000/kcp
https://github.com/xtaci/kcp-go
*/
func main() {
	tools.InitLogSimple()
	// if agent, err := kcp.DialWithOptions("127.0.0.1:8083", nil, 10, 3); err == nil { // 直接连接后端 kcp
	if agent, err := kcp.DialWithOptions("127.0.0.1:8082", nil, 10, 3); err == nil { // 通过网关连接后端tcp
		data := tools.CalSendData("127.0.0.1:7083")
		agent.Write(data)
		gopool.Go(func() {
			buf := make([]byte, 0)
			for {
				temp := make([]byte, 256)
				readLen, err := agent.Read(temp)
				if err != nil {
					agent.Close()
					fmt.Println(err)
					return
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
		gopool.Go(func() {
			ticker := time.NewTicker(3 * time.Second)
			for {
				select {
				case <-ticker.C:
					writeData := "client hello worldclient hello worldclient hello worldclient hello worldclient hello worldclient hello worldclient hello worldclient hello worldclient hello worldclient hello worldclient hello worldclient hello worldclient hello worldclient hello worldclient hello worldclient hello worldclient hello worldclient hello world\n"
					// writeData := "client hello world\n"
					if writeData == "" {
						continue
					}
					if writeData == "exit" {
						agent.Close()
						break
					}
					data := tools.CalSendData(writeData)
					_, err := agent.Write(data)
					if err != nil {
						fmt.Println(err)
						return
					}
				}
			}
		})
		select {}
	} else {
		tools.Error(err)
	}
}
