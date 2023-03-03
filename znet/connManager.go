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
