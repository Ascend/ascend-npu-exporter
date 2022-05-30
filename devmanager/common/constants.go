//  Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.

// Package common define common variable
package common

// DeviceType device type enum
type DeviceType int32

const (
	// Memory  Ascend310 & Ascend910
	Memory DeviceType = iota + 1
	// AICore Ascend310 & Ascend910
	AICore
	// AICPU  Ascend310 & Ascend910
	AICPU
	// CTRLCPU  Ascend310 & Ascend910
	CTRLCPU
	// MEMBandWidth memory bandwidth Ascend310 & Ascend910
	MEMBandWidth
	// HBM Ascend910 only
	HBM
	// AICoreCurrentFreq AI core current frequency Ascend910 only
	AICoreCurrentFreq
	// DDR now is not supported
	DDR
	// AICoreNormalFreq AI core normal frequency Ascend910 only
	AICoreNormalFreq
	// HBMBandWidth Ascend910 only
	HBMBandWidth
	// VectorCore now is not supported
	VectorCore DeviceType = 12
)

const (
	// Success for interface return code
	Success = 0
	// RetError return error when the function failed
	RetError = -1
	// Percent constant of 100
	Percent = 100
	// FuncNotFound is not found in interface
	FuncNotFound = -99998
	// MaxErrorCodeCount number of error codes
	MaxErrorCodeCount = 128
	// UnRetError return unsigned int error
	UnRetError = 100
	// OneKilo for unit change kb to mb
	OneKilo = 1024

	// HiAIMaxCardNum max card number
	HiAIMaxCardNum = 16

	// NpuType present npu chip
	NpuType = 0

	// AiCoreNum1 1 ai core
	AiCoreNum1 = 1
	// AiCoreNum2 2 ai core
	AiCoreNum2 = 2
	// AiCoreNum4 4 ai core
	AiCoreNum4 = 4
	// AiCoreNum8 8 ai core
	AiCoreNum8 = 8
	// AiCoreNum16 16 ai core
	AiCoreNum16 = 16

	// ReduceOnePercent for calculation reduce one percent
	ReduceOnePercent = 0.01
	// ReduceTenth for calculation reduce one tenth
	ReduceTenth = 0.1
	// DefaultTemperatureWhenQueryFailed when get temperature failed, use this value
	DefaultTemperatureWhenQueryFailed = -275
	// MaxChipNum chip number max value
	MaxChipNum = 64
)
