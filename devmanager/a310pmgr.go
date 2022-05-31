//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package devmanager this Ascend310P device manager
package devmanager

import "huawei.com/npu-exporter/devmanager/dcmi"

// A310PManager Ascend310P device manager
type A310PManager struct {
	dcmi.DcManager
}
