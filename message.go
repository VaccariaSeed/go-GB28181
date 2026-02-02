package go_GB28181

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ghettovoice/gosip/sip"
)

// 从request中获取一个指定的tag标识
func selectRequestHandler(req sip.Request, k string) (string, error) {
	for _, handler := range req.Headers() {
		if handler.Name() == k {
			return handler.Value(), nil
		}
	}
	return "", fmt.Errorf("no such tag %s", k)
}

// CreateNoMsgTypeResponse SIP/2.0 200 OK 创建这种
func CreateNoMsgTypeResponse(req sip.Request, statusCode int, conf *Config) sip.Message {
	resp := sip.NewResponseFromRequest(req.MessageID(), req, sip.StatusCode(statusCode), "OK", "")
	contact, flag := req.Contact()
	if flag {
		resp.AppendHeader(contact)
	}
	ua := sip.UserAgentHeader(conf.UserAgent)
	resp.AppendHeader(&ua)
	sip.CopyHeaders(expires, req, resp)
	if !strings.Contains(req.StartLine(), strings.TrimSpace(conf.ServerSIP)) {
		resp.SetStatusCode(sip.StatusCode(http.StatusForbidden))
	}
	return resp
}

func createServerRequestByRequest(messID sip.MessageID, method sip.RequestMethod, uri sip.Uri, hdrs []sip.Header, body string, req sip.Request, conf *Config) sip.Message {
	request := sip.NewRequest(messID, method, uri, conf.SipVersion, hdrs, body, nil)
	//via
	via, flag := req.Via()
	if flag {
		request.AppendHeader(via)
	}
	//from
	from, flag := req.From()
	if flag {
		request.AppendHeader(from)
	}
	//to
	to, flag := req.To()
	if flag {
		request.AppendHeader(to)
	}
	//Call-ID
	callId, flag := req.CallID()
	if flag {
		request.AppendHeader(callId)
	}
	seq, flag := req.CSeq()
	if flag {
		request.AppendHeader(seq)
	}
	ua := sip.UserAgentHeader(conf.UserAgent)
	request.AppendHeader(&ua)
	//Content-Length
	cl := sip.ContentLength(len(body))
	request.AppendHeader(&cl)
	return request
}

func CreateServerRequestNoReq(messID sip.MessageID, method sip.RequestMethod, hdrs []sip.Header, body string, conf *Config, cameraSIP string, cameraHost string) sip.Message {
	port := sip.Port(conf.Port)
	uri := &sip.SipUri{FUser: sip.String{Str: conf.ServerSIP}, FHost: conf.ServerIP, FPort: &port}
	req := sip.NewRequest(messID, method, uri, conf.SipVersion, hdrs, body, nil)
	vt := generateSIPBranchID(6)
	callIdKet := conf.ServerSIP + "_" + vt
	callId := sip.CallID(callIdKet)
	req.AppendHeader(&callId)
	//CSeq
	seq := sip.CSeq{SeqNo: flushSeqNo(), MethodName: method}
	req.AppendHeader(&seq)
	//From
	req.AppendHeader(serverFrom(conf))
	//to
	req.AppendHeader(serverTo(cameraSIP, cameraHost))
	//Max-Forwards
	mf := sip.MaxForwards(conf.MaxForwards)
	req.AppendHeader(&mf)
	//Contact
	contact := sip.ContactHeader{
		Address: &sip.SipUri{FUser: sip.String{Str: conf.ServerSIP}, FHost: conf.ServerSIP[:10]},
	}
	req.AppendHeader(&contact)
	//Via
	req.AppendHeader(serverVia("tag", conf))
	cl := sip.ContentLength(len(body))
	req.AppendHeader(&cl)
	return req
}

func serverVia(key string, conf *Config) sip.Header {
	viaHandler := sip.ViaHeader{}
	port := sip.Port(conf.Port)
	params := sip.NewParams()
	branch := generateSIPBranchID(6)
	params.Add(key, sip.String{Str: branch})
	hop := &sip.ViaHop{ProtocolName: conf.ProtocolName, ProtocolVersion: conf.ProtocolVersion, Transport: string(conf.Trans), Host: conf.ServerIP, Port: &port, Params: params}
	return append(viaHandler, hop)
}

func serverFrom(conf *Config) sip.Header {
	params := sip.NewParams()
	branch := generateSIPBranchID(6)
	params.Add("tag", sip.String{Str: branch})
	return &sip.FromHeader{Address: &sip.SipUri{FIsEncrypted: false, FUser: sip.String{Str: conf.ServerSIP}, FHost: conf.ServerSIP[:10]}, Params: params}
}

func serverTo(cameraSIP string, cameraHost string) sip.Header {
	return &sip.ToHeader{Address: &sip.SipUri{FUser: sip.String{Str: cameraSIP}, FHost: cameraHost}}
}
