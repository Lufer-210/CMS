package logger

import (
	"log"
	"os"
	"sync"
)

type Logger struct {
	logger *log.Logger // 单一日志实例，所有级别共用
	mu     sync.Mutex
}

var instance *Logger
var once sync.Once

// GetLogger 获取日志实例
func GetLogger() *Logger {
	once.Do(func() {
		instance = newLogger()
	})
	return instance
}

func newLogger() *Logger {
	return &Logger{
		logger: log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (l *Logger) Info(v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Println(append([]interface{}{"INFO:"}, v...)...)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Printf("INFO: "+format, v...)
}

func (l *Logger) Error(v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Println(append([]interface{}{"ERROR:"}, v...)...)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Printf("ERROR: "+format, v...)
}

func (l *Logger) Debug(v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Println(append([]interface{}{"DEBUG:"}, v...)...)
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Printf("DEBUG: "+format, v...)
}
