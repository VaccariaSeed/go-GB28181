package go_GB28181

import (
	"fmt"

	"github.com/ghettovoice/gosip/sip"
)

var _ sip.Uri = (*NoneUri)(nil)
var _ sip.Header = (*Authenticate)(nil)

type NoneUri struct {
}

func (n *NoneUri) Equals(other interface{}) bool {
	return false
}

func (n *NoneUri) String() string {
	return ""
}

func (n *NoneUri) Clone() sip.Uri {
	return &NoneUri{}
}

func (n *NoneUri) IsEncrypted() bool {
	return false
}

func (n *NoneUri) SetEncrypted(flag bool) {
	return
}

func (n *NoneUri) User() sip.MaybeString {
	return nil
}

func (n *NoneUri) SetUser(user sip.MaybeString) {
	return
}

func (n *NoneUri) Password() sip.MaybeString {
	return nil
}

func (n *NoneUri) SetPassword(pass sip.MaybeString) {
	return
}

func (n *NoneUri) Host() string {
	return ""
}

func (n *NoneUri) SetHost(host string) {
	return
}

func (n *NoneUri) Port() *sip.Port {
	return nil
}

func (n *NoneUri) SetPort(port *sip.Port) {
	return
}

func (n *NoneUri) UriParams() sip.Params {
	return nil
}

func (n *NoneUri) SetUriParams(params sip.Params) {
	return
}

func (n *NoneUri) Headers() sip.Params {
	return nil
}

func (n *NoneUri) SetHeaders(params sip.Params) {
	return
}

func (n *NoneUri) IsWildcard() bool {
	return false
}

type Authenticate struct {
	Auth string
}

func (a Authenticate) Name() string {
	return "WWW-Authenticate"
}

func (a Authenticate) Value() string {
	return a.Auth
}

func (a Authenticate) Clone() sip.Header {
	return &Authenticate{Auth: a.Auth}
}

func (a Authenticate) String() string {
	return fmt.Sprintf("WWW-Authenticate: %s", a.Auth)
}

func (a Authenticate) Equals(other interface{}) bool {
	return false
}
