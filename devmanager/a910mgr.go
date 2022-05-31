//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package devmanager this Ascend910 device manager
package devmanager

import "huawei.com/npu-exporter/devmanager/dcmi"

// A910Manager Ascend910 device manager
type A910Manager struct {
	dcmi.DcManager
}
