//  Copyright(C) 2020. Huawei Technologies Co.,Ltd. All rights reserved.

// Package collector for Prometheus
package collector

import (
	"huawei.com/npu-exporter/dsmi"
	"time"
)

// HealthEnum enum
type HealthEnum string

const (
	// Healthy status of  Health
	Healthy HealthEnum = "Healthy"
	// UnHealthy status of unhealth
	UnHealthy HealthEnum = "UnHealthy"
	// convert base
	base = 10
	// log level
	five = 5
	// cache key
	key = "npu-exporter-npu-list"
	// cache key for parsing-device result
	containersDevicesInfoKey = "npu-exporter-containers-devices"
)

// HuaWeiAIChip chip info
type HuaWeiAIChip struct {
	// the memoryInfo of the chip
	Meminf *dsmi.MemoryInfo `json:"memory_info"`
	// the chip info
	ChipIfo *dsmi.ChipInfo `json:"chip_info"`
	// the hbm info
	HbmInfo *dsmi.HbmInfo `json:"hbm_info"`
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
