package zinterface

type IConnManager interface {

	//添加链接方法
	AddConn(conn IConnection)
	//查找链接方法
	GetConnById(cId uint32) (IConnection, error)
	//删除链接方法
	RemoveConn(conn IConnection)
	//获取总链接数量方法
	GetConnCount() int
	//销毁全部链接方法
	ClearConn()
}
