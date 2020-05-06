package flog

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	LevelError = iota
	LevelWarn
	LevelInfo
	LevelDebug
)

// 获取logrus.level映射关系
func getLevelMapping(level uint32) logrus.Level {
	switch level {
	case LevelError:
		return logrus.ErrorLevel
	case LevelWarn:
		return logrus.WarnLevel
	case LevelInfo:
		return logrus.InfoLevel
	case LevelDebug:
		return logrus.DebugLevel
	default:
		return logrus.DebugLevel
	}
}

type LoggerInterface interface {
	Init(config string) error
	WriteMsg(msg string, level int) error
	Destroy()
	Flush()
}

type LogWriter struct {
	*log.Logger
	mw *MuxWriter
	startLock sync.Mutex

	// 配置
	Filename	string	`json:"filename"`
	Maxlines	int		`json:"maxlines"`
	Maxsize		int		`json:"maxsize"`
	Daily		bool	`json:"daily"`
	Maxdays		int64	`json:"maxdays"`
	Rotate		bool	`json:"rotate"`
	Level		int		`json:"level"`

	// 记录
	maxlinesCurlines int
	maxsizeCursize int
	dailyOpendate int
}

type MuxWriter struct {
	sync.Mutex
	fd *os.File
}

func (l *MuxWriter) Write(b []byte) (int, error) {
	l.Lock()
	defer l.Unlock()
	return l.fd.Write(b)
}

func (l *MuxWriter) SetFd(fd *os.File) {
	if l.fd != nil {
		_ = l.fd.Close()
	}
	l.fd = fd
}

func NewFileWriter() LoggerInterface {
	w := &LogWriter{
		Filename: "server",
		Maxlines: 1000000,
		Maxsize:  1 << 28, //256 MB
		Daily:    true,
		Maxdays:  1,
		Rotate:   true,
		Level:    LevelDebug,
	}
	w.mw = new(MuxWriter)
	w.Logger = log.New(w.mw, "", log.Ldate|log.Ltime)
	return w
}

func (w *LogWriter) Init(config string) error {
	err := json.Unmarshal([]byte(config), w)
	if err != nil {
		return err
	}

	return w.startLogger()
}

func (w *LogWriter) startLogger() error {
	fd, err := w.createLogFile()
	if err != nil {
		return err
	}

	w.mw.SetFd(fd)
	err = w.initFd()
	if err != nil {
		return err
	}

	return nil
}

func (w *LogWriter) docheck(size int) {
	w.startLock.Lock()
	defer w.startLock.Unlock()
	if w.Rotate && ((w.Maxlines > 0 && w.maxlinesCurlines >= w.Maxlines) ||
		(w.Maxsize > 0 && w.maxsizeCursize >= w.Maxsize) ||
		(w.Daily && time.Now().Day() != w.dailyOpendate)) {
		if err := w.DoNewFile(); err != nil {
			fmt.Println(os.Stderr, "DoNewFile error:", err)
			return
		}
	}
	w.maxlinesCurlines++
	w.maxsizeCursize += size
}

func (w *LogWriter) WriteMsg(msg string, level int) error {
	if level > w.Level {
		return nil
	}
	n := 24 + len(msg)
	w.docheck(n)
	w.Logger.Print(msg)
	return nil
}

func (w *LogWriter) getFilename() string {
	t := time.Now()
	return fmt.Sprintf("%s_%04d%02d%02d_%02d%02d%02d.log",
		w.Filename, t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
}

func (w *LogWriter) createLogDir() error {
	dir := "./log"
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.Mkdir(dir, os.ModePerm)
			if err != nil {
				return fmt.Errorf("mkdir failed![%v]", err)
			}
			return nil
		}
		return fmt.Errorf("stat file error")
	}
	return nil
}

func (w *LogWriter) createLogFile() (*os.File, error) {
	err := w.createLogDir()
	if err != nil {
		return nil, err
	}
	fd, err := os.OpenFile("./log/"+w.getFilename(), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	return fd, err
}

func (w *LogWriter) initFd() error {
	fd := w.mw.fd
	finfo, err := fd.Stat()
	if err != nil {
		return fmt.Errorf("get stat err: %s", err)
	}
	w.maxsizeCursize = int(finfo.Size())
	w.dailyOpendate = time.Now().Day()
	if finfo.Size() > 0 {
		content, err := ioutil.ReadFile(w.Filename)
		if err != nil {
			return err
		}
		w.maxlinesCurlines = len(strings.Split(string(content), "\n"))
	} else {
		w.maxlinesCurlines = 0
	}
	return nil
}

// 新建一个日志文件
func (w *LogWriter) DoNewFile() error {
	fmt.Println("Need create a new file")

	w.mw.Lock()
	defer w.mw.Unlock()

	err := w.mw.fd.Close()
	if err != nil {
		return fmt.Errorf("DoNewFile fd.Close err:%s", err)
	}

	err = w.startLogger()
	if err != nil {
		return fmt.Errorf("DoNewFile startLogger err:%s", err)
	}
	return nil
}

func (w *LogWriter) Destroy() {
	_ = w.mw.fd.Close()
}

func (w *LogWriter) Flush() {
	_ = w.mw.fd.Sync()
}
