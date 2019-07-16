package serverfinder

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
