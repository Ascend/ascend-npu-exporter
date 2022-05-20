//  Copyright(C) 2020. Huawei Technologies Co.,Ltd. All rights reserved.

// Package dsmi interface
package dsmi

// ChipType chip type enum
type ChipType string

const (
	// HiAIMaxDeviceNum the max device num
	HiAIMaxDeviceNum = 64
	// HIAIMaxCardNum the max card num
	HIAIMaxCardNum = 16
	// Ascend910 Enum
	Ascend910 ChipType = "Ascend910"
	// Ascend310P chip type enum
	Ascend310P ChipType = "Ascend310P"
	// Ascend310 chip type enum
	Ascend310 ChipType = "Ascend310"
	// DefaultErrorValue default error value
	DefaultErrorValue = -1
)
