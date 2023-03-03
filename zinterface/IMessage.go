package zinterface

// 请求的消息封装到Message结构体中，定义Message的getter和setter方法
type IMessage interface {
	GetMsgId() uint32
	GetMsgLen() uint32
	GetMsgContent() []byte

	SetMsgId(uint32)
	SetMsgLen(uint32)
	SetMsgContent([]byte)
}
