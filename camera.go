package go_GB28181

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ghettovoice/gosip/sip"
	"github.com/ghettovoice/gosip/sip/parser"
)

type Camera struct {
	host string
	sip  string

	//连接相关
	registered bool //连接状态
	client     Clienter

	//配置相关
	*Config

	//业务相关
	cameraLock sync.Mutex
	LatestTime int64 //最后一次通讯时间,毫秒值
	close      chan string
	pip        *MessageCar
	keepAlive  int64
	CameraUa   string
}

func (c *Camera) Write(frame sip.Message) error {
	_, err := c.client.Write(frame, c.WriteTimeout)
	c.Handle.SentHandle(c.host, c.sip, frame.String(), err)
	return err
}

// FlushRegistered 刷新注册状态
func (c *Camera) FlushRegistered(flag bool, err error) {
	c.registered = flag
	if c.registered {
		c.Handle.CameraOnHandle(c.host, c.sip)
	} else {
		c.Handle.CameraDownHandle(c.host, c.sip, err)
	}
}

func (c *Camera) closeByErr(err error) {
	_ = c.client.Close()
	c.close <- c.host
	c.FlushRegistered(false, err)
}

func (c *Camera) Close() error {
	err := c.client.Close()
	c.close <- c.host
	c.FlushRegistered(false, nil)
	return err
}

func (c *Camera) tcpFlow(_ []byte, ctx context.Context) {
	var n int
	var err error
	buffer := make([]byte, c.BuffSize)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_ = c.client.SetDeadline(c.ReadTimeout)
			n, err = c.client.Read(buffer)
			if err != nil && c.Handle.ErrorVerifyHandle(err) {
				c.closeByErr(err)
				return
			}
			if n > 0 && err == nil {
				c.msgFlow(buffer[:n])
			}
		}
	}
}

func (c *Camera) Registered() bool {
	return c.registered == true
}

func (c *Camera) msgFlow(data []byte) {
	if data == nil || len(data) == 0 {
		return
	}
	msg, err := parser.ParseMessage(data, logger)
	if err != nil {
		c.Handle.ReceivedHandle(c.host, c.sip, string(data), err)
		return
	}
	c.Handle.ReceivedHandle(c.host, c.sip, msg.String(), err)
	atomic.StoreInt64(&c.LatestTime, time.Now().UnixMilli())
	switch msg.(type) {
	case sip.Request:
		err = c.requestHandle(msg.(sip.Request))
	case sip.Response:
		err = c.responseHandle(msg.(sip.Response))
	default:
		err = errors.New("invalid message type")
	}
	if err != nil {
		c.Handle.ReceivedHandle(c.host, c.sip, msg.String(), err)
	}
}

// 解析收到的请求
func (c *Camera) requestHandle(req sip.Request) error {
	var err error
	switch req.Method() {
	case sip.REGISTER:
		err = c.registerHandle(req) //注册
	case sip.MESSAGE:
		err = c.messageHandle(req) //消息
	default:
		err = fmt.Errorf("unsupported method: %s", req.Method())
	}
	return err
}

// FetchRealTimeStream 拉取已经注册摄像头的实时流
func (c *Camera) FetchRealTimeStream(serverIp string, serverPort int, timeout time.Duration, pvrStr string) error {
	if pvrStr == "" {
		pvrStr = pvr(serverIp, fmt.Sprintf("%d", serverPort), c.ServerSIP)
	}
	req := CreateServerRequestNoReq("", sip.INVITE, nil, pvrStr, c.Config, c.sip, c.host)
	ct := sip.ContentType("application/sdp")
	req.AppendHeader(&ct)
	//等待消息
	c.pip = NewMessageCar(2, timeout)
	defer func() { c.pip = nil }()
	err := c.Write(req)
	if err != nil {
		return err
	}
	result, err := c.pip.Down()
	if err != nil {
		return err
	}
	if len(result) < 2 {
		return errors.New("camera no response")
	}
	if slices.Contains(result, TryingType) && slices.Contains(result, okType) {
		return nil
	}
	return errors.New("camera response lost")
}

// CatalogueSearch 目录检索
func (c *Camera) CatalogueSearch(timeout time.Duration) ([]*CatalogueResponse, error) {
	body, err := modelToXml(&Query{
		XMLName:  xml.Name{Space: "", Local: "Query"},
		CmdType:  "Catalog",
		SN:       fmt.Sprintf("%d", flushSeqNo()),
		DeviceID: c.sip,
	})
	if err != nil {
		return nil, err
	}
	req := CreateServerRequestNoReq("", sip.MESSAGE, nil, body, c.Config, c.sip, c.host)
	ct := sip.ContentType("Application/MANSCDP+xml")
	req.AppendHeader(&ct)
	c.pip = NewMessageCar(0, timeout)
	defer func() { c.pip = nil }()
	err = c.Write(req)
	if err != nil {
		return nil, err
	}
	logs, err := c.pip.Down()
	if err != nil {
		return nil, err
	}
	var resp []*CatalogueResponse
	for _, l := range logs {
		if ll, ok := l.(*CatalogueResponse); ok {
			resp = append(resp, ll)
		}
	}
	return resp, nil
}

func (c *Camera) StopFetchStream(timeout time.Duration) error {
	req := CreateServerRequestNoReq("", sip.BYE, nil, "", c.Config, c.sip, c.host)
	c.pip = NewMessageCar(1, timeout)
	defer func() { c.pip = nil }()
	err := c.Write(req)
	if err != nil {
		return err
	}
	result, err := c.pip.Down()
	if err != nil {
		return err
	}
	if len(result) == 0 {
		return errors.New("camera no response")
	}
	if slices.Contains(result, okType) {
		return nil
	}
	return errors.New("camera no response")
}

