// @file:   sink_console.go
// @author: caibo
// @email:  caibo86@gmail.com
// @desc:   标准输出

package logger

import (
	"fmt"

	"github.com/mgutz/ansi"
)

var console consoleSink

func init() {
	console.fatalColor = ansi.ColorFunc("red+u")
	console.errorColor = ansi.ColorFunc("red")
	console.warnColor = ansi.ColorFunc("yellow")
	console.infoColor = ansi.ColorFunc("white")
	console.debugColor = ansi.ColorFunc("cyan")
	console.format = "2006-01-02 15:04:05" // 精度自己选择
	// console.format = "2006-01-02 15:04:05.999"
}

// consoleSink 标准输出
type consoleSink struct {
	format     string
	fatalColor func(string) string
	errorColor func(string) string
	warnColor  func(string) string
	infoColor  func(string) string
	debugColor func(string) string
}

func (sink *consoleSink) Recv(msg *Msg) {
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
	s := fmt.Sprintf("%s (%20s:%5d) [%s] %12s -- %s",
		msg.TS.Format(sink.format),
		msg.File,
		msg.Line,
		msg.Flag,
		msg.Log,
		msg.Content)
	fmt.Println(color(s))
}
