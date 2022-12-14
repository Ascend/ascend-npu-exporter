/* Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package devmanager this for device driver manager
package devmanager

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/npu-exporter/devmanager/common"
	"huawei.com/npu-exporter/devmanager/dcmi"
)

// DeviceInterface for common device interface
type DeviceInterface interface {
	Init() error
	ShutDown() error
	GetDeviceCount() (int32, error)
	GetCardList() (int32, []int32, error)
	GetDeviceNumInCard(cardID int32) (int32, error)
	GetDeviceList() (int32, []int32, error)
	GetDeviceHealth(logicID int32) (uint32, error)
	GetDeviceNetWorkHealth(logicID int32) (uint32, error)
	GetDeviceUtilizationRate(logicID int32, deviceType common.DeviceType) (uint32, error)
	GetDeviceTemperature(logicID int32) (int32, error)
	GetDeviceVoltage(logicID int32) (float32, error)
	GetDevicePowerInfo(logicID int32) (float32, error)
	GetMcuPowerInfo(cardID int32) (float32, error)
	GetDeviceFrequency(logicID int32, deviceType common.DeviceType) (int32, error)
	GetDeviceMemoryInfo(logicID int32) (*common.MemoryInfo, error)
	GetDeviceHbmInfo(logicID int32) (*common.HbmInfo, error)
	GetDeviceErrorCode(logicID int32) (int32, int64, error)
	GetChipInfo(logicID int32) (*common.ChipInfo, error)
	GetPhysicIDFromLogicID(logicID int32) (int32, error)
	GetLogicIDFromPhysicID(physicID int32) (int32, error)
	GetDeviceLogicID(cardID, deviceID int32) (int32, error)
	GetCardIDDeviceID(logicID int32) (int32, int32, error)
	GetDeviceIPAddress(logicID int32) (string, error)
	CreateVirtualDevice(logicID int32, vDevInfo common.CgoCreateVDevRes) (common.CgoCreateVDevOut, error)
	GetVirtualDeviceInfo(logicID int32) (common.VirtualDevInfo, error)
	DestroyVirtualDevice(logicID int32, vDevID uint32) error
	GetDevType() string
	GetProductType() (string, error)
}

// DeviceManager common device manager for Ascend910/310P/310
type DeviceManager struct {
	// DcMgr for common dev manager
	DcMgr dcmi.DcDriverInterface
	// DevType the value is the same as the device type corresponding to the DcMgr variable.
	// Options: common.Ascend310,common.Ascend310P,common.Ascend910
	DevType string
}

// GetDevType return dev type
func (d *DeviceManager) GetDevType() string {
	return d.DevType
}

// AutoInit auto detect npu chip type and return the corresponding processing object
func AutoInit(dType string) (*DeviceManager, error) {
	chipInfo, err := getChipInfoForInit()
	if err != nil {
		return nil, fmt.Errorf("auto init failed, err: %s", err)
	}
	devManager := &DeviceManager{}
	devType := common.GetDeviceTypeByChipName(chipInfo.Name)
	switch devType {
	case common.Ascend910:
		devManager.DcMgr = &A910Manager{}
	case common.Ascend310P:
		devManager.DcMgr = &A310PManager{}
	case common.Ascend310:
		devManager.DcMgr = &A310Manager{}
	default:
		return nil, fmt.Errorf("unsupport device type (%s)", devType)
	}
	if dType != "" && devType != dType {
		return nil, fmt.Errorf("the value of dType(%s) is inconsistent with the actual chip type(%s)",
			dType, devType)
	}
	devManager.DevType = devType
	if err = devManager.Init(); err != nil {
		return nil, fmt.Errorf("deviceManager init failed, err: %#v", err)
	}
	return devManager, nil
}

func getChipInfoForInit() (common.ChipInfo, error) {
	dcMgr := dcmi.DcManager{}
	if err := dcMgr.DcInit(); err != nil {
		return common.ChipInfo{}, fmt.Errorf("dc init failed, err: %#v", err)
	}
	defer func() {
		if err := dcMgr.DcShutDown(); err != nil {
			hwlog.RunLog.Error(err)
		}
	}()
	// get card list
	carNum, cardList, err := dcMgr.DcGetCardList()
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.ChipInfo{}, fmt.Errorf("get card list failed for init")
	}
	if carNum == 0 {
		return common.ChipInfo{}, fmt.Errorf("get chip info failed, no card found")
	}
	// get device in card, then get chip info by cardID and deviceID
	for _, cardID := range cardList {
		devNum, err := dcMgr.DcGetDeviceNumInCard(cardID)
		if err != nil || devNum == 0 {
			hwlog.RunLog.Debugf("get device num by cardID(%d) failed, error: %#v", cardID, err)
			continue
		}
		for devID := int32(0); devID < devNum; devID++ {
			chipInfo, err := dcMgr.DcGetChipInfo(cardID, devID)
			if err != nil {
				hwlog.RunLog.Debugf("get chip info failed by cardID(%d), deviceID(%d), error: %#v", cardID, devID,
					err)
				continue
			}
			if !common.IsValidChipInfo(chipInfo) {
				hwlog.RunLog.Debugf("invalid chip info by cardID(%d), deviceID(%d), error: %#v", cardID, devID,
					err)
				continue
			}
			return *chipInfo, nil
		}
	}

	return common.ChipInfo{}, errors.New("cannot get valid chip info")
}

// Init load symbol and initialize dcmi
func (d *DeviceManager) Init() error {
	return d.DcMgr.DcInit()
}

// ShutDown clean the dynamically loaded resource
func (d *DeviceManager) ShutDown() error {
	return d.DcMgr.DcShutDown()
}

// GetDeviceCount get npu device count
func (d *DeviceManager) GetDeviceCount() (int32, error) {
	return d.DcMgr.DcGetDeviceCount()
}

// GetCardList  get all card list
func (d *DeviceManager) GetCardList() (int32, []int32, error) {
	return d.DcMgr.DcGetCardList()
}

// GetDeviceNumInCard  get all device list in one card
func (d *DeviceManager) GetDeviceNumInCard(cardID int32) (int32, error) {
	return d.DcMgr.DcGetDeviceNumInCard(cardID)
}

// GetDeviceList get all device logicID list
func (d *DeviceManager) GetDeviceList() (int32, []int32, error) {
	return d.DcMgr.DcGetLogicIDList()
}

// GetDeviceHealth query npu device health status
func (d *DeviceManager) GetDeviceHealth(logicID int32) (uint32, error) {
	cardID, deviceID, err := d.DcMgr.DcGetCardIDDeviceID(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.UnRetError, fmt.Errorf("failed to get health code by logicID(%d)", logicID)
	}
	healthCode, err := d.DcMgr.DcGetDeviceHealth(cardID, deviceID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.UnRetError, fmt.Errorf("failed to get health code by logicID(%d)", logicID)
	}

	return uint32(healthCode), nil
}

// GetDeviceNetWorkHealth query npu device network health status
func (d *DeviceManager) GetDeviceNetWorkHealth(logicID int32) (uint32, error) {
	cardID, deviceID, err := d.DcMgr.DcGetCardIDDeviceID(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.UnRetError, fmt.Errorf("failed to get network health code by logicID(%d)", logicID)
	}
	healthCode, err := d.DcMgr.DcGetDeviceNetWorkHealth(cardID, deviceID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.UnRetError, fmt.Errorf("failed to get network health code by logicID(%d)", logicID)
	}

	return healthCode, nil
}

// GetDeviceUtilizationRate get npu device utilization
func (d *DeviceManager) GetDeviceUtilizationRate(logicID int32, deviceType common.DeviceType) (uint32, error) {
	cardID, deviceID, err := d.DcMgr.DcGetCardIDDeviceID(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.UnRetError, fmt.Errorf("failed to get utilization by logicID(%d)", logicID)
	}
	rate, err := d.DcMgr.DcGetDeviceUtilizationRate(cardID, deviceID, deviceType)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.UnRetError, fmt.Errorf("failed to get utilization by logicID(%d)", logicID)
	}

	return uint32(rate), nil
}

// GetDeviceTemperature get npu device temperature
func (d *DeviceManager) GetDeviceTemperature(logicID int32) (int32, error) {
	cardID, deviceID, err := d.DcMgr.DcGetCardIDDeviceID(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.RetError, fmt.Errorf("failed to get temperature by logicID(%d)", logicID)
	}
	temp, err := d.DcMgr.DcGetDeviceTemperature(cardID, deviceID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.RetError, fmt.Errorf("failed to get temperature by logicID(%d)", logicID)
	}

	return temp, nil
}

// GetDeviceVoltage get npu device voltage
func (d *DeviceManager) GetDeviceVoltage(logicID int32) (float32, error) {
	cardID, deviceID, err := d.DcMgr.DcGetCardIDDeviceID(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.UnRetError, fmt.Errorf("failed to get voltage by logicID(%d)", logicID)
	}
	voltage, err := d.DcMgr.DcGetDeviceVoltage(cardID, deviceID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.UnRetError, fmt.Errorf("failed to get voltage by logicID(%d)", logicID)
	}

	return voltage, nil
}

// GetDevicePowerInfo get npu device power info
func (d *DeviceManager) GetDevicePowerInfo(logicID int32) (float32, error) {
	cardID, deviceID, err := d.DcMgr.DcGetCardIDDeviceID(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.UnRetError, fmt.Errorf("failed to get power by logicID(%d)", logicID)
	}
	power, err := d.DcMgr.DcGetDevicePowerInfo(cardID, deviceID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.UnRetError, fmt.Errorf("failed to get power by logicID(%d)", logicID)
	}

	return power, nil
}

// GetDeviceFrequency get npu device work frequency
func (d *DeviceManager) GetDeviceFrequency(logicID int32, deviceType common.DeviceType) (int32, error) {
	cardID, deviceID, err := d.DcMgr.DcGetCardIDDeviceID(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.RetError, fmt.Errorf("failed to get frequency by logicID(%d)", logicID)
	}
	frequency, err := d.DcMgr.DcGetDeviceFrequency(cardID, deviceID, deviceType)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.RetError, fmt.Errorf("failed to get frequency by logicID(%d)", logicID)
	}

	return frequency, nil
}

// GetDeviceMemoryInfo get npu memory information
func (d *DeviceManager) GetDeviceMemoryInfo(logicID int32) (*common.MemoryInfo, error) {
	cardID, deviceID, err := d.DcMgr.DcGetCardIDDeviceID(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return nil, fmt.Errorf("failed to get memory info by logicID(%d)", logicID)
	}
	memInfo, err := d.DcMgr.DcGetMemoryInfo(cardID, deviceID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return nil, fmt.Errorf("failed to get memory info by logicID(%d)", logicID)
	}

	return memInfo, nil
}

// GetDeviceHbmInfo get npu HBM module memory and frequency information
func (d *DeviceManager) GetDeviceHbmInfo(logicID int32) (*common.HbmInfo, error) {
	cardID, deviceID, err := d.DcMgr.DcGetCardIDDeviceID(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return nil, fmt.Errorf("failed to get hbm info by logicID(%d)", logicID)
	}
	hbmInfo, err := d.DcMgr.DcGetHbmInfo(cardID, deviceID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return nil, fmt.Errorf("failed to get hbm info by logicID(%d)", logicID)
	}

	return hbmInfo, nil
}

// GetDeviceErrorCode get npu device error code
func (d *DeviceManager) GetDeviceErrorCode(logicID int32) (int32, int64, error) {
	cardID, deviceID, err := d.DcMgr.DcGetCardIDDeviceID(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.RetError, common.RetError, fmt.Errorf("failed to get device error code by logicID(%d)",
			logicID)
	}
	errCount, errCode, err := d.DcMgr.DcGetDeviceErrorCode(cardID, deviceID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.RetError, common.RetError, fmt.Errorf("failed to get device error code by logicID(%d)",
			logicID)
	}

	return errCount, errCode, nil
}

// GetChipInfo get npu device error code
func (d *DeviceManager) GetChipInfo(logicID int32) (*common.ChipInfo, error) {
	cardID, deviceID, err := d.DcMgr.DcGetCardIDDeviceID(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return nil, fmt.Errorf("failed to get chip info code by logicID(%d)", logicID)
	}
	chipInfo, err := d.DcMgr.DcGetChipInfo(cardID, deviceID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return nil, fmt.Errorf("failed to get chip info code by logicID(%d)", logicID)
	}

	return chipInfo, nil
}

// GetPhysicIDFromLogicID get device physic id from logic id
func (d *DeviceManager) GetPhysicIDFromLogicID(logicID int32) (int32, error) {
	physicID, err := d.DcMgr.DcGetPhysicIDFromLogicID(logicID)
	if err != nil {
		return common.RetError, fmt.Errorf("failed to get physicID by logicID(%d)", logicID)
	}

	return physicID, nil
}

// GetLogicIDFromPhysicID get device logic id from physic id
func (d *DeviceManager) GetLogicIDFromPhysicID(physicID int32) (int32, error) {
	logicID, err := d.DcMgr.DcGetLogicIDFromPhysicID(physicID)
	if err != nil {
		return common.RetError, fmt.Errorf("failed to get logicID by physicID(%d)", physicID)
	}

	return logicID, nil
}

// GetDeviceLogicID get device logic id from card id and device id
func (d *DeviceManager) GetDeviceLogicID(cardID, deviceID int32) (int32, error) {
	return d.DcMgr.DcGetDeviceLogicID(cardID, deviceID)
}

// GetDeviceIPAddress get device ip address
func (d *DeviceManager) GetDeviceIPAddress(logicID int32) (string, error) {
	cardID, deviceID, err := d.DcMgr.DcGetCardIDDeviceID(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return "", fmt.Errorf("failed to get ip address by logicID(%d)", logicID)
	}
	ipAddr, err := d.DcMgr.DcGetDeviceIPAddress(cardID, deviceID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return "", fmt.Errorf("failed to get ip address by logicID(%d)", logicID)
	}

	return ipAddr, nil
}

// CreateVirtualDevice create virtual device
func (d *DeviceManager) CreateVirtualDevice(logicID int32, vDevInfo common.CgoCreateVDevRes) (common.
	CgoCreateVDevOut, error) {
	if !common.IsValidTemplateName(d.DevType, vDevInfo.TemplateName) {
		return common.CgoCreateVDevOut{}, fmt.Errorf("input invalid template name: %s", vDevInfo.TemplateName)
	}
	return d.DcMgr.DcCreateVDevice(logicID, vDevInfo)
}

// GetVirtualDeviceInfo get virtual device info
func (d *DeviceManager) GetVirtualDeviceInfo(logicID int32) (common.VirtualDevInfo, error) {
	cgoVDevInfo, err := d.DcMgr.DcGetVDeviceInfo(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return common.VirtualDevInfo{}, fmt.Errorf("get virtual device info failed, error is: %#v "+
			"and vdev num is: %d", err, int32(cgoVDevInfo.TotalResource.VDevNum))
	}
	for _, vDevInfo := range cgoVDevInfo.VDevInfo {
		if !common.IsValidTemplateName(d.DevType, vDevInfo.QueryInfo.Name) {
			return common.VirtualDevInfo{}, fmt.Errorf("vdevice id %d, it's template name is invalid: %s",
				vDevInfo.VDevID, vDevInfo.QueryInfo.Name)
		}
	}
	return cgoVDevInfo, nil
}

// DestroyVirtualDevice destroy virtual device
func (d *DeviceManager) DestroyVirtualDevice(logicID int32, vDevID uint32) error {
	return d.DcMgr.DcDestroyVDevice(logicID, vDevID)
}

// GetMcuPowerInfo get mcu power info for cardID
func (d *DeviceManager) GetMcuPowerInfo(cardID int32) (float32, error) {
	return d.DcMgr.DcGetMcuPowerInfo(cardID)
}

// GetCardIDDeviceID get cardID and deviceID by logicID
func (d *DeviceManager) GetCardIDDeviceID(logicID int32) (int32, int32, error) {
	return d.DcMgr.DcGetCardIDDeviceID(logicID)
}

// GetProductType get product type
func (d *DeviceManager) GetProductType() (string, error) {
	cardNum, cardList, err := d.GetCardList()
	if cardNum == 0 || err != nil {
		hwlog.RunLog.Errorf("failed to get card list, err: %#v", err)
		return "", err
	}
	for _, cardID := range cardList {
		devNum, err := d.GetDeviceNumInCard(cardID)
		if err != nil {
			hwlog.RunLog.Debugf("get device num by cardID(%d) failed, error: %#v", cardID, err)
			continue
		}
		if devNum == 0 {
			hwlog.RunLog.Debugf("not found device on card %d", cardID)
			continue
		}
		for devID := int32(0); devID < devNum; devID++ {
			productType, err := d.DcMgr.DcGetProductType(cardID, devID)
			if err != nil {
				hwlog.RunLog.Debugf("get product type by card %d deviceID %d failed, err: %#v", cardID, devID, err)
				continue
			}
			return productType, nil
		}
	}
	return "", nil
}
