// @file:   log.go
// @author: caibo
// @email:  caibo86@gmail.com
// @desc :  日志句柄

package logger

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// LEVEL 日志级别
type LEVEL int32

// 内建日志级别
const (
	FATAL LEVEL = (1 << iota) // 致命
	ERROR                     // 错误
	WARN                      // 警告
	INFO                      // 信息
	DEBUG                     // 调试
)

func (level LEVEL) String() string {
	switch level {
	case FATAL:
		return "F"
	case ERROR:
		return "E"
	case WARN:
		return "W"
	case INFO:
		return "I"
	case DEBUG:
		return "D"
	}
	return "Unknown"
}

type baseLog struct {
	sync.RWMutex
	flags   LEVEL
	name    string
	sinks   []ISink
	service *logService
}

func (log *baseLog) write(flag LEVEL, format string, v ...interface{}) {
	file, line := stacktrace(3)
	msg := &Message{
		Flag:      flag,
		Timestamp: time.Now(),
		Log:       log,
		File:      file,
		Line:      line,
		Content:   fmt.Sprintf(format, v...),
		Format:    log.service.format,
	}
	log.service.dispath(msg)
}

func (log *baseLog) String() string {
	return log.name
}

func (log *baseLog) Flags() LEVEL {
	return log.flags
}

func (log *baseLog) D(format string, v ...interface{}) {
	if log.flags&DEBUG != 0 {
		log.write(DEBUG, format, v...)
	}
}

func (log *baseLog) I(format string, v ...interface{}) {
	if log.flags&INFO != 0 {
		log.write(INFO, format, v...)
	}
}

func (log *baseLog) W(format string, v ...interface{}) {
	if log.flags&WARN != 0 {
		log.write(WARN, format, v...)
	}
}

func (log *baseLog) E(format string, v ...interface{}) {
	if log.flags&ERROR != 0 {
		log.write(ERROR, format, v...)
	}
}

func (log *baseLog) F(format string, v ...interface{}) {
	if log.flags&FATAL != 0 {
		log.write(FATAL, format, v...)
	}
}

func (log *baseLog) Sinks() []ISink {
	return log.sinks
}

// 取文件名和行数
func stacktrace(skip int) (string, int) {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		file = "???"
		line = 0
	}
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			file = file[i+1:]
			break
		}
	}
	return file, line
}
