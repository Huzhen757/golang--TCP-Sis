package zinterface

// 路由接口IRouter, 路由中的数据都是自定义的IRequest

type IRouter interface {
	//处理业务之前的方法
	PreHandle(request IRequest)
	//业务处理主方法
	Handle(request IRequest)
	//处理业务之后的方法
	PostHandle(request IRequest)
}
