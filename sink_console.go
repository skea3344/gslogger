// @file:   sink_console.go
// @author: caibo
// @email:  caibo86@gmail.com
// @desc:   标准输出

package logger

import (
	"fmt"
)

var console consoleSink

// consoleSink 标准输出
type consoleSink struct {
	timeFormat string
	fatalColor func(string) string
	errorColor func(string) string
	warnColor  func(string) string
	infoColor  func(string) string
	debugColor func(string) string
}

func (sink *consoleSink) Recv(msg *Message) {
	var color func(string) string
	switch msg.Flag {
	case FATAL:
		color = sink.fatalColor
	case ERROR:
		color = sink.errorColor
	case WARN:
		color = sink.warnColor
	case INFO:
		color = sink.infoColor
	case DEBUG:
		color = sink.debugColor
	}
	fmt.Println(color(msg.To_string(sink.timeFormat)))
}

func (sink *consoleSink) Destroy() {}
