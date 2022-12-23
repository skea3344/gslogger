// @file:   service.go
// @author: caibo
// @email:  caibo86@gmail.com
// @desc :  日志后台服务

package logger

import (
	"sync"

	"github.com/mgutz/ansi"
)

const (
	DefaultCacheSize = 512
)

const (
	DefaultFormat = 1
	JSONFormat    = 2
)

// 虽然可以自己创建LogService 但是建议全局使用一个默认服务即可
var global *logService

func init() {
	console.fatalColor = ansi.ColorFunc("red+u")
	console.errorColor = ansi.ColorFunc("red")
	console.warnColor = ansi.ColorFunc("yellow")
	console.infoColor = ansi.ColorFunc("white")
	console.debugColor = ansi.ColorFunc("cyan")
	console.timeFormat = "2006-01-02 15:04:05" // 精度自己选择
	// console.format = "2006-01-02 15:04:05.999"
	global = NewService(DefaultCacheSize)
}

func SetFlags(flags LEVEL) {
	global.SetFlags(flags)
}

func SetSinks(sinks ...ISink) {
	global.SetSinks(sinks...)
}

func AddSink(sink ISink) {
	global.AddSink(sink)
}

func ResetSinks() {
	global.ResetSinks()
}

func Get(name string) ILog {
	return global.Get(name)
}

func SetFormat(format int) {
	switch format {
	case JSONFormat:
		global.format = JSONFormat
	default:
	}
}

func Logoff(name string) {
	global.Logoff(name)
}

// Join 日志服务在系统中应该是最后关闭的 以保证所有生产日志都已经完成输出
func Join() {
	global.Join()
}

// LogService 日志服务
type logService struct {
	sync.Mutex
	msgChan chan *Message   // 消息缓冲队列
	flags   LEVEL           // 级别
	sinks   []ISink         // 输出后台列表
	logs    map[string]ILog // 注册的日志对象
	exit    chan bool       // 关闭服务
	format  int             // 日志格式
}

func NewService(cachesize int) *logService {
	service := &logService{
		msgChan: make(chan *Message, cachesize),
		flags:   FATAL | ERROR | WARN | INFO | DEBUG,
		logs:    make(map[string]ILog),
		exit:    make(chan bool, 1),
		sinks:   []ISink{&console},
		format:  DefaultFormat,
	}
	go service.start()
	return service
}

func (service *logService) dispath(msg *Message) {
	service.msgChan <- msg
}

func (service *logService) SetFlags(flags LEVEL) {
	service.Lock()
	defer service.Unlock()
	service.flags = flags
}

func (service *logService) SetSinks(sinks ...ISink) {
	service.Lock()
	defer service.Unlock()
	service.sinks = sinks
}

func (service *logService) AddSink(sink ISink) {
	service.Lock()
	defer service.Unlock()
	service.sinks = append(service.sinks, sink)
}

func (service *logService) ResetSinks() {
	service.Lock()
	defer service.Unlock()
	service.sinks = []ISink{&console}
}

func (service *logService) Get(name string) ILog {
	service.Lock()
	defer service.Unlock()
	if log, ok := service.logs[name]; ok {
		return log
	}
	log := &baseLog{
		flags:   service.flags,
		name:    name,
		sinks:   service.sinks,
		service: service,
	}
	service.logs[name] = log
	return log
}

func (service *logService) Logoff(name string) {
	service.Lock()
	defer service.Unlock()
	delete(service.logs, name)
}

func (service *logService) start() {
	for msg := range service.msgChan {
		for _, sink := range msg.Log.Sinks() {
			sink.Recv(msg)
		}
	}
	for _, log := range service.logs {
		for _, sink := range log.Sinks() {
			sink.Destroy()
		}
	}
	close(service.exit)
}

func (service *logService) Join() {
	close(service.msgChan)
	for range service.exit {
	}
}
