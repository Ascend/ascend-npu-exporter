//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package devmanager this Ascend310P device manager
package devmanager

import (
	"huawei.com/npu-exporter/devmanager/dcmi"
)

// A310PManager Ascend310P device manager
type A310PManager struct {
	dcmi.DcManager
}

// DcGetMcuPowerInfo this function is only for Ascend310P
func (d *A310PManager) DcGetMcuPowerInfo(cardID int32) (float32, error) {
	return dcmi.FuncDcmiMcuGetPowerInfo(cardID)
}
