package go_GB28181

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ghettovoice/gosip/sip"
)

const (
	keepAliveType     = "Keepalive" //心跳
	catalogType       = "Catalog"
	DeviceControlType = "DeviceControl"
)

func (c *Camera) messageHandle(req sip.Request) error {
	body := req.Body()
	cmdType, err := parseWithXmlData(body)
	if err != nil {
		return err
	}
	if cmdType == "" {
		return errors.New("camera message_request cmdType not found")
	}
	from, flag := req.From()
	if !flag {
		return errors.New("keepAlive not support from")
	}
	cameraSip := from.Address.User().String()
	if cameraSip == "" {
		return errors.New("keepAlive not support cameraSip")
	}
	switch cmdType {
	case keepAliveType:
		//保活
		err = c.keepAliveHandle(req)
	case catalogType:
		//目录
		err = c.catalogHandle(req)
	case DeviceControlType:
		err = c.deviceControlHandle(req)
	default:
		err = errors.New(fmt.Sprintf("unsupported command: %s", cmdType))
	}
	return err
}

func (c *Camera) deviceControlHandle(req sip.Request) error {
	var control *ControlResponse
	err := gb2312XMLDecoder([]byte(req.Body()), &control)
	if err != nil {
		return err
	}
	if c.pip != nil {
		c.pip.OccValue(control.Result)
	}
	return nil
}

func (c *Camera) keepAliveHandle(req sip.Request) error {
	resp := sip.NewResponseFromRequest("", req, sip.StatusCode(http.StatusOK), "OK", "")
	ua := sip.UserAgentHeader(c.UserAgent)
	resp.AppendHeader(&ua)
	return c.Write(resp)
}

func (c *Camera) catalogHandle(req sip.Request) error {
	var catalogue *CatalogueResponse
	err := gb2312XMLDecoder([]byte(req.Body()), &catalogue)
	if err != nil {
		return err
	}
	resp := CreateNoMsgTypeResponse(req, http.StatusOK, c.Config)
	if c.pip != nil {
		c.pip.OccValue(catalogue)
	}
	return c.Write(resp)
}
