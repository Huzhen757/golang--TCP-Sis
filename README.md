# golang-TCP-Server
# 基于golang实现的轻量级TCP服务器框架
原作者github地址：https://github.com/aceld/zinx/
## 整体框架
### 1.zinterface文件夹 -> 各个模块的抽象层接口
### 2. znet文件夹 -> 每个抽象层接口对应的实现：结构体(实现类)定义，成员方法，初始化方法等
### 3. utils文件夹 -> 公共组件
### 4. Client.go 基于框架实现的客户端应用程序
### 5. Server.go 基于框架实现的服务端应用程序
#### 6. zinx.json 主配置文件

# golang--TCP-Zinx学习
# 基于golang实现的轻量级TCP服务器框架

原作者：https://github.com/aceld/zinx/

# 1. 基础Server

+ Server接口方法
  + 停止服务器Stop：做一些资源的回收和状态的回执
  + 启动服务器Start：获取TCP addr；监听服务器地址,创建listener；阻塞的等待客户端连接，处理客户端业务
  + 运行服务器Server：调用Start方法，调用之后做阻塞处理，期间可做一些扩展的功能
  + 初始化server

~~~go
type IServer interface {
	// 启动服务器
	Start()
	// 停止服务器
	Stop()
	// 运行服务器
	Serve()
}
~~~

+ Server结构体属性，Server接口的方法实现
  + server名
  + 监听的ip
  + 监听的端口号

# 2. 简单的链接封装和业务绑定

+ Connection接口方法
  + 启动链接Start
  + 停止链接Stop
  + 获取当前链接的conn对象（套接字）
  + 得到链接Id
  + 获取客户端连接的地址和端口
  + 发送数据方法Send
  + Connection所绑定的业务处理函数类型

~~~go
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
	// 发送数据给客户端
	Send(data []byte) error
}

// 处理链接业务的方法类型: 当前链接，处理的数据，数据长度
type HandleFunc func(*net.TCPConn, []byte, int) error
~~~

+ Connection结构体属性
  + socket TCP套接字
  + 链接的id
  + 当前链接状态（是否已经关闭）
  + 与当前链接所绑定的业务处理方法
  + 需要将待退出的链接写入到exit channel

~~~go
package znet

import (
	"fmt"
	"net"
	"zinx/zinterface"
)

type Connection struct {
	//socket TCP套接字
	Conn *net.TCPConn
	//链接的id
	ConnId uint32
	//当前链接状态（是否已经关闭）=> false开启 true关闭
	ConnStatus bool
	//与当前链接所绑定的业务处理方法
	Handler zinterface.HandleFunc
	//需要将待退出的链接写入到exit channel
	ExitChan chan bool
}

// 初始化Connection实例
func NewConnection(conn *net.TCPConn, connId uint32, handler zinterface.HandleFunc) *Connection {
	return &Connection{
		Conn:       conn,
		ConnId:     connId,
		ConnStatus: false,
		Handler:    handler,
		ExitChan:   make(chan bool, 1),
	}
}

func (c *Connection) Start() {
	fmt.Println("[Server] connection is starting..., Connection id: ", c.ConnId)
	// 开启协程读取客户端数据
	go c.StartRead()
	// 开启协程写入数据到客户端

}

func (c *Connection) Stop() {
	fmt.Println("[Server] connection is shutting down..., Connection id: ", c.ConnId)
	if c.ConnStatus == true {
		return
	}
	c.ConnStatus = true

	c.Conn.Close()
	close(c.ExitChan)

}

func (c *Connection) GetTcpConnection() *net.TCPConn {
	return c.Conn
}

func (c *Connection) GetConnId() uint32 {
	return c.ConnId
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c *Connection) Send(data []byte) error {
	return nil
}

func (c *Connection) StartRead() {
	fmt.Println("[Server] reader data goroutine is running...")
	defer fmt.Println("[Server] reader data goroutine is exit, client address: ", c.RemoteAddr().String())
	defer c.Stop()
	// 循环读取客户端的数据
	for {
		buf := make([]byte, 1024)
		len, err := c.Conn.Read(buf)
		if err != nil {
			fmt.Println("[Server] read buffer data failed...", err)
			continue
		}
		if err = c.Handler(c.Conn, buf, len); err != nil {
			fmt.Println("[Server] process data of client failed...", err)
			break
		}
	}
}
~~~

# 3. 基础router模块

## a. Request请求封装

+ request接口方法
  + 获取自定义封装的IConnection
  + 获取客户端请求数据

