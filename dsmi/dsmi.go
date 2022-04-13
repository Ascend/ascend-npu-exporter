//  Copyright(C) 2020. Huawei Technologies Co.,Ltd. All rights reserved.

// Package dsmi convert the npu driver interface to go style
package dsmi

// #cgo LDFLAGS: -ldl
/*
#include <stddef.h>
#include <dlfcn.h>
#include <stdlib.h>
#include <stdio.h>

#include "dsmi_common_interface.h"
#include "dcmi_interface_api.h"

// dsmiHandle is the handle for dynamically loaded libdrvdsmi_host.so
void *dsmiHandle;
void *dcmiHandle;
#define SO_NOT_FOUND -99999
#define FUNCTION_NOT_FOUND -99998
#define SUCCESS 0
#define ERROR_UNKNOWN -99997
#define CALL_FUNC(func_name,...)					\
	if(func_name##_func == NULL){					\
		return FUNCTION_NOT_FOUND;					\
	}												\
	return func_name##_func(__VA_ARGS__);			\

int (*dsmi_get_device_count_func)(int *device_count);
int dsmi_get_device_count(int *device_count){
    CALL_FUNC(dsmi_get_device_count,device_count)
}

int (*dsmi_list_device_func)(int device_id_list[], int count);
int dsmi_list_device(int device_id_list[], int count){
	CALL_FUNC(dsmi_list_device,device_id_list,count)
}

int (*dsmi_get_device_health_func)(int device_id, unsigned int *phealth);
int dsmi_get_device_health(int device_id, unsigned int *phealth){
	CALL_FUNC(dsmi_get_device_health,device_id,phealth)
}

int (*dsmi_get_device_utilization_rate_func)(int device_id, int device_type, unsigned int *putilization_rate);
int dsmi_get_device_utilization_rate(int device_id,int device_type, unsigned int *putilization_rate){
	CALL_FUNC(dsmi_get_device_utilization_rate,device_id, device_type,putilization_rate)
}

int (*dsmi_get_phyid_from_logicid_func)(unsigned int logicid, unsigned int *phyid);
int dsmi_get_phyid_from_logicid(unsigned int logicid, unsigned int *phyid){
	CALL_FUNC(dsmi_get_phyid_from_logicid,logicid,phyid)
}

int (*dsmi_get_logicid_from_phyid_func)(unsigned int phyid, unsigned int *logicid);
int dsmi_get_logicid_from_phyid(unsigned int phyid, unsigned int *logicid){
	CALL_FUNC(dsmi_get_logicid_from_phyid,phyid,logicid)
}

int (*dsmi_get_device_temperature_func)(int device_id,  int *ptemperature);
int dsmi_get_device_temperature(int device_id,  int *ptemperature){
	CALL_FUNC(dsmi_get_device_temperature,device_id,ptemperature)
}

int (*dsmi_get_device_voltage_func)(int device_id, unsigned int *pvoltage);
int dsmi_get_device_voltage(int device_id, unsigned int *pvoltage){
	CALL_FUNC(dsmi_get_device_voltage,device_id,pvoltage)
}

int (*dsmi_get_device_power_info_func)(int device_id, struct dsmi_power_info_stru *pdevice_power_info);
int dsmi_get_device_power_info(int device_id, struct dsmi_power_info_stru *pdevice_power_info){
	CALL_FUNC(dsmi_get_device_power_info,device_id,pdevice_power_info)
}

int (*dsmi_get_device_frequency_func)(int device_id, int device_type,unsigned int *pfrequency);
int dsmi_get_device_frequency(int device_id, int device_type,unsigned int *pfrequency){
	CALL_FUNC(dsmi_get_device_frequency,device_id,device_type,pfrequency)
}

int (*dsmi_get_hbm_info_func)(int device_id, struct dsmi_hbm_info_stru *pdevice_hbm_info);
int dsmi_get_hbm_info(int device_id, struct dsmi_hbm_info_stru *pdevice_hbm_info){
	CALL_FUNC(dsmi_get_hbm_info,device_id,pdevice_hbm_info)
}

int (*dsmi_get_memory_info_func)(int device_id, struct dsmi_memory_info_stru *pdevice_memory_info);
int dsmi_get_memory_info(int device_id, struct dsmi_memory_info_stru *pdevice_memory_info){
	CALL_FUNC(dsmi_get_memory_info,device_id,pdevice_memory_info)
}

int (*dsmi_get_device_errorcode_func)(int device_id, int *errorcount,unsigned int *perrorcode);
int dsmi_get_device_errorcode(int device_id, int *errorcount,unsigned int *perrorcode){
	CALL_FUNC(dsmi_get_device_errorcode,device_id,errorcount,perrorcode)
}

int (*dsmi_get_chip_info_func)(int device_id, struct dsmi_chip_info_stru *chip_info);
int dsmi_get_chip_info(int device_id, struct dsmi_chip_info_stru *chip_info){
	CALL_FUNC(dsmi_get_chip_info,device_id,chip_info)
}
//dcmi

int (*dcmi_init_func)();
int dcmi_init(){
	CALL_FUNC(dcmi_init)
}

int (*dcmi_get_card_num_list_func)(int *card_num, int *card_list, int list_length);
int dcmi_get_card_num_list(int *card_num, int *card_list, int list_length){
	CALL_FUNC(dcmi_get_card_num_list,card_num,card_list,list_length)
}

int (*dcmi_get_device_num_in_card_func)(int card_id, int *device_num);
int dcmi_get_device_num_in_card(int card_id, int *device_num){
	CALL_FUNC(dcmi_get_device_num_in_card,card_id,device_num)
}

int (*dcmi_mcu_get_power_info_func)(int card_id,int *power);
int dcmi_mcu_get_power_info(int card_id,int *power){
	CALL_FUNC(dcmi_mcu_get_power_info,card_id,power)
}
int (*dcmi_get_device_logic_id_func)(int *device_logic_id, int card_id, int device_id);
int dcmi_get_device_logic_id(int *device_logic_id, int card_id, int device_id){
    CALL_FUNC(dcmi_get_device_logic_id,device_logic_id,card_id,device_id)
}

// load .so files and functions
int dsmiInit_dl(void){
	dsmiHandle = dlopen("libdrvdsmi_host.so",RTLD_LAZY | RTLD_GLOBAL);
	if (dsmiHandle == NULL){
		return SO_NOT_FOUND;
	}

	dsmi_list_device_func = dlsym(dsmiHandle,"dsmi_list_device");

	dsmi_get_device_count_func = dlsym(dsmiHandle,"dsmi_get_device_count");

	dsmi_get_device_health_func = dlsym(dsmiHandle,"dsmi_get_device_health");

	dsmi_get_device_utilization_rate_func = dlsym(dsmiHandle,"dsmi_get_device_utilization_rate");

	dsmi_get_phyid_from_logicid_func = dlsym(dsmiHandle,"dsmi_get_phyid_from_logicid");

	dsmi_get_logicid_from_phyid_func = dlsym(dsmiHandle,"dsmi_get_logicid_from_phyid");

	dsmi_get_device_temperature_func = dlsym(dsmiHandle,"dsmi_get_device_temperature");

	dsmi_get_device_voltage_func = dlsym(dsmiHandle,"dsmi_get_device_voltage");

	dsmi_get_device_power_info_func = dlsym(dsmiHandle,"dsmi_get_device_power_info");

	dsmi_get_device_frequency_func = dlsym(dsmiHandle,"dsmi_get_device_frequency");

	dsmi_get_hbm_info_func = dlsym(dsmiHandle,"dsmi_get_hbm_info");

	dsmi_get_memory_info_func = dlsym(dsmiHandle,"dsmi_get_memory_info");

	dsmi_get_device_errorcode_func = dlsym(dsmiHandle,"dsmi_get_device_errorcode");

	dsmi_get_chip_info_func = dlsym(dsmiHandle,"dsmi_get_chip_info");

	dlopen("libm.so",RTLD_LAZY | RTLD_GLOBAL);
	dcmiHandle = dlopen("libdcmi.so",RTLD_LAZY | RTLD_GLOBAL);
	if (dcmiHandle == NULL){
		fprintf (stderr,"%s\n",dlerror());
		return SO_NOT_FOUND;
	}

	dcmi_init_func = dlsym(dcmiHandle,"dcmi_init");

	dcmi_get_card_num_list_func = dlsym(dcmiHandle,"dcmi_get_card_num_list");

	dcmi_get_device_num_in_card_func = dlsym(dcmiHandle,"dcmi_get_device_num_in_card");

	dcmi_mcu_get_power_info_func = dlsym(dcmiHandle,"dcmi_mcu_get_power_info");

	dcmi_get_device_logic_id_func = dlsym(dcmiHandle,"dcmi_get_device_logic_id");

	return SUCCESS;
}

int dsmiShutDown(void){
	if (dsmiHandle == NULL && dcmiHandle == NULL){
		return SUCCESS;
	}
	return (dlclose(dsmiHandle) && dlclose(dcmiHandle) ? ERROR_UNKNOWN : SUCCESS);
}
*/
import "C"
import (
	"bufio"
	"fmt"
	"huawei.com/npu-exporter/utils"
	"math"
	"os"
	"regexp"
	"strings"
	"sync"

	"huawei.com/npu-exporter/hwlog"
)

