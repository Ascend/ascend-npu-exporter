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
