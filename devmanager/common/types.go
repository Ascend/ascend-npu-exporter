/* Copyright(C) 2021-2023. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package common define common types
package common

// MemoryInfo memory information struct
type MemoryInfo struct {
	MemorySize      uint64 `json:"memory_size"`
	MemoryAvailable uint64 `json:"memory_available"`
	Frequency       uint32 `json:"memory_frequency"`
	Utilization     uint32 `json:"memory_utilization"`
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

// CgoCreateVDevOut create virtual device output info
type CgoCreateVDevOut struct {
	VDevID     uint32
	PcieBus    uint32
	PcieDevice uint32
	PcieFunc   uint32
	VfgID      uint32
	Reserved   []uint8
}

// CgoCreateVDevRes create virtual device input info
type CgoCreateVDevRes struct {
	VDevID       uint32
	VfgID        uint32
	TemplateName string
	Reserved     []uint8
}

// CgoBaseResource base resource info
type CgoBaseResource struct {
	Token       uint64
	TokenMax    uint64
	TaskTimeout uint64
	VfgID       uint32
	VipMode     uint8
	Reserved    []uint8
}

// CgoComputingResource compute resource info
type CgoComputingResource struct {
	// accelator resource
	Aic     float32
	Aiv     float32
	Dsa     uint16
	Rtsq    uint16
	Acsq    uint16
	Cdqm    uint16
	CCore   uint16
	Ffts    uint16
	Sdma    uint16
	PcieDma uint16

	// memory resource, MB as unit
	MemorySize uint64

	// id resource
	EventID  uint32
	NotifyID uint32
	StreamID uint32
	ModelID  uint32

	// cpu resource
	TopicScheduleAicpu uint16
	HostCtrlCPU        uint16
	HostAicpu          uint16
	DeviceAicpu        uint16
	TopicCtrlCPUSlot   uint16

	Reserved []uint8
}

// CgoMediaResource media resource info
type CgoMediaResource struct {
	Jpegd    float32
	Jpege    float32
	Vpc      float32
	Vdec     float32
	Pngd     float32
	Venc     float32
	Reserved []uint8
}

// CgoVDevQueryInfo virtual resource special info
type CgoVDevQueryInfo struct {
	Name            string
	Status          uint32
	IsContainerUsed uint32
	Vfid            uint32
	VfgID           uint32
	ContainerID     uint64
	Base            CgoBaseResource
	Computing       CgoComputingResource
	Media           CgoMediaResource
}

// CgoVDevQueryStru virtual resource info
type CgoVDevQueryStru struct {
	VDevID    uint32
	QueryInfo CgoVDevQueryInfo
}

// CgoSocFreeResource soc free resource info
type CgoSocFreeResource struct {
	VfgNum    uint32
	VfgBitmap uint32
	Base      CgoBaseResource
	Computing CgoComputingResource
	Media     CgoMediaResource
}

// CgoSocTotalResource soc total resource info
type CgoSocTotalResource struct {
	VDevNum   uint32
	VDevID    []uint32
	VfgNum    uint32
	VfgBitmap uint32
	Base      CgoBaseResource
	Computing CgoComputingResource
	Media     CgoMediaResource
}

// VirtualDevInfo virtual device infos
type VirtualDevInfo struct {
	TotalResource CgoSocTotalResource
	FreeResource  CgoSocFreeResource
	VDevInfo      []CgoVDevQueryStru
}

// DevFaultInfo device's fault info
type DevFaultInfo struct {
	EventID         int64
	LogicID         int32
	Severity        int8
	Assertion       int8
	AlarmRaisedTime int64
}

// DevProcessInfo device process info
type DevProcessInfo struct {
	DevProcArray []DevProcInfo
	ProcNum      int32
}

// DevProcInfo process info in device side
type DevProcInfo struct {
	Pid int32
	// the total amount of memory occupied by the device side OS and allocated by the business, unit is MB
	MemUsage float64
}
