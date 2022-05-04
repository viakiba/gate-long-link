## 轻量的透明网关

## 应用场景

   可以根据客户端传入节点信息，把客户端连接转发到对应的后端结点上。对于有状态的服务有利于性能优化与安全保护。
   例如，游戏中可以进行逻辑分服，而域名只配置一个即可。

## 能力

用于支持 tcp, kcp 连接类型的透明对等网关以及 websocket, kcp 转 后端 tcp 的非对等透明网关。tcp与kcp连接方式，支持零拷贝。

|  客户端连接协议   | 网关监听协议  |  后端服务监听协议  |
|  ----  | ----  | ----  |
| tcp  | tcp | tcp |
| kcp  | kcp | tcp |
| kcp  | kcp | kcp |
| websocket  | websocket | tcp |

## 使用
### examples

1. clent  简单的各协议的客户端实现，可以直接 run 起来.
2. server 简单的各协议的服务端实现，可以直接 run 起来.
   
### 后端获取策略

tcp 与 kcp 是基于连接后的第一个数据包的值来获取后端的 ip 和 port ，而 websocket 是基于 http header 的 值 获取的后端的 ip 和 port。

内置策略

- host-ip : tcp, kcp 第一个带长度的数据包的值。websocket 的 server-identify 。
    - 假如需要传递的是 192.168.1.1:9090 的值
    - tcp, kcp : 四字节的长度 length 16 + 192.168.1.1:9090 字符串转数组。
    - websocekt : header对应的值 是 HOST:PORT 格式的值
- JWT ：把上述的 host:ip 换成 JWT 的值，传输规则一致。
- 可以自己根据需求添加，基于 etcd， zookeeper 等策略。
  
### 配置文件

config/config.toml
 
```toml
[log]
infofilePath = "/Users/dd/Documents/vkgateway/log/info.log"
errorfilePath = "/Users/dd/Documents/vkgateway/log/error.log"

[tcpgate]
port = 7082 # 监听端口
enabled = true # 是否开启 tcp 协议 网关
dial_timeout = 5 # 后端连接超时 秒
zero_copy = false # 零拷贝


[kcpgate]
port = 8082 # 监听端口
enabled = true # 是否开启 kcp 协议 网关
dial_timeout = 5 # 后端连接超时 秒
zero_copy = false # 零拷贝

[wsgate]
port = 9082 # 监听端口
enabled = true # 是否开启 ws 协议 网关
dial_timeout = 5 # 后端连接超时 秒

[common]
dial_retry = 3 # 后端连接失败重试次数
header_identify = "ipHost" # 后段地址获取策略 可选 origin jwt
success_resp_message = "b2s=" # success_resp_message  ok
connect_timeout = 10 # connect_timeout  握手等待超时时间 秒 seconds
kcp_reciprocity = true # kcp_reciprocity  kcp 网关是否 后端连接是否对等
```

### tcp

#### 对等连接
1. 启动 examples/server/tcp/main.go 的后端服务器例子，监听端口 7083.
   1. 他会对连入的客户端每隔 3 秒，发送一个带换行的 hello world 字符串。
   2. 收到的客户端请求，会以首段的四字节获取长度，解析后面的内容。
2. 根据配置文件注释，开启 tcp 网关，并监听端口7082，接收到客户端的连接请求，转发到后端。
3. 启动 examples/client/tcp/main.go 的客户端，连接到 tcp 网关，发送按照 host:ip 策略的后端地址后，网关回应成功。


### kcp

#### 对等连接

1. 启动 examples/server/kcp/main.go 的后端服务器例子，监听端口 8083.
   1. 他会对连入的客户端每隔 3 秒，发送一个带换行的 hello world 字符串。
   2. 收到的客户端请求，会以首段的四字节获取长度，解析后面的内容。
2. 根据配置文件注释，开启 kcp 网关，并监听端口 8082 ，接收到客户端的连接请求，转发到后端。
3. 启动 examples/client/kcp/main.go 的客户端，连接到 kcp 网关，发送按照 host:ip 策略的后端地址后，网关回应成功。

#### 非对等连接

1. 启动 examples/server/tcp/main.go 的后端服务器例子，监听端口 7083.
   1. 他会对连入的客户端每隔 3 秒，发送一个带换行的 hello world 字符串。
   2. 收到的客户端请求，会以首段的四字节获取长度，解析后面的内容。
2. 根据配置文件注释，开启 kcp 网关，并监听端口 8082 ，接收到客户端的连接请求，转发到后端。
3. 启动 examples/client/kcp/main.go 的客户端，连接到 kcp 网关，发送按照 host:ip 策略的后端地址后，网关回应成功。
   
### websocket

#### 非对等连接
1. 启动 examples/server/tcp/main.go 的后端服务器例子，监听端口 7083.
   1. 他会对连入的客户端每隔 3 秒，发送一个带换行的 hello world 字符串。
   2. 收到的客户端请求，会以首段的四字节获取长度，解析后面的内容。
2. 根据配置文件注释，开启 websocket 网关，并监听端口9082，接收到客户端的连接请求，转发到后端。
3. 启动 examples/client/kcp/main.go 的客户端，连接到 kcp 网关，发送按照 host:ip 策略的后端地址后，网关回应成功。

## 开发

1. 采用 go 1.18 版本实现。
2. go mod tidy 获取项目依赖。
3. 启动，cmd/main.go 文件，他会读取 config/config.toml 配置文件进行启动。
4. example 是测试使用的例子。
5. log 采用 zap 。
6. kcpgate, tcpgate, wsgate 是网关监听实现。所有核心逻辑在 tools 文件夹下的 common.go 文件中。
7. 支持 零拷贝，io.copy 。
8. 支持 bytes 池化 。

## 思考
### Q1: kcp网关如何控制协程结束

```txt
客户端以kcp协议连接到gate上，gate以tcp方式连接到后端业务server。
业务server依靠心跳方式检修活跃状态判断，长时间没有心跳的关闭连接，gate就会关闭与客户端的kcp连接。
进而释放所有资源。
```

### Q2: 如何安全的获取客户端传入的流数据，避免一直等待或者长度传入不多的情况下，导致协程阻塞

```txt
for 循环里面增加个时间判断  tools/common.go HandlerShake 方法
```