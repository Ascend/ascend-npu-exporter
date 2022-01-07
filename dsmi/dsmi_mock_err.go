//  Copyright(C) 2020. Huawei Technologies Co.,Ltd. All rights reserved.

// Package dsmi interface
package dsmi

import "fmt"

// DeviceManagerMockErr  struct definition
type DeviceManagerMockErr struct {
}

var errorMsg = "mock error"

// NewDeviceManagerMockErr new DeviceManagerMockErr instance
func NewDeviceManagerMockErr() *DeviceManagerMockErr {
	return &DeviceManagerMockErr{}
}

// GetDeviceCount get ascend910 device quantity
func (d *DeviceManagerMockErr) GetDeviceCount() (int32, error) {

	return 0, fmt.Errorf(errorMsg)
}

// GetDeviceList  get device list
func (d *DeviceManagerMockErr) GetDeviceList() (int32, []int32, error) {
	return 0, []int32{}, fmt.Errorf(errorMsg)
}

// GetDeviceHealth get device health by id
func (d *DeviceManagerMockErr) GetDeviceHealth(logicID int32) (int32, error) {
	return int32(0), fmt.Errorf(errorMsg)

}

// GetDeviceUtilizationRate get device utils rate by id
// DeviceType  Ascend910 1,2,3,4,5,6,10  Ascend310 1,2,3,4,5
func (d *DeviceManagerMockErr) GetDeviceUtilizationRate(logicID int32, deviceType DeviceType) (int32, error) {
	return int32(0), fmt.Errorf(errorMsg)
}

// GetDeviceTemperature get the device temperature
func (d *DeviceManagerMockErr) GetDeviceTemperature(logicID int32) (int32, error) {
	return int32(0), fmt.Errorf(errorMsg)
}

// GetDeviceVoltage get the device voltage
func (d *DeviceManagerMockErr) GetDeviceVoltage(logicID int32) (float32, error) {
	return 0.00025, fmt.Errorf(errorMsg)
}

// GetDevicePower get the power info of the device, the result like : 8.2w
func (d *DeviceManagerMockErr) GetDevicePower(logicID int32) (float32, error) {
	return 0.0007, fmt.Errorf(errorMsg)

}

// GetDeviceFrequency get device frequency, unit MHz
// Ascend910 1,6,7,9
// Ascend310 1,2,3,4,5
// subType enum:  Memory,6HBM,AI_Core_Current_Fre,AI_Core_Normal_Fre(1,6,7,9)    see DeviceType
func (d *DeviceManagerMockErr) GetDeviceFrequency(logicID int32, subType DeviceType) (int32, error) {
	return int32(0), fmt.Errorf(errorMsg)
}

// createMemoryInfoObj create Memory information object
func (d *DeviceManagerMockErr) createMemoryInfoObj(cmInfo *CStructDsmiMemoryInfo) *MemoryInfo {
	return nil
}

// GetDeviceMemoryInfo get memory information
func (d *DeviceManagerMockErr) GetDeviceMemoryInfo(logicID int32) (*MemoryInfo, error) {

	return nil, fmt.Errorf(errorMsg)
}

// GetDeviceHbmInfo get HBM information , only for Ascend910
func (d *DeviceManagerMockErr) GetDeviceHbmInfo(logicID int32) (*HbmInfo, error) {
	return nil, fmt.Errorf(errorMsg)
}

// GetDeviceErrCode get the error count and errorcode of the device
func (d *DeviceManagerMockErr) GetDeviceErrCode(logicID int32) (int32, int64, error) {
	return int32(0), int64(0), fmt.Errorf(errorMsg)
}

// GetChipInfo get chip info
func (d *DeviceManagerMockErr) GetChipInfo(logicID int32) (*ChipInfo, error) {
	return nil, fmt.Errorf(errorMsg)
}

// GetPhyIDFromLogicID convert the device physicalID to logicID
func (d *DeviceManagerMockErr) GetPhyIDFromLogicID(logicID uint32) (int32, error) {
	return int32(0), fmt.Errorf(errorMsg)
}

// GetLogicIDFromPhyID convert npu device logicId to physicalID
func (d *DeviceManagerMockErr) GetLogicIDFromPhyID(phyID uint32) (int32, error) {
	return int32(0), fmt.Errorf(errorMsg)
}

// GetNPUMajorID query the MajorID of NPU devices
func (d *DeviceManagerMockErr) GetNPUMajorID() (string, error) {
	return "", fmt.Errorf(errorMsg)
}

// GetCardList get npu card array
func (d *DeviceManagerMockErr) GetCardList() (int32, []int32, error) {
	return 0, []int32{}, fmt.Errorf(errorMsg)
}

// GetDeviceNumOnCard get device number on the npu card
func (d *DeviceManagerMockErr) GetDeviceNumOnCard(cardID int32) (int32, error) {
	return 1, fmt.Errorf(errorMsg)
}

// GetCardPower get card power
func (d *DeviceManagerMockErr) GetCardPower(cardID int32) (float32, error) {
	return 1, fmt.Errorf(errorMsg)
}

// GetDeviceLogicID get device logic ID
func (d *DeviceManagerMockErr) GetDeviceLogicID(cardID, deviceID int32) (int32, error) {
	return 1, fmt.Errorf(errorMsg)
}
