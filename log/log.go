package log

import (
	"beacon/common"
	"fmt"
	"os"
	"runtime"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	DefaultLogLevel = 1
)

var zapLogger *zap.SugaredLogger
var atomicLevel zap.AtomicLevel

func SetLogLevel(level string) error {
	initLogger()
	return atomicLevel.UnmarshalText([]byte(level))
}

func initLog() {
	execName, err := common.GetExecName()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := os.Mkdir("./logs", 0755); err != nil {
		if !os.IsExist(err) {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	logFile := fmt.Sprintf("./logs/%s.log", execName)
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    1024,
		MaxBackups: 50,
		Compress:   false,
		LocalTime:  true,
	})

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000000")
	encoderConfig.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		pc, file, line, ok := runtime.Caller(7)
		caller = zapcore.NewEntryCaller(pc, file, line, ok)
		if !ok {
			enc.AppendString("unknown")
			return
		}
		enc.AppendString(caller.TrimmedPath())
	}
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	atomicLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		w,
		atomicLevel,
	)
	zapLogger = zap.New(core, zap.AddCaller()).Sugar()
}

var once sync.Once

func initLogger() {
	once.Do(initLog)
}

func Info(args ...interface{})                    { initLogger(); zapLogger.Info(args...) }
func Infof(template string, args ...interface{})  { initLogger(); zapLogger.Infof(template, args...) }
func Debug(args ...interface{})                   { initLogger(); zapLogger.Debug(args...) }
func Debugf(template string, args ...interface{}) { initLogger(); zapLogger.Debugf(template, args...) }
func Warn(args ...interface{})                    { initLogger(); zapLogger.Warn(args...) }
func Warnf(template string, args ...interface{})  { initLogger(); zapLogger.Warnf(template, args...) }
func Error(args ...interface{})                   { initLogger(); zapLogger.Error(args...) }
func ErrorIfNeeded(do bool, args ...interface{}) {
	initLogger()
	if do {
		zapLogger.Error(args...)
	}
}
func Errorf(template string, args ...interface{}) { initLogger(); zapLogger.Errorf(template, args...) }
func ErrorfIfNeeded(do bool, template string, args ...interface{}) {
	initLogger()
	if do {
		zapLogger.Errorf(template, args...)
	}
}
func Fatal(args ...interface{})                   { initLogger(); zapLogger.Fatal(args...) }
func Fatalf(template string, args ...interface{}) { initLogger(); zapLogger.Fatalf(template, args...) }
