// @file:   sink_file.go
// @author: caibo
// @email:  caibo86@gmail.com
// @desc :  文件日志

package logger

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DefaultCutSize = 20000000 // 默认切片大小
	CompressDay    = 0        // 压缩几天前的日志
)

// 日志文件目录
var LogDir = os.Getenv("HOME") + "/go/log"

type FileSink struct {
	sync.RWMutex
	flag       byte     // 标志
	logfile    *os.File // 系统文件指针
	date       string   // 日志日期
	size       uint64   // 文件分片大小
	num        uint32   // 分片计数器
	logname    string   // 日志名字
	desc       string   // 日志描述
	timeformat string   // 日志时间格式
}

// SetLogDir 如果要使用文件日志，需要先调用此接口设置日志目录
func SetLogDir(dir string) {
	if len(dir) > 0 {
		LogDir = dir
	}
	makedir()
}

func NewFileSink(ln string, desc string, cutsize uint64) *FileSink {
	sink := &FileSink{
		num:        0,
		size:       cutsize,
		logname:    ln,
		desc:       desc,
		flag:       0xFF,
		timeformat: "2006-01-02 15:04:05.000000",
		date:       getDate(),
	}
	if sink.size == 0 {
		sink.size = DefaultCutSize
	}
	return sink
}

// 日志全路径文件名
func (sink *FileSink) filename() string {
	return fmt.Sprintf("%s/%s_%s_%s_%d.log", LogDir, sink.logname, sink.desc, sink.date, sink.num)
}

// Recv 实现ISink
func (sink *FileSink) Recv(msg *Message) {
	sink.Lock()
	defer sink.Unlock()
	if sink.flag != 0xFF {
		log.Fatal("filelog not available")
	}
	sink.getlogger()
	s := fmt.Sprintf("%s (%20s:%5d):[%s] %12s - %s",
		msg.Timestamp.Format(sink.timeformat),
		msg.File,
		msg.Line,
		msg.Flag,
		msg.Log,
		msg.Content)
	if sink.logfile != nil {
		fmt.Fprintln(sink.logfile, s)
	} else {
		fmt.Fprintln(os.Stdout, s)
	}
}

func (sink *FileSink) Destroy() {
	sink.Lock()
	defer sink.Unlock()
	if sink.flag == 0x00 {
		return
	}
	if sink.logfile != nil {
		sink.logfile.Close()
		sink.logfile = nil
	}
	sink.flag = 0x00
}

// getlogger 为文件日志打开正确的文件句柄
func (sink *FileSink) getlogger() {
	var cutsize uint64
	if cutsize = sink.size; cutsize == 0 {
		cutsize = DefaultCutSize
	}
	date := getDate()
	if sink.logfile != nil {
		if sink.date != date {
			_ = sink.logfile.Sync()
			sink.logfile.Close()
			sink.logfile = nil
		} else if fi, err := sink.logfile.Stat(); err != nil {
			_ = sink.logfile.Sync()
			sink.logfile.Close()
			sink.logfile = nil
		} else if uint64(fi.Size()) > cutsize {
			_ = sink.logfile.Sync()
			sink.logfile.Close()
			sink.logfile = nil
		}
	}
	if sink.logfile == nil {
		if sink.date != date {
			sink.date = date
			sink.num = 0
		}
		for {
			sink.num++
			var err error
			sink.logfile, err = open(sink.filename(), cutsize)
			if err != nil {
				log.Fatal(err)
			}
			if sink.logfile != nil {
				break
			}
		}
	}
}

// open 取文件
func open(filename string, cutsize uint64) (*os.File, error) {
	logfile, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	fi, err := logfile.Stat()
	if err != nil {
		logfile.Close()
		return nil, err
	}
	if uint64(fi.Size()) > cutsize {
		logfile.Close()
		return nil, nil
	}
	return logfile, nil
}

func getDate() string {
	now := time.Now()
	return fmt.Sprintf("%04d%02d%02d", now.Year(), now.Month(), now.Day())
}

// makedir 创建日志目录
func makedir() {
	var temp = LogDir
	if err := os.Mkdir(temp, os.ModePerm); err != nil {
		if !os.IsExist(err) {
			log.Fatal(err)
		}
	}
}

// CompressLog 压缩给定名字日志文件，现在仅将时间线以前的日志文件压到一起，不区分名字。可添加根据名字分别压缩的功能
func CompressLog(logname, description string) bool {
	// 目录
	dirname := LogDir
	// 当前时间
	date := time.Now()
	// 指定时间
	date = date.AddDate(0, 0, CompressDay)
	// 指定日期
	theday := fmt.Sprintf("%04d%02d%02d", date.Year(), date.Month(), date.Day())
	// 打开或者创建压缩文件
	fw, err := os.OpenFile(dirname+"/"+theday+".tar.gz", os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		log.Println("create file failed.", err)
		return false
	}
	defer fw.Close()
	// gzip writer
	gw := gzip.NewWriter(fw)
	defer gw.Close()
	// tar writer
	tw := tar.NewWriter(gw)
	defer tw.Close()
	// 打开目录
	dir, err := os.Open(dirname)
	if err != nil {
		log.Println(err)
		return false
	}
	defer dir.Close()
	// 读取目录
	fis, err := dir.Readdir(0)
	if err != nil {
		log.Println(err)
		return false
	}
	for _, fi := range fis {
		s := fi.Name()
		// 找到有两个_的文件
		index := strings.IndexByte(s, '_')
		if index == -1 {
			continue
		}
		s = s[index+1:]
		index = strings.IndexByte(s, '_')
		if index == -1 {
			continue
		}
		s = s[index+1 : index+9]
		x, _ := strconv.Atoi(s)
		y, _ := strconv.Atoi(theday)
		// 如果此文件日期大于需要处理的日期则跳过
		if x > y {
			continue
		}
		fr, err := os.Open(dirname + "/" + fi.Name())
		if err != nil {
			log.Println("open log file failed.", err)
			return false
		}
		defer fr.Close()
		h := new(tar.Header)
		h.Name = fi.Name()
		h.Size = fi.Size()
		h.Mode = int64(fi.Mode())
		h.ModTime = fi.ModTime()
		err = tw.WriteHeader(h)
		if err != nil {
			log.Println("compress log file failed.", err)
			return false
		}
		_, err = io.Copy(tw, fr)
		if err != nil {
			log.Println("compress log file failed.", err)
			return false
		}
		tw.Flush()
		err = os.Remove(dirname + "/" + fi.Name())
		if err != nil {
			log.Println("remove log file failed", err)
		}
	}
	return true
}

// UncompressLog 解压缩指定日期日志 eg.theday:20140808
func UncompressLog(theday string) bool {
	fr, err := os.Open(LogDir + "/" + theday + ".tar.gz")
	if err != nil {
		log.Println("open tar file failed.", err)
		return false
	}
	defer fr.Close()
	gr, err := gzip.NewReader(fr)
	if err != nil {
		log.Println(err)
		return false
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		h, e := tr.Next()
		if e == io.EOF {
			break
		}
		if err != nil {
			log.Println(e)
			return false
		}
		fw, e := os.OpenFile(LogDir+"/"+h.Name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if e != nil {
			log.Println(e)
			return false
		}
		defer fw.Close()
		_, e = io.Copy(fw, tr)
		if e != nil {
			log.Println("uncompress log failed.", e)
			return false
		}
	}
	err = os.Remove(LogDir + "/" + theday + ".tar.gz")
	if err != nil {
		log.Println("remove file failed.", err)
		return false
	}
	return true
}
