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

// Package dsmi convert the npu driver interface to go style
package dsmi

// #cgo LDFLAGS: -ldl
/*
 #include <stddef.h>
#include <dlfcn.h>
#include <stdlib.h>

#include "dsmi_common_interface.h"

// dsmiHandle is the handle for dynamically loaded libdrvdsmi_host.so
void *dsmiHandle;
#define SO_NOT_FOUND  -99999
#define FUNCTION_NOT_FOUND  -99998
#define SUCCESS  0
#define ERROR_UNKNOWN  -99997
#define CALL_FUNC(func_name, ...) 						\
	if (func_name##_func == NULL){  					\
			return FUNCTION_NOT_FOUND; 					\
		}  												\
    return func_name##_func(__VA_ARGS__);				\

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
int dsmi_get_device_utilization_rate(int device_id, int device_type, unsigned int *putilization_rate){
	CALL_FUNC(dsmi_get_device_utilization_rate,device_id,device_type,putilization_rate)
}

int (*dsmi_get_phyid_from_logicid_func)(unsigned int logicid, unsigned int *phyid);
int dsmi_get_phyid_from_logicid(unsigned int logicid, unsigned int *phyid){
	CALL_FUNC(dsmi_get_phyid_from_logicid,logicid,phyid)
}

int (*dsmi_get_logicid_from_phyid_func)(unsigned int phyid, unsigned int *logicid);
int dsmi_get_logicid_from_phyid(unsigned int phyid, unsigned int *logicid){
	CALL_FUNC(dsmi_get_logicid_from_phyid,phyid,logicid)
}

int (*dsmi_get_device_temperature_func)(int device_id, int *ptemperature);
int dsmi_get_device_temperature(int device_id, int *ptemperature){
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

int (*dsmi_get_device_frequency_func)(int device_id, int device_type, unsigned int *pfrequency);
int dsmi_get_device_frequency(int device_id, int device_type, unsigned int *pfrequency){
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

int (*dsmi_get_device_errorcode_func)(int device_id, int *errorcount, unsigned int *perrorcode);
int dsmi_get_device_errorcode(int device_id, int *errorcount, unsigned int *perrorcode){
	CALL_FUNC(dsmi_get_device_errorcode,device_id,errorcount,perrorcode)
}

int (*dsmi_get_chip_info_func)(int device_id, struct dsmi_chip_info_stru *chip_info);
int dsmi_get_chip_info(int device_id, struct dsmi_chip_info_stru *chip_info){
	CALL_FUNC(dsmi_get_chip_info,device_id,chip_info)
}

// load .so files and functions
int dsmiInit_dl(void){
	dsmiHandle = dlopen("libdrvdsmi_host.so",RTLD_LAZY);
	if (dsmiHandle == NULL) {
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

	return SUCCESS;
}

int dsmiShutDown(void){
	if (dsmiHandle == NULL) {
   	 	return SUCCESS;
  	}
	return (dlclose(dsmiHandle) ? ERROR_UNKNOWN : SUCCESS);
}
*/
import "C"
import (
	"bufio"
	"fmt"
	"io"
	"k8s.io/klog"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

const (
	// the error value when the function failed
	retError = -1
	// max chip name length
	maxChipName = 32
	// Percent constant of 100
	Percent = 100
	// OneKilo for unit change kb to mb
	OneKilo = 1024
)

// HbmInfo HBM info
type HbmInfo struct {
	MemorySize              uint32 `json:"memory_size"`        // HBM total size,KB
	MemoryFrequency         uint32 `json:"hbm_frequency"`      // HBM frequncy MHz
	MemoryUsage             uint32 `json:"memory_usage"`       // HBM memory usagem,KB
	MemoryTemp              int32  `json:"hbm_temperature"`    // HBM temperature
	MemoryBandWidthUtilRate uint32 `json:"hbm_bandwidth_util"` // HBM brandwidth utilization

}

// MemoryInfo memory infomation struct
type MemoryInfo struct {
	MemorySize  uint32 `json:"memory_size"`
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
func NewMemInfo(memorySize, frequency, utilization uint32) *MemoryInfo {
	return &MemoryInfo{
		MemorySize:  memorySize,
		Frequency:   frequency,
		Utilization: utilization,
	}
}

// NewHbmInfo new HbmInfo
func NewHbmInfo(memorySize, memoryFrequency, memoryUsage uint32, memoryTemp int32,
	memoryBandWidthUtilRate uint32) *HbmInfo {
	return &HbmInfo{
		MemorySize:              memorySize,
		MemoryFrequency:         memoryFrequency,
		MemoryUsage:             memoryUsage,
		MemoryTemp:              memoryTemp,
		MemoryBandWidthUtilRate: memoryBandWidthUtilRate,
	}
}

// DeviceMgrInterface interface for dsmi
type DeviceMgrInterface interface {
	// GetDeviceCount get npu device count
	GetDeviceCount() (int32, error)
	// GetDeviceList get npu device array
	GetDeviceList() (int32, []int32, error)
	// GetDeviceHealth query npu device health status
	GetDeviceHealth(logicID int32) (int32, error)
	// GetDeviceUtilizationRate get npu device utilization
	GetDeviceUtilizationRate(logicID int32, deviceType DeviceType) (int32, error)
	// GetDeviceTemperature get npu device tempetature
	GetDeviceTemperature(logicID int32) (int32, error)
	// GetDeviceVoltage get npu device voltage
	GetDeviceVoltage(logicID int32) (float32, error)
	// GetDevicePower get npu device power
	GetDevicePower(logicID int32) (float32, error)
	// GetDeviceFrequency get npu device work frequency
	GetDeviceFrequency(logicID int32, subType DeviceType) (int32, error)
	// GetDeviceMemoryInfo get npu device memory information
	GetDeviceMemoryInfo(logicID int32) (*MemoryInfo, error)
	// GetDeviceHbmInfo get npu HBM module memory and frequency information
	GetDeviceHbmInfo(logicID int32) (*HbmInfo, error)
	// GetDeviceErrCode get npu device error code
	GetDeviceErrCode(logicID int32) (int32, int32, error)
	// GetChipInfo get npu device ascend chip information
	GetChipInfo(logicID int32) (*ChipInfo, error)
	// GetPhyIDFromLogicID convert npu device physicalID to logicID
	GetPhyIDFromLogicID(logicID uint32) (int32, error)
	// GetLogicIDFromPhyID convert npu device logicID to physicalID
	GetLogicIDFromPhyID(phyID uint32) (int32, error)
	// GetNPUMajorID query the NPU majorId from OS
	GetNPUMajorID() (string, error)
}

// please use GetDeviceManager to get the singleton instance of baseDeviceManager
type baseDeviceManager struct {
	// abstract method
	GetDeviceHbmInfo func(logicID int32) (*HbmInfo, error)
}

type deviceManager910 struct {
	baseDeviceManager
}

type deviceManager310 struct {
	baseDeviceManager
}

var instance DeviceMgrInterface
var once sync.Once

// GetDeviceManager new baseDeviceManager singleton instance
func GetDeviceManager() DeviceMgrInterface {
	once.Do(func() {
		instance = &deviceManager310{}
		num, _, err := instance.GetDeviceList()
		if err != nil || num == 0 {
			klog.Error("This is no device on this machine")
			return
		}
		chip, err := instance.GetChipInfo(0)
		if err != nil {
			klog.Error(err)
			return
		}
		if IsAscend910(chip.ChipName) {
			klog.Info("change the instance to deviceManager910")
			instance = &deviceManager910{}
		}
	})
	return instance
}

// GetDeviceCount get ascend910 device quantity
func (d *baseDeviceManager) GetDeviceCount() (int32, error) {
	var count C.int
	if err := C.dsmi_get_device_count(&count); err != 0 {
		errInfo := fmt.Errorf("get device quantity failed, error code: %d", int32(err))
		klog.Error(errInfo)
		return retError, errInfo
	}
	return int32(count), nil
}

// GetDeviceList  get device list
func (d *baseDeviceManager) GetDeviceList() (int32, []int32, error) {
	var devices []int32
	devNum, err := d.GetDeviceCount()
	if err != nil {
		return devNum, devices, err
	}
	var ids [HiAIMaxDeviceNum]C.int
	if err := C.dsmi_list_device(&ids[0], C.int(devNum)); err != 0 {
		errInfo := fmt.Errorf("unable to get device list, return error: %d", int32(err))
		klog.Error(errInfo)
		return retError, devices, errInfo
	}
	// transfer device list
	var i int32
	for i = 0; i < devNum && i < int32(len(ids)-1); i++ {
		devices = append(devices, int32(ids[i]))
	}
	return devNum, devices, nil
}

// GetDeviceHealth get device health by id
func (d *baseDeviceManager) GetDeviceHealth(logicID int32) (int32, error) {
	var health C.uint
	if err := C.dsmi_get_device_health(C.int(logicID), &health); err != 0 {
		errInfo := fmt.Errorf("get device %d health state failed, error code: %d", logicID, int32(err))
		klog.Error(errInfo)
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
		klog.Errorf("Get Device utilize rate failed, error code: %d, try again ... ", int32(err))
		for i := 0; i < 3; i++ {
			klog.Errorf("try again %d", i)
			err = C.dsmi_get_device_utilization_rate(C.int(logicID), C.int(deviceType), &utilRate)
			if err == 0 {
				return int32(utilRate), nil
			}
		}
		return retError, fmt.Errorf("get Device utilize rate failed, error code: %d", int32(err))

	}
	return int32(utilRate), nil
}

// GetPhyIDFromLogicID get physic id form logic id
func (d *baseDeviceManager) GetPhyIDFromLogicID(logicID uint32) (int32, error) {
	var phyID C.uint
	if err := C.dsmi_get_phyid_from_logicid(C.uint(logicID), &phyID); err != 0 {
		errInfo := fmt.Errorf("get phy id failed ,error code is: %d", int32(err))
		klog.Error(errInfo)
		return retError, errInfo
	}
	return int32(phyID), nil
}

// GetLogicIDFromPhyID get logic id form physic id
func (d *baseDeviceManager) GetLogicIDFromPhyID(phyID uint32) (int32, error) {
	var logicID C.uint
	if err := C.dsmi_get_logicid_from_phyid(C.uint(phyID), &logicID); err != 0 {
		errInfo := fmt.Errorf("get logic id failed ,error code is : %d", int32(err))
		klog.Error(errInfo)
		return retError, errInfo
	}
	return int32(logicID), nil
}

// GetDeviceTemperature get the device temperature
func (d *baseDeviceManager) GetDeviceTemperature(logicID int32) (int32, error) {
	var temp C.int
	if err := C.dsmi_get_device_temperature(C.int(logicID), &temp); err != 0 {
		errInfo := fmt.Errorf("get temperature failed ,error code is : %d", int32(err))
		klog.Error(errInfo)
		return retError, errInfo
	}
	return int32(temp), nil
}

// GetDeviceVoltage get the device voltage
func (d *baseDeviceManager) GetDeviceVoltage(logicID int32) (float32, error) {
	var vol C.uint
	if err := C.dsmi_get_device_voltage(C.int(logicID), &vol); err != 0 {
		errInfo := fmt.Errorf("get temperature failed ,error code is : %d", int32(err))
		klog.Error(errInfo)
		return retError, errInfo
	}
	voltage := float32(vol) * 0.01
	return voltage, nil
}

// GetDevicePower get the power info of the device, the result like : 8.2w
func (d *baseDeviceManager) GetDevicePower(logicID int32) (float32, error) {
	var cpower C.struct_dsmi_power_info_stru
	if err := C.dsmi_get_device_power_info(C.int(logicID), &cpower); err != 0 {
		errInfo := fmt.Errorf("get device power failed, error code: %d", int32(err))
		klog.Error(errInfo)
		return retError, errInfo
	}
	power := float32(cpower.power) * 0.1
	return power, nil

}

// GetDeviceFrequency get device frequency, unit MHz
// Ascend910 1,6,7,9
// Ascend310 1,2,3,4,5
// subType enum:  Memory,6HBM,AI_Core_Current_Fre,AI_Core_Normal_Fre(1,6,7,9)    see DeviceType
func (d *baseDeviceManager) GetDeviceFrequency(logicID int32, subType DeviceType) (int32, error) {
	var cFrequency C.uint
	if err := C.dsmi_get_device_frequency(C.int(logicID), C.int(subType), &cFrequency); err != 0 {
		errInfo := fmt.Errorf("get device frequency failed, error code: %d", int32(err))
		klog.Error(errInfo)
		return retError, errInfo
	}
	return int32(cFrequency), nil
}

// GetDeviceHbmInfo get HBM information , only for Ascend910
func (d *deviceManager910) GetDeviceHbmInfo(logicID int32) (*HbmInfo, error) {
	var cHbmInfo C.struct_dsmi_hbm_info_stru
	if err := C.dsmi_get_hbm_info(C.int(logicID), &cHbmInfo); err != 0 {
		errInfo := fmt.Errorf("get device HBM information failed, error code: %d", int32(err))
		klog.Error(errInfo)
		return nil, errInfo
	}
	hbmInfo := NewHbmInfo(uint32(cHbmInfo.memory_size/OneKilo), uint32(cHbmInfo.freq),
		uint32(cHbmInfo.memory_usage/OneKilo), int32(cHbmInfo.temp), uint32(cHbmInfo.bandwith_util_rate))

	return hbmInfo, nil
}

// GetDeviceHbmInfo mock this function on Ascend310
func (d *deviceManager310) GetDeviceHbmInfo(logicID int32) (*HbmInfo, error) {
	hbmInfo := NewHbmInfo(0, 0, 0, 0, 0)
	return hbmInfo, nil
}

// GetDeviceMemoryInfo get memory information(310 MB  910 KB)
func (d *baseDeviceManager) GetDeviceMemoryInfo(logicID int32) (*MemoryInfo, error) {
	var cmInfo C.struct_dsmi_memory_info_stru
	if err := C.dsmi_get_memory_info(C.int(logicID), &cmInfo); err != 0 {
		errInfo := fmt.Errorf("get device memory information failed, error code: %d", int32(err))
		klog.Error(errInfo)
		return nil, errInfo
	}
	chip, errs := d.GetChipInfo(logicID)
	var memInfo *MemoryInfo
	if errs == nil && IsAscend910(chip.ChipName) {
		// change unit to MB
		memInfo = NewMemInfo(uint32(cmInfo.memory_size/OneKilo), uint32(cmInfo.freq), uint32(cmInfo.utiliza))
	} else {
		memInfo = NewMemInfo(uint32(cmInfo.memory_size), uint32(cmInfo.freq), uint32(cmInfo.utiliza))
	}
	return memInfo, nil
}

// GetDeviceErrCode get the error count and errorcode of the device
func (d *baseDeviceManager) GetDeviceErrCode(logicID int32) (int32, int32, error) {
	var errCount C.int
	var errCode C.uint
	if err := C.dsmi_get_device_errorcode(C.int(logicID), &errCount, &errCode); err != 0 {
		errInfo := fmt.Errorf("get device error code failed, error code: %d", int32(err))
		klog.Error(errInfo)
		return retError, retError, errInfo
	}
	return int32(errCount), int32(errCode), nil
}

// GetChipInfo get chip info
func (d *baseDeviceManager) GetChipInfo(logicID int32) (*ChipInfo, error) {
	var chipInfo C.struct_dsmi_chip_info_stru
	if err := C.dsmi_get_chip_info(C.int(logicID), &chipInfo); err != 0 {
		errInfo := fmt.Errorf("get device HBM information failed, error code: %d", int32(err))
		klog.Error(errInfo)
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
	return chip, nil
}

// GetNPUMajorId query the NPU majorId from OS
func (d *baseDeviceManager) GetNPUMajorID() (string, error) {
	command := "ls -al /dev/davinci*"
	cmd := exec.Command("/bin/sh", "-c", command)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		klog.Errorf("command exec failed:%s", command)
		return "", fmt.Errorf("ls command exec failed")
	}
	err = cmd.Start()
	if err != nil {
		klog.Errorf("start cmd error:%v", err)
	}
	reader := bufio.NewReader(stdout)
	var firstResult string
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			klog.Errorf("err:%v", err2)
			break
		}
		klog.Infof("LINE:%v", line)
		if line != "" {
			firstResult = line
			break
		}
	}
	if err := cmd.Wait(); err != nil {
		klog.Errorf("command exec failed，%v", err)
		return "", err
	}
	if firstResult == "" {
		klog.Errorf("can't to find NPU majorId")
		return "", fmt.Errorf("can't find NPU majorId")
	}
	// for example like this : crw-rw---- 1 HwHiAiUser HwHiAiUser 239, 0 Sep 28 21:56 /dev/davinci0
	reg := regexp.MustCompile("\\d{3},")
	res := reg.FindString(firstResult)
	if res == "" {
		klog.Errorf("regex match error,original string is %s", firstResult)
		return "", fmt.Errorf("regex match error")
	}
	return strings.TrimSuffix(res, ","), nil
}

func init() {
	C.dsmiInit_dl()
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

// ShutDown clean the dynamically loaded resource
func ShutDown() {
	C.dsmiShutDown()
}
