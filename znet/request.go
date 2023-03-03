package znet

import "zinx/zinterface"

type Request struct {
	// 与客户端建立的链接对象
	conn zinterface.IConnection
	// 客户端的请求数据
	msg zinterface.IMessage
}

func (request *Request) GetConn() zinterface.IConnection {
	return request.conn
}

func (request *Request) GetData() []byte {
	return request.msg.GetMsgContent()
}

func (request *Request) GetMsgId() uint32 {
	return request.msg.GetMsgId()
}
