package znet

import (
	"fmt"
	"strconv"
	"zinx/utils"
	"zinx/zinterface"
)

type MsgHandler struct {
	// 消息id和对应的router
	Apis map[uint32]zinterface.IRouter

	// 消息队列，一个worker中最大的消息数量
	TaskQueue []chan zinterface.IRequest
	// worker工作池数量 能开辟多少个worker(goroutine)
	WorkerPoolSize uint32
}

// 调度
func (mh *MsgHandler) DoMsgHandler(request zinterface.IRequest) {

	router, ok := mh.Apis[request.GetMsgId()]
	if !ok {
		fmt.Println("[Server] can't find router for %d msg! Need Register a New Router\n", request.GetMsgId())
	}
	router.PreHandle(request)
	router.Handle(request)
	router.PostHandle(request)
}

// 添加新路由
func (mh *MsgHandler) AddRouter(msgId uint32, router zinterface.IRouter) {
	if _, ok := mh.Apis[msgId]; ok {
		panic("[Server] repeat add routers, " + strconv.Itoa(int(msgId)))
	}
	mh.Apis[msgId] = router
	fmt.Printf("[Server] add routers success for %d msg\n", msgId)
}

func NewMsgHandler() *MsgHandler {
	return &MsgHandler{
		Apis:           make(map[uint32]zinterface.IRouter),
		WorkerPoolSize: utils.GlobalObject.WorkerPoolSize,                                   // 能开辟多少个worker(goroutine)
		TaskQueue:      make([]chan zinterface.IRequest, utils.GlobalObject.WorkerPoolSize), // 一个worker中最大的消息数量
	}
}

// 启动唯一的worker工作池: 对外暴露
func (mh *MsgHandler) StartWorkerPool() {
	fmt.Println("[Server] global worker pool is starting...")
	// 根据WorkerPoolSize数量去创建worker，每一个worker对应一个goroutine
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		// 一个消息中能处理的最大用户请求数
		mh.TaskQueue[i] = make(chan zinterface.IRequest, utils.GlobalObject.MaxTaskSize)
		go mh.StartWorker(i, mh.TaskQueue[i])
	}
	fmt.Println("[Server] global worker pool is started...")
}

// 启动一个具体的worker: 对外隐藏
func (mh *MsgHandler) StartWorker(workerId int, taskQueue chan zinterface.IRequest) {
	fmt.Println("[Server] current worker id: ", workerId, " is started...")
	// 阻塞等待着当前worker对应的channel发送来的消息，当接受到消息，worker处理消息对应的业务，调用DoMsgHandler方法
	for {
		select {
		case request := <-taskQueue:
			mh.DoMsgHandler(request)
		}
	}
}

// 将request请求发送给消息队列TaskQueue(一致性hash分配): 根据对应的workerId，存放到TaskQueue对应的channel中
func (mh *MsgHandler) SendMsgToTaskQueue(request zinterface.IRequest) {
	// 找到对应的workId
	workerId := request.GetConn().GetConnId() % mh.WorkerPoolSize
	fmt.Println("[Server] add connection id: ", request.GetConn().GetConnId,
		" msg id: ", request.GetMsgId(),
		" request to worker id: ", workerId)
	// 存到workerId对应的channel中
	mh.TaskQueue[workerId] <- request

}
