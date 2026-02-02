package go_GB28181

// GB28181Handler 服务端回调
type GB28181Handler interface {
	ErrorVerifyHandle(err error) bool                                            //错误校验，如果返回true，就代表服务端出问题，会关闭服务端
	ServerClosedHandle(err error)                                                //服务器已经关闭啦
	ReceivedHandle(clientHost string, clientSIP string, frame string, err error) //收到了消息
	SentHandle(clientHost string, clientSIP string, frame string, err error)     //发送了消息
	PasswordHandle(clientHost string, cameraSIP string) string                   //获取认证密码
	CameraOnHandle(clientHost string, clientSIP string)                          //摄像头上线
	CameraDownHandle(clientHost string, clientSIP string, err error)             //摄像头下线
}
