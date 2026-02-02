package go_GB28181

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ghettovoice/gosip/sip"
)

const (
	expires      = "Expires"
	cameraLogOut = "0" //退出登录

	cameraUa = "User-Agent"
	auth     = "authorization"
)

func (c *Camera) registerHandle(req sip.Request) error {
	from, flag := req.From()
	if !flag {
		return errors.New("not support from")
	}
	cameraSip := from.Address.User().String()
	if cameraSip == "" {
		return errors.New("not support cameraSip")
	}
	c.sip = cameraSip
	expiresValue, err := selectRequestHandler(req, expires)
	if err != nil {
		return err
	}
	switch expiresValue {
	case cameraLogOut:
		//退出登录
		err = c.cameraLogOutHandle(req)
	default:
		//登录
		err = c.cameraLogInHandle(req, cameraSip)
	}
	return err
}

// 退出登录
func (c *Camera) cameraLogOutHandle(req sip.Request) error {
	resp := CreateNoMsgTypeResponse(req, http.StatusOK, c.Config)
	err := c.Write(resp)
	if err == nil {
		err = c.Close()
	}
	return err
}

// 登录
func (c *Camera) cameraLogInHandle(req sip.Request, cameraSip string) error {
	if !c.Auth {
		//不需要鉴权
		resp := CreateNoMsgTypeResponse(req, http.StatusOK, c.Config)
		err := c.Write(resp)
		c.FlushRegistered(true, nil)
		c.login(req)
		return err
	}
	password := c.Handle.PasswordHandle(c.host, cameraSip)
	authorization := req.GetHeaders(auth)
	if authorization == nil || len(authorization) < 1 {
		//第一次登录
		authItem := "Digest realm=\"%s\", nonce=\"%s\", opaque=\"%s\", algorithm=MD5"
		nonce := generateSIPBranchID(6)
		pwd := hashWithSalt(password, nonce)
		authItem = fmt.Sprintf(authItem, cameraSip[:8], nonce, pwd)
		//第一次的注册信息,服务端请求鉴权
		serverReq := createServerRequestByRequest("", "SIP/2.0 401 Unauthorized", &NoneUri{}, []sip.Header{&Authenticate{Auth: authItem}}, "", req, c.Config)
		return c.Write(serverReq)
	}
	//进入鉴权步骤
	statusCode := cameraAuth(authorization[0].Value(), password)
	reap := CreateNoMsgTypeResponse(req, statusCode, c.Config)
	err := c.Write(reap)
	c.FlushRegistered(true, nil)
	c.login(req)
	return err
}

func cameraAuth(value string, password string) int {
	da, err := parseDigestAuth(value)
	if err != nil {
		return http.StatusForbidden
	}
	key1 := fmt.Sprintf("%s:%s:%s", da.Username, da.Realm, password)
	key1Md5Str := md5Encode(key1)
	key2Md5Str := md5Encode("REGISTER:" + da.URI)
	key3 := fmt.Sprintf("%s:%s:%s", key1Md5Str, da.Nonce, key2Md5Str)
	key3Md5Str := md5Encode(key3)
	flag := da.Response == key3Md5Str
	if !flag {
		return http.StatusForbidden
	}
	return http.StatusOK
}

func parseDigestAuth(authStr string) (*digestAuth, error) {
	if !strings.HasPrefix(authStr, "Digest ") {
		return nil, fmt.Errorf("不是有效的 Digest 认证字符串")
	}
	// 移除 "Digest " 前缀
	authStr = strings.TrimPrefix(authStr, "Digest ")
	// 解析键值对
	params := make(map[string]string)
	parts := strings.Split(authStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		// 分割键值
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		// 移除值可能的引号
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}
		params[key] = value
	}
	// 创建 DigestAuth 对象
	da := &digestAuth{
		Username:  params["username"],
		Realm:     params["realm"],
		Nonce:     params["nonce"],
		Response:  params["response"],
		URI:       params["uri"],
		Opaque:    params["opaque"],
		Algorithm: params["algorithm"],
	}
	return da, nil
}
