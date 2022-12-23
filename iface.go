// @file:   iface.go
// @author: caibo
// @email:  caibo86@gmail.com
// @desc :  interface

package logger

// ILog 日志接口
type ILog interface {
	Flags() LEVEL                      // 日志级别
	Sinks() []ISink                    // 输出列表
	D(format string, v ...interface{}) // 输出调试级别日志
	I(format string, v ...interface{}) // 输出信息级别日志
	W(format string, v ...interface{}) // 输出警告级别日志
	E(format string, v ...interface{}) // 输出错误级别日志
	F(format string, v ...interface{}) // 输出致命级别日志
	String() string
}

// ISink 日志输出后台接口
type ISink interface {
	Recv(msg *Message) // 接收并处理日志消息
	Destroy()
}
