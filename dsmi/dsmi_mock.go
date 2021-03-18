//  Copyright(C) 2020. Huawei Technologies Co.,Ltd. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package dsmi interface
package dsmi

import "C"

// DeviceManagerMock struct definition
type DeviceManagerMock struct {
}

// NewDeviceManagerMock new deviceManagerMock instance
func NewDeviceManagerMock() *DeviceManagerMock {
	return &DeviceManagerMock{}
}

// GetDeviceCount get ascend910 device quantity
func (d *DeviceManagerMock) GetDeviceCount() (int32, error) {
	return int32(1), nil
}

// GetDeviceList  get device list
func (d *DeviceManagerMock) GetDeviceList() (int32, []int32, error) {
	return int32(1), []int32{0}, nil
}

// GetDeviceHealth get device health by id
func (d *DeviceManagerMock) GetDeviceHealth(logicID int32) (int32, error) {
	return int32(0), nil

}

// GetDeviceUtilizationRate get device utils rate by id
// DeviceType  Ascend910 1,2,3,4,5,6,10  Ascend310 1,2,3,4,5
func (d *DeviceManagerMock) GetDeviceUtilizationRate(logicID int32, deviceType DeviceType) (int32, error) {
	return int32(1), nil
}

// GetDeviceTemperature get the device temperature
func (d *DeviceManagerMock) GetDeviceTemperature(logicID int32) (int32, error) {
	return int32(1), nil
}

// GetDeviceVoltage get the device voltage
func (d *DeviceManagerMock) GetDeviceVoltage(logicID int32) (float32, error) {
	return 1, nil
}

// GetDevicePower get the power info of the device, the result like : 8.2w
func (d *DeviceManagerMock) GetDevicePower(logicID int32) (float32, error) {
	return 1, nil

}

// GetDeviceFrequency get device frequency, unit MHz
// Ascend910 1,6,7,9
// Ascend310 1,2,3,4,5
// subType enum:  Memory,6HBM,AICoreCurrentFreq,AICoreNormalFreq(1,6,7,9)    see DeviceType
func (d *DeviceManagerMock) GetDeviceFrequency(logicID int32, subType DeviceType) (int32, error) {
	return int32(1), nil
}

// GetDeviceMemoryInfo get memory information
func (d *DeviceManagerMock) GetDeviceMemoryInfo(logicID int32) (*MemoryInfo, error) {
	hbmInfo := NewMemInfo(uint32(1), uint32(1), uint32(1))
	return hbmInfo, nil
}

// GetDeviceHbmInfo get HBM information , only for Ascend910
func (d *DeviceManagerMock) GetDeviceHbmInfo(logicID int32) (*HbmInfo, error) {
	hbmInfo := NewHbmInfo(uint32(1), uint32(1), uint32(1), 1, 1)
	return hbmInfo, nil
}

// GetDeviceErrCode get the error count and errorcode of the device
func (d *DeviceManagerMock) GetDeviceErrCode(logicID int32) (int32, int32, error) {

	return int32(0), int32(0), nil
}

// GetChipInfo get chip info
func (d *DeviceManagerMock) GetChipInfo(logicID int32) (*ChipInfo, error) {
	chip := &ChipInfo{
		ChipType: "ascend",
		ChipName: "910own",
		ChipVer:  "v1",
	}
	return chip, nil
}

// GetPhyIDFromLogicID convert the device physicalID to logicId
func (d *DeviceManagerMock) GetPhyIDFromLogicID(logicID uint32) (int32, error) {
	return int32(1), nil
}

// GetLogicIDFromPhyID convert npu device logicId to physicalID
func (d *DeviceManagerMock) GetLogicIDFromPhyID(phyID uint32) (int32, error) {
	return int32(1), nil
}

// GetNPUMajorID query the MajorID of NPU devices
func (d *DeviceManagerMock) GetNPUMajorID() (string, error) {
	return "239", nil
}

// GetCardList get npu card array
func (d *DeviceManagerMock) GetCardList() (int32, []int32, error) {
	return int32(1), []int32{0}, nil
}

// GetDeviceNumOnCard get device number on the npu card
func (d *DeviceManagerMock) GetDeviceNumOnCard(cardID int32) (int32, error) {
	return int32(1), nil
}

// GetCardPower get card power
func (d *DeviceManagerMock) GetCardPower(cardID int32) (float32, error) {
	return 1, nil
}
