/* Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package collector for Prometheus
package collector

import (
	"time"

	"huawei.com/npu-exporter/v5/devmanager/common"
)

// HealthEnum enum
type HealthEnum string

const (
	// Healthy status of  Health
	Healthy HealthEnum = "Healthy"
	// UnHealthy status of unhealth
	UnHealthy HealthEnum = "UnHealthy"
	// convert base
	base             = 10
	containerNameLen = 3
	// cache key
	key = "npu-exporter-npu-list"
	// cache key for parsing-device result
	containersDevicesInfoKey = "npu-exporter-containers-devices"
	initSize                 = 8
)

// HuaWeiAIChip chip info
type HuaWeiAIChip struct {
	// the memoryInfo of the chip
	Meminf *common.MemoryInfo `json:"memory_info"`
	// the chip info
	ChipIfo *common.ChipInfo `json:"chip_info"`
	// the hbm info
	HbmInfo *common.HbmInfo `json:"hbm_info"`
	// the healthy status of the  AI chip
	HealthStatus HealthEnum `json:"health_status"`
	// the error code of the chip
	ErrorCode int64 `json:"error_code"`
	// the utiliaiton of the chip
	Utilization int `json:"utilization"`
	// the temperature of the chip
	Temperature int `json:"temperature"`
	// the work power of the chip
	Power float32 `json:"power"`
	// the work voltage of the chip
	Voltage float32 `json:"voltage"`
	// the AI core frequency of the chip
	Frequency int `json:"frequency"`
	// the chip physic ID
	DeviceID int `json:"device_id"`
}

// HuaWeiNPUCard device
type HuaWeiNPUCard struct {
	// The chip list on the card
	DeviceList []*HuaWeiAIChip `json:"device_list"`
	// Timestamp
	Timestamp time.Time `json:"timestamp"`
	// The id of the NPU card
	CardID int `json:"card_id"`
}
