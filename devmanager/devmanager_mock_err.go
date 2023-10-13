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

// Package devmanager this for device driver manager error mock
package devmanager

import (
	"errors"

	"huawei.com/npu-exporter/v5/devmanager/common"
	"huawei.com/npu-exporter/v5/devmanager/dcmi"
)

var errorMsg = "mock error"

// DeviceManagerMockErr common device manager mock error for Ascend910/310P/310
type DeviceManagerMockErr struct {
}

// Init load symbol and initialize dcmi
func (d *DeviceManagerMockErr) Init() error {
	return errors.New(errorMsg)
}

// ShutDown clean the dynamically loaded resource
func (d *DeviceManagerMockErr) ShutDown() error {
	return errors.New(errorMsg)
}

// GetDevType return mock type
func (d *DeviceManagerMockErr) GetDevType() string {
	return common.Ascend910
}

// GetDeviceCount get npu device count
func (d *DeviceManagerMockErr) GetDeviceCount() (int32, error) {
	return 1, errors.New(errorMsg)
}

// GetCardList  get all card list
func (d *DeviceManagerMockErr) GetCardList() (int32, []int32, error) {
	return 1, []int32{0}, errors.New(errorMsg)
}

// GetDeviceNumInCard  get all device list in one card
func (d *DeviceManagerMockErr) GetDeviceNumInCard(cardID int32) (int32, error) {
	return 1, errors.New(errorMsg)
}

// GetDeviceList get all device logicID list
func (d *DeviceManagerMockErr) GetDeviceList() (int32, []int32, error) {
	return 1, []int32{0}, errors.New(errorMsg)
}

// GetDeviceHealth query npu device health status
func (d *DeviceManagerMockErr) GetDeviceHealth(logicID int32) (uint32, error) {
	return 0, errors.New(errorMsg)
}

// GetDeviceNetWorkHealth query npu device network health status
func (d *DeviceManagerMockErr) GetDeviceNetWorkHealth(logicID int32) (uint32, error) {
	return 0, errors.New(errorMsg)
}

// GetDeviceUtilizationRate get npu device utilization
func (d *DeviceManagerMockErr) GetDeviceUtilizationRate(logicID int32, deviceType common.DeviceType) (uint32, error) {
	return 1, errors.New(errorMsg)
}

// GetDeviceTemperature get npu device temperature
func (d *DeviceManagerMockErr) GetDeviceTemperature(logicID int32) (int32, error) {
	return 1, errors.New(errorMsg)
}

// GetDeviceVoltage get npu device voltage
func (d *DeviceManagerMockErr) GetDeviceVoltage(logicID int32) (float32, error) {
	return 1, errors.New(errorMsg)
}

// GetDevicePowerInfo get npu device power info
func (d *DeviceManagerMockErr) GetDevicePowerInfo(logicID int32) (float32, error) {
	return 1, errors.New(errorMsg)
}

// GetDeviceFrequency get npu device work frequency
func (d *DeviceManagerMockErr) GetDeviceFrequency(logicID int32, deviceType common.DeviceType) (uint32, error) {
	return 1, errors.New(errorMsg)
}

// GetDeviceMemoryInfo get npu memory information
func (d *DeviceManagerMockErr) GetDeviceMemoryInfo(logicID int32) (*common.MemoryInfo, error) {
	return &common.MemoryInfo{
		MemorySize:      1,
		MemoryAvailable: 1,
		Frequency:       1,
		Utilization:     1,
	}, errors.New(errorMsg)
}

// GetDeviceHbmInfo get npu HBM module memory and frequency information
func (d *DeviceManagerMockErr) GetDeviceHbmInfo(logicID int32) (*common.HbmInfo, error) {
	return &common.HbmInfo{
		MemorySize:        1,
		Frequency:         1,
		Usage:             1,
		Temp:              1,
		BandWidthUtilRate: 1,
	}, errors.New(errorMsg)
}

// GetDeviceErrorCode get npu device error code
func (d *DeviceManagerMockErr) GetDeviceErrorCode(logicID int32) (int32, int64, error) {
	return int32(0), int64(0), errors.New(errorMsg)
}

// GetChipInfo get npu device error code
func (d *DeviceManagerMockErr) GetChipInfo(logicID int32) (*common.ChipInfo, error) {
	chip := &common.ChipInfo{
		Type:    "ascend",
		Name:    "910",
		Version: "v1",
	}
	return chip, errors.New(errorMsg)
}

