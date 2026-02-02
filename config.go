package go_GB28181

import "time"

type LinkType string

// 两个连接类型
const (
	TcpTrans LinkType = "tcp"
	UdpTrans LinkType = "udp"
)

// NewDefaultConfig 创建SIP服务端的配置
func NewDefaultConfig(port int16, trans LinkType, serverIp string, serverSip string, handle GB28181Handler) *Config {
	return &Config{
		Port:            port,
		Trans:           trans,
		ServerIP:        serverIp,
		ServerSIP:       serverSip,
		ProtocolName:    "SIP",
		ProtocolVersion: "2.0",
		SipVersion:      "SIP/2.0",
		BuffSize:        1024,
		Handle:          handle,
		ReadTimeout:     time.Second * 5,
		WriteTimeout:    time.Second * 5,
		UserAgent:       "GB28181 SERVER",
		MaxForwards:     70,
	}
}

// Config 服务端配置
type Config struct {
	Port            int16          //端口号
	WriteTimeout    time.Duration  //写超时
	ReadTimeout     time.Duration  //读超时
	Trans           LinkType       //连接类型
	ServerIP        string         //服务端IP
	ServerSIP       string         //服务端的sip
	MaxForwards     int            // 最大转发跳数,默认70
	ProtocolVersion string         //默认“2.0”
	ProtocolName    string         //默认“SIP”
	SipVersion      string         //默认"SIP/2.0"
	UserAgent       string         //服务端名称
	BuffSize        int            //缓冲区大小，建议不要小于1024
	Handle          GB28181Handler //回调处理器
	Auth            bool           //是否需要针对摄像头鉴权
}
