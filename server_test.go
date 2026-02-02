package go_GB28181

import (
	"fmt"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	conf := NewDefaultConfig(15060, UdpTrans, "127.0.0.1", "34020000002000000001", &CameraHandler{})
	server := NewServer(conf)
	err := server.Open()
	if err != nil {
		panic(err)
	}
	time.Sleep(10 * time.Second)
	//目录检索
	taglogs, err := server.CatalogueSearch("34020000001110739997", 5*time.Second)
	fmt.Println(taglogs)
	fmt.Println(err)
	//拉取实时流
	err = server.FetchRealTimeStream("34020000001110739997", "127.0.0.1", 589, 5*time.Second, "")
	fmt.Println(err)
	//关流
	err = server.StopFetchStream("34020000001110739997", 5*time.Second)
	fmt.Println(err)
	back := &PlayBack{
		TargetNumber:    "1000000001",
		SourceNumber:    "1000000001",
		ChanNumber:      0,
		StartTime:       time.Now().Add(-2 * time.Hour).Unix(),
		EndTime:         time.Now().Add(-1 * time.Hour).Unix(),
		VideoServerIp:   "127.0.0.1",
		videoServerPort: 9978,
	}
	err = server.FetchPlayBackStream("34020000001110789924", back, 5*time.Second)
	//回放暂停
	err = server.ControlPlayBackStop("34020000001110794199", 1, 5*time.Second)
	start := time.Now().Add(-3 * time.Hour)
	now := time.Now()
	err = server.ControlPlayBackPlay("34020000001110794199", 1, 5*time.Second, &start, &now)
	//回放定位
	err = server.PlayBackPositioning("34020000001110794199", -211, 1, 1.0, 5*time.Second)
	//控制，以ptz举例，ptz有聚焦、光圈的方法
	ptz := &PTZ{
		Version: 0,
		Address: 0,
	}
	cmd, _ := ptz.BuildPTZ(ZoomIn, 1, TiltDown, 1, PanIdle, 1)
	err = server.PTZControl("34020000001110004271", cmd, 5*time.Second)
	//获取所有在线摄像头
	cameras, err := server.CameraList()
	fmt.Println("cameras", cameras)
	time.Sleep(10 * time.Second)
	cameras, err = server.CameraList()
	fmt.Println("cameras", cameras)
	time.Sleep(1 * time.Hour)
}

var _ GB28181Handler = (*CameraHandler)(nil)

type CameraHandler struct{}

func (c *CameraHandler) ErrorVerifyHandle(err error) bool {
	if err == nil {
		return false
	}
	fmt.Println("套接字错误：" + err.Error())
	return false
}

func (c *CameraHandler) ServerClosedHandle(err error) {
	fmt.Println("服务器关闭：" + err.Error())
}

func (c *CameraHandler) ReceivedHandle(clientHost string, clientSIP string, frame string, err error) {
	fmt.Println("-------收到了消息--------")
	fmt.Println("clientHost:" + clientHost)
	fmt.Println("clientSIP:" + clientSIP)
	fmt.Println(frame)
	if err != nil {
		fmt.Println("解析错误：" + err.Error())
	}
}

func (c *CameraHandler) SentHandle(clientHost string, clientSIP string, frame string, err error) {
	fmt.Println("-------发送了消息--------")
	fmt.Println(clientHost)
	fmt.Println(clientSIP)
	fmt.Println(frame)
	if err != nil {
		fmt.Println("发送错误：" + err.Error())
	}
}

func (c *CameraHandler) PasswordHandle(clientHost string, cameraSIP string) string {
	fmt.Println("------获取密码-------")
	fmt.Println(clientHost)
	fmt.Println(cameraSIP)
	return "12345678"
}

func (c *CameraHandler) CameraOnHandle(clientHost string, clientSIP string) {
	fmt.Println("Camera上线：", clientHost, "  ", clientSIP)
}

func (c *CameraHandler) CameraDownHandle(clientHost string, clientSIP string, err error) {
	fmt.Println("Camera下线：", clientHost, "  ", clientSIP)
}
