//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package common define common types
package common

// MemoryInfo memory information struct
type MemoryInfo struct {
	MemorySize  uint64 `json:"memory_size"`
	Frequency   uint32 `json:"memory_frequency"`
	Utilization uint32 `json:"memory_utilization"`
}

// HbmInfo HBM info
type HbmInfo struct {
	MemorySize        uint64 `json:"memory_size"`        // HBM total size,KB
	Frequency         uint32 `json:"hbm_frequency"`      // HBM frequency MHz
	Usage             uint64 `json:"memory_usage"`       // HBM memory usage,KB
	Temp              int32  `json:"hbm_temperature"`    // HBM temperature
	BandWidthUtilRate uint32 `json:"hbm_bandwidth_util"` // HBM bandwidth utilization
}

// ChipInfo chip info
type ChipInfo struct {
	Type    string `json:"chip_type"`
	Name    string `json:"chip_name"`
	Version string `json:"chip_version"`
}

// CgoDsmiSubVDevInfo single VDevInfo info
type CgoDsmiSubVDevInfo struct {
	Status uint32
	VDevID uint32
	VfID   uint32
	CID    uint64
	Spec   CgoDsmiVdevSpecInfo
}

// CgoDsmiVdevSpecInfo is special info
type CgoDsmiVdevSpecInfo struct {
	CoreNum  string
	Reserved string
}

// VirtualDevInfo virtual device infos
type VirtualDevInfo struct {
	VDevNum             uint32
	CoreNumUnused       uint32
	CoreCount           uint32
	CgoDsmiSubVDevInfos []CgoDsmiSubVDevInfo
}