const (
	// the error value  when the function failed
	retError = -1
	// max chip name length
	maxChipName = 32
	// Percent constant of 100
	Percent = 100
	// OneKilo for unit change kb to mb
	OneKilo = 1024
	// MaxErrorCodeCount number of error codes
	MaxErrorCodeCount = 128
	// DefaultTemperatureWhenQueryFailed when get temperature failed, use this value
	DefaultTemperatureWhenQueryFailed = -275
	maxChipNum                        = 64
	unitChange100                     = 0.01
	unitChange10                      = 0.1
	retryTime                         = 3
)

// HbmInfo HBM info
type HbmInfo struct {
	MemorySize              uint64 `json:"memory_size"`        // HBM total size,KB
	MemoryFrequency         uint32 `json:"hbm_frequency"`      // HBM frequncy MHz
	MemoryUsage             uint64 `json:"memory_usage"`       // HBM memory usagem,KB
	MemoryTemp              int32  `json:"hbm_temperature"`    // HBM temperature
	MemoryBandWidthUtilRate uint32 `json:"hbm_bandwidth_util"` // HBM brandwidth utilization

}

// MemoryInfo memory infomation struct
type MemoryInfo struct {
	MemorySize  uint64 `json:"memory_size"`
	Frequency   uint32 `json:"memory_frequency"`
	Utilization uint32 `json:"memory_utilization"`
}

