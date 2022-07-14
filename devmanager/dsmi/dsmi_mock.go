//  Copyright(C) 2020. Huawei Technologies Co.,Ltd. All rights reserved.

// Package dsmi interface
package dsmi

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
// subType enum:  Memory,6HBM,AI_Core_Current_Fre,AI_Core_Normal_Fre(1,6,7,9)    see DeviceType
func (d *DeviceManagerMock) GetDeviceFrequency(logicID int32, subType DeviceType) (int32, error) {
	return int32(1), nil
}

// GetDeviceErrCode get the error count and errorcode of the device
func (d *DeviceManagerMock) GetDeviceErrCode(logicID int32) (int32, int64, error) {

	return int32(0), int64(0), nil
}

// GetPhyIDFromLogicID convert the device physicalID to logicID
func (d *DeviceManagerMock) GetPhyIDFromLogicID(logicID uint32) (int32, error) {
	return int32(1), nil
}

// GetLogicIDFromPhyID convert npu device logicId to physicalID
func (d *DeviceManagerMock) GetLogicIDFromPhyID(phyID uint32) (int32, error) {
	return int32(1), nil
}

// GetNPUMajorID query the MajorID of NPU devices
func (d *DeviceManagerMock) GetNPUMajorID() ([]string, error) {
	return []string{"239", "238"}, nil
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

// GetDeviceLogicID get device logic ID
func (d *DeviceManagerMock) GetDeviceLogicID(cardID, deviceID int32) (int32, error) {
	return 1, nil
}
