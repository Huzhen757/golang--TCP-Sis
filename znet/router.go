package znet

import "zinx/zinterface"

// router应该由用户自定义实现，可以先嵌入BaseRouter基类，然后进行方法的重写
// BaseRouter其实是一种模板方法的体现，方法具体实现留给子类进行扩展
type BaseRouter struct {
}

//处理业务之前的方法
func (br *BaseRouter) PreHandle(request zinterface.IRequest) {}

//业务处理主方法
func (br *BaseRouter) Handle(request zinterface.IRequest) {}

//处理业务之后的方法
func (br *BaseRouter) PostHandle(request zinterface.IRequest) {}
