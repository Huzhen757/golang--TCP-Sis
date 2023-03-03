package zinterface

type IServer interface {
	// 启动服务器
	Start()
	// 停止服务器
	Stop()
	// 运行服务器
	Serve()

	// 路由功能
	AddRouter(msgId uint32, handler IRouter)

	// 获取ConnectionManager管理器
	GetConnManager() IConnManager

	// 注册OnConnStart hook函数的方法
	SetOnConnStart(func(conn IConnection))
	// 注册OnConnStop hook函数的方法
	SetOnConnStop(func(conn IConnection))
	// 调用OnConnStart hook函数的方法
	CallOnConnStart(conn IConnection)
	// 调用OnConnStart hook函数的方法
	CallOnConnStop(conn IConnection)
}
