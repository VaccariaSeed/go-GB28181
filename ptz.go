package go_GB28181

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	startChar = 0xa5 //指令首字节
)

// 镜头变倍（Zoom）          云台垂直方向控制（Tilt）    云台水平方向控制（Pan）
// 缩小（OUT） 放大（IN）     上（Up） 下（Down）        左（Left） 右（Right）

type Zoom byte
type Tilt byte

type Pan byte

type Iris byte

type Focus byte

const (
	ZoomOut  Zoom = 2 //镜头变倍，缩小
	ZoomIn   Zoom = 1 //镜头变倍，方法
	ZoomIdle Zoom = 0 //镜头变倍，静置

	TiltUp   Tilt = 2 //云台垂直方向控制，上
	TiltDown Tilt = 1 //云台垂直方向控制，下
	TiltIdle Tilt = 0 //云台垂直方向控制，静置

	PanLeft  Pan = 2 //云台水平方向控制，左
	PanRight Pan = 1 //云台水平方向控制，右
	PanIdle  Pan = 0 //云台水平方向控制，静置

	IrisLittle Iris = 2 //光圈缩小
	IrisBig    Iris = 1 //光圈方法
	IrisIdle   Iris = 0 //光圈不动

	FocusLittle Focus = 2 //聚焦近
	FocusBig    Focus = 1 //聚焦远
	FocusIdle   Focus = 0 //聚焦不动
)

// PTZ 云台控制
type PTZ struct {
	Version byte   //版本信息，不能大于0x0F
	Address uint16 //地址
	data    uint16 //数据
}

// BuildPTZ 构建一个控制镜头变倍、垂直方向控制、水平方向控制的PTZ
func (p *PTZ) BuildPTZ(zoom Zoom, zoomValue byte, tilt Tilt, tiltValue byte, pan Pan, panValue byte) (string, error) {
	// 检查zoom
	if zoom != ZoomOut && zoom != ZoomIn && zoom != ZoomIdle {
		return "", errors.New("invalid zoom value")
	}
	// 检查tilt
	if tilt != TiltUp && tilt != TiltDown && tilt != TiltIdle {
		return "", errors.New("invalid tilt value")
	}
	// 检查pan
	if pan != PanLeft && pan != PanRight && pan != PanIdle {
		return "", errors.New("invalid pan value")
	}
	//ztp
	cmdStr := "00" + fmt.Sprintf("%02b", zoom) + fmt.Sprintf("%02b", tilt) + fmt.Sprintf("%02b", pan)
	cmd, err := strconv.ParseUint(cmdStr, 2, 8)
	if err != nil {
		return "", err
	}
	arr := []byte{startChar, p.comCode1(), byte(p.Address & 0xFF), byte(cmd), panValue, tiltValue, p.byte7(zoomValue)}
	var sum uint16 = 0
	// 遍历所有字节并求和
	for _, b := range arr {
		sum += uint16(b)
	}
	// 取低8位，即对256取模
	arr = append(arr, byte(sum%256))
	return strings.ToLower(hex.EncodeToString(arr)), nil
}

// BuildFI 聚焦，光圈
func (p *PTZ) BuildFI(iris Iris, irisValue byte, focus Focus, focusValue byte) (string, error) {
	if iris != IrisLittle && iris != IrisBig && iris != IrisIdle {
		return "", errors.New("invalid iris value")
	}
	// 检查tilt
	if focus != FocusLittle && focus != FocusBig && focus != FocusIdle {
		return "", errors.New("invalid focus value")
	}
	cmdStr := "0100" + fmt.Sprintf("%02b", iris) + fmt.Sprintf("%02b", focus)
	cmd, err := strconv.ParseUint(cmdStr, 2, 8)
	if err != nil {
		return "", err
	}
	arr := []byte{startChar, p.comCode1(), byte(p.Address & 0xFF), byte(cmd), focusValue, irisValue, p.byte7(0)}
	var sum uint16 = 0
	// 遍历所有字节并求和
	for _, b := range arr {
		sum += uint16(b)
	}
	// 取低8位，即对256取模
	arr = append(arr, byte(sum%256))
	return strings.ToLower(hex.EncodeToString(arr)), nil
}

func (p *PTZ) byte7(zoomValue byte) byte {
	// 获取高4位并右移4位
	high4Bits := p.Address >> 4
	// 获取 byteLow 的低4位
	low4Bits := zoomValue & 0x0F
	// 将高4位左移4位后与低4位组合
	return byte(high4Bits<<4) | low4Bits
}

// 获取组合码1
func (p *PTZ) comCode1() byte {
	byte1High := startChar >> 4  // 0xA
	byte1Low := startChar & 0x0F // 0x5
	checkSum := (byte1High + byte1Low + int(p.Version)) % 16
	return byte((int(p.Version) << 4) | checkSum)
}
