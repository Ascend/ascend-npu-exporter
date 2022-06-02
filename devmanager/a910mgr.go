//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package devmanager this Ascend910 device manager
package devmanager

import (
	"huawei.com/npu-exporter/devmanager/common"
	"huawei.com/npu-exporter/devmanager/dcmi"
)

// A910Manager Ascend910 device manager
type A910Manager struct {
	dcmi.DcManager
}

// DcGetHbmInfo get HBM information, only for Ascend910
func (d *A910Manager) DcGetHbmInfo(cardID, deviceID int32) (*common.HbmInfo, error) {
	return dcmi.FuncDcmiGetDeviceHbmInfo(cardID, deviceID)
}