// GetPhysicIDFromLogicID get device physic id from logic id
func (d *DeviceManagerMockErr) GetPhysicIDFromLogicID(logicID int32) (int32, error) {
	return 1, errors.New(errorMsg)
}

// GetLogicIDFromPhysicID get device logic id from physic id
func (d *DeviceManagerMockErr) GetLogicIDFromPhysicID(physicID int32) (int32, error) {
	return 1, errors.New(errorMsg)
}

// GetDeviceLogicID get device logic id from card id and device id
func (d *DeviceManagerMockErr) GetDeviceLogicID(cardID, deviceID int32) (int32, error) {
	return 1, errors.New(errorMsg)
}

// GetDeviceIPAddress get device ip address
func (d *DeviceManagerMockErr) GetDeviceIPAddress(logicID, ipType int32) (string, error) {
	return "127.0.0.1", errors.New(errorMsg)
}

// CreateVirtualDevice create virtual device
func (d *DeviceManagerMockErr) CreateVirtualDevice(logicID int32, vDevInfo common.CgoCreateVDevRes) (common.
	CgoCreateVDevOut, error) {
	return common.CgoCreateVDevOut{}, errors.New(errorMsg)
}

// GetVirtualDeviceInfo get virtual device info
func (d *DeviceManagerMockErr) GetVirtualDeviceInfo(logicID int32) (common.VirtualDevInfo, error) {
	return common.VirtualDevInfo{}, errors.New(errorMsg)
}

// DestroyVirtualDevice destroy virtual device
func (d *DeviceManagerMockErr) DestroyVirtualDevice(logicID int32, vDevID uint32) error {
	return errors.New(errorMsg)
}

// GetMcuPowerInfo get mcu power info for cardID
func (d *DeviceManagerMockErr) GetMcuPowerInfo(cardID int32) (float32, error) {
	return 1, errors.New(errorMsg)
}

// GetCardIDDeviceID get cardID and deviceID by logicID
func (d *DeviceManagerMockErr) GetCardIDDeviceID(logicID int32) (int32, int32, error) {
	return 0, 0, errors.New(errorMsg)
}

// GetProductType get product type failed
func (d *DeviceManagerMockErr) GetProductType(cardID, deviceID int32) (string, error) {
	return "", errors.New("not found product type name")
}

// GetAllProductType get all product type failed
func (d *DeviceManagerMockErr) GetAllProductType() ([]string, error) {
	return []string{}, errors.New("not found product type name")
}

// GetNpuWorkMode get npu work mode failed
func (d *DeviceManagerMockErr) GetNpuWorkMode() string {
	return ""
}

// SetDeviceReset set device reset failed
func (d *DeviceManagerMockErr) SetDeviceReset(cardID, deviceID int32) error {
	return errors.New(errorMsg)
}

// GetDeviceBootStatus get device boot status failed
func (d *DeviceManagerMockErr) GetDeviceBootStatus(logicID int32) (int, error) {
	return common.RetError, errors.New(errorMsg)
}

// GetDeviceAllErrorCode get device all error code failed
func (d *DeviceManagerMockErr) GetDeviceAllErrorCode(logicID int32) (int32, []int64, error) {
	return common.RetError, nil, errors.New(errorMsg)
}

// SubscribeDeviceFaultEvent subscribe device fault event failed
func (d *DeviceManagerMockErr) SubscribeDeviceFaultEvent(logicID int32) error {
	return errors.New(errorMsg)
}

// SetFaultEventCallFunc set fault event call func failed
func (d *DeviceManagerMockErr) SetFaultEventCallFunc(businessFunc func(common.DevFaultInfo)) error {
	return errors.New(errorMsg)
}

// GetDieID get die id failed
func (d *DeviceManagerMockErr) GetDieID(logicID int32, dcmiDieType dcmi.DcmiDieType) (string, error) {
	return "", errors.New(errorMsg)
}

// GetDevProcessInfo get process info
func (d *DeviceManagerMockErr) GetDevProcessInfo(logicID int32) (*common.DevProcessInfo, error) {
	return nil, errors.New(errorMsg)
}

func (d *DeviceManagerMockErr) GetPCIeBusInfo(logicID int32) (string, error) {
	return "", errors.New(errorMsg)
}

func (d *DeviceManagerMockErr) GetBoardInfo(logicID int32) (common.BoardInfo, error) {
	return common.BoardInfo{}, errors.New(errorMsg)
}

// GetProductTypeArray test for get empty product type array
func (d *DeviceManagerMockErr) GetProductTypeArray() []string {
	return nil
}