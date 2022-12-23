// @file:   message.go
// @author: caibo
// @email:  caibo86@gmail.com
// @desc :  日志消息

package logger

import "time"

// message 日志消息
type Message struct {
	Flag      LEVEL     // 日志级别
	Timestamp time.Time // 时间戳
	Log       ILog      // 产生此消息的日志对象
	File      string    // 源文件名
	Line      int       // 代码行数
	Content   string    // 内容
}