~~~go
type IRequest interface {

	// 获取当前链接
	GetConn() IConnection
	// 获取当前数据
	GetData() []byte
}
~~~

+ request结构体属性
  +  客户端建立的链接对象
  +  客户端的请求数据

~~~go
import "zinx/zinterface"

type Request struct {
	// 与客户端建立的链接对象
	conn zinterface.IConnection
	// 客户端的请求数据
	data []byte
}

func (request *Request) GetConn() zinterface.IConnection {
	return request.conn
}

func (request *Request) GetData() []byte {
	return request.data
}
~~~

## b. Router模块

+ router接口方法
  + 处理业务之前的方法
  + 业务处理主方法
  + 处理业务之后的方法
+ baseRouter实现类
  + 处理业务之前的方法
  + 业务处理主方法
  + 处理业务之后的方法

~~~go
// 路由接口IRouter, 路由中的数据都是自定义的IRequest

type IRouter interface {
	//处理业务之前的方法
	PreHandle(request IRequest)
	//业务处理主方法
	Hnadle(request IRequest)
	//处理业务之后的方法
	PostHandle(request IRequest)
}
~~~

~~~go
import "zinx/zinterface"

// router应该由用户自定义实现，可以先嵌入BaseRouter基类，然后进行方法的重写
// BaseRouter其实是一种模板方法的体现，方法具体实现留给子类进行扩展
type BaseRouter struct {
}

//处理业务之前的方法
func (br *BaseRouter) PreHandle(request zinterface.IRequest) {}

//业务处理主方法
func (br *BaseRouter) Hnadle(request zinterface.IRequest) {}

//处理业务之后的方法
func (br *BaseRouter) PostHandle(request zinterface.IRequest) {}
~~~

## c. zinx集成router模块

+ Server接口添加路由功能
+ Server结构体增加router成员属性 => 
+ Connection结构体绑定一个router成员 => 替换之前的`Handler zinterface.HandleFunc`
+ 在Connection调用已注册的router处理业务

**Server服务器端**

+ 创建一个Server句柄
+ 给当前server句柄增加一个自定义router：`s.AddRouter(&MyRouter{})`，其中自定义MyRouter需要继承BaseRouter，重写其方法
+ 启动server

# 4. 全局配置

+ 服务器应用文件夹下/conf/zinx.json （用户自定义）
+ 创建一个zinx的全局配置模块utils/globalObj.go，读取用户配置的zinx.json文件，封装成globalObj对象
+ 将zinx框架中的硬编码参数用globalObj对象的属性进行填充

~~~go
package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"zinx/zinterface"
)

// 存储zinx框架的全局参数，部分参数由用户自行配置zinx.json

var GlobalObject *GlobalObj

type GlobalObj struct {
	TcpServer zinterface.IServer

	Host    string // 监听ip
	TcpPort uint32 // 端口号
	Name    string // 服务器名
	Version string // zinx版本号
	MaxConn int16  // server的最大连接数

	MaxTransDataSize uint32 // 数据传输包的最大值(字节)

}

// 初始化GlobalObject对象,后续可通过用户配置的zinx.json更新部分参数
func init() {
	GlobalObject = &GlobalObj{
		Name:             "Zinx-TcpServer-App",
		Version:          "V2",
		TcpPort:          8999,
		Host:             "0.0.0.0",
		MaxConn:          1000,
		MaxTransDataSize: 4096,
	}

	GlobalObject.LoadArg()
}

// 加载zinx.json配置文件
func (g *GlobalObj) LoadArg() {
	data, err := ioutil.ReadFile("conf\\zinx.json")
	if err != nil {
		fmt.Println("read configuration file failed...", err)
		return
	}
	err = json.Unmarshal(data, &GlobalObject)
	if err != nil {
		fmt.Println("read configuration file failed...", err)
		panic(err)
	}
}
~~~

~~~json
{
    "Name":"Zinx-ServerApp-V4",
    "Version":"tcp4",
    "Host":"0.0.0.0",
    "TcpPort":8999,
    "MaxConn":256,
    "MaxTransDataSize": 4096
}
~~~



# 5. Message消息模块

## 消息Message结构体的属性以及getter，setter方法

+ 消息id
+ 消息长度
+ 消息内容

~~~go
type IMessage interface {
	GetMsgId() uint32
	GetMsgLen() uint32
	GetMsgContent() []byte

	SetMsgId(uint32)
	SetMsgLen(uint32)
	SetMsgContent([]byte)
}
///////////////////////////////////////////////////////////////
type Message struct {
	MsgId      uint32
	MsgLen     uint32
	MsgContent []byte
}

