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

// Package hwlog adapts to different logger
package hwlog

// KmcLoggerApdaptor is used to adapt to the log module of the KMC.
// it implements the CryptoLogger interface of KMC.
// it will invoke the method of hwlog .
type KmcLoggerApdaptor struct {
}

// Error print error log
func (kla *KmcLoggerApdaptor) Error(msg string) {
	Error(msg)
}

// Warn print warning log
func (kla *KmcLoggerApdaptor) Warn(msg string) {
	Warn(msg)
}

// Info print info log
func (kla *KmcLoggerApdaptor) Info(msg string) {
	Info(msg)
}

// Debug print debug log
func (kla *KmcLoggerApdaptor) Debug(msg string) {
	Debug(msg)
}

// Trace print trace log
func (kla *KmcLoggerApdaptor) Trace(msg string) {
	Debug(msg)
}

// Log print log
func (kla *KmcLoggerApdaptor) Log(msg string) {
	Info(msg)
}
