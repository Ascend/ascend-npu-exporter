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
	"sync"
)

const (
	defaultFileMaxSize = 20        // 默认单个日志文件大小最大为20M
	defaultMinSaveAge  = 7         // 备份日志文件最小保存时间7天
	logFileMode        = 0640      // 日志文件权限
	backupLogFileMode  = 0400      // 备份日志文件权限
)

var (
	logger *zap.Logger
	mutex sync.Mutex
)

// LogConfig log module config
type LogConfig struct {
	LogFileName string              // 日志文件路径
	LogLevel    int                 // 日志级别，-1-debug, 0-info, 1-warning, 2-error, 3-dpanic, 4-panic, 5-fatal
	LogMode     os.FileMode         // 日志文件权限
	FileMaxSize int                 // 单个日志文件大小(MB)
	MaxBackups  int                 // 最多保存多少个日志文件
	MaxAge      int                 // 最多保存多少天
	IsCompress  bool                // 是否压缩
	SignalCh    chan struct{}
}

// GetLogger to get Logger
func GetLogger(config LogConfig) (*zap.Logger, error) {
	mutex.Lock()
	if logger == nil {
		err := initLogger(config)
		if err != nil {
			return nil, err
		}
	}
	mutex.Unlock()
	return logger, nil
}

func initLogger(config LogConfig) error {
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

func createLogger(config LogConfig) *zap.Logger {
	logWriter := getLogWriter(config)
	logEncoder := getEncoder()
	core := zapcore.NewCore(logEncoder, zapcore.NewMultiWriteSyncer(
		zapcore.AddSync(os.Stdout), logWriter), zapcore.Level(config.LogLevel))
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
	isAbs := path.IsAbs(filePath)
	if !isAbs {
		return fmt.Errorf("file path is not absolute path")
	}
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("convert to absolute path failed")
	}
	fileRealPath, _ := filepath.EvalSymlinks(absPath)
	if absPath != fileRealPath {
		return fmt.Errorf("can not use Symlinks path")
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
	if logger == nil {
		fmt.Println("workerWatcher logger is nil")
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