// ChipInfo chip info
type ChipInfo struct {
	ChipType string `json:"chip_type"`
	ChipName string `json:"chip_name"`
	ChipVer  string `json:"chip_version"`
}

// NewMemInfo new meminfo struct
func NewMemInfo(memorySize uint64, frequency, utilization uint32) *MemoryInfo {
	return &MemoryInfo{
		MemorySize:  memorySize,
		Frequency:   frequency,
		Utilization: utilization,
	}
}

// NewHbmInfo new HbmInfo
func NewHbmInfo(memorySize uint64, memoryFrequency uint32, memoryUsage uint64, memoryTemp int32,
	memoryBandWidthUtilRate uint32) *HbmInfo {
	return &HbmInfo{
		MemorySize:              memorySize,
		MemoryFrequency:         memoryFrequency,
		MemoryUsage:             memoryUsage,
		MemoryTemp:              memoryTemp,
		MemoryBandWidthUtilRate: memoryBandWidthUtilRate,
	}
}

// 'getDeviceInfoInterface' is used to obtain device information
// if device information meets the requirements, it will return directly.
// otherwise, one or more methods in 'handleDeviceInfoInterface' will be invoked
// to handle the device information before return.
type getDeviceInfoInterface interface {
	// GetDeviceCount get npu device count
	GetDeviceCount() (int32, error)
	// GetDeviceList get npu device array
	GetDeviceList() (int32, []int32, error)
	// GetDeviceHealth query npu device health status
	GetDeviceHealth(logicID int32) (int32, error)
	// GetDeviceUtilizationRate get npu device utilization
	GetDeviceUtilizationRate(logicID int32, deviceType DeviceType) (int32, error)
	// GetDeviceTemperature get npu device temperature
	GetDeviceTemperature(logicID int32) (int32, error)
	// GetDeviceVoltage get npu device voltage
	GetDeviceVoltage(logicID int32) (float32, error)
	// GetDevicePower  get npu device power
	GetDevicePower(logicID int32) (float32, error)
	// GetDeviceFrequency get npu device work frequency
	GetDeviceFrequency(logicID int32, subType DeviceType) (int32, error)
	// GetDeviceMemoryInfo get npu memory information
	GetDeviceMemoryInfo(logicID int32) (*MemoryInfo, error)
	// GetDeviceHbmInfo get npu HBM module memory and frequency information
	GetDeviceHbmInfo(logicID int32) (*HbmInfo, error)
	// GetDeviceErrCode get npu device error code
	GetDeviceErrCode(logicID int32) (int32, int64, error)
	// GetChipInfo get npu device ascend chip information
	GetChipInfo(logicID int32) (*ChipInfo, error)
	// GetPhyIDFromLogicID convert npu device physicalID to logicId
	GetPhyIDFromLogicID(logicID uint32) (int32, error)
	// GetLogicIDFromPhyID convert npu device logicId to physicalID
	GetLogicIDFromPhyID(phyID uint32) (int32, error)
	// GetNPUMajorID query the MajorID of NPU devices
	GetNPUMajorID() ([]string, error)
	// GetCardList get npu card array
	GetCardList() (int32, []int32, error)
	// GetDeviceNumOnCard get device number on the npu card
	GetDeviceNumOnCard(cardID int32) (int32, error)
	// GetCardPower get card power
	GetCardPower(cardID int32) (float32, error)
	// GetDeviceLogicID get device logic ID
	GetDeviceLogicID(cardID, deviceID int32) (int32, error)
}

