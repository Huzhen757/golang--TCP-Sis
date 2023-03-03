package znet

import (
	"fmt"
	"io"
	"net"
	"testing"
)

// datapack方法的单元测试
func TestDataPack(t *testing.T) {

	// 1.创建tcp socket
	listner, err := net.Listen("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("[Test-Server] create tcp socker failed...", err)
	}
	// 2.开启协程模拟server对Message的unpack
	go func() {
		for {
			conn, err := listner.Accept()
			if err != nil {
				fmt.Println("[Test-Server] server accept connetion failed...", err)
			}

			go func(c net.Conn) {
				dp := &DataPack{}
				for {
					// 1.第一次读取Message head
					head := make([]byte, dp.GetHeadLen())
					len, err := io.ReadFull(c, head)
					if len != int(dp.GetHeadLen()) || err != nil {
						fmt.Println("[Test-Server] first read message failed...", err)
						break
					}
					msgInterface, err := dp.Unpack(head)
					if err != nil {
						fmt.Println("[Test-Server] first read message failed...", err)
						return
					}
					// 2.第二次读取Message
					if msgInterface.GetMsgLen() > 0 {
						msg := msgInterface.(*Message)
						msg.MsgContent = make([]byte, msg.GetMsgLen())

						_, err = io.ReadFull(c, msg.MsgContent)
						if err != nil {
							fmt.Println("[Test-Server] second read message failed...", err)
							return
						}
						fmt.Println("[Test-Server] full message read complete: ", msg.MsgId, "=>", msg.MsgLen, "=>", string(msg.MsgContent))
					}
				}
			}(conn)
		}
	}()

	// 2.模拟客户端对message的pack
	conn, err := net.Dial("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("[Test-Client] create client failed...", err)
		return
	}
	dp := &DataPack{}
	m1 := &Message{
		MsgId:      1001,
		MsgLen:     4,
		MsgContent: []byte{'z', 'i', 'n', 'x'},
	}
	r1, err := dp.Pack(m1)
	if err != nil {
		fmt.Println("[Test-Client] client first pack msg failed...", err)
		return
	}

	m2 := &Message{
		MsgId:      1002,
		MsgLen:     7,
		MsgContent: []byte{'g', 'o', 'l', 'a', 'n', 'g', '!'},
	}
	r2, err := dp.Pack(m2)
	if err != nil {
		fmt.Println("[Test-Client] client second pack msg failed...", err)
		return
	}
	r1 = append(r1, r2...)
	if _, err = conn.Write(r1); err != nil {
		fmt.Println("[Test-Client] client trans msg failed...", err)
		return
	}
	// 客户端阻塞等待
	select {}
}
