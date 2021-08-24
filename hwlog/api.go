//  Copyright(c) 2021. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package hwlog provides the capability of processing Huawei log rules.
package hwlog

import (
	"fmt"
	"go.uber.org/zap"
)

// Logger logger struct
type Logger struct {
	ZapLogger *zap.Logger
}

// NewLogger create Logger
func NewLogger() *Logger {
	newLogger := new(Logger)
	return newLogger
}

// InitLogger initialize run logger
func (lg *Logger)InitLogger(config *LogConfig, stopCh <-chan struct{}) error {
	if lg != nil && lg.ZapLogger != nil {
		lg.Warn("logger is been initialized.")
		return nil
	}
	zapLog, err := Init(config, stopCh)
	if err != nil {
		return err
	}
	lg.ZapLogger = zapLog
	return nil
}

// IsInit check logger initialized
func (lg *Logger) IsInit() bool {
	return lg.ZapLogger != nil
}

// Debug record debug not format
func (lg *Logger) Debug(args ...interface{}) {
	if lg.ZapLogger == nil {
		fmt.Println("Debug function's logger is nil")
		return
	}
	printHelper(lg.ZapLogger.Debug, fmt.Sprint(args...))
}

// Debugf record debug
func (lg *Logger) Debugf(format string, args ...interface{}) {
	if lg.ZapLogger == nil {
		fmt.Println("Debugf function's logger is nil")
		return
	}
	printHelper(lg.ZapLogger.Debug, fmt.Sprintf(format, args...))
}

// Info record info not format
func (lg *Logger) Info(args ...interface{}) {
	if lg.ZapLogger == nil {
		fmt.Println("Info function's logger is nil")
		return
	}
	printHelper(lg.ZapLogger.Info, fmt.Sprint(args...))
}

// Infof record info
func (lg *Logger) Infof(format string, args ...interface{}) {
	if lg.ZapLogger == nil {
		fmt.Println("Infof function's logger is nil")
		return
	}
	printHelper(lg.ZapLogger.Info, fmt.Sprintf(format, args...))
}

// Warn record warn not format
func (lg *Logger) Warn(args ...interface{}) {
	if lg.ZapLogger == nil {
		fmt.Println("Warn function's logger is nil")
		return
	}
	printHelper(lg.ZapLogger.Warn, fmt.Sprint(args...))
}

// Warnf record warn
func (lg *Logger) Warnf(format string, args ...interface{}) {
	if lg.ZapLogger == nil {
		fmt.Println("Warnf function's logger is nil")
		return
	}
	printHelper(lg.ZapLogger.Warn, fmt.Sprintf(format, args...))
}

// Error record error not format
func (lg *Logger) Error(args ...interface{}) {
	if lg.ZapLogger == nil {
		fmt.Println("Error function's logger is nil")
		return
	}
	printHelper(lg.ZapLogger.Error, fmt.Sprint(args...))
}

// Errorf record error
func (lg *Logger) Errorf(format string, args ...interface{}) {
	if lg.ZapLogger == nil {
		fmt.Println("Errorf function's logger is nil")
		return
	}
	printHelper(lg.ZapLogger.Error, fmt.Sprintf(format, args...))
}

// Dpanic record panic not format
func (lg *Logger) Dpanic(args ...interface{}) {
	if lg.ZapLogger == nil {
		fmt.Println("Dpanic function's logger is nil")
		return
	}
	printHelper(lg.ZapLogger.DPanic, fmt.Sprint(args...))
}

// Dpanicf record panic
func (lg *Logger) Dpanicf(format string, args ...interface{}) {
	if lg.ZapLogger == nil {
		fmt.Println("Dpanicf function's logger is nil")
		return
	}
	printHelper(lg.ZapLogger.DPanic, fmt.Sprintf(format, args...))
}

// Panic record panic not format
func (lg *Logger) Panic(args ...interface{}) {
	if lg.ZapLogger == nil {
		fmt.Println("Panic function's logger is nil")
		return
	}
	printHelper(lg.ZapLogger.Panic, fmt.Sprint(args...))
}

// Panicf record panic
func (lg *Logger) Panicf(format string, args ...interface{}) {
	if lg.ZapLogger == nil {
		fmt.Println("Panicf function's logger is nil")
		return
	}
	printHelper(lg.ZapLogger.Panic, fmt.Sprintf(format, args...))
}

// Fatal record fatal not format
func (lg *Logger) Fatal(args ...interface{}) {
	if lg.ZapLogger == nil {
		fmt.Println("Fatal function's logger is nil")
		return
	}
	printHelper(lg.ZapLogger.Fatal, fmt.Sprint(args...))
}

// Fatalf record fatal
func (lg *Logger) Fatalf(format string, args ...interface{}) {
	if lg.ZapLogger == nil {
		fmt.Println("Fatalf function's logger is nil")
		return
	}
	printHelper(lg.ZapLogger.Fatal, fmt.Sprintf(format, args...))
}
