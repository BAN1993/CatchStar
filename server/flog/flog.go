package flog

import (
	"sync"

	"github.com/sirupsen/logrus"
)

// 单例
var once sync.Once
var glog *logrus.Logger
func GetInstance() *logrus.Logger {
	once.Do(func() {
		glog = logrus.New()
	})
	return glog
}

// 初始化
func InitLog(filename string, level uint32) error {
	hook := NewHook(filename, level)
	GetInstance().AddHook(hook)
	GetInstance().SetLevel(getLevelMapping(level))
	return nil
}
