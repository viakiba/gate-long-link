package tools

import (
	"github.com/BurntSushi/toml"
)

type tomlConfig struct {
	LogConfig    log          `toml:"log"`
	Tcpgate      Gate         `toml:"tcpgate"`
	Kcpgate      Gate         `toml:"kcpgate"`
	Wsgate       Gate         `toml:"wsgate"`
	CommonConfig CommonConfig `toml:"common"`
}

type CommonConfig struct {
	DialRetry          int    `toml:"dial_retry"`
	HeaderIdentify     string `toml:"header_identify"`
	SuccessRespMessage string `toml:"success_resp_message"`
	ConnectTimeout     int64  `toml:"connect_timeout"`
	KcpIs              int64  `toml:"connect_timeout"`
	KcpReciprocity     bool   `toml:"kcp_reciprocity"`
}

// Gate 配置
type Gate struct {
	Port        int  `toml:"port"`
	Enabled     bool `toml:"enabled"`
	DialTimeout int  `toml:"dial_timeout"`
	ZeroCopy    bool `toml:"zero_copy"`
}

type log struct {
	InfofilePath  string `toml:"infofilePath"`
	ErrorfilePath string `toml:"errorfilePath"`
}

// TomlConfig 配置
var TomlConfig = &tomlConfig{}

// InitConfig 初始化配置
func InitConfig() {
	_, err := toml.DecodeFile("../config/config.toml", TomlConfig)
	if err != nil {
		Error("toml.DecodeFile err:", err)
	}
}
