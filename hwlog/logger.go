package hwlog

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path"
	"path/filepath"
	"regexp"
)

const (
	defaultFileMaxSize = 20
	defaultMinSaveAge  = 7
	logFileMode        = 0640
	backupLogFileMode  = 0400
)

var logger *zap.Logger

type LogConfig struct {
	LogFileName string
	LogMode     os.FileMode
	FileMaxSize int
	MaxBackups  int
	MaxAge      int
	IsCompress  bool
	SignalCh    chan struct{}
}

// NewLogger to create logger
func InitLogger(config LogConfig) error {
	err := validateLogConfigFiled(config)
	if err != nil {
		return err
	}
	logger = createLogger(config)
	if logger == nil {
		return fmt.Errorf("create logger error")
	}
	logger.Info("logger start")
	err = os.Chmod(config.LogFileName, config.LogMode)
	if err != nil {
		logger.Error("config log path error")
		return fmt.Errorf("set log file mode failed")
	}
	go workerWatcher(config)
	return nil
}

// GetLogger to get Logger
func GetLogger() *zap.Logger {
	return logger
}

func createLogger(config LogConfig) *zap.Logger {
	logWriter := getLogWriter(config)
	logEncoder := getEncoder()
	core := zapcore.NewCore(logEncoder, zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), logWriter), zap.InfoLevel)
	return zap.New(core, zap.AddCaller())
}

// getEncoder get zap encoder
func getEncoder() zapcore.Encoder {
	encoderConfig := zapcore.EncoderConfig{
		// Keys can be anything except the empty string.
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	return zapcore.NewConsoleEncoder(encoderConfig)
}

// getLogWriter get zap log writer
func getLogWriter(config LogConfig) zapcore.WriteSyncer {
	lumberjackLogger := &lumberjack.Logger{
		Filename:   config.LogFileName,
		MaxSize:    config.FileMaxSize, // megabytes
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge, // days
		Compress:   config.IsCompress,
	}
	return zapcore.AddSync(lumberjackLogger)
}

func validate(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("file path is empty")
	}
	realpath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("it's error when converted to an absolute path")
	}
	pattern := `^/*`
	reg := regexp.MustCompile(pattern)
	if !reg.MatchString(realpath) {
		return fmt.Errorf("regexp match failed")
	}
	return nil
}

func validateLogConfigFiled(config LogConfig) error {
	err := validate(config.LogFileName)
	if err != nil {
		return err
	}
	if config.FileMaxSize > defaultFileMaxSize {
		return fmt.Errorf("maximum size of a single file is 20 MB")
	}
	if config.MaxAge < defaultMinSaveAge {
		return fmt.Errorf("the storage duration must be greater than 7 days")
	}
	return nil
}

func workerWatcher(config LogConfig) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Error("NewWatcher failed", zap.String("err", err.Error()))
		return
	}
	defer watcher.Close()
	logPath := path.Dir(config.LogFileName)
	err = watcher.Add(logPath)
	if err != nil {
		logger.Error("watcher add log path failed")
		return
	}
	for {
		select {
		case <-config.SignalCh:
			logger.Error("recv stop signal")
			return
		case event, ok := <-watcher.Events:
			if !ok {
				logger.Error("watcher event failed, exit")
				return
			}
			if event.Op&fsnotify.Create == 0 {
				break
			}
			changeFileMode(event, config.LogFileName)
		case errWatcher, ok := <-watcher.Errors:
			if !ok {
				logger.Error("watcher error failed, exit")
				return
			}
			logger.Error("watcher error", zap.String("err", errWatcher.Error()))
			return
		}
	}
}

func changeFileMode(event fsnotify.Event, logFileFullPath string) {
	var logMode os.FileMode = backupLogFileMode
	logFileName := path.Base(logFileFullPath)
	logPath := path.Dir(logFileFullPath)
	changedFileName := path.Base(event.Name)
	if changedFileName == logFileName {
		logMode = logFileMode
	}
	changedLogFilePath := path.Join(logPath, changedFileName)
	errChmod := os.Chmod(changedLogFilePath, logMode)
	if errChmod != nil {
		logger.Error("set file mode failed", zap.String("filename", changedFileName))
	}
}
