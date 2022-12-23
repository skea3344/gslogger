// @file:   service.go
// @author: caibo
// @email:  caibo86@gmail.com
// @desc :  日志后台服务

package logger

import "sync"

const (
	DefaultCacheSize = 512
)

// 虽然可以自己创建LogService 但是建议全局使用一个默认服务即可
var global *LogService

func init() {
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

func Logoff(name string) {
	global.Logoff(name)
}

// Join 日志服务在系统中应该是最后关闭的 以保证所有生产日志都已经完成输出
func Join() {
	global.Join()
}

// LogService 日志服务
type LogService struct {
	sync.Mutex
	MsgChan chan *Msg       // 消息缓冲队列
	Flags   LEVEL           // 级别
	Sinks   []ISink         // 输出后台列表
	Logs    map[string]ILog // 注册的日志对象
	Exit    chan bool       // 关闭服务
}

func NewService(cachesize int) *LogService {
	service := &LogService{
		MsgChan: make(chan *Msg, cachesize),
		Flags:   FATAL | ERROR | WARN | INFO | DEBUG,
		Logs:    make(map[string]ILog),
		Exit:    make(chan bool, 1),
		Sinks:   []ISink{&console},
	}
	go service.start()
	return service
}

func (service *LogService) dispath(msg *Msg) {
	service.MsgChan <- msg
}

func (service *LogService) SetFlags(flags LEVEL) {
	service.Lock()
	defer service.Unlock()
	service.Flags = flags
	for _, log := range service.Logs {
		log.SetFlags(flags)
	}
}

func (service *LogService) SetSinks(sinks ...ISink) {
	service.Lock()
	defer service.Unlock()
	service.Sinks = sinks
	for _, log := range service.Logs {
		log.SetSinks(sinks...)
	}
}

func (service *LogService) AddSink(sink ISink) {
	service.Lock()
	defer service.Unlock()
	service.Sinks = append(service.Sinks, sink)
	for _, log := range service.Logs {
		log.AddSink(sink)
	}
}

func (service *LogService) ResetSinks() {
	service.Lock()
	defer service.Unlock()
	service.Sinks = []ISink{&console}
	for _, log := range service.Logs {
		log.SetSinks(service.Sinks...)
	}
}

func (service *LogService) Get(name string) ILog {
	service.Lock()
	defer service.Unlock()
	if log, ok := service.Logs[name]; ok {
		return log
	}
	log := &baseLog{
		flags:   service.Flags,
		name:    name,
		sinks:   service.Sinks,
		service: service,
	}
	service.Logs[name] = log
	return log
}

func (service *LogService) Logoff(name string) {
	service.Lock()
	defer service.Unlock()
	delete(service.Logs, name)
}

func (service *LogService) start() {
	for msg := range service.MsgChan {
		for _, sink := range msg.Log.Sinks() {
			sink.Recv(msg)
		}
	}
	close(service.Exit)
}

func (service *LogService) Join() {
	close(service.MsgChan)
	for range service.Exit {
	}
}
