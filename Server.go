package main

import (
	"fmt"
	"zinx/zinterface"
	"zinx/znet"
	"zinxDemo/zinxV1/common"
)

/*
基于Zinx框架开发的服务端应用程序
*/

// 定义创建链接之后的hook函数
func AfterCreateConnTask(conn zinterface.IConnection) {
	fmt.Println("[Server] ========>AfterCreateConnTask is called...")
	if err := conn.SendMsg(101, []byte("Create Connection hook task")); err != nil {
		fmt.Println(err)
	}

	// 给当前Connection添加属性
	conn.SetProperty("author", "huzhen")
	conn.SetProperty("email", "huzhen114514@163.com")
	conn.SetProperty("csdn", "https://blog.csdn.net/m0_45971439?type=blog")
	conn.SetProperty("github", "https://github.com/huzhen757")

}

// 定义断开链接之前的hook函数
func BeforeDestroyConnTask(conn zinterface.IConnection) {
	fmt.Println("[Server] ========>BeforeDestroyConnTask is called...")
	fmt.Println("[Server] connection id: ", conn.GetConnId(), " is destory")

	// 获取当前Connection的属性
	fmt.Println("============> get property...")
	if name, err := conn.GetProperty("author"); err == nil {
		fmt.Println("[Server] author: ", name)
	}
	if email, err := conn.GetProperty("email"); err == nil {
		fmt.Println("[Server] email: ", email)
	}
	if csdn, err := conn.GetProperty("csdn"); err == nil {
		fmt.Println("[Server] csdn: ", csdn)
	}
	if github, err := conn.GetProperty("github"); err == nil {
		fmt.Println("[Server] github: ", github)
	}
}

func main() {
	// 1.创建server句柄
	s := znet.NewServer("[zinx V1]")
	// 2.添加自定义的路由管理模块
	s.AddRouter(0, &MyRouter{})
	s.AddRouter(1, &Zinx5Router{})
	// 3.注册链接需要的Hook函数
	s.SetOnConnStart(AfterCreateConnTask)
	s.SetOnConnStop(BeforeDestroyConnTask)
	// 4.启动server
	s.Serve()

}

// 自定义Router
type MyRouter struct {
	znet.BaseRouter
}

func (this *MyRouter) PreHandle(r zinterface.IRequest) {
	fmt.Println("[Server] Execute pre Handle")

}

func (this *MyRouter) Handle(r zinterface.IRequest) {
	fmt.Println("[Server] MyRouter -> Execute Handle!")
	fmt.Printf("[Server] receive client msg id:%d, msg:%s\n", r.GetMsgId(), r.GetData())
	// 读取客户端的数据，再回写hello, this is server!
	id, _ := common.GetRandomId()
	if err := r.GetConn().SendMsg(uint32(id), []byte("hello, this is server!")); err != nil {
		fmt.Println("[Server] write back to client msg failed...", err)
	}

}

func (this *MyRouter) PostHandle(r zinterface.IRequest) {
	fmt.Println("[Server] Execute post Handle")

}

type Zinx5Router struct {
	znet.BaseRouter
}

func (this *Zinx5Router) Handle(r zinterface.IRequest) {
	fmt.Println("[Server] Zinx5Router -> Execute Handle!")

	fmt.Printf("[Server] receive client msg id:%d, msg:%s\n", r.GetMsgId(), r.GetData())
	// 读取客户端的数据，再回写hello, this is zinx5 router!
	id, _ := common.GetRandomId()
	if err := r.GetConn().SendMsg(uint32(id), []byte("hello, this is zinx5 router!")); err != nil {
		fmt.Println("[Server] write back to client msg failed...", err)
	}

}
