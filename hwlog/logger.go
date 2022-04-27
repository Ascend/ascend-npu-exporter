//  Copyright(C) 2021. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package hwlog provides the capability of processing Huawei log rules.
package hwlog

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
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
	LogDirMode     = 0750
	backUpLogRegex = `^.+-[0-9]{4}-[0-9]{2}-[0-9T]{5}-[0-9]{2}-[0-9]{2}\.[0-9]{2,4}`
	bitsize        = 64
	stackDeep      = 3
	pathLen        = 2
	minLogLevel    = -1
	maxLogLevel    = 5
)

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

var reg = regexp.MustCompile(backUpLogRegex)

type validateFunc func(config *LogConfig) error

// Init initialize and return the logger
func Init(config *LogConfig, stopCh <-chan struct{}) (*zap.Logger, error) {
	if err := validateLogConfigFiled(config); err != nil {
		return nil, err
	}
	zapLogger := create(*config)
	if zapLogger == nil {
		return nil, fmt.Errorf("create logger error")
	}
	msg := fmt.Sprintf("%s's logger init success.", path.Base(config.LogFileName))
	zapLogger.Info(msg)
	// skip change file mode and fs notify
	if config.OnlyToStdout {
		return zapLogger, nil
	}
	if err := os.Chmod(config.LogFileName, config.LogMode); err != nil {
		zapLogger.Error("config log path error")
		return zapLogger, fmt.Errorf("set log file mode failed")
	}
	go workerWatcher(zapLogger, *config, stopCh)
	return zapLogger, nil
}

func create(config LogConfig) *zap.Logger {
	logEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
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
		if err != nil {
			return fmt.Errorf("create file(%s) failed", fileName)
		}
		defer f.Close()
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
	if !isExist(path) {
		return path[len(path)-1:] == "/"
	}
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
	if config.MaxBackups <= 0 || config.MaxBackups > DefaultMaxBackups {
		return fmt.Errorf("the number of backup log file range is (0, 30]")
	}
	return nil
}

func validateLogConfigMaxAge(config *LogConfig) error {
	if config.MaxAge < DefaultMinSaveAge {
		return fmt.Errorf("the maxage should be greater than 7 days")
	}
	return nil
}

func validateLogLevel(config *LogConfig) error {
	if config.LogLevel < minLogLevel || config.LogLevel > maxLogLevel {
		return fmt.Errorf("the log level range should be [-1, 5]")
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
		validateLogConfigMaxAge, validateLogConfigFileMode, validateLogLevel)
	return funcList
}

func validateLogConfigFiled(config *LogConfig) error {
	if config.OnlyToStdout {
		return nil
	}
	if _, err := CheckPath(config.LogFileName); err != nil {
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

func workerWatcher(l *zap.Logger, config LogConfig, stopCh <-chan struct{}) {
	if l == nil {
		fmt.Println("workerWatcher logger is nil")
		return
	}
	if stopCh == nil {
		l.Error("stop channel is nil")
		return
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		l.Error("NewWatcher failed", zap.String("err", err.Error()))
		return
	}
	defer watcher.Close()
	logPath := path.Dir(config.LogFileName)
	if err = watcher.Add(logPath); err != nil {
		l.Error("watcher add log path failed")
		return
	}
	for {
		select {
		case _, ok := <-stopCh:
			if !ok {
				l.Error("recv stop signal")
				return
			}
		case event, ok := <-watcher.Events:
			if !ok {
				l.Error("watcher event failed, exit")
				return
			}
			if event.Op&fsnotify.Create == 0 {
				break
			}
			changeFileMode(l, event, config.LogFileName)
		case errWatcher, ok := <-watcher.Errors:
			if !ok {
				l.Error("watcher error failed, exit")
				return
			}
			l.Error("watcher error", zap.String("err", errWatcher.Error()))
			return
		}
	}
}

func changeFileMode(l *zap.Logger, event fsnotify.Event, logFileFullPath string) {
	if l == nil {
		fmt.Println("changeFileMode logger is nil")
		return
	}
	var logMode = LogFileMode
	logPath := path.Dir(logFileFullPath)
	changedFileName := path.Base(event.Name)
	if isTargetLog(changedFileName) {
		logMode = BackupLogFileMode
	}
	changedLogFilePath := path.Join(logPath, changedFileName)
	if !isExist(changedLogFilePath) {
		return
	}
	path, err := CheckPath(changedLogFilePath)
	if err != nil {
		return
	}
	if errChmod := os.Chmod(path, logMode); errChmod != nil {
		l.Error("set file mode failed", zap.String("filename", changedFileName))
	}
}
func isTargetLog(fileName string) bool {
	return reg.MatchString(fileName)
}

// CheckPath  validate path
func CheckPath(path string) (string, error) {
	if path == "" {
		return path, nil
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", errors.New("get the absolute path failed")
	}
	resoledPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			return "", os.ErrNotExist
		}
		return "", errors.New("get the symlinks path failed")
	}
	if absPath != resoledPath {
		return "", errors.New("can't support symlinks")
	}
	return resoledPath, nil
}
