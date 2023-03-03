package znet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"zinx/utils"
	"zinx/zinterface"
)

type DataPack struct{}

func NewDataPack() *DataPack {
	return &DataPack{}
}

func (dp *DataPack) GetHeadLen() uint32 {
	return 8 // MsgLen(4B) + MsgId(4B)
}

// Message序列化成TVL: 先写len，再写id
func (dp *DataPack) Pack(msg zinterface.IMessage) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	err := binary.Write(buf, binary.LittleEndian, msg.GetMsgLen())
	if err != nil {
		fmt.Println("[Server] message len transfer to tvl failed: ", err)
		return nil, err
	}
	err = binary.Write(buf, binary.LittleEndian, msg.GetMsgId())
	if err != nil {
		fmt.Println("[Server] message type transfer to tvl failed: ", err)
		return nil, err
	}

	if err = binary.Write(buf, binary.LittleEndian, msg.GetMsgContent()); err != nil {
		fmt.Println("[Server] message content transfer to tvl failed: ", err)
		return nil, err
	}
	return buf.Bytes(), nil
}

// TVL反序列化成Message对象
func (dp *DataPack) Unpack(binaryData []byte) (zinterface.IMessage, error) {
	//创建一个从输入二进制数据的ioReader
	dataBuff := bytes.NewReader(binaryData)

	//只解压head的信息，得到dataLen和msgID
	msg := &Message{}

	//读dataLen
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.MsgLen); err != nil {
		return nil, err
	}

	//读msgID
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.MsgId); err != nil {
		return nil, err
	}

	//判断dataLen的长度是否超出我们允许的最大包长度
	if utils.GlobalObject.MaxTransDataSize > 0 && msg.MsgLen > utils.GlobalObject.MaxTransDataSize {
		return nil, errors.New("too large msg data received")
	}

	//这里只需要把head的数据拆包出来就可以了，然后再通过head的长度，再从conn读取一次数据
	return msg, nil
}
