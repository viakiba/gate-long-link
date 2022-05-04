package main

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/viakiba/gate-long-link/tools"
)

func main() {
	tools.InitLogSimple()
	agent, err := net.DialTimeout("tcp", "127.0.0.1:7082", 5*time.Second) // gate 中转
	// agent, err := net.DialTimeout("tcp", "127.0.0.1:7083", 5*time.Second) // server 直连
	if err != nil {
		tools.Error("Error: ", err)
		return
	}
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

}
