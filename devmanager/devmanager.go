//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package devmanager this for device driver manager
package devmanager

import (
	"huawei.com/npu-exporter/devmanager/common"
)

// Init load symbol and initialize dcmi or dsmi
func Init() error {
	return nil
}

// ShutDown clean the dynamically loaded resource
func ShutDown() {}

// GetDeviceCount get npu device count
func GetDeviceCount() (uint32, error) {
	return 0, nil
}

// GetDeviceHealth query npu device health status
func GetDeviceHealth(logicID int32) (uint32, error) {
	return 0, nil
}

// GetDeviceNetWorkHealth query npu device network health status
func GetDeviceNetWorkHealth(logicID int32) (uint32, error) {
	return 0, nil
}

// GetDeviceUtilizationRate get npu device utilization
func GetDeviceUtilizationRate(logicID int32, deviceType common.DeviceType) (uint32, error) {
	return 0, nil
}

// GetDeviceTemperature get npu device temperature
func GetDeviceTemperature(logicID int32) (int32, error) {
	return 0, nil
}

// GetDeviceVoltage get npu device voltage
func GetDeviceVoltage(logicID int32) (float32, error) {
	return 0, nil
}

// GetDevicePowerInfo get npu device power info
func GetDevicePowerInfo(logicID int32) (float32, error) {
	return 0, nil
}

// GetDeviceFrequency get npu device work frequency
func GetDeviceFrequency(logicID int32, deviceType common.DeviceType) (uint32, error) {
	return 0, nil
}

// GetDeviceMemoryInfo get npu memory information
func GetDeviceMemoryInfo(logicID int32) (common.MemoryInfo, error) {
	return common.MemoryInfo{}, nil
}

// GetDeviceHbmInfo get npu HBM module memory and frequency information
func GetDeviceHbmInfo(logicID int32) (common.HbmInfo, error) {
	return common.HbmInfo{}, nil
}

// GetDeviceErrorCode get npu device error code
func GetDeviceErrorCode(logicID int32) (int32, int64, error) {
	return 0, 0, nil
}

// GetChipInfo get npu device error code
func GetChipInfo(logicID int32) (common.ChipInfo, error) {
	return common.ChipInfo{}, nil
}

// GetPhysicIDFromLogicID get device physic id from logic id
func GetPhysicIDFromLogicID(logicID uint32) (uint32, error) {
	return 0, nil
}

// GetLogicIDFromPhysicID get device logic id from physic id
func GetLogicIDFromPhysicID(physicID uint32) (uint32, error) {
	return 0, nil
}

// GetDeviceLogicID get device logic id from card id and device id
func GetDeviceLogicID(cardID, deviceID int32) (int32, error) {
	return 0, nil
}

// GetDeviceIPAddress get device ip address
func GetDeviceIPAddress(logicID int32) (string, error) {
	return "", nil
}

// CreateVirtualDevice create virtual device
func CreateVirtualDevice(logicID, aiCore int32) (uint32, error) {
	return 0, nil
}

// GetVirtualDeviceInfo get virtual device info
func GetVirtualDeviceInfo(logicID int32) (common.VirtualDevInfo, error) {
	return common.VirtualDevInfo{}, nil
}

// DestroyVirtualDevice destroy virtual device
func DestroyVirtualDevice(logicID int32) error {
	return nil
}