// handleDeviceInfoInterface is used to process the device information before return
// different device types can have different implementations of the methods in the interface
type handleDeviceInfoInterface interface {
	createMemoryInfoObj(cmInfo *CStructDsmiMemoryInfo) *MemoryInfo
}

// DeviceMgrInterface is used to obtain the device information that meets the requirements.
type DeviceMgrInterface interface {
	getDeviceInfoInterface
	handleDeviceInfoInterface
}

// please use GetDeviceManager to get the singleton instance of baseDeviceManager
type baseDeviceManager struct{}

type deviceManager910 struct {
	baseDeviceManager
}
type deviceManager710 struct {
	baseDeviceManager
}
type deviceManager310 struct {
	baseDeviceManager
}

var instance DeviceMgrInterface
var once sync.Once
var chipType = Ascend310

// CStructDsmiMemoryInfo the c struct of memoryInfo
type CStructDsmiMemoryInfo = C.struct_dsmi_memory_info_stru

// GetChipTypeNow get the chip type on this machine
func GetChipTypeNow() ChipType {
	return chipType
}

// GetDeviceManager new baseDeviceManager singleton instance
func GetDeviceManager() DeviceMgrInterface {
	once.Do(func() {
		instance = &deviceManager310{}
		num, _, err := instance.GetDeviceList()
		if err != nil || num == 0 {
			hwlog.RunLog.Error("This is no device on this machine")
			return
		}
		var chipinfo *ChipInfo
		for i := int32(0); i < num; i++ {
			chipinfo, err = instance.GetChipInfo(i)
			if err == nil {
				break
			}
			if i == num-1 {
				hwlog.RunLog.Error("get chipInfo failed")
				return
			}
		}

		if err != nil {
			hwlog.RunLog.Error(err)
			return
		}
		if chipinfo == nil {
			hwlog.RunLog.Error("chip info is nil")
			return
		}
		if IsAscend710(chipinfo.ChipName) {
			hwlog.RunLog.Info("change the instance to deviceManager710")
			instance = &deviceManager710{}
			chipType = Ascend710
		}
		if IsAscend910(chipinfo.ChipName) {
			hwlog.RunLog.Info("change the instance to deviceManager910")
			instance = &deviceManager910{}
			chipType = Ascend910
		}
	})
	return instance
}

// GetDeviceCount get ascend910 device quantity
func (d *baseDeviceManager) GetDeviceCount() (int32, error) {
	var count C.int
	if err := C.dsmi_get_device_count(&count); err != 0 {
		errInfo := fmt.Errorf("get device quantity failed, error code: %d", int32(err))
		hwlog.RunLog.Error(errInfo)
		return retError, errInfo
	}
	// Invalid number of devices.
	if count < 0 || count > maxChipNum {
		errInfo := fmt.Errorf("get device quantity failed, the number of devices is: %d", int32(count))
		hwlog.RunLog.Error(errInfo)
		return retError, errInfo
	}
	return int32(count), nil
}

// GetDeviceList  get device list
func (d *baseDeviceManager) GetDeviceList() (int32, []int32, error) {
	var devices []int32
	devNum, err := d.GetDeviceCount()
	if err != nil || devNum == 0 {
		return devNum, devices, err
	}

	var ids [HiAIMaxDeviceNum]C.int
	if err := C.dsmi_list_device(&ids[0], C.int(devNum)); err != 0 {
		errInfo := fmt.Errorf("unable to get device list, return error: %d", int32(err))
		hwlog.RunLog.Error(errInfo)
		return retError, devices, errInfo
	}
	// transfer device list
	var i int32
	for i = 0; i < devNum && i < int32(len(ids)-1); i++ {
		deviceId := int32(ids[i])
		if deviceId < 0 || deviceId > maxChipNum {
			errInfo := fmt.Errorf("the device ids array has invalid id(%d)", deviceId)
			hwlog.RunLog.Error(errInfo)
			continue
		}
		devices = append(devices, deviceId)
	}

	return devNum, devices, nil
}