func (m *Message) GetMsgId() uint32 {
	return m.MsgId
}

func (m *Message) GetMsgLen() uint32 {
	return m.MsgLen
}

func (m *Message) GetMsgContent() []byte {
	return m.MsgContent
}

func (m *Message) SetMsgId(msgId uint32) {
	m.MsgId = msgId
}
func (m *Message) SetMsgLen(msgLen uint32) {
	m.MsgLen = msgLen
}
func (m *Message) SetMsgContent(content []byte) {
	m.MsgContent = content
}
~~~

## 消息的**TLV**（type-len-value）格式

+ 消息头：消息长度(4B) + 消息类型(4B)（消息头固定8字节）
+ 消息主体

+ 消息传输模块
  + 将Message封装成（序列化）TLV格式
  + 接收到TLV消息反序列化成Message对象：读取固定长度消息头，根据消息长度再向后读取消息主体

## 消息封装机制集成到zinx框架

+ 修改request的data属性为Message
+ 修改Connection链接读取数据的机制，之前的简单读取byte修改成unpack形式的读取
+ 给Connection提供一个发包机制：将发送的消息pack之后再发送

~~~go
// request.go
type Request struct {
	// 与客户端建立的链接对象
	conn zinterface.IConnection
	// 客户端的请求数据
	// data []byte
	msg zinterface.IMessage
}

// message.go 
// 创建消息实例的方法
func NewMsg(id uint32, data []byte) *Message {
	return &Message{id, uint32(len(data)), data}
}
~~~



~~~go
// connection.go
func (c *Connection) StartRead() {
	fmt.Println("[Server] reader data goroutine is running...")
	defer fmt.Println("[Server] reader data goroutine is exit, client address: ", c.RemoteAddr().String())
	defer c.Stop()
	// 循环读取客户端的数据
	for {
		// 直接读取msg
		// buf := make([]byte, utils.GlobalObject.MaxTransDataSize)
		// _, err := c.Conn.Read(buf)
		// if err != nil {
		// 	fmt.Println("[Server] read buffer data failed...", err)
		// 	continue
		// }
		// if err = c.Handler(c.Conn, buf, len); err != nil {
		// 	fmt.Println("[Server] process data of client failed...", err)
		// 	break
		// }

		// 创建一个dataPack对象:读取客户端的Msg head(8B)=>msg id和msg len
		// 再次读取msglen长度的数据存放在msg.MsgContent中
		dp := &DataPack{}
		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTcpConnection(), headData); err != nil {
			fmt.Println("[Server] read msg head failed...", err)
			break
		}
		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("[Server] unpack msg head failed...", err)
			break
		}
		var data []byte
		if msg.GetMsgLen() > 0 {
			data = make([]byte, msg.GetMsgLen())
			if _, err = io.ReadFull(c.GetTcpConnection(), data); err != nil {
				fmt.Println("[Server] read msg content failed...", err)
				break
			}
		}
		msg.SetMsgContent(data)

		// 调用router路由方法
		request := Request{c, msg}
		go func(request zinterface.IRequest) {
			c.Router.PreHandle(request)
			c.Router.Handle(request)
			c.Router.PostHandle(request)
		}(&request)

	}
}

// 提供一个send msg方法，将消息进行pack再发送给客户端
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.ConnStatus == true {
		return errors.New("[Server] connection closed when send msg!")
	}

	dp := &DataPack{}
	binaryMsg, err := dp.Pack(NewMsg(msgId, data))
	if err != nil {
		fmt.Println("[Server] msg pack failed...", err)
		return errors.New("[Server] msg pack error")
	}
	if _, err = c.Conn.Write(binaryMsg); err != nil {
		fmt.Println("[Server] connection write msg failed...", err)
		return errors.New("[Server] connection write msg error")
	}

	return nil
}
~~~



# 6. 多路由模式

## 消息管理模块设计（支持多路由业务api调度管理）

+ 消息id和对应的router对应（map结构）
+ 根据msgId来索引调度路由的方法
+ 添加路由到map集合中的方法

~~~go
// 消息管理抽象层
type IMsgHandler interface {

	//根据msgId来索引调度路由的方法
	DoMsgHandler(request IRequest)
	//添加路由到map集合中的方法
	AddRouter(msgId uint32, router IRouter)
	
}
~~~

~~~go
import (
	"fmt"
	"strconv"
	"zinx/zinterface"
)

