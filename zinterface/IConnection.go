package zinterface

import "net"

type IConnection interface {
	// 启动链接
	Start()
	// 停止链接
	Stop()
	// 获取当前链接的conn对象
	GetTcpConnection() *net.TCPConn
	// 得到链接Id
	GetConnId() uint32
	// 获取客户端连接的Tcp状态和端口
	RemoteAddr() net.Addr
	// 提供一个send msg方法，将发送给客户端的消息进行pack
	SendMsg(msgId uint32, data []byte) error

	// 设置Connection属性方法
	SetProperty(key string, val interface{})
	//获取Connection属性方法
	GetProperty(key string) (interface{}, error)
	//删除Connection属性方法
	RemoveProperty(key string)
}

// 处理链接业务的方法类型: 当前链接，处理的数据，数据长度
type HandleFunc func(*net.TCPConn, []byte, int) error
