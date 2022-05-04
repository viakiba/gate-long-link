package tools

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

// IKeyIdentify 接口
type IKeyIdentify interface {
	// getIpHost 获取 ip:port
	getIpHost(key string) (string, error)
}

// JwtKeyIdentify jwt token
type JwtKeyIdentify struct {
}

// getIpHost 获取 ip:port
func (j JwtKeyIdentify) getIpHost(key string) (string, error) {
	s := strings.Split(key, ".")
	if len(s) != 3 {
		return "", errors.New("jwt token error")
	}
	s1 := s[1]
	b, err := base64.StdEncoding.DecodeString(s1)
	if err != nil {
		return "", err
	}
	s1 = string(b)
	m := make(map[string]interface{})
	json.Unmarshal([]byte(s1), &m)
	host := m["host"].(string)
	portStr := m["port"].(int32)
	port := strconv.FormatInt(int64(portStr), 10)
	return host + ":" + port, nil
}

// IpHostIdentify ip:port
type IpHostIdentify struct {
}

// 直接返回 xxx.xxx.xxx.xxx:xxxx ip:port
func (j IpHostIdentify) getIpHost(key string) (string, error) {
	return key, nil
}
