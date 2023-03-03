package main

import (
	"fmt"
	"io"
	"net"
	"time"
	"zinx/znet"
)

/*
客户端
*/

func main() {
	fmt.Println("[Client] client start...")

	time.Sleep(time.Second * 1)

	// conn, err := net.Dial("tcp",
	// 	fmt.Sprintf("%s:%s", utils.GlobalObject.Host,
	// 		utils.GlobalObject.TcpPort))
	conn, err := net.Dial("tcp", "127.0.0.1:8999")

	if err != nil {
		fmt.Println("[Client] client start error, exit: ", err)
		return
	}

	for {
		// 1.数据进行pack再发给Server
		dp := &znet.DataPack{}
		// id, _ := common.GetRandomId()
		bMsg, err := dp.Pack(znet.NewMsg(1, []byte("Hello, this is Zinx client!")))
		if err != nil {
			fmt.Println("[Client] pack msg failed: ", err)
			return
		}

		if _, err := conn.Write(bMsg); err != nil {
			fmt.Println("[Client] send msg to srever failed: ", err)
			return
		}
		// 2.接收服务器给客户端的回复
		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(conn, headData); err != nil {
			fmt.Println("[Client] read msg head from server failed: ", err)
			break
		}
		msgHead, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("[Client] unpack msg head from server failed: ", err)
			break
		}
		if msgHead.GetMsgLen() > 0 {
			msg := msgHead.(*znet.Message)
			msg.MsgContent = make([]byte, msg.GetMsgLen())
			if _, err := io.ReadFull(conn, msg.MsgContent); err != nil {
				fmt.Println("[Client] read msg content from server failed: ", err)
				break
			}

			fmt.Printf("[Client] receive server feedback msg success! msg id:%d, msg len:%d, msg content:%s\n",
				msg.GetMsgId(), msg.GetMsgLen(), string(msg.MsgContent))
		}

		// 阻塞
		time.Sleep(time.Second * 1)
	}
}
