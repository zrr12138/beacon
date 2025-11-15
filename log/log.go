package log

import (
	"beacon/common"
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/gorm/logger"
)

const (
	DefaultLogLevel = 1
)

var zapLogger *zap.SugaredLogger
var atomicLevel zap.AtomicLevel

const SqlCtxLogKey string = "SqlCtxLogKey"

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
	err = os.Mkdir("./logs", 0755)
	if err != nil {
		if !os.IsExist(err) {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	logFile := fmt.Sprintf("./logs/%s.log", execName)
	//ln -s
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    1024,
		MaxBackups: 50,
		Compress:   false, // 是否压缩
		LocalTime:  true,
	})

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000000")

	encoderConfig.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		pc, file, line, ok := runtime.Caller(7)
		caller = zapcore.NewEntryCaller(pc, file, line, ok)
		if !ok {
			enc.AppendString("unknown")
			//enc.AppendString("unknown")
			return
		}
		//fn := runtime.FuncForPC(pc)
		//if fn == nil {
		//	enc.AppendString("unknown")
		//	enc.AppendString("unknown")
		//	return
		//}
		enc.AppendString(caller.TrimmedPath())
		//enc.AppendString(fn.Name())
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

func Info(args ...interface{}) {
	initLogger()
	zapLogger.Info(args...)
}

func Infof(template string, args ...interface{}) {
	initLogger()
	zapLogger.Infof(template, args...)
}

func Debug(args ...interface{}) {
	initLogger()
	zapLogger.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	initLogger()
	zapLogger.Debugf(template, args...)
}

func Warn(args ...interface{}) {
	initLogger()
	zapLogger.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	initLogger()
	zapLogger.Warnf(template, args...)
}

func Error(args ...interface{}) {
	initLogger()
	zapLogger.Error(args...)
}

func ErrorIfNeeded(log bool, args ...interface{}) {
	initLogger()
	if log {
		zapLogger.Error(args...)
	}
}

func Errorf(template string, args ...interface{}) {
	initLogger()
	zapLogger.Errorf(template, args...)
}

func ErrorfIfNeeded(log bool, template string, args ...interface{}) {
	initLogger()
	if log {
		zapLogger.Errorf(template, args...)
	}
}

func Fatal(args ...interface{}) {
	initLogger()
	zapLogger.Fatal(args...)

}

func Fatalf(template string, args ...interface{}) {
	initLogger()
	zapLogger.Fatalf(template, args...)
}

type Logger struct {
}

func (K *Logger) LogMode(level logger.LogLevel) logger.Interface {
	Info("log level: %s", level)
	return K
}

func (K *Logger) Info(ctx context.Context, s string, i ...interface{}) {
	initLogger()
	Infof(s, i)
}

func (K *Logger) Warn(ctx context.Context, s string, i ...interface{}) {
	initLogger()
	Warnf(s, i)
}

func (K *Logger) Error(ctx context.Context, s string, i ...interface{}) {
	initLogger()
	Errorf(s, i)
}

func (K *Logger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	initLogger()
	sql, rows := fc()
	level, ok := ctx.Value(SqlCtxLogKey).(zapcore.Level)
	if !ok {
		level = zapcore.InfoLevel
	}
	zapLogger.Logf(level, "execute [%v] affect rows num [%v]", sql, rows)
}

func (K *Logger) Flush() {
	return
}
