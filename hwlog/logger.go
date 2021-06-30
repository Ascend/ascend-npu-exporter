//  Copyright(C) 2021. Huawei Technologies Co.,Ltd.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package hwlog provides the capability of processing Huawei log rules.
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
)

const (
	defaultFileMaxSize             = 20   // the default maximum size of a single log file is 20 MB
	defaultMinSaveAge              = 7    // the minimum storage duration of backup logs is 7 days
	defaultMaxBackups              = 30   // the default number of backup log
	logFileMode        os.FileMode = 0640 // log file mode
	backupLogFileMode  os.FileMode = 0400 // backup log file mode
)

var logger *zap.Logger

// LogConfig log module config
type LogConfig struct {
	// log file path
	LogFileName string
	// only write to std out, default value: false
	OnlyToStdout bool
	// log level, -1-debug, 0-info, 1-warning, 2-error, 3-dpanic, 4-panic, 5-fatal, default value: 0
	LogLevel int
	// log file mode, default value: 0640
	LogMode os.FileMode
	// backup log file mode, default value: 0440
	BackupLogMode os.FileMode
	// size of a single log file (MB), default value: 20MB
	FileMaxSize int
	// maximum number of backup log files, default value: 30
	MaxBackups int
	// maximum number of days for backup log files, default value: 7
	MaxAge int
	// whether backup files need to be compressed, default value: false
	IsCompress bool
}

// IsInitLogger check logger initialized
func IsInitLogger() bool {
	return logger != nil
}

// InitLogger to get Logger
func InitLogger(config *LogConfig, stopCh <-chan struct{}) error {
	if logger != nil {
		return fmt.Errorf("the logger has been initialized and does not need to be initialized again")
	}
	err := validateLogConfigFiled(config)
	if err != nil {
		return err
	}
	logger = createLogger(*config)
	if logger == nil {
		return fmt.Errorf("create logger error")
	}
	logger.Info("logger init success")
	// skip change file mode and fs notify
	if config.OnlyToStdout {
		return nil
	}
	err = os.Chmod(config.LogFileName, config.LogMode)
	if err != nil {
		logger.Error("config log path error")
		return fmt.Errorf("set log file mode failed")
	}
	go workerWatcher(*config, stopCh)
	return nil
}

func createLogger(config LogConfig) *zap.Logger {
	logEncoder := getEncoder()
	var writeSyncer zapcore.WriteSyncer
	if config.OnlyToStdout {
		writeSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout))
	} else {
		logWriter := getLogWriter(config)
		writeSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), logWriter)
	}
	core := zapcore.NewCore(logEncoder, writeSyncer, zapcore.Level(config.LogLevel))
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
	isAbs := path.IsAbs(filePath)
	if !isAbs {
		return fmt.Errorf("file path is not absolute path")
	}
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("convert to absolute path failed")
	}
	fileRealPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return fmt.Errorf("eval Symlinks path failed")
	}
	if absPath != fileRealPath {
		return fmt.Errorf("can not use Symlinks path")
	}
	return nil
}

func validateLogConfigFileMaxSize(config *LogConfig) error {
	if config.FileMaxSize == 0 {
		config.FileMaxSize = defaultFileMaxSize
		return nil
	}
	if config.FileMaxSize < 0 || config.FileMaxSize > defaultFileMaxSize {
		return fmt.Errorf("the size of a single log file range is (0, 20] MB")
	}

	return nil
}

func validateLogConfigBackups(config *LogConfig) error {
	if config.MaxBackups == 0 {
		config.MaxBackups = defaultMaxBackups
		return nil
	}
	if config.MaxBackups < 0 || config.MaxBackups > defaultMaxBackups {
		return fmt.Errorf("the number of backup log file range is (0, 30]")
	}
	return nil
}

func validateLogConfigMaxAge(config *LogConfig) error {
	if config.MaxAge == 0 {
		config.MaxAge = defaultMinSaveAge
		return nil
	}
	if config.MaxAge < defaultMinSaveAge {
		return fmt.Errorf("the maxage should be greater than 7 days")
	}
	return nil
}

