package zinterface

type IDataPack interface {
	GetHeadLen() uint32 // 获取消息长度

	Pack(msg IMessage) ([]byte, error) // 消息Message序列化=>TLV格式

	Unpack([]byte) (IMessage, error) // TLV => 反序列化 => 消息Message
}
