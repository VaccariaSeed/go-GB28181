package go_GB28181

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"
)

// NewServer 创建GB28181服务端
func NewServer(config *Config) *Server {
	return &Server{Config: config, clis: &clientSnap{snap: make(map[string]*Camera), ch: make(chan string, 10)}}
}

// Server GB28181服务端
type Server struct {
	*Config

	//TCP套接字相关
	listener *net.TCPListener

	//UDP套接字相关
	udpAddr *net.UDPAddr
	udpConn *net.UDPConn

	//流程相关
	ctx    context.Context
	cancel context.CancelFunc

	//客户端相关
	clis *clientSnap
}

func (s *Server) close(err error) {
	_ = s.Close()
	s.Handle.ServerClosedHandle(err)
}

// Close 关闭服务端
func (s *Server) Close() error {
	var err error
	if s.cancel != nil {
		s.cancel()
		s.cancel, s.ctx = nil, nil
	}
	if s.listener != nil {
		err = s.listener.Close()
		s.listener = nil
	}
	if s.udpConn != nil {
		err = s.udpConn.Close()
		s.udpConn, s.udpAddr = nil, nil
	}
	
	s.clis.clear()
	return err
}

// Open 打开服务端
func (s *Server) Open() error {
	_ = s.Close()
	if s.Config == nil {
		return errors.New("config is nil")
	}
	if s.Handle == nil {
		return errors.New("handle is nil")
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	var err error
	//会根据连接类型进行区分，并以此来创建连接
	switch s.Trans {
	case TcpTrans:
		err = s.tcpOpen()
	case UdpTrans:
		err = s.udpOpen()
	default:
		err = fmt.Errorf("unsupport trans type: %s", s.Trans)
	}
	if err == nil {
		//处理client相关的业务
		s.clis.run(s.ctx)
		s.clis.monitor(s.ctx)
	}
	return err
}

// 打开tcp服务端
func (s *Server) tcpOpen() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		return err
	}
	s.listener = listener.(*net.TCPListener)
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			default:
				_ = s.listener.SetDeadline(time.Now().Add(s.ReadTimeout))
				conn, connErr := s.listener.Accept()
				if connErr != nil && s.Handle.ErrorVerifyHandle(connErr) {
					s.close(connErr)
				}
				if connErr == nil {
					s.tcpClientHandle(conn)
				}
			}
		}
	}()
	return nil
}

func (s *Server) udpOpen() error {
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		return err
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	s.udpAddr, s.udpConn = udpAddr, udpConn
	go func() {
		buffer := make([]byte, s.BuffSize)
		for {
			select {
			case <-s.ctx.Done():
				return
			default:
				_ = s.udpConn.SetReadDeadline(time.Now().Add(s.ReadTimeout))
				n, clientAddr, readErr := s.udpConn.ReadFromUDP(buffer)
				if readErr != nil && s.Handle.ErrorVerifyHandle(readErr) {
					s.close(readErr)
				}
				if readErr == nil {
					camera := s.clis.flushUdpCamera(clientAddr, s.Config, s.udpConn)
					camera.msgFlow(buffer[:n])
				}
			}
		}
	}()
	return nil
}

func (s *Server) tcpClientHandle(conn net.Conn) {
	camera := s.clis.flushTcpCamera(conn, s.Config)
	go camera.tcpFlow(nil, s.ctx)
}