func validateLogConfigFileMode(config *LogConfig) error {
	if config.LogMode > logFileMode {
		return fmt.Errorf("log file mode greater than default value")
	}
	if config.BackupLogMode > backupLogFileMode {
		return fmt.Errorf("backup log file mode greater than default value")
	}
	if config.LogMode == 0 {
		config.LogMode = logFileMode
	}
	if config.BackupLogMode == 0 {
		config.BackupLogMode = backupLogFileMode
	}
	return nil
}

func validateLogConfigFiled(config *LogConfig) error {
	if !config.OnlyToStdout {
		err := validate(config.LogFileName)
		if err != nil {
			return err
		}
	}
	err := validateLogConfigFileMaxSize(config)
	if err != nil {
		return err
	}
	err = validateLogConfigBackups(config)
	if err != nil {
		return err
	}
	err = validateLogConfigMaxAge(config)
	if err != nil {
		return err
	}
	err = validateLogConfigFileMode(config)
	if err != nil {
		return err
	}

	return nil
}

func workerWatcher(config LogConfig, stopCh <-chan struct{}) {
	if logger == nil {
		fmt.Println("workerWatcher logger is nil")
		return
	}
	if stopCh == nil {
		logger.Error("stop channel is nil")
		return
	}
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
		case <-stopCh:
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
	if logger == nil {
		fmt.Println("changeFileMode logger is nil")
		return
	}
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

// Debug record debug
func Debug(format string, args ...interface{}) {
	Debugf(format, args...)
}

// Debugf record debug
func Debugf(format string, args ...interface{}) {
	msgInfo := fmt.Sprintf(format, args...)
	if logger == nil {
		fmt.Println("Debugf function's logger is nil")
		return
	}
	logger.Debug(msgInfo)
}

// Info record info
func Info(format string, args ...interface{}) {
	Infof(format, args...)
}

// Infof record info
func Infof(format string, args ...interface{}) {
	msgInfo := fmt.Sprintf(format, args...)
	if logger == nil {
		fmt.Println("Infof function's logger is nil")
		return
	}
	logger.Info(msgInfo)
}

// Warn record warn
func Warn(format string, args ...interface{}) {
	Warnf(format, args...)
}

// Warnf record warn
func Warnf(format string, args ...interface{}) {
	msgInfo := fmt.Sprintf(format, args...)
	if logger == nil {
		fmt.Println("Warnf function's logger is nil")
		return
	}
	logger.Warn(msgInfo)
}

// Error record error
func Error(format string, args ...interface{}) {
	Errorf(format, args...)
}

// Errorf record error
func Errorf(format string, args ...interface{}) {
	msgInfo := fmt.Sprintf(format, args...)
	if logger == nil {
		fmt.Println("Errorf function's logger is nil")
		return
	}
	logger.Error(msgInfo)
}

// Dpanic record panic
func Dpanic(format string, args ...interface{}) {
	Dpanicf(format, args...)
}

// Dpanicf record panic
func Dpanicf(format string, args ...interface{}) {
	msgInfo := fmt.Sprintf(format, args...)
	if logger == nil {
		fmt.Println("Dpanicf function's logger is nil")
		return
	}
	logger.DPanic(msgInfo)
}

// Panic record panic
func Panic(format string, args ...interface{}) {
	Panicf(format, args...)
}

// Panicf record panic
func Panicf(format string, args ...interface{}) {
	msgInfo := fmt.Sprintf(format, args...)
	if logger == nil {
		fmt.Println("Panicf function's logger is nil")
		return
	}
	logger.Panic(msgInfo)
}

// Fatal record fatal
func Fatal(format string, args ...interface{}) {
	Fatalf(format, args...)
}

// Fatalf record fatal
func Fatalf(format string, args ...interface{}) {
	msgInfo := fmt.Sprintf(format, args...)
	if logger == nil {
		fmt.Println("Fatalf function's logger is nil")
		return
	}
	logger.Fatal(msgInfo)
}
