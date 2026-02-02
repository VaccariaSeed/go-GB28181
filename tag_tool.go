package go_GB28181

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"net"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
)

func gb2312XMLDecoder(xmlData []byte, v interface{}) error {
	xmlStr := string(xmlData)
	if strings.Contains(xmlStr, `encoding="GB2312"`) {
		xmlStr = strings.Replace(xmlStr, `encoding="GB2312"`, "", 1)
	}
	if strings.Contains(xmlStr, `encoding='GB2312'`) {
		xmlStr = strings.Replace(xmlStr, `encoding='GB2312'`, "", 1)
	}

	// 创建GBK解码器
	decoder := simplifiedchinese.GBK.NewDecoder()

	// 将GB2312字节转换为UTF-8读取器
	utf8Reader := decoder.Reader(strings.NewReader(xmlStr))

	// 创建XML解码器
	xmlDecoder := xml.NewDecoder(utf8Reader)

	// 解析XML
	return xmlDecoder.Decode(v)
}

func backPvr(back *PlayBack, serverSIP, serverIP, cameraSIP string) string {
	str := "v=0\no=%s 0 0 IN IP4 %s\ns=Playback\nu=%s:%d\nc=IN IP4 %s\nt=%d %d\nm=video %d RTP/AVP 96 98 97\na=recvonly\na=rtpmap:96 PS/90000\na=rtpmap:98 H264/90000\na=rtpmap:97 MPEG4/90000\ny=%s"
	return fmt.Sprintf(str, serverSIP, serverIP, cameraSIP, back.ChanNumber, back.VideoServerIp, back.StartTime, back.EndTime, back.videoServerPort, back.TargetNumber)
}

func pvr(ip, port, serverSIP string) string {
	//fmt.Println(`备注:
	//v=SDP协议版本号，总是0
	//o=用户名/会话ID 会话版本 增量版本 网络类型(Internet) 地址类型(IPv4) 主叫方希望接收媒体流的IP地址
	//s=会话名称。这里简单命名为“Play”
	//c=网络类型(Internet) 地址类型(IPv4) 主叫方希望接收媒体流的IP地址
	//t=0 0 会话时间描述。0 0表示会话没有时间限制，是永久的，或者说是“随时可用”的
	//m=video(媒体类型是视频) 希望对方将视频流发送到的UDP端口号 RTP/AVP 96 98 97
	//a=recvonly
	//a=rtpmap:96 PS/90000
	//a=rtpmap:98 H264/90000
	//a=rtpmap:97 MPEG4/90000
	//y=0100000001
	//f=
	//`)
	str := "v=0\no=%s 0 0 IN IP4 %s\ns=Play\nc=IN IP4 %s\nt=0 0\nm=video %s RTP/AVP 96 98 97\na=recvonly\na=rtpmap:96 PS/90000\na=rtpmap:98 H264/90000\na=rtpmap:97 MPEG4/90000\ny=0100000001\nf="
	return fmt.Sprintf(str, serverSIP, ip, ip, port)
}

func modelToXml(msg interface{}) (string, error) {
	output, err := xml.MarshalIndent(msg, "", "  ")
	if err != nil {
		return "", err
	}
	xmlData := `<?xml version="1.0"?>` + "\n" + string(output)
	return xmlData, nil

}

func createUDPClient(host string) (*net.UDPAddr, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", host)
	if err != nil {
		return nil, err
	}
	return udpAddr, nil
}

func parseWithXmlData(body string) (string, error) {
	if strings.Contains(body, `encoding="GB2312"`) {
		body = strings.Replace(body, `encoding="GB2312"`, "", 1)
	}
	if strings.Contains(body, `encoding='GB2312'`) {
		body = strings.Replace(body, `encoding='GB2312'`, "", 1)
	}
	var root Node
	err := xml.Unmarshal([]byte(body), &root)
	if err != nil {
		return "", fmt.Errorf("XML 解析失败: %v", err)
	}
	// 开始递归搜索
	if cmdTypeValue, findErr := root.findCmdType(); findErr == nil {
		return cmdTypeValue, nil
	} else {
		return "", findErr
	}

}

func xmlToModel(body string, target interface{}) error {
	if strings.Contains(body, `encoding="GB2312"`) {
		body = strings.Replace(body, `encoding="GB2312"`, "", 1)
	}
	if strings.Contains(body, `encoding='GB2312'`) {
		body = strings.Replace(body, `encoding='GB2312'`, "", 1)
	}
	return xml.Unmarshal([]byte(body), target)
}

func md5Encode(key string) string {
	hash := md5.Sum([]byte(key))
	keyMd5Str := hex.EncodeToString(hash[:])
	return keyMd5Str
}

func hashWithSalt(password, salt string) string {
	data := []byte(password + salt)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func generateSIPBranchID(size int) string {
	randomBytes := make([]byte, size)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().Unix())[:size]
	}
	randomSuffix := hex.EncodeToString(randomBytes)
	return randomSuffix
}
