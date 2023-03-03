package znet

import (
	"fmt"
	"net"
	"zinx/utils"
	"zinx/zinterface"
)

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

func (s *Server) Start() {
	fmt.Printf("[Server config info] name:%s, ip:%s, port:%d\n", utils.GlobalObject.Name, utils.GlobalObject.Host, utils.GlobalObject.TcpPort)

	// fmt.Printf("[Server] Server Listener at IP:%s, port:%d, is starting\n", s.IP, s.Port)
	go func() {
		// 开启worker pool以及task queue
		s.MsgHandler.StartWorkerPool()

		// 1.获取TCP addr
		addr, err := net.ResolveTCPAddr(s.Version, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("[Server] resolve tcp address error: ", err)
			return
		}
		// 2.监听服务器地址,创建listener
		listener, err := net.ListenTCP(s.Version, addr)
		if err != nil {
			fmt.Println("[Server] listen tcp server address error: ", err)
			return
		}
		fmt.Printf("[Server] start %s server success\n", s.Name)
		// 3.阻塞的等待客户端连接，处理客户端业务
		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				fmt.Println("[Server] Accept TCP error: ", err)
				continue
			}
			// go func() {
			// 	for {
			// 		buf := make([]byte, 1024)
			// 		len, err := conn.Read(buf)
			// 		if err != nil {
			// 			fmt.Println("[Server] receive buffer data to client error: ", err)
			// 			continue
			// 		}
			// 		fmt.Printf("[Server] receive client data: %s, len: %d\n", buf, len)
			// 		if _, err = conn.Write(buf[:len]); err != nil {
			// 			fmt.Println("[Server] write data to client error: ", err)
			// 			continue
			// 		}
			// 	}
			// }()

			// 超过链接数阈值，拒绝链接
			if utils.GlobalObject.MaxConn <= int16(s.ConnMgr.GetConnCount()) {
				fmt.Println("[Server] connection number more than ", utils.GlobalObject.MaxConn)
				conn.Close()
				continue
			}

			var cid uint32
			// 封装为自定义的Connection对象
			dealConn := NewConnection(conn, cid, s.MsgHandler, s)
			cid++
			go dealConn.Start()

		}
	}()
}

func (s *Server) Stop() {
	fmt.Println("[Server] clear all connection, connections len: ", s.ConnMgr.GetConnCount())
	s.ConnMgr.ClearConn()
}

func (s *Server) Serve() {
	s.Start()
	select {} // 阻塞
}

func (s *Server) AddRouter(msgId uint32, router zinterface.IRouter) {
	s.MsgHandler.AddRouter(msgId, router)
}

// 初始化Server模块的方法
func NewServer(name string) zinterface.IServer {
	server := &Server{
		Name:       utils.GlobalObject.Name,
		Version:    "tcp4",
		IP:         utils.GlobalObject.Host,
		Port:       utils.GlobalObject.TcpPort,
		MsgHandler: NewMsgHandler(),
		ConnMgr:    NewConnManager(),
	}
	return server
}

// 获取server中的ConnManager管理器
func (s *Server) GetConnManager() zinterface.IConnManager {
	return s.ConnMgr
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