// GetDeviceHealth get device health by id
func (d *baseDeviceManager) GetDeviceHealth(logicID int32) (int32, error) {
	var health C.uint

	if err := C.dsmi_get_device_health(C.int(logicID), &health); err != 0 {
		errInfo := fmt.Errorf("get device%d health state failed, error code: %d", logicID, int32(err))
		hwlog.RunLog.Error(errInfo)
		return retError, errInfo
	}
	if isGreaterThanOrEqualInt32(int64(health)) {
		errInfo := fmt.Errorf("get wrong health state , device: %d health: %d", logicID, int64(health))
		hwlog.RunLog.Error(errInfo)
		return retError, errInfo
	}

	return int32(health), nil

}

// GetDeviceUtilizationRate get device utils rate by id
// DeviceType  Ascend910 1,2,3,4,5,6,10  Ascend310 1,2,3,4,5
func (d *baseDeviceManager) GetDeviceUtilizationRate(logicID int32, deviceType DeviceType) (int32, error) {
	var utilRate C.uint

	err := C.dsmi_get_device_utilization_rate(C.int(logicID), C.int(deviceType), &utilRate)
	if err != 0 {
		hwlog.RunLog.Errorf("get device%d utilize rate failed, error code: %d, try again ... ", logicID, int32(err))
		for i := 0; i < retryTime; i++ {
			hwlog.RunLog.Errorf("try again %d", i)
			err = C.dsmi_get_device_utilization_rate(C.int(logicID), C.int(deviceType), &utilRate)
			if err == 0 && isValidUtilizationRate(uint32(utilRate)) {
				return int32(utilRate), nil
			}
		}
		return retError, fmt.Errorf("get device%d utilize rate failed, error code: %d", logicID, int32(err))
	}

	if !isValidUtilizationRate(uint32(utilRate)) {
		return retError, fmt.Errorf("get wrong device utilize rate, device: %d utilize rate: %d", logicID,
			uint32(utilRate))
	}

	return int32(utilRate), nil
}

// GetPhyIDFromLogicID get physic id form logic id
func (d *baseDeviceManager) GetPhyIDFromLogicID(logicID uint32) (int32, error) {
	var phyID C.uint

	if err := C.dsmi_get_phyid_from_logicid(C.uint(logicID), &phyID); err != 0 {
		errInfo := fmt.Errorf("get device%d phy id failed ,error code is: %d", logicID, int32(err))
		hwlog.RunLog.Error(errInfo)
		return retError, errInfo
	}
	// check whether phyID is too big
	if uint32(phyID) > uint32(math.MaxInt8) {
		errInfo := fmt.Errorf("get error phyID from logicID, phyID is: %d, logicID is: %d", uint32(phyID), logicID)
		hwlog.RunLog.Error(errInfo)
		return retError, errInfo
	}

	return int32(phyID), nil
}

// GetLogicIDFromPhyID get logic id form physic id
func (d *baseDeviceManager) GetLogicIDFromPhyID(phyID uint32) (int32, error) {
	var logicID C.uint

	if err := C.dsmi_get_logicid_from_phyid(C.uint(phyID), &logicID); err != 0 {
		errInfo := fmt.Errorf("get device%d logic id failed ,error code is : %d", phyID, int32(err))
		hwlog.RunLog.Error(errInfo)
		return retError, errInfo
	}
	// check whether logicID is too big
	if uint32(logicID) > uint32(math.MaxInt8) {
		errInfo := fmt.Errorf("get error logicID from phyID, logicID is: %d, phyID is: %d", uint32(logicID), phyID)
		hwlog.RunLog.Error(errInfo)
		return retError, errInfo
	}

	return int32(logicID), nil

}

// GetDeviceTemperature get the device temperature
func (d *baseDeviceManager) GetDeviceTemperature(logicID int32) (int32, error) {
	var temp C.int
	if err := C.dsmi_get_device_temperature(C.int(logicID), &temp); err != 0 {
		errInfo := fmt.Errorf("get device%d temperature failed ,error code is : %d", logicID, int32(err))
		return retError, errInfo
	}
	parsedTemp := int32(temp)
	if parsedTemp < int32(DefaultTemperatureWhenQueryFailed) {
		errInfo := fmt.Errorf("get wrong device temperature, devcie: %d, temperature: %d", logicID, parsedTemp)
		return retError, errInfo
	}

	return parsedTemp, nil
}

