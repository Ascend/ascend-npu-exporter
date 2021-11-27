//  Copyright(C) 2021. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package hwlog provides the capability of processing Huawei log rules.
package hwlog

var (
	// BuildName build name
	BuildName string
	// BuildVersion build version
	BuildVersion string
)

// ContextKey especially for context value
// to solve problem of "should not use basic type untyped string as key in context.WithValue"
type ContextKey string

// String  the implement of String method
func (c ContextKey) String() string {
	return string(c)
}

const (
	// UserID used for context value key of "ID"
	UserID ContextKey = "UserID"
	// ReqID used for context value key of "requestID"
	ReqID ContextKey = "RequestID"
)
