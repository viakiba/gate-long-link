package tools

type ConnectType int

const (
	CONNECT_TYPE_TCP ConnectType = 0
	CONNECT_TYPE_KCP ConnectType = 1
	CONNECT_TYPE_WS  ConnectType = 2
)
