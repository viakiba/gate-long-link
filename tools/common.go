package tools

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/xtaci/kcp-go/v5"
)

//BytesPool 字节数据流 池化
var BytesPool = sync.Pool{
	New: func() interface{} {
		var buffer = bytes.NewBuffer(make([]byte, 2048))
		return buffer
	},
}

var ikeyIdentify IKeyIdentify

// CommonInit 初始化 统一的配置
func CommonInit() {
	var once sync.Once
	once.Do(
		func() {
			InitLogSimple()
			//logs.InitLog(configs.TomlConfig.LogConfig.InfofilePath, configs.TomlConfig.LogConfig.ErrorfilePath)
			InitConfig()
			initGOMAXPROCS()
			pidInit()
			hackManyOpenFiles()
			ikeyIdentify = getKeyIdentify(TomlConfig.CommonConfig.HeaderIdentify)
			gopool.Go(func() {
				for {
					time.Sleep(time.Second * 5)
					Info("协程数 : ", runtime.NumGoroutine())
				}
			})
		},
	)
}

func getKeyIdentify(param string) IKeyIdentify {
	switch param {
	case "origin":
		return &IpHostIdentify{}
	case "jwt":
		return &JwtKeyIdentify{}
	default:
		return &IpHostIdentify{}
	}
}

// HandlerShake 握手
func HandlerShake(conn net.Conn, dialTimeout time.Duration, conType ConnectType) (net.Conn, error) {
	conn.SetReadDeadline(time.Now().Add(dialTimeout))
	buf := make([]byte, 0)
	readBufLength := int32(4)
	payloadLength := uint32(0)
	connectStartTime := time.Now().Unix()
	for {
		connectEndTime := time.Now().Unix()
		if connectEndTime-connectStartTime > TomlConfig.CommonConfig.ConnectTimeout {
			return nil, errors.New("连接超时")
		}
		temp := make([]byte, readBufLength)
		readLen, err := conn.Read(temp)
		if err != nil {
			return nil, err
		}
		if readLen == 0 {
			continue
		}
		buf = append(buf, temp[:readLen]...)
		if len(buf) < 4 {
			continue
		}
		if len(buf) >= 4 && payloadLength == 0 {
			payloadLength = binary.LittleEndian.Uint32(buf[0:4])
		}
		readBufLength = 4 + int32(payloadLength) - int32(len(buf))
		if readBufLength == 0 {
			break
		}
	}
	keyIdent := string(buf[4:])
	Info("解析目标地址: ", keyIdent, "解析目标长度: ", payloadLength)
	addr, err := ikeyIdentify.getIpHost(keyIdent)
	if err != nil {
		return nil, err
	}
	agent, done := connBackendAgent(conn, dialTimeout, addr, conType)
	if done {
		conn.Write([]byte("error"))
		return nil, errors.New("server error")
	}
	resp := Base64SuccMessage()
	conn.Write(resp)
	conn.SetReadDeadline(time.Time{})
	return agent, nil
}

// WsHandshake 客户端握手 websocket
func WsHandshake(conn net.Conn, dialTimeout time.Duration) (net.Conn, error) {
	headerMap := make(map[string]string)
	u := ws.Upgrader{
		OnHeader: func(key, value []byte) error {
			headerMap[strings.ToLower(string(key))] = string(value)
			return nil
		},
	}
	_, err := u.Upgrade(conn)
	if err != nil {
		Error("Upgrade error:", err)
		return nil, err
	}

	keyIdent := headerMap["server-identify"]
	addr, err := ikeyIdentify.getIpHost(keyIdent)
	if err != nil {
		return nil, err
	}
	Info("解析目标地址: ", addr)
	agent, b := connBackendAgent(conn, dialTimeout, addr, CONNECT_TYPE_WS)
	if b {
		conn.Write([]byte("error"))
		return nil, errors.New("server error")
	}
	resp, _ := base64.StdEncoding.DecodeString(TomlConfig.CommonConfig.SuccessRespMessage)
	writeToWs(conn, resp)
	return agent, nil
}

func connBackendAgent(conn net.Conn, dialTimeout time.Duration, addr string, conType ConnectType) (net.Conn, bool) {
	var agent net.Conn
	var err error
	for i := uint(0); i < 3; i++ {
		if TomlConfig.CommonConfig.KcpReciprocity && conType == CONNECT_TYPE_KCP {
			agent, err = kcp.DialWithOptions("127.0.0.1:8082", nil, 10, 3)
		} else {
			agent, err = net.DialTimeout("tcp", addr, dialTimeout)
		}
		if err == nil {
			break
		}
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			continue
		}
		conn.Close()
		Error("Error connecting:", err)
		return nil, true
	}
	return agent, false
}

// Handle 客户端请求处理
func Handle(conn net.Conn, agent net.Conn, gate *Gate) {
	if gate.ZeroCopy {
		Info("零拷贝模式")
		gopool.Go(func() {
			defer conn.Close()
			defer agent.Close()
			io.Copy(conn, agent)
			Error("断线了--------2")
		})
		gopool.Go(func() {
			defer conn.Close()
			defer agent.Close()
			io.Copy(agent, conn)
			Error("断线了--------1")
		})
		return
	}
	Info("非零拷贝模式")
	gopool.Go(func() {
		copy(conn, agent, "conn")
	})
	gopool.Go(func() {
		copy(agent, conn, "agent")
	})
}

