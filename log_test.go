// @file:   log_test.go
// @author: caibo
// @email:  caibo86@gmail.com
// @desc :  测试

package logger

import (
	"testing"
)

func TestLogger(t *testing.T) {
	SetLogDir("")
	AddSink(NewFileSink("logger", "test", 0))
	log := Get("test")
	log.F("This is a fatal log")
	log.E("This is an error log")
	log.W("This is a warn log")
	log.I("This is an info log")
	log.D("This is a debug log")
	Join()
}
