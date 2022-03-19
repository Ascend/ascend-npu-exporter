//  Copyright(C) 2021. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package kmclog adapts to different logger
package kmclog

import (
	"huawei.com/npu-exporter/hwlog"
)

// LoggerAdaptor is used to adapt to the log module of the KMC.
// it implements the CryptoLogger interface of KMC.
// it will invoke the method of hwlog .
type LoggerAdaptor struct {
}

// Error print error log
func (kla *LoggerAdaptor) Error(msg string) {
	hwlog.RunLog.Error(msg)
}

// Warn print warning log
func (kla *LoggerAdaptor) Warn(msg string) {
	hwlog.RunLog.Warn(msg)
}

// Info print info log
func (kla *LoggerAdaptor) Info(msg string) {
	hwlog.RunLog.Info(msg)
}

// Debug print debug log
func (kla *LoggerAdaptor) Debug(msg string) {
	hwlog.RunLog.Debug(msg)
}

// Trace print trace log
func (kla *LoggerAdaptor) Trace(msg string) {
	hwlog.RunLog.Debug(msg)
}

// Log print log
func (kla *LoggerAdaptor) Log(msg string) {
	hwlog.RunLog.Info(msg)
}
