package go_GB28181

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type clientSnap struct {
	lock sync.Mutex
	snap map[string]*Camera
	ch   chan string
}

func (s *clientSnap) flushUdpCamera(addr *net.UDPAddr, conf *Config, udpConn *net.UDPConn) *Camera {
	s.lock.Lock()
	defer s.lock.Unlock()
	key := addr.String()
	if _, ok := s.snap[key]; !ok {
		cli := &UDPClient{
			addr: addr,
			conn: udpConn,
		}
		camera := &Camera{
			host:   key,
			client: cli,
			close:  s.ch,
			Config: conf,
		}
		s.snap[key] = camera
	}
	return s.snap[key]
}

func (s *clientSnap) flushTcpCamera(conn net.Conn, conf *Config) *Camera {
	s.lock.Lock()
	defer s.lock.Unlock()
	key := conn.RemoteAddr().String()
	if cli, ok := s.snap[key]; ok {
		return cli
	}
	cli := &TCPClient{
		conn: conn,
	}
	camera := &Camera{
		host:   key,
		client: cli,
		close:  s.ch,
		Config: conf,
	}
	s.snap[key] = camera
	return camera
}

func (s *clientSnap) clear() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.snap = make(map[string]*Camera)
}

func (s *clientSnap) run(ctx context.Context) {
	//关闭客户端
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-s.ch:
				s.lock.Lock()
				delete(s.snap, msg)
				s.lock.Unlock()
			default:
				continue
			}
		}
	}()
}

func (s *clientSnap) ObtainCamera(sip string) (*Camera, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, camera := range s.snap {
		if camera.sip == sip {
			return camera, nil
		}
	}
	return nil, errors.New("camera not found by: " + sip)
}

func (s *clientSnap) monitor(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.lock.Lock()
				for _, camera := range s.snap {
					latest := atomic.LoadInt64(&camera.LatestTime)
					if !camera.Registered() {
						//没有注册，判定是不是半个小时
						if time.Now().UnixMilli() >= latest+30*60*1000 {
							camera.closeByErr(errors.New("if no registration message has been received for the camera within half an hour, it is determined to be offline"))
						}
					} else {
						//注册了，是不是心跳3倍依旧没通讯
						exp := camera.keepAlive * 3
						if time.Now().UnixMilli() >= latest+exp {
							camera.closeByErr(errors.New("no communication within 3 times the heartbeat time, judged as offline"))
						}
					}
				}
				s.lock.Unlock()
			default:
				continue
			}
		}
	}()
}

func (s *clientSnap) CameraList() ([]*CameraModel, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if len(s.snap) == 0 {
		return nil, errors.New("no camera registed")
	}
	var cm []*CameraModel
	for _, camera := range s.snap {
		cm = append(cm, &CameraModel{
			SIP:        camera.sip,
			UserAgent:  camera.CameraUa,
			Host:       camera.host,
			KeepAlive:  camera.keepAlive,
			LatestTime: atomic.LoadInt64(&camera.LatestTime),
		})
	}
	return cm, nil
}
