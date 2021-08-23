// Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.

// Package container for monitoring containers' npu allocation
package container

// NameFetcher fetches name for the given-ID container under different situations
type NameFetcher interface {
	Init() error
	Name(id string) string
	Close() error
}