// GetDeviceVoltage get the device voltage
func (d *baseDeviceManager) GetDeviceVoltage(logicID int32) (float32, error) {
	var vol C.uint
	if err := C.dsmi_get_device_voltage(C.int(logicID), &vol); err != 0 {
		errInfo := fmt.Errorf("get device%d voltage failed ,error code is : %d", logicID, int32(err))
		hwlog.RunLog.Error(errInfo)
		return retError, errInfo
	}
	// the voltage's value is error if it's greater than or equal to MaxInt32(1<<31 - 1)
	if isGreaterThanOrEqualInt32(int64(vol)) {
		errInfo := fmt.Errorf("get wrong device voltage, device: %d, voltage: %d", logicID, int64(vol))
		hwlog.RunLog.Error(errInfo)
		return retError, errInfo
	}
	voltage := float32(vol) * unitChange100
	return voltage, nil
}

// GetDevicePower get the power info of the device, the result like : 8.2w
func (d *baseDeviceManager) GetDevicePower(logicID int32) (float32, error) {
	var cpower C.struct_dsmi_power_info_stru
	if err := C.dsmi_get_device_power_info(C.int(logicID), &cpower); err != 0 {
		errInfo := fmt.Errorf("get device%d power failed, error code: %d", logicID, int32(err))
		hwlog.RunLog.Error(errInfo)
		return retError, errInfo
	}
	parsedPower := float32(cpower.power)
	if parsedPower < 0 {
		errInfo := fmt.Errorf("get wrong device power, device: %d, power: %f", logicID, parsedPower)
		hwlog.RunLog.Error(errInfo)
		return retError, errInfo
	}
	power := parsedPower * unitChange10
	return power, nil

}

// GetDeviceFrequency get device frequency, unit MHz
// Ascend910 1,6,7,9
// Ascend310 1,2,3,4,5
// subType enum:  Memory,6HBM,AICoreCurrentFreq,AICoreNormalFreq(1,6,7,9)    see DeviceType
func (d *baseDeviceManager) GetDeviceFrequency(logicID int32, subType DeviceType) (int32, error) {
	var cFrequency C.uint
	if err := C.dsmi_get_device_frequency(C.int(logicID), C.int(subType), &cFrequency); err != 0 {
		errInfo := fmt.Errorf("get device%d frequency failed, error code: %d", logicID, int32(err))
		hwlog.RunLog.Error(errInfo)
		return retError, errInfo
	}
	// check whether cFrequency is too big
	if isGreaterThanOrEqualInt32(int64(cFrequency)) {
		errInfo := fmt.Errorf("get wrong device frequency, device: %d, frequency: %d", logicID, int64(cFrequency))
		hwlog.RunLog.Error(errInfo)
		return retError, errInfo
	}
	return int32(cFrequency), nil
}

// GetDeviceHbmInfo mock this function on Ascend310 or Ascend710
func (d *baseDeviceManager) GetDeviceHbmInfo(logicID int32) (*HbmInfo, error) {
	hbmInfo := NewHbmInfo(0, 0, 0, 0, 0)
	return hbmInfo, nil
}

// GetDeviceHbmInfo get HBM information , only for Ascend910
func (d *deviceManager910) GetDeviceHbmInfo(logicID int32) (*HbmInfo, error) {
	var cHbmInfo C.struct_dsmi_hbm_info_stru
	if err := C.dsmi_get_hbm_info(C.int(logicID), &cHbmInfo); err != 0 {
		errInfo := fmt.Errorf("get device%d HBM information failed, error code: %d", logicID, int32(err))
		hwlog.RunLog.Error(errInfo)
		return nil, errInfo
	}
	hbmTemp := int32(cHbmInfo.temp)
	if hbmTemp < 0 {
		errInfo := fmt.Errorf("get wrong device HBM information, device: %d, HBM.temp: %d", logicID, hbmTemp)
		hwlog.RunLog.Error(errInfo)
		return nil, errInfo
	}
	hbmInfo := NewHbmInfo(uint64(cHbmInfo.memory_size)/uint64(OneKilo), uint32(cHbmInfo.freq),
		uint64(cHbmInfo.memory_usage)/uint64(OneKilo), hbmTemp, uint32(cHbmInfo.bandwith_util_rate))

	return hbmInfo, nil
}

// GetDevicePower mock this function on Ascend710
func (d *deviceManager710) GetDevicePower(logicID int32) (float32, error) {
	// Ascend710 not support chip power
	return 0, nil

}

