//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package devmanager this for device driver manager mock
package devmanager

import (
	"huawei.com/npu-exporter/devmanager/common"
)

// DeviceManagerMock common device manager mock for Ascend910/310P/310
type DeviceManagerMock struct {
}

// Init load symbol and initialize dcmi
func (d *DeviceManagerMock) Init() error {
	return nil
}

// ShutDown clean the dynamically loaded resource
func (d *DeviceManagerMock) ShutDown() error {
	return nil
}

// GetDevType return mock type
func (d *DeviceManagerMock) GetDevType() string {
	return common.Ascend910
}

// GetDeviceCount get npu device count
func (d *DeviceManagerMock) GetDeviceCount() (int32, error) {
	return 1, nil
}

// GetCardList  get all card list
func (d *DeviceManagerMock) GetCardList() (int32, []int32, error) {
	return 1, []int32{0}, nil
}

// GetDeviceNumInCard  get all device list in one card
func (d *DeviceManagerMock) GetDeviceNumInCard(cardID int32) (int32, error) {
	return 1, nil
}

// GetDeviceList get all device logicID list
func (d *DeviceManagerMock) GetDeviceList() (int32, []int32, error) {
	return 1, []int32{0}, nil
}

// GetDeviceHealth query npu device health status
func (d *DeviceManagerMock) GetDeviceHealth(logicID int32) (uint32, error) {
	return 0, nil
}

// GetDeviceNetWorkHealth query npu device network health status
func (d *DeviceManagerMock) GetDeviceNetWorkHealth(logicID int32) (uint32, error) {
	return 0, nil
}

// GetDeviceUtilizationRate get npu device utilization
func (d *DeviceManagerMock) GetDeviceUtilizationRate(logicID int32, deviceType common.DeviceType) (uint32, error) {
	return 1, nil
}

// GetDeviceTemperature get npu device temperature
func (d *DeviceManagerMock) GetDeviceTemperature(logicID int32) (int32, error) {
	return 1, nil
}

// GetDeviceVoltage get npu device voltage
func (d *DeviceManagerMock) GetDeviceVoltage(logicID int32) (float32, error) {
	return 1, nil
}

// GetDevicePowerInfo get npu device power info
func (d *DeviceManagerMock) GetDevicePowerInfo(logicID int32) (float32, error) {
	return 1, nil
}

// GetDeviceFrequency get npu device work frequency
func (d *DeviceManagerMock) GetDeviceFrequency(logicID int32, deviceType common.DeviceType) (int32, error) {
	return 1, nil
}

// GetDeviceMemoryInfo get npu memory information
func (d *DeviceManagerMock) GetDeviceMemoryInfo(logicID int32) (*common.MemoryInfo, error) {
	return &common.MemoryInfo{
		MemorySize:      1,
		MemoryAvailable: 1,
		Frequency:       1,
		Utilization:     1,
	}, nil
}

// GetDeviceHbmInfo get npu HBM module memory and frequency information
func (d *DeviceManagerMock) GetDeviceHbmInfo(logicID int32) (*common.HbmInfo, error) {
	return &common.HbmInfo{
		MemorySize:        1,
		Frequency:         1,
		Usage:             1,
		Temp:              1,
		BandWidthUtilRate: 1,
	}, nil
}

// GetDeviceErrorCode get npu device error code
func (d *DeviceManagerMock) GetDeviceErrorCode(logicID int32) (int32, int64, error) {
	return int32(0), int64(0), nil
}

// GetChipInfo get npu device error code
func (d *DeviceManagerMock) GetChipInfo(logicID int32) (*common.ChipInfo, error) {
	chip := &common.ChipInfo{
		Type:    "ascend",
		Name:    "910",
		Version: "v1",
	}
	return chip, nil
}

// GetPhysicIDFromLogicID get device physic id from logic id
func (d *DeviceManagerMock) GetPhysicIDFromLogicID(logicID int32) (int32, error) {
	return 1, nil
}

// GetLogicIDFromPhysicID get device logic id from physic id
func (d *DeviceManagerMock) GetLogicIDFromPhysicID(physicID int32) (int32, error) {
	return 1, nil
}

// GetDeviceLogicID get device logic id from card id and device id
func (d *DeviceManagerMock) GetDeviceLogicID(cardID, deviceID int32) (int32, error) {
	return 1, nil
}

// GetDeviceIPAddress get device ip address
func (d *DeviceManagerMock) GetDeviceIPAddress(logicID int32) (string, error) {
	return "127.0.0.1", nil
}

// CreateVirtualDevice create virtual device
func (d *DeviceManagerMock) CreateVirtualDevice(logicID int32, vDevInfo common.CgoCreateVDevRes) (common.
	CgoCreateVDevOut, error) {
	return common.CgoCreateVDevOut{}, nil
}

// GetVirtualDeviceInfo get virtual device info
func (d *DeviceManagerMock) GetVirtualDeviceInfo(logicID int32) (common.VirtualDevInfo, error) {
	return common.VirtualDevInfo{}, nil
}

// DestroyVirtualDevice destroy virtual device
func (d *DeviceManagerMock) DestroyVirtualDevice(logicID int32, vDevID uint32) error {
	return nil
}

// GetMcuPowerInfo get mcu power info for cardID
func (d *DeviceManagerMock) GetMcuPowerInfo(cardID int32) (float32, error) {
	return 1, nil
}

// GetCardIDDeviceID get cardID and deviceID by logicID
func (d *DeviceManagerMock) GetCardIDDeviceID(logicID int32) (int32, int32, error) {
	return 0, 0, nil
}
