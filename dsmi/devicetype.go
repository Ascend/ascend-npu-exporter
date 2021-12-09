//  Copyright(C) 2020. Huawei Technologies Co.,Ltd. All rights reserved.

// Package dsmi interface
package dsmi

// DeviceType devive type enum
type DeviceType int32

const (
	// Memory  Ascend310 & Ascend910
	Memory DeviceType = 1
	// AICore Ascend310 & Ascend910
	AICore DeviceType = 2
	// AICPU  Ascend310 & Ascend910
	AICPU DeviceType = 3
	// CTRLCPU  Ascend310 & Ascend910
	CTRLCPU DeviceType = 4
	// MEMBandWidth memory brandwidth  Ascend310 & Ascend910
	MEMBandWidth DeviceType = 5
	// HBM             Ascend910 only
	HBM DeviceType = 6
	// AICoreCurrentFreq AI core current frequency  Ascend910 only
	AICoreCurrentFreq DeviceType = 7
	// AICoreNormalFreq AI core normal frequency  Ascend910 only
	AICoreNormalFreq DeviceType = 9
	// HBMBandWidth Ascend910 only
	HBMBandWidth DeviceType = 10
)
