package go_GB28181

import (
	"net"
	"time"

	"github.com/ghettovoice/gosip/log"
	"github.com/ghettovoice/gosip/sip"
)

var logger = log.NewDefaultLogrusLogger()

func init() {
	logger.SetLevel(uint32(log.PanicLevel))
}

var _ Clienter = (*TCPClient)(nil) //TCP客户端
var _ Clienter = (*UDPClient)(nil) //UDP客户端

type Clienter interface {
	Read(buffer []byte) (int, error)
	SetDeadline(timeout time.Duration) error
	Write(frame sip.Message, writeTimeout time.Duration) (n int, err error)
	Close() error
}

type TCPClient struct {
	conn net.Conn
}

func (T *TCPClient) Read(buffer []byte) (int, error) {
	return T.conn.Read(buffer)
}

func (T *TCPClient) SetDeadline(timeout time.Duration) error {
	return T.conn.SetDeadline(time.Now().Add(timeout))
}

func (T *TCPClient) Write(frame sip.Message, writeTimeout time.Duration) (n int, err error) {
	err = T.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	if err != nil {
		return
	}
	return T.conn.Write([]byte(frame.String()))
}

func (T *TCPClient) Close() error {
	var err error
	if T.conn != nil {
		err = T.conn.Close()
	}
	return err
}

type UDPClient struct {
	addr *net.UDPAddr
	conn *net.UDPConn
}

func (U *UDPClient) Read(buffer []byte) (int, error) {
	return U.conn.Read(buffer)
}

func (U *UDPClient) SetDeadline(timeout time.Duration) error {
	return U.conn.SetDeadline(time.Now().Add(timeout))
}

func (U *UDPClient) Write(frame sip.Message, writeTimeout time.Duration) (n int, err error) {
	err = U.conn.SetDeadline(time.Now().Add(writeTimeout))
	if err != nil {
		return
	}
	return U.conn.WriteTo([]byte(frame.String()), U.addr)
}

func (U *UDPClient) Close() error {
	return nil
}
