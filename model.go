package go_GB28181

import (
	"encoding/xml"
	"errors"
	"strings"
)

type Node struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Content []byte     `xml:",innerxml"`
	Nodes   []Node     `xml:",any"`
}

// findCmdType 递归搜索节点及其子节点，寻找 CmdType 属性或元素
func (n *Node) findCmdType() (string, error) {
	// 首先检查当前节点的属性
	for _, attr := range n.Attrs {
		// 查找名为 "CmdType" 的属性
		if attr.Name.Local == "CmdType" {
			return attr.Value, errors.New("not implement")
		}
	}

	// 如果当前节点本身就叫 <CmdType>，则返回其内部文本内容
	if n.XMLName.Local == "CmdType" {
		return strings.TrimSpace(string(n.Content)), nil
	}

	// 递归检查所有子节点
	for _, childNode := range n.Nodes {
		if value, foundErr := childNode.findCmdType(); foundErr == nil {
			return value, nil
		}
	}
	return "", errors.New("not implement")
}

type digestAuth struct {
	Username  string
	Realm     string
	Nonce     string
	Response  string
	URI       string
	Opaque    string
	Algorithm string
}

// Query 请求目录检索
type Query struct {
	XMLName  xml.Name `xml:"Query"`
	CmdType  string   `xml:"CmdType"`
	SN       string   `xml:"SN"`
	DeviceID string   `xml:"DeviceID"`
}

// CatalogueResponse 目录检索响应
type CatalogueResponse struct {
	XMLName    xml.Name   `xml:"Response"`
	CmdType    string     `xml:"CmdType"`
	SN         int        `xml:"SN"`
	DeviceID   string     `xml:"DeviceID"`
	SumNum     int        `xml:"SumNum"`
	DeviceList DeviceList `xml:"DeviceList"`
}

type ControlResponse struct {
	XMLName  xml.Name `xml:"Response"`
	CmdType  string   `xml:"CmdType"`
	SN       int      `xml:"SN"`
	DeviceID string   `xml:"DeviceID"`
	Result   string   `xml:"Result"`
}

type DeviceList struct {
	Num   int    `xml:"Num,attr"`
	Items []Item `xml:"Item"`
}

type Item struct {
	DeviceID     string `xml:"DeviceID"`
	Name         string `xml:"Name"`
	Manufacturer string `xml:"Manufacturer"`
	Model        string `xml:"Model"`
	Owner        string `xml:"Owner"`
	CivilCode    string `xml:"CivilCode"`
	Address      string `xml:"Address"`
	Parental     int    `xml:"Parental"`
	ParentID     string `xml:"ParentID"`
	RegisterWay  int    `xml:"RegisterWay"`
	Secrecy      int    `xml:"Secrecy"`
	Status       string `xml:"Status"`
}
