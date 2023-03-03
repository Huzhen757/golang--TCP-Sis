package znet

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

// 创建消息实例的方法
func NewMsg(id uint32, data []byte) *Message {
	return &Message{id, uint32(len(data)), data}
}