// GetDeviceMemoryInfo get memory information(310 MB  910 KB)
func (d *baseDeviceManager) GetDeviceMemoryInfo(logicID int32) (*MemoryInfo, error) {
	var cmInfo CStructDsmiMemoryInfo
	if err := C.dsmi_get_memory_info(C.int(logicID), &cmInfo); err != 0 {
		errInfo := fmt.Errorf("get device%d memory information failed, error code: %d", logicID, int32(err))
		hwlog.RunLog.Error(errInfo)
		return nil, errInfo
	}
	if !isValidUtilizationRate(uint32(cmInfo.utiliza)) {
		errInfo := fmt.Errorf("get wrong memory utilization, device: %d, utilization: %d", logicID,
			uint32(cmInfo.utiliza))
		hwlog.RunLog.Error(errInfo)
		return nil, errInfo
	}

	dmgr := GetDeviceManager()
	memInfo := dmgr.createMemoryInfoObj(&cmInfo)
	return memInfo, nil
}

// Unit of Ascend310: MB
func (d *deviceManager310) createMemoryInfoObj(cmInfo *CStructDsmiMemoryInfo) *MemoryInfo {
	return NewMemInfo(uint64(cmInfo.memory_size), uint32(cmInfo.freq), uint32(cmInfo.utiliza))
}

// The unit of Ascend910 and Ascend710 is KB. Therefore, you need to convert the unit to MB.
func (d *baseDeviceManager) createMemoryInfoObj(cmInfo *CStructDsmiMemoryInfo) *MemoryInfo {
	return NewMemInfo(
		uint64(cmInfo.memory_size)/uint64(OneKilo),
		uint32(cmInfo.freq),
		uint32(cmInfo.utiliza))
}

// GetDeviceErrCode get the error count and errorcode of the device,only return the first errorcode
func (d *baseDeviceManager) GetDeviceErrCode(logicID int32) (int32, int64, error) {
	var errCount C.int
	var errCodeArray [MaxErrorCodeCount]C.uint
	if err := C.dsmi_get_device_errorcode(C.int(logicID), &errCount, &errCodeArray[0]); err != 0 {
		errInfo := fmt.Errorf("get device%d error code failed, error code: %d", logicID, int32(err))
		hwlog.RunLog.Error(errInfo)
		return retError, retError, errInfo
	}
	errorCodeCount := int32(errCount)
	if errorCodeCount < 0 || errorCodeCount > MaxErrorCodeCount {
		errInfo := fmt.Errorf("get wrong errorcode count, device: %d, errorcode count: %d", logicID, int32(errCount))
		hwlog.RunLog.Error(errInfo)
		return retError, retError, errInfo
	}

	return errorCodeCount, int64(errCodeArray[0]), nil
}

// GetChipInfo get chip info
func (d *baseDeviceManager) GetChipInfo(logicID int32) (*ChipInfo, error) {
	var chipInfo C.struct_dsmi_chip_info_stru
	if err := C.dsmi_get_chip_info(C.int(logicID), &chipInfo); err != 0 {
		errInfo := fmt.Errorf("get device%d ChipIno information failed, error code: %d", logicID, int32(err))
		hwlog.RunLog.Error(errInfo)
		return nil, errInfo
	}
	var name []rune
	var cType []rune
	var ver []rune
	name = convertToCharArr(name, chipInfo.chip_name)
	cType = convertToCharArr(cType, chipInfo.chip_type)
	ver = convertToCharArr(ver, chipInfo.chip_ver)
	chip := &ChipInfo{
		ChipName: string(name),
		ChipType: string(cType),
		ChipVer:  string(ver),
	}
	// If the name, type, and version are both empty, false is returned
	if !isValidChipInfo(chip) {
		errInfo := fmt.Errorf("get device%d ChipIno information failed, chip info is empty", logicID)
		hwlog.RunLog.Error(errInfo)
		return nil, errInfo
	}
	return chip, nil
}

