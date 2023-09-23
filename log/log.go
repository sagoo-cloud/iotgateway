package log

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
)

// Level 定义日志级别
type Level uint8

const (
	DEBUG   Level = iota // 调试级别
	INFO                 // 信息级别
	WARNING              // 警告级别
	ERROR                // 错误级别
	FATAL                // 致命错误级别
)

// Logger 是日志结构体
type Logger struct {
	level Level // 日志级别
}

// 全局默认logger
var std = New(INFO)

// New 创建一个logger
func New(level Level) *Logger {
	return &Logger{level: level}
}

// 写入日志
func (l *Logger) log(lv Level, format string, a ...any) {
	if lv < l.level {
		return
	}
	timeStr := time.Now().Format("2006-01-02 15:04:05.000") // 当前时间
	msg := fmt.Sprintf(format, a...)

	// 构建日志内容
	var log string
	switch lv {
	case DEBUG:
		log = timeStr + color.BlueString(" [DEBU] ") + msg

	case INFO:
		log = timeStr + color.GreenString(" [INFO] ") + msg

	case WARNING:
		log = timeStr + color.YellowString(" [WARN] ") + msg

	case ERROR:
		log = timeStr + color.RedString(" [ERRO] ") + msg

	case FATAL:
		log = timeStr + color.RedString(" [FATA] ") + msg

	}

	fmt.Fprintln(os.Stderr, log)
}

// Debug 调试级别日志
func (l *Logger) Debug(format string, a ...any) {
	l.log(DEBUG, format, a...)
}

// Info 信息级别日志
func (l *Logger) Info(format string, a ...any) {
	l.log(INFO, format, a...)
}

// Warn 警告级别日志
func (l *Logger) Warn(format string, a ...any) {
	l.log(WARNING, format, a...)
}

// Error 错误级别日志
func (l *Logger) Error(format string, a ...any) {
	l.log(ERROR, format, a...)
}

// Fatal 致命级别日志并退出
func (l *Logger) Fatal(format string, a ...any) {
	l.log(FATAL, format, a...)
	os.Exit(1)
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(lv Level) {
	l.level = lv
}

// GetLevel 获取日志级别
func (l *Logger) GetLevel() Level {
	return l.level
}

// Debug 包级别日志函数
func Debug(format string, a ...any) {
	std.Debug(format, a)
}

func Info(format string, a ...any) {
	std.Info(format, a)
}

func Warn(format string, a ...any) {
	std.Warn(format, a)
}

func Error(format string, a ...any) {
	std.Error(format, a)
}

func Fatal(format string, a ...any) {
	std.Fatal(format, a)
}