type MsgHandler struct {
	// 消息id和对应的router
	Apis map[uint32]zinterface.IRouter
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
	return &MsgHandler{make(map[uint32]zinterface.IRouter)}
}
~~~

## 消息管理模块

+ 修改server结构体中的Router属性，修改初始化方法
+ 修改connection结构体中的Router属性，修改初始化方法
+ 修改connection中调度router的业务方法

# 7. 读写协程分离

+ 添加一个reader和writer之间通信的channel
+ 添加一个writer goroutine
+ reader由之前直接发送给客户端方式修改为：发送给通信channel
+ 启动reader和writer同时工作

# 8. 消息队列以及多任务

+ 创建一个消息队列：在`MsgHandler`结构体中增加两个属性

~~~go
type MsgHandler struct {
	// 消息id和对应的router
	Apis map[uint32]zinterface.IRouter
	// 下面两个属性均可以在配置文件中配置
	// 消息队列：channel切片，每一个切片对应一个具体的用户请求
	TaskQueue []chan zinterface.IRequest
	// worker工作池数量
	WorkerPoolSize uint32
}

func NewMsgHandler() *MsgHandler {
	return &MsgHandler{
		Apis:           make(map[uint32]zinterface.IRouter),
		WorkerPoolSize: utils.GlobalObject.WorkerPoolSize, 
		TaskQueue:      make([]chan zinterface.IRequest, utils.GlobalObject.MaxWorkerPoolSize),
	}
}
~~~

+ 创建多任务worker的工作池并且启动
  + 根据WorkerPoolSize数量去创建worker
  + 每个worker都需要由一个goroutine去负责，阻塞等待着当前worker对应的channel发送来的消息，当接受到消息，worker处理消息对应的业务，调用`DoMsgHandler`方法
+ server端不再直接发送消息给client，而是交给消息队列和worker工作池处理

~~~go
// 启动唯一的worker工作池: 对外暴露
func (mh *MsgHandler) StartWorkerPool() {
	// 根据WorkerPoolSize数量去创建worker，每一个worker对应一个goroutine
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		// 一个消息中能处理的最大用户请求数
		mh.TaskQueue[i] = make(chan zinterface.IRequest, utils.GlobalObject.MaxTaskSize)
		go mh.StartWorker(i, mh.TaskQueue[i])
	}
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
~~~

+ 在server端开启消息队列和work pool，将client的request请求发送给当前的worker pool处理

~~~go
// server.go
func (s *Server) Start() {
	fmt.Printf("[Server config info] name:%s, ip:%s, port:%d\n", utils.GlobalObject.Name, utils.GlobalObject.Host, utils.GlobalObject.TcpPort)

	// fmt.Printf("[Server] Server Listener at IP:%s, port:%d, is starting\n", s.IP, s.Port)
	go func() {
		// 开启worker pool以及task queue
		s.MsgHandler.StartWorkerPool()
		
		// 1.获取TCP addr...........
    }
}

// connection.go$StartReader()
func (c *Connection) StartReader() {
	fmt.Println("[Server] reader data goroutine is running...")
	defer fmt.Println("[Server] ", c.RemoteAddr().String(), " reader goroutine is exit")
	defer c.Stop()
	// 循环读取客户端的数据
	for {
		// 创建一个dataPack对象:读取客户端的Msg head(8B)=>msg id和msg len
		// 再次读取msglen长度的数据存放在msg.MsgContent中
		dp := &DataPack{}
		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.Conn, headData); err != nil {
			fmt.Println("[Server] read msg head failed...", err)
			break
		}

		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("[Server] unpack msg head failed...", err)
			break
		}
		var data []byte
		if msg.GetMsgLen() > 0 {
			data = make([]byte, msg.GetMsgLen())
			if _, err = io.ReadFull(c.Conn, data); err != nil {
				fmt.Println("[Server] read msg content failed...", err)
				break
			}
		}
		msg.SetMsgContent(data)

		// 调用router路由方法
		request := Request{c, msg}

		if utils.GlobalObject.WorkerPoolSize > 0 {
			// worker pool已经开启,将消息发送给task queue，交给work pool处理
			c.MsgHandler.SendMsgToTaskQueue(&request)
		} else {
			// 从路由中找到注册绑定的conn对应msgHandler(router集合)调用
			go c.MsgHandler.DoMsgHandler(&request)
		}
	}
}    
~~~

# 9. 链接管理

给框架增加链接个数的限定，如果超过一定数量的client链接个数，为保护后端的及时响应，需要拒绝链接请求

