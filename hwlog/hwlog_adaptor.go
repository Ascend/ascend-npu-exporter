//  Copyright(c) 2021. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package hwlog provides the capability of processing Huawei log rules.
package hwlog

import "fmt"

// RunLog run logger
var RunLog *Logger
// OpLog operate logger
var OpLog *Logger
// SecLog security logger
var SecLog *Logger

// InitRunLogger initialize run logger
func InitRunLogger (config *LogConfig, stopCh <-chan struct{}) error {
	if RunLog != nil && RunLog.ZapLogger != nil {
		RunLog.Warn("run logger is been initialized.")
		return nil
	}
	RunLog = NewLogger()
	err := RunLog.InitLogger(config, stopCh)
	if err != nil {
		return err
	}
	if !RunLog.IsInit() {
		return fmt.Errorf("run logger not init")
	}
	return nil
}

// InitOperateLogger initialize operate logger
func InitOperateLogger (config *LogConfig, stopCh <-chan struct{}) error {
	if OpLog != nil && OpLog.ZapLogger != nil {
		OpLog.Warn("operate logger is been initialized.")
		return nil
	}
	OpLog = NewLogger()
	err := OpLog.InitLogger(config, stopCh)
	if err != nil {
		return err
	}
	if !OpLog.IsInit() {
		return fmt.Errorf("op logger not init")
	}
	return nil
}

// InitSecurityLogger initialize security logger
func InitSecurityLogger (config *LogConfig, stopCh <-chan struct{}) error {
	if SecLog != nil && SecLog.ZapLogger != nil {
		SecLog.Warn("security logger is been initialized.")
		return nil
	}
	SecLog = NewLogger()
	err := SecLog.InitLogger(config, stopCh)
	if err != nil {
		return err
	}
	if !SecLog.IsInit() {
		return fmt.Errorf("security logger not init")
	}
	return nil
}