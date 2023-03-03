package znet

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
		property:   make(map[string]interface{}),
	}

	// 每次初始化一个Connection，将其加入到ConnManager管理器中
	c.TcpServer.GetConnManager().AddConn(c)
	return c
}

func (c *Connection) Start() {
	fmt.Println("[Server] connection is starting..., Connection id: ", c.ConnId)
	// 开启协程读取客户端数据
	go c.StartReader()
	// 开启协程写入数据到客户端
	go c.StartWriter()

	// 根据开发端传递的hook函数执行
	c.TcpServer.CallOnConnStart(c)
}

func (c *Connection) Stop() {
	fmt.Println("[Server] connection is shutting down..., Connection id: ", c.ConnId)
	if c.ConnStatus == true {
		return
	}
	c.ConnStatus = true

	// 关闭链接之前调用hook方法
	c.TcpServer.CallOnConnStop(c)

	c.Conn.Close()     // 关闭链接
	c.ExitChan <- true // 告知writer协程关闭
	// 将当前链接从链接管理器中删除
	c.TcpServer.GetConnManager().RemoveConn(c)
	// 回收资源
	close(c.ExitChan)
	close(c.msgChan)
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
	// 需要写回的数据存入channel，由写协程负责发送给客户端
	c.msgChan <- binaryMsg
	return nil
}

// 发送给客户端的写模块
func (c *Connection) StartWriter() {
	fmt.Println("[Server] writer data goroutine is running...")
	defer fmt.Println("[Server] ", c.RemoteAddr().String(), " writer goroutine is exit")
	// 阻塞等待channel中的数据，一旦channel有数据就发送给client
	for {
		select {
		case data := <-c.msgChan:
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("[Server] send data to client error...", err)
				return
			}
		case <-c.ExitChan:
			// ExitChan有数据,当前链接已经退出了=>写协程退出
			return
		}
	}
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
