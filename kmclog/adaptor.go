//  Copyright(C) 2021. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package kmclog adapts to different logger
package kmclog

import (
	"huawei.com/npu-exporter/hwlog"
)

// KmcLoggerAdaptor is used to adapt to the log module of the KMC.
// it implements the CryptoLogger interface of KMC.
// it will invoke the method of hwlog .
type KmcLoggerAdaptor struct {
}

// Error print error log
func (kla *KmcLoggerAdaptor) Error(msg string) {
	hwlog.RunLog.Error(msg)
}

// Warn print warning log
func (kla *KmcLoggerAdaptor) Warn(msg string) {
	hwlog.RunLog.Warn(msg)
}

// Info print info log
func (kla *KmcLoggerAdaptor) Info(msg string) {
	hwlog.RunLog.Info(msg)
}

// Debug print debug log
func (kla *KmcLoggerAdaptor) Debug(msg string) {
	hwlog.RunLog.Debug(msg)
}

// Trace print trace log
func (kla *KmcLoggerAdaptor) Trace(msg string) {
	hwlog.RunLog.Debug(msg)
}

// Log print log
func (kla *KmcLoggerAdaptor) Log(msg string) {
	hwlog.RunLog.Info(msg)
}
