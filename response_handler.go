package go_GB28181

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ghettovoice/gosip/sip"
)

const (
	okType       = "OK"
	TryingType   = "Trying"
	registerType = string(sip.REGISTER)
)

// 解析response
func (c *Camera) responseHandle(resp sip.Response) error {
	callId, flag := resp.CallID()
	if !flag {
		return errors.New("bad call id")
	}
	to, flag := resp.To()
	if !flag {
		return errors.New("response not support to")
	}
	cameraSip := to.Address.User().String()
	if strings.TrimSpace(cameraSip) == "" {
		return errors.New("response not support camera sip")
	}
	cseq, flag := resp.CSeq()
	if !flag {
		return errors.New("response not support seq")
	}
	reason := resp.Reason()
	var err error
	switch reason {
	case okType:
		err = c.okHandle(resp, callId, cseq)
	case TryingType:
		err = c.tryingHandle(resp)
	}

	return err
}

func (c *Camera) tryingHandle(resp sip.Response) error {
	statusCode := resp.StatusCode()
	if statusCode == http.StatusContinue {
		if c.pip != nil {
			c.pip.OccValue(TryingType)
		}
	}
	return nil
}

func (c *Camera) okHandle(resp sip.Response, callId *sip.CallID, cseq *sip.CSeq) error {
	var err error
	if resp.StatusCode() == http.StatusOK && cseq.MethodName == sip.INVITE {
		//发送ack
		ackFrame := CreateServerRequestNoReq(resp.MessageID(), sip.ACK, nil, "", c.Config, c.sip, c.host)
		ackFrame.RemoveHeader("Call-ID")
		ackFrame.AppendHeader(callId)
		err = c.Write(ackFrame)
	}
	if c.pip != nil {
		if resp.StatusCode() != http.StatusOK {
			c.pip.OccError(errors.New(fmt.Sprintf("camera response code:%v", resp.StatusCode())))
		} else {
			c.pip.OccValue(okType)
		}
	}
	return err
}