## 创建一个链接管理模块

+ 已创建的Connection集合（map）
+ 解决map访问的同步问题：读写锁
+ 添加链接方法
+ 查找链接方法
+ 删除链接方法
+ 获取总链接数量方法
+ 销毁全部链接方法

~~~go
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
~~~

~~~go
package znet

import (
	"errors"
	"fmt"
	"sync"
	"zinx/zinterface"
)

type ConnManager struct {
	ConnMap map[uint32]zinterface.IConnection

	ConnLock sync.RWMutex
}

func NewConnManager() *ConnManager {
	return &ConnManager{
		ConnMap: make(map[uint32]zinterface.IConnection),
	}
}

// 添加链接方法
func (cm *ConnManager) AddConn(conn zinterface.IConnection) {
	// 使用写锁同步
	cm.ConnLock.Lock()
	defer cm.ConnLock.Unlock()

	// 链接添加到map集合中
	cm.ConnMap[conn.GetConnId()] = conn
	fmt.Println("[Server] add connectionto:: ", conn.GetConnId(), " to ConnManager success.")

}

// 查找链接方法
func (cm *ConnManager) GetConnById(cId uint32) (zinterface.IConnection, error) {
	// 使用读锁同步
	cm.ConnLock.RLock()
	defer cm.ConnLock.RUnlock()

	if conn, ok := cm.ConnMap[cId]; ok {
		return conn, nil
	} else {
		return nil, errors.New("[Server] current connection not find.")
	}
}

// 删除链接方法
func (cm *ConnManager) RemoveConn(conn zinterface.IConnection) {
	// 使用写锁同步
	cm.ConnLock.Lock()
	defer cm.ConnLock.Unlock()

	delete(cm.ConnMap, conn.GetConnId())
	fmt.Println("[Server] remove connectionto:: ", conn.GetConnId(), " from ConnManager success.")
}

// 获取总链接数量方法
func (cm *ConnManager) GetConnCount() int {
	return len(cm.ConnMap)
}

// 销毁全部链接方法
func (cm *ConnManager) ClearConn() {
	// 使用写锁同步
	cm.ConnLock.Lock()
	defer cm.ConnLock.Unlock()

	for cid, conn := range cm.ConnMap {
		conn.Stop()
		delete(cm.ConnMap, cid)
	}

	fmt.Println("[Server] clear all connection success. connections len: ", cm.GetConnCount())
}
~~~

## 链接管理模块集成到zinx框架

+ 将ConnManger加入到Server模块中，在server初始化方法中对该属性初始化

~~~go
type Server struct {
	Name    string // 服务名
	Version string // 版本
	IP      string // ip
	Port    uint32 // 端口号

	// 当前Server的router => 链接处理业务
	// Router zinterface.IRouter
	// 当前server的消息管理模块
	MsgHandler zinterface.IMsgHandler

	// 添加链接管理器
	ConnMgr zinterface.IConnManager
}

// 获取server中的ConnManager管理器
func (s *Server) GetConnManager() zinterface.IConnManager {
	return s.ConnMgr
}
~~~

+ 每次与client成功链接后，判断当前连接数是否超过阈值，再添加链接
+ 每次与client断开链接后，删除链接

~~~go
type Connection struct {
	//socket TCP套接字
	Conn *net.TCPConn
	//链接的id
	ConnId uint32
	//当前链接状态（是否已经关闭）=> false开启 true关闭
	ConnStatus bool
	//与当前链接所绑定的业务处理方法
	// Handler zinterface.HandleFunc
	//需要将待退出的链接写入到exit channel
	ExitChan chan bool
	// router
	MsgHandler zinterface.IMsgHandler

	// 无缓冲channel 存放需要发送的数据，由写协程发送给client
	msgChan chan []byte

	// 当前conn属于哪个server
	TcpServer zinterface.IServer
}

// 初始化Connection实例
func NewConnection(conn *net.TCPConn, connId uint32 /*handler zinterface.HandleFunc,*/, handler zinterface.IMsgHandler, server zinterface.IServer) *Connection {
	c := &Connection{
		Conn:       conn,
		ConnId:     connId,
		ConnStatus: false,
		// Handler:    handler,
		ExitChan:   make(chan bool, 1),
		MsgHandler: handler,
		msgChan:    make(chan []byte),
		TcpServer:  server,
	}

	// 每次初始化一个Connection，将其加入到ConnManager管理器中
	c.TcpServer.GetConnManager().AddConn(c)
	return c
}

