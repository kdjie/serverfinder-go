package serverfinder

import (
	"fmt"
	"time"
	"encoding/json"
)

const (
	// 服务器Key前缀，方便使用redis keys SK*搜索。
	// 完整格式："SK_服务名_IP:PORT"
	SERVER_KEY_PREFIX = "SK"
)

// 服务器节点格式定义
type ServerNode struct {
	Name string		`json:"name"`
	Ip string		`json:"ip"`
	Port int		`json:"port"`
	Tags []string	`json:"tags"`
	LiveTick int64	`json:"livetick"`
}

func (s *ServerNode) toKey() string {
	return fmt.Sprintf("%s_%s_%s:%d", SERVER_KEY_PREFIX, s.Name, s.Ip, s.Port)
}

func (s *ServerNode) toJson() string {
	s.LiveTick = time.Now().Unix()

	data, err := json.Marshal(s)
	if err != nil {
		return ""
	}

	return string(data)
}

func (s *ServerNode) fromJson(strJson string) bool {
	err := json.Unmarshal([]byte(strJson), s)
	if err != nil {
		return false
	}

	return true
}
