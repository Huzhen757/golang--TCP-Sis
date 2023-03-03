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
	WorkerPoolSize   uint32 // worker工作池数量
	MaxTaskSize      uint32 // worker工作池数量的上限
}

// 初始化GlobalObject对象,后续可通过用户配置的zinx.json更新部分参数
func init() {
	GlobalObject = &GlobalObj{
		Name:             "Zinx-TcpServer-App",
		Version:          "V3",
		TcpPort:          8999,
		Host:             "0.0.0.0",
		MaxConn:          1000,
		MaxTransDataSize: 4096,
		WorkerPoolSize:   16,
		MaxTaskSize:      1024,
	}

	GlobalObject.LoadArg()
}

// 加载zinx.json配置文件
func (g *GlobalObj) LoadArg() {
	data, err := ioutil.ReadFile("D:\\goproject\\src\\zinxDemo\\zinxV1\\conf\\zinx.json")
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
