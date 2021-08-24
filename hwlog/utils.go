//  Copyright(C) 2021. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package hwlog provides the capability of processing Huawei log rules.
package hwlog

import (
	"fmt"
	"go.uber.org/zap"
	"runtime/debug"
	"strconv"
	"strings"
)

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
	goroutineID := getGoroutineID(paths)

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
func getGoroutineID(paths []string) int {
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
