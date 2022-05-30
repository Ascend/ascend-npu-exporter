//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package common this for util method
package common

import (
	"math"
)

// IsGreaterThanOrEqualInt32 check num range
func IsGreaterThanOrEqualInt32(num int64) bool {
	if num >= int64(math.MaxInt32) {
		return true
	}

	return false
}

// IsValidUtilizationRate valid utilization rate is 0-100
func IsValidUtilizationRate(num uint32) bool {
	if num > uint32(Percent) || num < 0 {
		return false
	}

	return true
}

// IsValidChipInfo valid chip info is or not empty
func IsValidChipInfo(chip *ChipInfo) bool {
	return chip.Name != "" || chip.Type != "" || chip.Version != ""
}

// IsValidCardID valid card id
func IsValidCardID(cardID int32) bool {
	return cardID >= 0 && cardID < HiAIMaxCardNum
}

// IsValidDeviceID valid device id
func IsValidDeviceID(deviceID int32) bool {
	return deviceID >= 0 && deviceID < HiAIMaxDeviceNum
}

// IsValidLogicID valid logic id
func IsValidLogicID(logicID uint32) bool {
	return logicID < HiAIMaxCardNum*HiAIMaxDeviceNum
}

// IsValidPhysicID valid physic id
func IsValidPhysicID(physicID uint32) bool {
	return physicID < HiAIMaxCardNum*HiAIMaxDeviceNum
}
