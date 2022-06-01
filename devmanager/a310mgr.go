//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package devmanager this Ascend310 device manager
package devmanager

import "huawei.com/npu-exporter/devmanager/dcmi"

// A310Manager Ascend310 device manager
type A310Manager struct {
	dcmi.DcManager
}
