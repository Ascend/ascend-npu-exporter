//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package common this for util method
package common

import "math"

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
