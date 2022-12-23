// @file:   message.go
// @author: caibo
// @email:  caibo86@gmail.com
// @desc :  日志消息

package logger

import (
	"encoding/json"
	"fmt"
	"time"
)

// message 日志消息
type Message struct {
	Flag      LEVEL     // 日志级别
	Timestamp time.Time // 时间戳
	Log       ILog      // 产生此消息的日志对象
	File      string    // 源文件名
	Line      int       // 代码行数
	Content   string    // 内容
	Format    int       // 日志格式
}

func (msg *Message) To_string(timeformat string) string {
	var s string
	if msg.Format == JSONFormat {
		m := make(map[string]interface{})
		m["Timestamp"] = msg.Timestamp.Format(timeformat)
		m["File"] = msg.File
		m["Line"] = msg.Line
		m["Flag"] = msg.Flag.String()
		m["Log"] = msg.Log.String()
		m["Content"] = msg.Content
		ret, _ := json.Marshal(m)
		s = string(ret)
	} else {
		s = fmt.Sprintf("%s (%20s:%5d) [%s] %12s -- %s",
			msg.Timestamp.Format(timeformat),
			msg.File,
			msg.Line,
			msg.Flag,
			msg.Log,
			msg.Content)
	}
	return s
}
