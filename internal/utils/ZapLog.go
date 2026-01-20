package utils

import (
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var atomicLogger atomic.Value // 保存 *zap.Logger

// InitCustomLogger 初始化一个定制的 Logger
func initCustomLogger() {
	fmt.Printf("Logger Init\n")
	dir := "output/log/"
	// result := fileutils.DeleteChildFiles(dir)
	// fmt.Printf("清空日志目录结果:%t\n", result)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		panic("Logger System Error")
	}

	currentDate := time.Now().Format("2006-01-02")
	// 日志文件路径
	debugLogFile := fmt.Sprintf("%sdebug_%s.log", dir, currentDate)
	infoLogFile := fmt.Sprintf("%sinfo_%s.log", dir, currentDate)
	errorLogFile := fmt.Sprintf("%serror_%s.log", dir, currentDate)

	// 创建日志切割器：根据日志级别分别设置不同的日志文件
	debugWriteSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   debugLogFile, // debug 日志文件
		MaxSize:    10,           // 单个日志文件最大尺寸（MB）
		MaxBackups: 5,            // 最多保留5个备份
		MaxAge:     30,           // 日志保留最长天数
		Compress:   true,         //
	})

	infoWriteSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   infoLogFile, // info 日志文件
		MaxSize:    10,          // 单个日志文件最大尺寸（MB）
		MaxBackups: 5,           // 最多保留5个备份
		MaxAge:     30,          // 日志保留最长天数
		Compress:   true,        // 启用日志压缩
	})

	errorWriteSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   errorLogFile, // error 日志文件
		MaxSize:    10,           // 单个日志文件最大尺寸（MB）
		MaxBackups: 5,            // 最多保留5个备份
		MaxAge:     30,           // 日志保留最长天数
		Compress:   true,         // 启用日志压缩
	})

	// 配置日志编码器
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000")) // 格式化为年月日时分秒.毫秒
	}
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder // 日志级别大写
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	// 终端输出（跨平台）
	consoleWriteSyncer := zapcore.AddSync(os.Stdout)
	consoleCore := zapcore.NewCore(encoder, consoleWriteSyncer, zapcore.DebugLevel)

	// 创建日志 Core
	debugCore := zapcore.NewCore(encoder, debugWriteSyncer, zapcore.DebugLevel)
	infoCore := zapcore.NewCore(encoder, infoWriteSyncer, zapcore.InfoLevel)
	errorCore := zapcore.NewCore(encoder, errorWriteSyncer, zapcore.ErrorLevel)

	// 将多个 Core 合并
	core := zapcore.NewTee(consoleCore, debugCore, infoCore, errorCore)

	// zap.AddCaller() 会在日志中加入调用函数的文件名和行号
	// zap.AddCallerSkip(1) 会跳过调用函数的文件名和行号
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	atomicLogger.Store(logger) //这么做主要是原子存储，方式，在每天0成切换日志文件的时候，重新执行initCustomLogger 方法过程中，logger重新构建问题
}

// InitLogger 初始化日志器（目前为空）
func InitLogger() {
	// 每天 0 点刷新日志
	initCustomLogger()
	go func() {
		for {
			now := time.Now()
			// 计算下一个 0 点
			next := now.Add(24 * time.Hour).Truncate(24 * time.Hour)
			time.Sleep(time.Until(next))
			initCustomLogger()
		}
	}()
}

// LDebug 输出日志
func LDebug(message string, fild ...zapcore.Field) {
	if l, ok := atomicLogger.Load().(*zap.Logger); ok && l != nil {
		l.Debug(message, fild...)
	}
}

// LInfo 输出信息级别日志
func LInfo(message string, fild ...zapcore.Field) {
	if l, ok := atomicLogger.Load().(*zap.Logger); ok && l != nil {
		l.Info(message, fild...)
	}
}

// LError 输出错误级别日志
func LError(message string, fild ...zapcore.Field) {
	if l, ok := atomicLogger.Load().(*zap.Logger); ok && l != nil {
		l.Error(message, fild...)
	}
}

func Sync() {
	if l, ok := atomicLogger.Load().(*zap.Logger); ok && l != nil {
		l.Sync()
	}
}
