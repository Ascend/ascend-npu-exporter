//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package devmanager this for device model
package devmanager

// CgoDcmiCreateVDevOut create virtual device info
type CgoDcmiCreateVDevOut struct {
	VDevID     uint32
	PcieBus    uint32
	PcieDevice uint32
	PcieFunc   uint32
	VfgID      uint32
	Reserved   []uint8
}

// CgoDcmiBaseResource base resource info
type CgoDcmiBaseResource struct {
	token       uint64
	tokenMax    uint64
	taskTimeout uint64
	vfgID       uint32
	vipMode     uint8
	reserved    []uint8
}

// CgoDcmiComputingResource compute resource info
type CgoDcmiComputingResource struct {
	// accelator resource
	aic     float32
	aiv     float32
	dsa     uint16
	rtsq    uint16
	acsq    uint16
	cdqm    uint16
	cCore   uint16
	ffts    uint16
	sdma    uint16
	pcieDma uint16

	// memory resource, MB as unit
	memorySize uint64

	// id resource
	eventID  uint32
	notifyID uint32
	streamID uint32
	modelID  uint32

	// cpu resource
	topicScheduleAicpu uint16
	hostCtrlCPU        uint16
	hostAicpu          uint16
	deviceAicpu        uint16
	topicCtrlCPUSlot   uint16

	reserved []uint8
}

// CgoDcmiMediaResource media resource info
type CgoDcmiMediaResource struct {
	jpegd    float32
	jpege    float32
	vpc      float32
	vdec     float32
	pngd     float32
	venc     float32
	reserved []uint8
}

// CgoVDevQueryInfo virtual resource special info
type CgoVDevQueryInfo struct {
	name            string
	status          uint32
	isContainerUsed uint32
	vfid            uint32
	vfgID           uint32
	containerID     uint64
	base            CgoDcmiBaseResource
	computing       CgoDcmiComputingResource
	media           CgoDcmiMediaResource
}

// CgoVDevQueryStru virtual resource info
type CgoVDevQueryStru struct {
	vDevID    uint32
	queryInfo CgoVDevQueryInfo
}

// CgoDcmiSocFreeResource soc free resource info
type CgoDcmiSocFreeResource struct {
	vfgNum    uint32
	vfgBitmap uint32
	base      CgoDcmiBaseResource
	computing CgoDcmiComputingResource
	media     CgoDcmiMediaResource
}

// CgoDcmiSocTotalResource soc total resource info
type CgoDcmiSocTotalResource struct {
	vDevNum   uint32
	vDevID    []uint32
	vfgNum    uint32
	vfgBitmap uint32
	base      CgoDcmiBaseResource
	computing CgoDcmiComputingResource
	media     CgoDcmiMediaResource
}

// CgoVDevInfo virtual device infos
type CgoVDevInfo struct {
	VDevNum       uint32    // number of virtual devices
	CoreNumUnused float32   // number of unused cores
	Status        []uint32  // status of virtual devices
	VDevID        []uint32  // id of virtual devices
	VfID          []uint32  // id
	CID           []uint64  // container id
	CoreNum       []float32 // aicore num for virtual device
}

// dcmi model end

// MemoryInfo memory information struct
type MemoryInfo struct {
	MemorySize  uint64 `json:"memory_size"` // MemorySize unit:MB
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
