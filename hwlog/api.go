//  Copyright(c) 2021. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package hwlog provides the capability of processing Huawei log rules.
package hwlog

import (
	"context"
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
func (lg *Logger) InitLogger(config *LogConfig, stopCh <-chan struct{}) error {
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
	lg.DebugWithCtx(nil, args...)
}

// Debugf record debug
func (lg *Logger) Debugf(format string, args ...interface{}) {
	lg.DebugfWithCtx(nil, format, args...)
}

// DebugWithCtx record Debug not format
func (lg *Logger) DebugWithCtx(ctx context.Context, args ...interface{}) {
	if lg.validate() {
		printHelper(lg.ZapLogger.Debug, fmt.Sprint(args...), ctx)
	}
}

// DebugfWithCtx record Debug  format
func (lg *Logger) DebugfWithCtx(ctx context.Context, format string, args ...interface{}) {
	if lg.validate() {
		printHelper(lg.ZapLogger.Debug, fmt.Sprintf(format, args...), ctx)
	}
}

// Info record info not format
func (lg *Logger) Info(args ...interface{}) {
	lg.InfoWithCtx(nil, args...)
}

// Infof record info
func (lg *Logger) Infof(format string, args ...interface{}) {
	lg.InfofWithCtx(nil, format, args...)
}

// InfoWithCtx record Info not format with context, if you have no ctx, please use the method with not ctx
func (lg *Logger) InfoWithCtx(ctx context.Context, args ...interface{}) {
	if lg.validate() {
		printHelper(lg.ZapLogger.Info, fmt.Sprint(args...), ctx)
	}
}

// InfofWithCtx record Info  format with context, if you have no ctx, please use the method with not ctx
func (lg *Logger) InfofWithCtx(ctx context.Context, format string, args ...interface{}) {
	if lg.validate() {
		printHelper(lg.ZapLogger.Info, fmt.Sprintf(format, args...), ctx)
	}
}

// Warn record warn not format
func (lg *Logger) Warn(args ...interface{}) {
	lg.WarnWithCtx(nil, args...)
}

// Warnf record warn
func (lg *Logger) Warnf(format string, args ...interface{}) {
	lg.WarnfWithCtx(nil, format, args...)
}

// WarnWithCtx record Warn not format with context, if you have no ctx, please use the method with not ctx
func (lg *Logger) WarnWithCtx(ctx context.Context, args ...interface{}) {
	if lg.validate() {
		printHelper(lg.ZapLogger.Warn, fmt.Sprint(args...), ctx)
	}
}

// WarnfWithCtx record Warn  format with context, if you have no ctx, please use the method with not ctx
func (lg *Logger) WarnfWithCtx(ctx context.Context, format string, args ...interface{}) {
	if lg.validate() {
		printHelper(lg.ZapLogger.Warn, fmt.Sprintf(format, args...), ctx)
	}
}

// Error record error not format
func (lg *Logger) Error(args ...interface{}) {
	lg.ErrorWithCtx(nil, args...)
}

// Errorf record error
func (lg *Logger) Errorf(format string, args ...interface{}) {
	lg.ErrorfWithCtx(nil, format, args...)
}

// ErrorWithCtx record Error not format with context, if you have no ctx, please use the method with not ctx
func (lg *Logger) ErrorWithCtx(ctx context.Context, args ...interface{}) {
	if lg.validate() {
		printHelper(lg.ZapLogger.Error, fmt.Sprint(args...), ctx)
	}
}

// ErrorfWithCtx record Error  format with context, if you have no ctx, please use the method with not ctx
func (lg *Logger) ErrorfWithCtx(ctx context.Context, format string, args ...interface{}) {
	if lg.validate() {
		printHelper(lg.ZapLogger.Error, fmt.Sprintf(format, args...), ctx)
	}
}

// Dpanic record DPanic not format
func (lg *Logger) Dpanic(args ...interface{}) {
	lg.DPanicWithCtx(nil, args...)
}

// Dpanicf record DPanic
func (lg *Logger) Dpanicf(format string, args ...interface{}) {
	lg.DPanicfWithCtx(nil, format, args...)
}

// DPanicWithCtx record DPanic not format with context, if you have no ctx, please use the method with not ctx
func (lg *Logger) DPanicWithCtx(ctx context.Context, args ...interface{}) {
	if lg.validate() {
		printHelper(lg.ZapLogger.DPanic, fmt.Sprint(args...), ctx)
	}
}

// DPanicfWithCtx record DPanic  format with context, if you have no ctx, please use the method with not ctx
func (lg *Logger) DPanicfWithCtx(ctx context.Context, format string, args ...interface{}) {
	if lg.validate() {
		printHelper(lg.ZapLogger.DPanic, fmt.Sprintf(format, args...), ctx)
	}
}

// Panic record panic not format with context, if you have no ctx, please use the method with not ctx
func (lg *Logger) Panic(args ...interface{}) {
	lg.PanicWithCtx(nil, args...)
}

// Panicf record panic
func (lg *Logger) Panicf(format string, args ...interface{}) {
	lg.PanicfWithCtx(nil, format, args...)
}

// PanicWithCtx record panic not format with context, if you have no ctx, please use the method with not ctx
func (lg *Logger) PanicWithCtx(ctx context.Context, args ...interface{}) {
	if lg.validate() {
		printHelper(lg.ZapLogger.Panic, fmt.Sprint(args...), ctx)
	}
}

// PanicfWithCtx record panic  format with context, if you have no ctx, please use the method with not ctx
func (lg *Logger) PanicfWithCtx(ctx context.Context, format string, args ...interface{}) {
	if lg.validate() {
		printHelper(lg.ZapLogger.Panic, fmt.Sprintf(format, args...), ctx)
	}
}

// Fatal record fatal not format
func (lg *Logger) Fatal(args ...interface{}) {
	lg.FatalWithCtx(nil, args...)
}

// Fatalf record fatal
func (lg *Logger) Fatalf(format string, args ...interface{}) {
	lg.FatalfWithCtx(nil, format, args...)
}

// FatalWithCtx record fatal not format with context, if you have no ctx, please use the method with not ctx
func (lg *Logger) FatalWithCtx(ctx context.Context, args ...interface{}) {
	if lg.validate() {
		printHelper(lg.ZapLogger.Fatal, fmt.Sprint(args...), ctx)
	}
}

// FatalfWithCtx record fatal  format with context, if you have no ctx, please use the method with not ctx
func (lg *Logger) FatalfWithCtx(ctx context.Context, format string, args ...interface{}) {
	if lg.validate() {
		printHelper(lg.ZapLogger.Fatal, fmt.Sprintf(format, args...), ctx)
	}
}

func (lg *Logger) validate() bool {
	if lg.ZapLogger == nil {
		fmt.Println("Fatal function's logger is nil")
		return false
	}
	return true
}