// WsHandle ws 客户端请求处理
func WsHandle(conn net.Conn, agent net.Conn, gate *Gate) {
	gopool.Go(func() {
		wsCopy(conn, agent, "conn", false)
	})
	gopool.Go(func() {
		wsCopy(agent, conn, "agent", true)
	})
}

func copy(writer net.Conn, reader net.Conn, flag string) {
	defer func() {
		if r := recover(); r != nil {
			Error("Recovered in", r, ":", string(debug.Stack()))
		}
	}()
	defer writer.Close()
	defer reader.Close()
	for {
		pool := newBufferFromPool()
		buf1 := pool.Bytes()
		var err error
		n, err := reader.Read(buf1)
		if err == nil {
			buf1 = buf1[:n]
		}
		if err != nil {
			BytesPool.Put(pool)
			Error(flag + " 对等连接关闭")
			return
		}
		_, err = writer.Write(buf1)
		if err != nil {
			BytesPool.Put(pool)
			Error(flag + " 对等连接关闭")
			return
		}
		BytesPool.Put(pool)
	}
}

func wsCopy(writer net.Conn, reader net.Conn, flag string, readerWsFlag bool) {
	defer func() {
		if r := recover(); r != nil {
			Error("Recovered in", r, ":", string(debug.Stack()))
		}
	}()
	defer writer.Close()
	defer reader.Close()
	for {
		pool := newBufferFromPool()
		buf1 := pool.Bytes()
		var err error
		if readerWsFlag {
			buf1, _, err = readFromWs(reader, buf1)
		} else {
			n, err := reader.Read(buf1)
			if err == nil {
				buf1 = buf1[:n]
			}
		}
		if err != nil {
			BytesPool.Put(pool)
			Error(flag + " 对等连接关闭")
			return
		}
		if !readerWsFlag {
			err = writeToWs(writer, buf1)
		} else {
			if readerWsFlag {
				// 如果是ws则需要写入长度
				bs := make([]byte, 4)
				i := len(buf1)
				binary.LittleEndian.PutUint32(bs, uint32(i))
				writer.Write(bs)
				_, err = writer.Write(buf1)
			} else {
				_, err = writer.Write(buf1)
			}
		}
		if err != nil {
			BytesPool.Put(pool)
			Error(flag + " 对等连接关闭")
			return
		}
		BytesPool.Put(pool)
	}
}
func writeToWs(conn net.Conn, payload []byte) error {
	err := wsutil.WriteServerMessage(conn, ws.OpBinary, payload)
	return err
}

func readFromWs(conn net.Conn, buf []byte) ([]byte, int, error) {
	header, err2 := ws.ReadHeader(conn)
	if header.OpCode == ws.OpClose {
		return nil, 0, errors.New("")
	}
	if err2 != nil {
		Error("read header error:", err2)
		return nil, 0, errors.New("")
	}
	buf = buf[:header.Length]
	n, err := io.ReadFull(conn, buf)
	if err != nil {
		Error("read header error:", err)
		return nil, 0, errors.New("")
	}
	if header.Masked {
		ws.Cipher(buf, header.Mask, 0)
	}
	return buf, n, nil
}

// 通过Get来获得一个
func newBufferFromPool() *bytes.Buffer {
	return BytesPool.Get().(*bytes.Buffer)
}

// hackManyOpenFiles
func hackManyOpenFiles() {
	_MaxOpenfile := uint64(1024 * 1024 * 1024)
	var lim syscall.Rlimit
	syscall.Getrlimit(syscall.RLIMIT_NOFILE, &lim)
	if lim.Cur < _MaxOpenfile || lim.Max < _MaxOpenfile {
		lim.Cur = _MaxOpenfile
		lim.Max = _MaxOpenfile
		syscall.Setrlimit(syscall.RLIMIT_NOFILE, &lim)
	}
}

// initGOMAXPROCS
func initGOMAXPROCS() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	os.Setenv("GOTRACEBACK", "crash")
}

// pidInit 初始化pid
func pidInit() {
	pid := syscall.Getpid()
	if err := ioutil.WriteFile("gateway.pid", []byte(strconv.Itoa(pid)), 0644); err != nil {
		Error("Can't write pid file: %s", err)
	}
}

func CalSendData(writeData string) []byte {
	payload := []byte(writeData)
	return CalSendByteData(payload)
}

func CalSendByteData(writeData []byte) []byte {
	length := make([]byte, 4)
	binary.LittleEndian.PutUint32(length, uint32(len(writeData)))
	return append(length, writeData...)

}

func CalReadData(readLen int, buf []byte, temp []byte) (int32, []byte, []byte) {
	if readLen == 0 {
		return -1, nil, buf
	}
	buf = append(buf, temp[:readLen]...)
	if len(buf) < 4 {
		return -1, nil, buf
	}
	length := binary.LittleEndian.Uint32(buf[0:4])
	if len(buf)-4 < int(length) {
		return -1, nil, buf
	}
	payload := buf[4 : 4+length]
	return int32(length), payload, buf
}

func Base64SuccMessage() []byte {
	sDec, _ := base64.StdEncoding.DecodeString(TomlConfig.CommonConfig.SuccessRespMessage)
	return CalSendByteData(sDec)
}
