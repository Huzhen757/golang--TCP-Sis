package zinterface

type IRequest interface {

	// 获取当前链接
	GetConn() IConnection
	// 获取当前数据
	GetData() []byte

	GetMsgId() uint32
}
