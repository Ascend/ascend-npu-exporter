//  Copyright(C) 2021. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package hwlog provides the capability of processing Huawei log rules.
package hwlog

import (
	"bytes"
	"fmt"
	"go.uber.org/zap"
	"runtime"
	"strings"
)

// printHelper helper function for log printing
func printHelper(f func(string, ...zap.Field), msg string) {
	str := getCallerInfo()
	f(str + msg)
}

// getCallerInfo gets the caller's information
func getCallerInfo() string {
	var funcName string
	pc, codePath, codeLine, ok := runtime.Caller(stackDeep)
	if ok {
		funcName = runtime.FuncForPC(pc).Name()
	}
	p := strings.Split(codePath, "/")
	l := len(p)
	if l == pathLen {
		funcName = p[l-1]
	} else if l > pathLen {
		funcName = fmt.Sprintf("%s/%s", p[l-pathLen], p[l-1])
	}
	callerPath := fmt.Sprintf("%s:%d", funcName, codeLine)
	goroutineID := getGoroutineID()
	str := fmt.Sprintf("%-8s%s    ", goroutineID, callerPath)
	return str
}

// getCallerGoroutineID gets the goroutineID
func getGoroutineID() string {
	b := make([]byte, bitsize, bitsize)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	return string(b)
}