// GetNPUMajorID query the MajorID of NPU devices
func (d *baseDeviceManager) GetNPUMajorID() ([]string, error) {
	path, err := utils.CheckPath("/proc/devices")
	if err != nil {
		return nil, err
	}
	majorID := make([]string, 0, deviceCount)
	f, err := os.Open(path)
	if err != nil {
		return majorID, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	count := 0
	for s.Scan() {
		// prevent from searching too many lines
		if count > maxSearchLine {
			break
		}
		count++
		text := s.Text()
		matched, err := regexp.MatchString("^[0-9]{1,3}\\s[v]?devdrv-cdev$", text)
		if err != nil {
			return majorID, err
		}
		if !matched {
			continue
		}
		fields := strings.Fields(text)
		majorID = append(majorID, fields[0])
	}
	return majorID, nil
}

// GetCardList get npu card array
func (d *baseDeviceManager) GetCardList() (int32, []int32, error) {
	var ids [HIAIMaxCardNum]C.int
	var cNum C.int
	if err := C.dcmi_get_card_num_list(&cNum, &ids[0], HIAIMaxCardNum); err != 0 {
		errInfo := fmt.Errorf("get card list failed, error code: %d", int32(err))
		hwlog.RunLog.Error(errInfo)
		return retError, nil, errInfo
	}
	// checking card's quantity
	if cNum <= 0 {
		errInfo := fmt.Errorf("get error card quantity: %d", int32(cNum))
		hwlog.RunLog.Error(errInfo)
		return retError, nil, errInfo
	}
	var cardNum = int32(cNum)
	var i int32
	var cardIDList []int32
	for i = 0; i < cardNum && i < HIAIMaxCardNum; i++ {
		cardID := int32(ids[i])
		if cardID < 0 {
			errInfo := fmt.Errorf("get invalid card ID: %d", cardID)
			hwlog.RunLog.Error(errInfo)
			continue
		}
		cardIDList = append(cardIDList, cardID)
	}
	return cardNum, cardIDList, nil
}

// GetDeviceNumOnCard get device number on the npu card
func (d *baseDeviceManager) GetDeviceNumOnCard(cardID int32) (int32, error) {
	var deviceNum C.int
	if err := C.dcmi_get_device_num_in_card(C.int(cardID), &deviceNum); err != 0 {
		errInfo := fmt.Errorf("get device count on the card failed, error code: %d", int32(err))
		hwlog.RunLog.Error(errInfo)
		return retError, errInfo
	}
	if deviceNum <= 0 {
		errInfo := fmt.Errorf("the number of chips obtained is invalid, the number is: %d", int32(deviceNum))
		hwlog.RunLog.Error(errInfo)
		return retError, errInfo
	}
	return int32(deviceNum), nil
}

// GetCardPower get card power with Ascend710
func (d *baseDeviceManager) GetCardPower(cardID int32) (float32, error) {
	var power C.int
	if err := C.dcmi_mcu_get_power_info(C.int(cardID), &power); err != 0 {
		errInfo := fmt.Errorf("get card power failed, error code: %d", int32(err))
		hwlog.RunLog.Error(errInfo)
		return retError, errInfo
	}

	parsedPower := float32(power)
	if parsedPower < 0 {
		errInfo := fmt.Errorf("get wrong device power, cardID: %d, power: %f", int32(cardID), parsedPower)
		hwlog.RunLog.Error(errInfo)
		return retError, errInfo
	}
	return parsedPower * unitChange10, nil
}

// GetDeviceLogicID get device logicID
func (d *baseDeviceManager) GetDeviceLogicID(cardID, deviceID int32) (int32, error) {
	var logicID C.int
	if err := C.dcmi_get_device_logic_id(&logicID, C.int(cardID), C.int(deviceID)); err != 0 {
		errInfo := fmt.Errorf("get logicID failed, error code: %d", int32(err))
		hwlog.RunLog.Error(errInfo)
		return retError, errInfo
	}

	// check whether phyID is too big
	if uint32(logicID) > uint32(math.MaxInt8) {
		errInfo := fmt.Errorf("the logicID value is invalid,logicID is: %d", logicID)
		hwlog.RunLog.Error(errInfo)
		return retError, errInfo
	}
	return int32(logicID), nil
}

func init() {
	C.dsmiInit_dl()
	C.dcmi_init()
}

func convertToCharArr(charArr []rune, cgoArr [maxChipName]C.uchar) []rune {
	for _, v := range cgoArr {
		if v != 0 {
			charArr = append(charArr, rune(v))
		}
	}
	return charArr
}

// IsAscend910 check chipName
func IsAscend910(chipName string) bool {
	return strings.Contains(chipName, "910")
}

// IsAscend710 check chipName
func IsAscend710(chipName string) bool {
	return strings.Contains(chipName, "710")
}

// ShutDown clean the dynamically loaded resource
func ShutDown() {
	C.dsmiShutDown()
}

func isValidChipInfo(chip *ChipInfo) bool {
	chipName := chip.ChipName
	chipT := chip.ChipType
	chipVer := chip.ChipVer

	if chipName == "" && chipT == "" && chipVer == "" {
		return false
	}

	return true
}

// valid utilization rate is 0-100
func isValidUtilizationRate(num uint32) bool {
	if num > uint32(Percent) || num < 0 {
		return false
	}

	return true
}

func isGreaterThanOrEqualInt32(num int64) bool {
	if num >= int64(math.MaxInt32) {
		return true
	}

	return false
}
