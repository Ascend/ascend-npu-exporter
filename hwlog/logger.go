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
	"runtime/debug"
	"strconv"
	"strings"
)

const (
	// DefaultFileMaxSize  the default maximum size of a single log file is 20 MB
	DefaultFileMaxSize = 20
	// DefaultMinSaveAge the minimum storage duration of backup logs is 7 days
	DefaultMinSaveAge = 7
	// DefaultMaxBackups the default number of backup log
	DefaultMaxBackups = 30
	// LogFileMode log file mode
	LogFileMode os.FileMode = 0640
	// BackupLogFileMode backup log file mode
	BackupLogFileMode os.FileMode = 0400
	// LogDirMode log dir mode
	LogDirMode = 0750
	// IndexOfCallerFileInfo the index of the caller's file information
	IndexOfCallerFileInfo = 10
	// IndexOfGoroutineIDInfo the index of the goroutineID information
	IndexOfGoroutineIDInfo = 0
	// IndexOfGoroutineID the index of the goroutineID
	IndexOfGoroutineID = 1
	// LengthOfFileInfo the length of the log printed file information
	LengthOfFileInfo = 20
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
	// backup log file mode, default value: 0400
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

type validateFunc func(config *LogConfig) error

// IsInit check logger initialized
func IsInit() bool {
	return logger != nil
}

// Init to get Logger
func Init(config *LogConfig, stopCh <-chan struct{}) error {
	if logger != nil {
		return fmt.Errorf("the logger has been initialized and does not need to be initialized again")
	}
	if err := validateLogConfigFiled(config); err != nil {
		return err
	}
	logger = create(*config)
	if logger == nil {
		return fmt.Errorf("create logger error")
	}
	logger.Info("logger init success")
	// skip change file mode and fs notify
	if config.OnlyToStdout {
		return nil
	}
	if err := os.Chmod(config.LogFileName, config.LogMode); err != nil {
		logger.Error("config log path error")
		return fmt.Errorf("set log file mode failed")
	}
	go workerWatcher(*config, stopCh)
	return nil
}

func create(config LogConfig) *zap.Logger {
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

func checkDir(fileDir string) error {
	if !isExist(fileDir) {
		if err := os.MkdirAll(fileDir, LogDirMode); err != nil {
			return fmt.Errorf("create dirs failed")
		}
		return nil
	}
	if err := os.Chmod(fileDir, LogDirMode); err != nil {
		return fmt.Errorf("change log dir mode failed")
	}
	return nil
}

func createFile(filePath string) error {
	fileName := path.Base(filePath)
	if !isExist(filePath) {
		f, err := os.Create(filePath)
		defer f.Close()
		if err != nil {
			return fmt.Errorf("create file(%s) failed", fileName)
		}
	}
	return nil
}

func checkAndCreateLogFile(filePath string) error {
	if !isFile(filePath) {
		return fmt.Errorf("config path is not file")
	}
	fileDir := path.Dir(filePath)
	if err := checkDir(fileDir); err != nil {
		return err
	}
	if err := createFile(filePath); err != nil {
		return err
	}
	return nil
}

func isDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func isFile(path string) bool {
	return !isDir(path)
}

func isExist(filePath string) bool {
	if _, err := os.Stat(filePath); err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func validateLogConfigFileMaxSize(config *LogConfig) error {
	if config.FileMaxSize == 0 {
		config.FileMaxSize = DefaultFileMaxSize
		return nil
	}
	if config.FileMaxSize < 0 || config.FileMaxSize > DefaultFileMaxSize {
		return fmt.Errorf("the size of a single log file range is (0, 20] MB")
	}

	return nil
}

func validateLogConfigBackups(config *LogConfig) error {
	if config.MaxBackups == 0 {
		config.MaxBackups = DefaultMaxBackups
		return nil
	}
	if config.MaxBackups < 0 || config.MaxBackups > DefaultMaxBackups {
		return fmt.Errorf("the number of backup log file range is (0, 30]")
	}
	return nil
}

func validateLogConfigMaxAge(config *LogConfig) error {
	if config.MaxAge == 0 {
		config.MaxAge = DefaultMinSaveAge
		return nil
	}
	if config.MaxAge < DefaultMinSaveAge {
		return fmt.Errorf("the maxage should be greater than 7 days")
	}
	return nil
}

func validateLogConfigFileMode(config *LogConfig) error {
	if config.LogMode == 0 {
		config.LogMode = LogFileMode
	}
	if config.BackupLogMode == 0 {
		config.BackupLogMode = BackupLogFileMode
	}
	return nil
}

func getValidateFuncList() []validateFunc {
	var funcList []validateFunc
	funcList = append(funcList, validateLogConfigFileMaxSize, validateLogConfigBackups,
		validateLogConfigMaxAge, validateLogConfigFileMode)
	return funcList
}

func validateLogConfigFiled(config *LogConfig) error {
	if config.OnlyToStdout {
		return nil
	}
	if !path.IsAbs(config.LogFileName) {
		return fmt.Errorf("config log path is not absolute path")
	}

	if err := checkAndCreateLogFile(config.LogFileName); err != nil {
		return err
	}
	validateFuncList := getValidateFuncList()
	for _, vaFunc := range validateFuncList {
		if err := vaFunc(config); err != nil {
			return err
		}
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
	if err = watcher.Add(logPath); err != nil {
		logger.Error("watcher add log path failed")
		return
	}
	for {
		select {
		case _, ok := <-stopCh:
			if !ok {
				logger.Error("recv stop signal")
				return
			}
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
	var logMode = BackupLogFileMode
	logFileName := path.Base(logFileFullPath)
	logPath := path.Dir(logFileFullPath)
	changedFileName := path.Base(event.Name)
	if changedFileName == logFileName {
		logMode = LogFileMode
	}
	changedLogFilePath := path.Join(logPath, changedFileName)
	if !isExist(changedLogFilePath) {
		return
	}
	if errChmod := os.Chmod(changedLogFilePath, logMode); errChmod != nil {
		logger.Error("set file mode failed", zap.String("filename", changedFileName))
	}
}

// printHelper helper function for log printing
func printHelper(f func(string, ...zap.Field), msg string) {
	str := getCallerInfo()
	f(str + msg)
}

// getCallerInfo gets the caller's information
func getCallerInfo() string {
	path := string(debug.Stack())
	paths := strings.Split(path, "\n")

	callerPath := getCallerPath(paths)
	goroutineID := getGorouineID(paths)

	str := fmt.Sprintf("%s\t%-8d\t", callerPath, goroutineID)

	return str
}

// getCallerPath gets the file path and line number of the caller
func getCallerPath(paths []string) string {
	if len(paths) <= IndexOfCallerFileInfo {
		return ""
	}
	str := paths[IndexOfCallerFileInfo]
	spaceIndex := strings.LastIndex(str, " ")
	if spaceIndex != -1 {
		str = str[:spaceIndex]
	}
	slashIndex1 := strings.LastIndex(str, "/")
	if slashIndex1 == -1 {
		return ""
	}
	slashIndex2 := strings.LastIndex(str[:slashIndex1], "/")
	if slashIndex2 == -1 {
		return ""
	}

	if len(str)-slashIndex2-1 > LengthOfFileInfo {
		str = str[slashIndex1+1:]
	} else {
		str = str[slashIndex2+1:]
	}

	if len(str) < LengthOfFileInfo {
		str += strings.Repeat(" ", LengthOfFileInfo-len(str))
	}
	return str
}

// getCallerGoroutineID gets the goroutineID
func getGorouineID(paths []string) int {
	if len(paths) <= IndexOfGoroutineIDInfo {
		return -1
	}
	str := paths[IndexOfGoroutineIDInfo]
	strs := strings.Split(str, " ")
	if len(strs) <= IndexOfGoroutineID {
		return -1
	}
	curGoroutineID, err := strconv.Atoi(strs[IndexOfGoroutineID])
	if err != nil {
		return -1
	}
	return curGoroutineID
}

// Debug record debug not format
func Debug(args ...interface{}) {
	if logger == nil {
		fmt.Println("Debug function's logger is nil")
		return
	}
	printHelper(logger.Debug, fmt.Sprint(args...))
}

// Debugf record debug
func Debugf(format string, args ...interface{}) {
	if logger == nil {
		fmt.Println("Debugf function's logger is nil")
		return
	}
	printHelper(logger.Debug, fmt.Sprintf(format, args...))
}

// Info record info not format
func Info(args ...interface{}) {
	if logger == nil {
		fmt.Println("Info function's logger is nil")
		return
	}
	printHelper(logger.Info, fmt.Sprint(args...))
}

// Infof record info
func Infof(format string, args ...interface{}) {
	if logger == nil {
		fmt.Println("Infof function's logger is nil")
		return
	}
	printHelper(logger.Info, fmt.Sprintf(format, args...))
}

// Warn record warn not format
func Warn(args ...interface{}) {
	if logger == nil {
		fmt.Println("Warn function's logger is nil")
		return
	}
	printHelper(logger.Warn, fmt.Sprint(args...))
}

// Warnf record warn
func Warnf(format string, args ...interface{}) {
	if logger == nil {
		fmt.Println("Warnf function's logger is nil")
		return
	}
	printHelper(logger.Warn, fmt.Sprintf(format, args...))
}

// Error record error not format
func Error(args ...interface{}) {
	if logger == nil {
		fmt.Println("Error function's logger is nil")
		return
	}
	printHelper(logger.Error, fmt.Sprint(args...))
}

// Errorf record error
func Errorf(format string, args ...interface{}) {
	if logger == nil {
		fmt.Println("Errorf function's logger is nil")
		return
	}
	printHelper(logger.Error, fmt.Sprintf(format, args...))
}

// Dpanic record panic not format
func Dpanic(args ...interface{}) {
	if logger == nil {
		fmt.Println("Dpanic function's logger is nil")
		return
	}
	printHelper(logger.DPanic, fmt.Sprint(args...))
}

// Dpanicf record panic
func Dpanicf(format string, args ...interface{}) {
	if logger == nil {
		fmt.Println("Dpanicf function's logger is nil")
		return
	}
	printHelper(logger.DPanic, fmt.Sprintf(format, args...))
}

// Panic record panic not format
func Panic(args ...interface{}) {
	if logger == nil {
		fmt.Println("Panic function's logger is nil")
		return
	}
	printHelper(logger.Panic, fmt.Sprint(args...))
}

// Panicf record panic
func Panicf(format string, args ...interface{}) {
	if logger == nil {
		fmt.Println("Panicf function's logger is nil")
		return
	}
	printHelper(logger.Panic, fmt.Sprintf(format, args...))
}

// Fatal record fatal not format
func Fatal(args ...interface{}) {
	if logger == nil {
		fmt.Println("Fatal function's logger is nil")
		return
	}
	printHelper(logger.Fatal, fmt.Sprint(args...))
}

// Fatalf record fatal
func Fatalf(format string, args ...interface{}) {
	if logger == nil {
		fmt.Println("Fatalf function's logger is nil")
		return
	}
	printHelper(logger.Fatal, fmt.Sprintf(format, args...))
}
