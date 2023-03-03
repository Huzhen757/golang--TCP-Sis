package zinterface

// 消息管理抽象层
type IMsgHandler interface {

	//根据msgId来索引调度路由的方法
	DoMsgHandler(request IRequest)
	//添加路由到map集合中的方法
	AddRouter(msgId uint32, router IRouter)
	// 开启全局工作池
	StartWorkerPool()
	// 将request请求发送给消息队列TaskQueue
	SendMsgToTaskQueue(request IRequest)
}
