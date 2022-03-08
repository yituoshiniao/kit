package xlog

import (
	"os"
	"path/filepath"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/natefinch/lumberjack"
	"github.com/pkg/errors"
	"github.com/tchap/zapext/zapsentry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Set(conf Config) (func(), error) {
	logger, err := New(conf)
	if err != nil {
		return func() {}, err
	}
	return func() {
		_ = logger.Sync()
	}, nil
}

func New(conf Config) (*zap.Logger, error) {
	tee := []zapcore.Core{getBaseCore(conf)}
	errorCore := getErrorCore(conf)
	if errorCore != nil {
		tee = append(tee, errorCore)
	}
	warnCore := getWarnCore(conf)
	if warnCore != nil {
		tee = append(tee, warnCore)
	}

	logger := zap.New(zapcore.NewTee(tee...))

	_, _ = zap.RedirectStdLogAt(logger, conf.level()) //替换标准库的日志输出

	sentryCore, err := getSentryCore(conf.SentryDSN)
	if err != nil {
		logger.Error("获取Sentry client失败", zap.Error(errors.WithStack(err)))
	} else if sentryCore != nil {
		logger = logger.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewTee(core, sentryCore)
		}))
	}

	logger = logger.Named(conf.ServiceName)
	//替换全局logger
	zap.ReplaceGlobals(logger)
	if err = recordPanic(conf.File.Filename, conf.Stdout); err != nil {
		//S(context.Background()).Warn(err)
		return nil, err
	}

	return logger, nil
}

func getBaseCore(conf Config) zapcore.Core {
	var syncers []zapcore.WriteSyncer

	if conf.File.Filename != "" {
		if conf.File.BufSize == 0 {
			conf.File.BufSize = 1024 * 200
		}
		syncers = append(syncers, getRotatedSyncer(conf.File))
	}

	if conf.Stdout {
		//添加控制台打印
		syncers = append(syncers, getStdoutSyncer())
	}

	//zapcore.NewCore(zapcore.NewJSONEncoder(config), zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), logLevel),//同时将日志输出到控制台，NewJSONEncoder 是结构化输出

	return zapcore.NewCore(
		encoderFromFormat(conf.Format, conf.LevelColor, conf.CallerKey), // 编码器配置
		zapcore.NewMultiWriteSyncer(syncers...),         // 增加同步器
		zap.NewAtomicLevelAt(conf.level()),              // 日志级别

	)
}

func getErrorCore(conf Config) zapcore.Core {
	if conf.File.Filename != "" {
		file := conf.File
		file.Filename = filepath.Dir(file.Filename) + "/error/error.log"
		encoder := encoderFromFormat("plain", true, conf.CallerKey)
		return zapcore.NewCore(
			encoder, // 编码器配置
			zapcore.NewMultiWriteSyncer(getRotatedSyncer(file)), // 增加同步器
			zap.LevelEnablerFunc(func(level zapcore.Level) bool {
				return level == zapcore.ErrorLevel
			}),
		)
	}

	return nil
}

func getWarnCore(conf Config) zapcore.Core {
	if conf.File.Filename != "" {
		file := conf.File
		file.Filename = filepath.Dir(file.Filename) + "/error/warn.log"
		encoder := encoderFromFormat("plain", true, conf.CallerKey)
		return zapcore.NewCore(
			encoder, // 编码器配置
			zapcore.NewMultiWriteSyncer(getRotatedSyncer(file)), // 增加同步器
			zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
				return lvl == zapcore.WarnLevel
			}),
		)
	}

	return nil
}

func getSentryCore(sentryDSN string) (*zapsentry.Core, error) {
	if sentryDSN == "" {
		return nil, nil
	}
	client, err := raven.NewWithTags(sentryDSN, map[string]string{})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return zapsentry.NewCore(zapcore.WarnLevel, client), nil
}

func encoderFromFormat(format string, levelColor bool, callerKey string) zapcore.Encoder {
	ec := zap.NewProductionEncoderConfig()
	ec.EncodeTime = zapcore.ISO8601TimeEncoder
	if callerKey != "" {
		ec.CallerKey = callerKey //显示日志行号
	}
	ec.NameKey = "app"
	if levelColor {
		ec.EncodeLevel = zapcore.LowercaseColorLevelEncoder
	}
	if format == "json" {
		return zapcore.NewJSONEncoder(ec)
	} else {
		return zapcore.NewConsoleEncoder(ec)
	}
}

func getRotatedSyncer(flc FileLogConfig) zapcore.WriteSyncer {
	writer := &lumberjack.Logger{
		Filename:   flc.Filename,   // 日志文件路径
		MaxSize:    flc.MaxSize,    // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: flc.MaxBackups, // 日志文件最多保存多少个备份
		MaxAge:     flc.MaxDays,    // 文件最多保存多少天
	}
	go func() {
		for {
			<-time.After(time.Hour)
			_ = writer.Rotate()
		}
	}()

	return zapcore.AddSync(writer)
}

func getStdoutSyncer() zapcore.WriteSyncer {
	return zapcore.AddSync(os.Stdout)
}
