//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package devmanager this for device driver manager
package devmanager

import (
	"huawei.com/npu-exporter/devmanager/common"
)

// DeviceInterface for common device interface
type DeviceInterface interface {
	Init() error
	ShutDown() error
	GetDeviceCount() (uint32, error)
	GetDeviceHealth(logicID int32) (uint32, error)
	GetDeviceNetWorkHealth(logicID int32) (uint32, error)
	GetDeviceUtilizationRate(logicID int32, deviceType common.DeviceType) (uint32, error)
	GetDeviceTemperature(logicID int32) (int32, error)
	GetDeviceVoltage(logicID int32) (float32, error)
	GetDevicePowerInfo(logicID int32) (float32, error)
	GetDeviceFrequency(logicID int32, deviceType common.DeviceType) (uint32, error)
	GetDeviceMemoryInfo(logicID int32) (common.MemoryInfo, error)
	GetDeviceHbmInfo(logicID int32) (common.HbmInfo, error)
	GetDeviceErrorCode(logicID int32) (int32, int64, error)
	GetChipInfo(logicID int32) (common.ChipInfo, error)
	GetPhysicIDFromLogicID(logicID uint32) (uint32, error)
	GetLogicIDFromPhysicID(physicID uint32) (uint32, error)
	GetDeviceLogicID(cardID, deviceID int32) (int32, error)
	GetDeviceIPAddress(logicID int32) (string, error)
	CreateVirtualDevice(logicID, aiCore int32) (uint32, error)
	GetVirtualDeviceInfo(logicID int32) (common.VirtualDevInfo, error)
	DestroyVirtualDevice(logicID int32) error
}

// DeviceManager common device manager for Ascend910/310P/310
type DeviceManager struct{}

// Init load symbol and initialize dcmi or dsmi
func (d *DeviceManager) Init() error {
	return nil
}

// ShutDown clean the dynamically loaded resource
func (d *DeviceManager) ShutDown() error {
	return nil
}

// GetDeviceCount get npu device count
func (d *DeviceManager) GetDeviceCount() (uint32, error) {
	return 0, nil
}

// GetDeviceHealth query npu device health status
func (d *DeviceManager) GetDeviceHealth(logicID int32) (uint32, error) {
	return 0, nil
}

// GetDeviceNetWorkHealth query npu device network health status
func (d *DeviceManager) GetDeviceNetWorkHealth(logicID int32) (uint32, error) {
	return 0, nil
}

// GetDeviceUtilizationRate get npu device utilization
func (d *DeviceManager) GetDeviceUtilizationRate(logicID int32, deviceType common.DeviceType) (uint32, error) {
	return 0, nil
}

// GetDeviceTemperature get npu device temperature
func (d *DeviceManager) GetDeviceTemperature(logicID int32) (int32, error) {
	return 0, nil
}

// GetDeviceVoltage get npu device voltage
func (d *DeviceManager) GetDeviceVoltage(logicID int32) (float32, error) {
	return 0, nil
}

// GetDevicePowerInfo get npu device power info
func (d *DeviceManager) GetDevicePowerInfo(logicID int32) (float32, error) {
	return 0, nil
}

// GetDeviceFrequency get npu device work frequency
func (d *DeviceManager) GetDeviceFrequency(logicID int32, deviceType common.DeviceType) (uint32, error) {
	return 0, nil
}

// GetDeviceMemoryInfo get npu memory information
func (d *DeviceManager) GetDeviceMemoryInfo(logicID int32) (common.MemoryInfo, error) {
	return common.MemoryInfo{}, nil
}

// GetDeviceHbmInfo get npu HBM module memory and frequency information
func (d *DeviceManager) GetDeviceHbmInfo(logicID int32) (common.HbmInfo, error) {
	return common.HbmInfo{}, nil
}

// GetDeviceErrorCode get npu device error code
func (d *DeviceManager) GetDeviceErrorCode(logicID int32) (int32, int64, error) {
	return 0, 0, nil
}

// GetChipInfo get npu device error code
func (d *DeviceManager) GetChipInfo(logicID int32) (common.ChipInfo, error) {
	return common.ChipInfo{}, nil
}

// GetPhysicIDFromLogicID get device physic id from logic id
func (d *DeviceManager) GetPhysicIDFromLogicID(logicID uint32) (uint32, error) {
	return 0, nil
}

// GetLogicIDFromPhysicID get device logic id from physic id
func (d *DeviceManager) GetLogicIDFromPhysicID(physicID uint32) (uint32, error) {
	return 0, nil
}

// GetDeviceLogicID get device logic id from card id and device id
func (d *DeviceManager) GetDeviceLogicID(cardID, deviceID int32) (int32, error) {
	return 0, nil
}

// GetDeviceIPAddress get device ip address
func (d *DeviceManager) GetDeviceIPAddress(logicID int32) (string, error) {
	return "", nil
}

// CreateVirtualDevice create virtual device
func (d *DeviceManager) CreateVirtualDevice(logicID, aiCore int32) (uint32, error) {
	return 0, nil
}

// GetVirtualDeviceInfo get virtual device info
func (d *DeviceManager) GetVirtualDeviceInfo(logicID int32) (common.VirtualDevInfo, error) {
	return common.VirtualDevInfo{}, nil
}

// DestroyVirtualDevice destroy virtual device
func (d *DeviceManager) DestroyVirtualDevice(logicID int32) error {
	return nil
}
