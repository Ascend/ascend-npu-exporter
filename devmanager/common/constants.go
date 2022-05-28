//  Copyright(C) 2020. Huawei Technologies Co.,Ltd. All rights reserved.

// Package common define common variable
package common

// DeviceType device type enum
type DeviceType int32

const (
	// Percent constant of 100
	Percent = 100
	// RetryTime call func retry times
	RetryTime = 3
	// HiAIMaxDeviceNum the max device num
	HiAIMaxDeviceNum = 64
	// MaxChipNum max chip num
	MaxChipNum = 64
	// DefaultTemperatureWhenQueryFailed when get temperature failed, use this value
	DefaultTemperatureWhenQueryFailed = -275
)