// FetchPlayBackStream 回放
func (c *Camera) FetchPlayBackStream(back *PlayBack, timeout time.Duration) error {
	if back.Pvr == "" {
		back.Pvr = backPvr(back, c.Config.ServerSIP, c.Config.ServerIP, c.sip)
	}
	req := CreateServerRequestNoReq("", sip.INVITE, nil, back.Pvr, c.Config, c.sip, c.host)
	ct := sip.ContentType("application/sdp")
	req.AppendHeader(&ct)
	ua := sip.UserAgentHeader(c.UserAgent)
	req.AppendHeader(&ua)
	//Subject
	sub := fmt.Sprintf("%s:%s:%s:%s", c.sip, back.TargetNumber, c.ServerSIP, back.SourceNumber)
	subject := Subject{sub}
	c.pip = NewMessageCar(2, timeout)
	defer func() { c.pip = nil }()
	req.AppendHeader(&subject)
	err := c.Write(req)
	if err != nil {
		return err
	}
	result, err := c.pip.Down()
	if err != nil {
		return err
	}
	if len(result) < 2 {
		return errors.New("camera no response")
	}
	if slices.Contains(result, TryingType) && slices.Contains(result, okType) {
		return nil
	}
	return errors.New("camera response lost")
}

// ControlPlayBackStop 回放暂停
func (c *Camera) ControlPlayBackStop(PauseTime int, timeout time.Duration) error {
	puse := "PAUSE MANSRTSP/1.0\nCSeq: %d\nPauseTime: %d"
	puse = fmt.Sprintf(puse, flushSeqNo(), PauseTime)
	return c.controlPlayBack(puse, timeout)
}

func (c *Camera) controlPlayBack(control string, timeout time.Duration) error {
	req := CreateServerRequestNoReq("", sip.INFO, nil, control, c.Config, c.sip, c.host)
	ct := sip.ContentType("Application/MANSRTSP")
	req.AppendHeader(&ct)
	ua := sip.UserAgentHeader(c.UserAgent)
	req.AppendHeader(&ua)
	c.pip = NewMessageCar(1, timeout)
	defer func() { c.pip = nil }()
	err := c.Write(req)
	if err != nil {
		return err
	}
	result, err := c.pip.Down()
	if err != nil {
		return err
	}
	if len(result) != 1 {
		return errors.New("camera no response")
	}
	if slices.Contains(result, okType) {
		return nil
	}
	return errors.New("camera response lost")
}

// ControlPlayBackPlay 播放回放
func (c *Camera) ControlPlayBackPlay(scale float32, timeout time.Duration, startTime *time.Time, endTime *time.Time) error {
	play := "PLAY MANSRTSP/1.0\nCSeq: %d\nScale: %v"
	play = fmt.Sprintf(play, flushSeqNo(), math.Trunc(float64(scale)*10)/10)
	if startTime != nil {
		play = fmt.Sprintf("%s\nRange:clock=%s-", play, startTime.Format("20060102T150405"))
		if endTime != nil {
			play = play + endTime.Format("20060102T150405")
		}
	}
	return c.controlPlayBack(play, timeout)
}

// PlayBackPositioning 回放定位
func (c *Camera) PlayBackPositioning(startTime int, endTime int, scale float32, timeout time.Duration) error {
	play := "PLAY MANSRTSP/1.0\nCSeq: %d\nScale: %v\nRange: npt=%d-%d"
	play = fmt.Sprintf(play, flushSeqNo(), math.Trunc(float64(scale)*10)/10, startTime, endTime)
	return c.controlPlayBack(play, timeout)
}

// PTZControl 云台控制
func (c *Camera) PTZControl(cmd string, timeout time.Duration) error {
	play := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<Control>\n    <CmdType>DeviceControl</CmdType>\n    <SN>%d</SN>\n    <DeviceID>%s</DeviceID>\n    <PTZCmd>%s</PTZCmd>\n</Control>"
	play = fmt.Sprintf(play, flushSeqNo(), c.sip, cmd)
	req := CreateServerRequestNoReq("", sip.MESSAGE, nil, play, c.Config, c.sip, c.host)
	ct := sip.ContentType("Application/MANSCDP+xml")
	req.AppendHeader(&ct)
	ua := sip.UserAgentHeader(c.UserAgent)
	req.AppendHeader(&ua)
	c.pip = NewMessageCar(2, timeout)
	defer func() { c.pip = nil }()
	err := c.Write(req)
	if err != nil {
		return err
	}
	result, err := c.pip.Down()
	if err != nil {
		return err
	}
	if len(result) != 2 {
		return errors.New("camera response error")
	}

	if result[0] == okType && result[1] == okType {
		return nil
	}
	return errors.New("camera response lost")
}

func (c *Camera) login(req sip.Request) {
	//获取Expires，获取User-Agent
	expiresValue, err := selectRequestHandler(req, expires)
	if err != nil {
		return
	}
	ua, err := selectRequestHandler(req, "User-Agent")
	if err != nil {
		return
	}
	keepAlive, err := strconv.Atoi(expiresValue)
	if err != nil {
		return
	}
	c.keepAlive = int64(keepAlive * 1000)
	c.CameraUa = ua
}