func (c *Connection) Stop() {
	fmt.Println("[Server] connection is shutting down..., Connection id: ", c.ConnId)
	if c.ConnStatus == true {
		return
	}
	c.ConnStatus = true

	c.Conn.Close()     // 关闭链接
	c.ExitChan <- true // 告知writer协程关闭
	// 将当前链接从链接管理器中删除
	c.TcpServer.GetConnManager().RemoveConn(c)
	// 回收资源
	close(c.ExitChan)
	close(c.msgChan)
}
~~~

## 提供创建链接之后/销毁之前需要处理的业务

+ 提供给用户能够注册的Hook函数
  + 属性：server创建链接之后自动调用Hook函数，server销毁链接之前自动调用Hook函数
  + 注册OnConnStart方法
  + 注册OnConnStop方法

~~~go
type Server struct {
	Name    string // 服务名
	Version string // 版本
	IP      string // ip
	Port    uint32 // 端口号

	// 当前Server的router => 链接处理业务
	// Router zinterface.IRouter
	// 当前server的消息管理模块
	MsgHandler zinterface.IMsgHandler

	// 添加链接管理器
	ConnMgr zinterface.IConnManager

	// Connection创建之后调用OnConnStart方法
	OnConnStart func(conn zinterface.IConnection)
	// Connection销毁之前调用OnConnStop方法
	OnConnStop func(conn zinterface.IConnection)
}

// 注册OnConnStart hook函数的方法
func (s *Server) SetOnConnStart(startFunc func(conn zinterface.IConnection)) {
	s.OnConnStart = startFunc
}

// 注册OnConnStop hook函数的方法
func (s *Server) SetOnConnStop(stopFunc func(conn zinterface.IConnection)) {
	s.OnConnStop = stopFunc
}

// 调用OnConnStart hook函数的方法
func (s *Server) CallOnConnStart(conn zinterface.IConnection) {
	if s.OnConnStart != nil {
		fmt.Println("[Server] ===>call start hook method")
		s.OnConnStart(conn)
	}
}

// 调用OnConnStart hook函数的方法
func (s *Server) CallOnConnStop(conn zinterface.IConnection) {
	if s.OnConnStop != nil {
		fmt.Println("[Server] ===>call stop hook method")
		s.OnConnStop(conn)
	}
}
~~~

~~~go
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
~~~

+ 自定义hook方法，集成到zinx框架中

~~~go
// 定义创建链接之后的hook函数
func AfterCreateConnTask(conn zinterface.IConnection) {
	fmt.Println("[Server] ========>AfterCreateConnTask is called...")
	if err := conn.SendMsg(101, []byte("Create Connection hook task")); err != nil {
		fmt.Println(err)
	}

}

// 定义断开链接之前的hook函数
func BeforeDestroyConnTask(conn zinterface.IConnection) {
	fmt.Println("[Server] ========>BeforeDestroyConnTask is called...")
	fmt.Println("[Server] connection id: ", conn.GetConnId(), " is destory")
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
~~~



# 10. Connection属性配置

Connection结构体新增属性和方法

+ Connection属性集合（map）当前Connection有哪些属性：k-v键值对
+ 保护Connection属性的锁
+ 设置Connection属性方法
+ 获取Connection属性方法
+ 删除Connection属性方法

~~~go
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
~~~

~~~go
import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"zinx/utils"
	"zinx/zinterface"
)

type Connection struct {
	//socket TCP套接字
	Conn *net.TCPConn
	//链接的id
	ConnId uint32
	//当前链接状态（是否已经关闭）=> false开启 true关闭
	ConnStatus bool
	//与当前链接所绑定的业务处理方法
	// Handler zinterface.HandleFunc
	//需要将待退出的链接写入到exit channel
	ExitChan chan bool
	// router
	MsgHandler zinterface.IMsgHandler

	// 无缓冲channel 存放需要发送的数据，由写协程发送给client
	msgChan chan []byte

	// 当前conn属于哪个server
	TcpServer zinterface.IServer

	// 链接属性集合(map)=> 当前Connection有哪些属性: k-v键值对
	property map[string]interface{}
	// 保护链接属性的锁
	propertyLock sync.RWMutex
}
// 添加Connection的属性
func (c *Connection) SetProperty(key string, val interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.property[key] = val
}

// 获取Connection的属性
func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	if val, ok := c.property[key]; ok {
		return val, nil
	} else {
		return nil, errors.New("[Server] can't find property")
	}
}

// 删除Connection的属性
func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}
~~~


