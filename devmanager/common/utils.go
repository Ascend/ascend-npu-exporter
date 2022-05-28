//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package common this for util method
package common

import "math"

// IsValidUtilizationRate valid utilization rate is 0-100
func IsValidUtilizationRate(num uint32) bool {
	if num > uint32(Percent) || num < 0 {
		return false
	}

	return true
}

// IsGreaterThanOrEqualInt32 valid int32 max
func IsGreaterThanOrEqualInt32(num int64) bool {
	if num >= int64(math.MaxInt32) {
		return true
	}

	return false
}
