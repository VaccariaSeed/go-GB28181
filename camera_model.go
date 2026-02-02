package go_GB28181

type CameraModel struct {
	SIP        string `json:"sip"`
	UserAgent  string `json:"user_agent"`
	Host       string `json:"host"`
	KeepAlive  int64  `json:"keep_alive"`  //单位毫秒
	LatestTime int64  `json:"latest_time"` //最后一次通讯时间
}
