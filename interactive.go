package go_GB28181

import (
	"errors"
	"time"
)

// FetchRealTimeStream 拉取已经注册摄像头的实时流
// cameraSIP 摄像头的SIP
// videoServerIp 流媒体服务器的IP
// videoServerPort 接收流的接口
// timeout 超时
// pvr "v=0\no=%s 0 0 IN IP4 %s\ns=Play\nc=IN IP4 %s\nt=0 0\nm=video %s RTP/AVP 96 98 97\na=recvonly\na=rtpmap:96 PS/90000\na=rtpmap:98 H264/90000\na=rtpmap:97 MPEG4/90000\ny=0100000001\nf=" 这种，传入""的话会使用默认的
func (s *Server) FetchRealTimeStream(cameraSIP string, videoServerIp string, videoServerPort int, timeout time.Duration, pvr string) error {
	camera, err := s.clis.ObtainCamera(cameraSIP)
	if err != nil {
		return err
	}
	camera.cameraLock.Lock()
	defer camera.cameraLock.Unlock()
	return camera.FetchRealTimeStream(videoServerIp, videoServerPort, timeout, pvr)
}

// CatalogueSearch 目录检索
// cameraSIP 摄像头的SIP
// timeout 超时
func (s *Server) CatalogueSearch(cameraSIP string, timeout time.Duration) ([]*CatalogueResponse, error) {
	camera, err := s.clis.ObtainCamera(cameraSIP)
	if err != nil {
		return nil, err
	}
	camera.cameraLock.Lock()
	defer camera.cameraLock.Unlock()
	return camera.CatalogueSearch(timeout)
}

// StopFetchStream 停止拉流
// cameraSIP 摄像头的SIP
// timeout 超时
func (s *Server) StopFetchStream(cameraSIP string, timeout time.Duration) error {
	camera, err := s.clis.ObtainCamera(cameraSIP)
	if err != nil {
		return err
	}
	camera.cameraLock.Lock()
	defer camera.cameraLock.Unlock()
	return camera.StopFetchStream(timeout)
}

// PlayBack 回放请求参数
type PlayBack struct {
	TargetNumber    string //目标通道号
	SourceNumber    string //源设备通道号
	ChanNumber      int    //通道号
	StartTime       int64  //开始时间,秒级时间戳
	EndTime         int64  //结束时间,秒级时间戳
	VideoServerIp   string //接收IP
	videoServerPort int    //接收端口号
	Pvr             string //自定义的播放参数，如果=="",就是用默认的
}

// FetchPlayBackStream 回放
// cameraSIP 摄像头的SIP
// PlayBack 回放请求参数
// timeout 超时
func (s *Server) FetchPlayBackStream(cameraSIP string, back *PlayBack, timeout time.Duration) error {
	if back == nil {
		return errors.New("play back items is nil")
	}
	camera, err := s.clis.ObtainCamera(cameraSIP)
	if err != nil {
		return err
	}
	camera.cameraLock.Lock()
	defer camera.cameraLock.Unlock()
	return camera.FetchPlayBackStream(back, timeout)
}

// ControlPlayBackStop 回放暂停
// cameraSIP 摄像头的SIP
// PauseTime 回放的暂停延迟
// timeout 超时
func (s *Server) ControlPlayBackStop(cameraSIP string, PauseTime int, timeout time.Duration) error {
	camera, err := s.clis.ObtainCamera(cameraSIP)
	if err != nil {
		return err
	}
	camera.cameraLock.Lock()
	defer camera.cameraLock.Unlock()
	return camera.ControlPlayBackStop(PauseTime, timeout)
}

// ControlPlayBackPlay 播放回放
// cameraSIP 摄像头的SIP
// scale 倍速
// timeout 超时
// startTime 开始时间
// endTime 结束时间
func (s *Server) ControlPlayBackPlay(cameraSIP string, scale float32, timeout time.Duration, startTime *time.Time, endTime *time.Time) error {
	camera, err := s.clis.ObtainCamera(cameraSIP)
	if err != nil {
		return err
	}
	camera.cameraLock.Lock()
	defer camera.cameraLock.Unlock()
	return camera.ControlPlayBackPlay(scale, timeout, startTime, endTime)
}

// PlayBackPositioning 回放定位
// cameraSIP 摄像头的SIP
// startTime 开始时间,秒数，第几秒开始回放
// endTime 结束时间，秒数，第几秒结束
// scale 倍速
// timeout 超时
func (s *Server) PlayBackPositioning(cameraSIP string, startTime int, endTime int, scale float32, timeout time.Duration) error {
	camera, err := s.clis.ObtainCamera(cameraSIP)
	if err != nil {
		return err
	}
	camera.cameraLock.Lock()
	defer camera.cameraLock.Unlock()
	return camera.PlayBackPositioning(startTime, endTime, scale, timeout)
}

// PTZControl 云台控制
func (s *Server) PTZControl(cameraSIP, cmd string, timeout time.Duration) error {
	camera, err := s.clis.ObtainCamera(cameraSIP)
	if err != nil {
		return err
	}
	camera.cameraLock.Lock()
	defer camera.cameraLock.Unlock()
	return camera.PTZControl(cmd, timeout)
}

// CameraList 获取所有在线摄像头
func (s *Server) CameraList() ([]*CameraModel, error) {
	return s.clis.CameraList()
}
