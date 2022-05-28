//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package dcmi this for dcmi manager
package dcmi

// #cgo LDFLAGS: -ldl
/*
   #include <stddef.h>
   #include <dlfcn.h>
   #include <stdlib.h>
   #include <stdio.h>

   #include "dcmi_interface_api.h"

   void *dcmiHandle;
   #define SO_NOT_FOUND  -99999
   #define FUNCTION_NOT_FOUND  -99998
   #define SUCCESS  0
   #define ERROR_UNKNOWN  -99997
   #define	CALL_FUNC(name,...) if(name##_func==NULL){return FUNCTION_NOT_FOUND;}return name##_func(__VA_ARGS__);

   // dcmi
   int (*dcmi_init_func)();
   int dcmi_init(){
   	CALL_FUNC(dcmi_init)
   }

   int (*dcmi_get_card_num_list_func)(int *card_num,int *card_list,int list_length);
   int dcmi_get_card_num_list(int *card_num,int *card_list,int list_length){
   	CALL_FUNC(dcmi_get_card_num_list,card_num,card_list,list_length)
   }

   int (*dcmi_get_device_num_in_card_func)(int card_id,int *device_num);
   int dcmi_get_device_num_in_card(int card_id,int *device_num){
   	CALL_FUNC(dcmi_get_device_num_in_card,card_id,device_num)
   }

   int (*dcmi_get_device_logic_id_func)(int *device_logic_id,int card_id,int device_id);
   int dcmi_get_device_logic_id(int *device_logic_id,int card_id,int device_id){
   	CALL_FUNC(dcmi_get_device_logic_id,device_logic_id,card_id,device_id)
   }

   int (*dcmi_create_vdevice_func)(int card_id, int device_id, int vdev_id, const char *template_name,
   	struct dcmi_create_vdev_out *out);
   int dcmi_create_vdevice(int card_id, int device_id, int vdev_id, const char *template_name,
   	struct dcmi_create_vdev_out *out){
   	CALL_FUNC(dcmi_create_vdevice,card_id,device_id,vdev_id,template_name,out)
   }

   int (*dcmi_get_device_info_func)(int card_id, int device_id, enum dcmi_main_cmd main_cmd, unsigned int sub_cmd,
   	void *buf, unsigned int *size);
   int dcmi_get_device_info(int card_id, int device_id, enum dcmi_main_cmd main_cmd, unsigned int sub_cmd, void *buf,
   	unsigned int *size){
   	CALL_FUNC(dcmi_get_device_info,card_id,device_id,main_cmd,sub_cmd,buf,size)
   }

   int (*dcmi_set_destroy_vdevice_func)(int card_id,int device_id, unsigned int VDevid);
   int dcmi_set_destroy_vdevice(int card_id,int device_id, unsigned int VDevid){
   	CALL_FUNC(dcmi_set_destroy_vdevice,card_id,device_id,VDevid)
   }

   int (*dcmi_get_device_type_func)(int card_id,int device_id,enum dcmi_unit_type *device_type);
   int dcmi_get_device_type(int card_id,int device_id,enum dcmi_unit_type *device_type){
   	CALL_FUNC(dcmi_get_device_type,card_id,device_id,device_type)
   }

   int (*dcmi_get_device_health_func)(int card_id, int device_id, unsigned int *health);
   int dcmi_get_device_health(int card_id, int device_id, unsigned int *health){
   	CALL_FUNC(dcmi_get_device_health,card_id,device_id,health)
   }

   int (*dcmi_get_device_utilization_rate_func)(int card_id, int device_id, int input_type,
    unsigned int *utilization_rate);
   int dcmi_get_device_utilization_rate(int card_id, int device_id, int input_type, unsigned int *utilization_rate){
   	CALL_FUNC(dcmi_get_device_utilization_rate,card_id,device_id,input_type,utilization_rate)
   }

   int (*dcmi_get_device_temperature_func)(int card_id, int device_id, int *temperature);
   int dcmi_get_device_temperature(int card_id, int device_id, int *temperature){
    CALL_FUNC(dcmi_get_device_temperature,card_id,device_id,temperature)
   }

   int (*dcmi_get_device_voltage_func)(int card_id, int device_id, unsigned int *voltage);
   int dcmi_get_device_voltage(int card_id, int device_id, unsigned int *voltage){
    CALL_FUNC(dcmi_get_device_voltage,card_id,device_id,voltage)
   }

   int (*dcmi_get_device_power_info_func)(int card_id, int device_id, int *power);
   int dcmi_get_device_power_info(int card_id, int device_id, int *power){
    CALL_FUNC(dcmi_get_device_power_info,card_id,device_id,power)
   }

   int (*dcmi_get_device_frequency_func)(int card_id, int device_id, enum dcmi_freq_type input_type,
    unsigned int *frequency);
   int dcmi_get_device_frequency(int card_id, int device_id, enum dcmi_freq_type input_type, unsigned int *frequency){
    CALL_FUNC(dcmi_get_device_frequency,card_id,device_id,input_type,frequency)
   }

   int (*dcmi_get_device_memory_info_v3_func)(int card_id, int device_id,
    struct dcmi_get_memory_info_stru *memory_info);
   int dcmi_get_device_memory_info_v3(int card_id, int device_id, struct dcmi_get_memory_info_stru *memory_info){
    CALL_FUNC(dcmi_get_device_memory_info_v3,card_id,device_id,memory_info)
   }

   int (*dcmi_get_device_hbm_info_func)(int card_id, int device_id, struct dcmi_hbm_info *hbm_info);
   int dcmi_get_device_hbm_info(int card_id, int device_id, struct dcmi_hbm_info *hbm_info){
    CALL_FUNC(dcmi_get_device_hbm_info,card_id,device_id,hbm_info)
   }

   int (*dcmi_get_device_errorcode_v2_func)(int card_id, int device_id, int *error_count, unsigned int *error_code_list,
    unsigned int list_len);
   int dcmi_get_device_errorcode_v2(int card_id, int device_id, int *error_count,
    unsigned int *error_code_list, unsigned int list_len){
    CALL_FUNC(dcmi_get_device_errorcode_v2,card_id,device_id,error_count,error_code_list,list_len)
   }

   int (*dcmi_get_device_chip_info_func)(int card_id, int device_id, struct dcmi_chip_info *chip_info);
   int dcmi_get_device_chip_info(int card_id, int device_id, struct dcmi_chip_info *chip_info){
    CALL_FUNC(dcmi_get_device_chip_info,card_id,device_id,chip_info)
   }

   int (*dcmi_get_device_phyid_from_logicid_func)(unsigned int logicid, unsigned int *phyid);
   int dcmi_get_device_phyid_from_logicid(unsigned int logicid, unsigned int *phyid){
    CALL_FUNC(dcmi_get_device_phyid_from_logicid,logicid,phyid)
   }

   int (*dcmi_get_device_logicid_from_phyid_func)(unsigned int phyid, unsigned int *logicid);
   int dcmi_get_device_logicid_from_phyid(unsigned int phyid, unsigned int *logicid){
    CALL_FUNC(dcmi_get_device_logicid_from_phyid,phyid,logicid)
   }

   int (*dcmi_get_device_ip_func)(int card_id, int device_id, enum dcmi_port_type input_type, int port_id,
    struct dcmi_ip_addr *ip, struct dcmi_ip_addr *mask);
   int dcmi_get_device_ip(int card_id, int device_id, enum dcmi_port_type input_type, int port_id,
    struct dcmi_ip_addr *ip, struct dcmi_ip_addr *mask){
    CALL_FUNC(dcmi_get_device_ip,card_id,device_id,input_type,port_id,ip,mask)
   }

   int (*dcmi_get_device_network_health_func)(int card_id, int device_id, enum dcmi_rdfx_detect_result *result);
   int dcmi_get_device_network_health(int card_id, int device_id, enum dcmi_rdfx_detect_result *result){
    CALL_FUNC(dcmi_get_device_network_health,card_id,device_id,result)
   }

   int (*dcmi_get_card_list_func)(int *card_num, int *card_list, int list_len);
   int dcmi_get_card_list(int *card_num, int *card_list, int list_len){
    CALL_FUNC(dcmi_get_card_list,card_num,card_list,list_len)
   }

   int (*dcmi_get_device_id_in_card_func)(int card_id, int *device_id_max, int *mcu_id, int *cpu_id);
   int dcmi_get_device_id_in_card(int card_id, int *device_id_max, int *mcu_id, int *cpu_id){
    CALL_FUNC(dcmi_get_device_id_in_card,card_id,device_id_max,mcu_id,cpu_id)
   }

   // load .so files and functions
   int dcmiInit_dl(void){
   	dcmiHandle = dlopen("libdcmi.so",RTLD_LAZY | RTLD_GLOBAL);
   	if (dcmiHandle == NULL){
   		fprintf (stderr,"%s\n",dlerror());
   		return SO_NOT_FOUND;
   	}

   	dcmi_init_func = dlsym(dcmiHandle,"dcmi_init");

   	dcmi_get_card_num_list_func = dlsym(dcmiHandle,"dcmi_get_card_num_list");

   	dcmi_get_device_num_in_card_func = dlsym(dcmiHandle,"dcmi_get_device_num_in_card");

   	dcmi_get_device_logic_id_func = dlsym(dcmiHandle,"dcmi_get_device_logic_id");

   	dcmi_create_vdevice_func = dlsym(dcmiHandle,"dcmi_create_vdevice");

   	dcmi_get_device_info_func = dlsym(dcmiHandle,"dcmi_get_device_info");

   	dcmi_set_destroy_vdevice_func = dlsym(dcmiHandle,"dcmi_set_destroy_vdevice");

   	dcmi_get_device_type_func = dlsym(dcmiHandle,"dcmi_get_device_type");

   	dcmi_get_device_health_func = dlsym(dcmiHandle,"dcmi_get_device_health");

   	dcmi_get_device_utilization_rate_func = dlsym(dcmiHandle,"dcmi_get_device_utilization_rate");

   	dcmi_get_device_temperature_func = dlsym(dcmiHandle,"dcmi_get_device_temperature");

   	dcmi_get_device_voltage_func = dlsym(dcmiHandle,"dcmi_get_device_voltage");

   	dcmi_get_device_power_info_func = dlsym(dcmiHandle,"dcmi_get_device_power_info");

   	dcmi_get_device_frequency_func = dlsym(dcmiHandle,"dcmi_get_device_frequency");

   	dcmi_get_device_memory_info_v3_func = dlsym(dcmiHandle,"dcmi_get_device_memory_info_v3");

   	dcmi_get_device_hbm_info_func = dlsym(dcmiHandle,"dcmi_get_device_hbm_info");

   	dcmi_get_device_errorcode_v2_func = dlsym(dcmiHandle,"dcmi_get_device_errorcode_v2");

   	dcmi_get_device_chip_info_func = dlsym(dcmiHandle,"dcmi_get_device_chip_info");

   	dcmi_get_device_phyid_from_logicid_func = dlsym(dcmiHandle,"dcmi_get_device_phyid_from_logicid");

   	dcmi_get_device_logicid_from_phyid_func = dlsym(dcmiHandle,"dcmi_get_device_logicid_from_phyid");

   	dcmi_get_device_ip_func = dlsym(dcmiHandle,"dcmi_get_device_ip");

   	dcmi_get_device_network_health_func = dlsym(dcmiHandle,"dcmi_get_device_network_health");

   	dcmi_get_card_list_func = dlsym(dcmiHandle,"dcmi_get_card_list");

   	dcmi_get_device_id_in_card_func = dlsym(dcmiHandle,"dcmi_get_device_id_in_card");

   	return SUCCESS;
   }

   int dcmiShutDown(void){
   	if (dcmiHandle == NULL) {
   		return SUCCESS;
   	}
   	return (dlclose(dcmiHandle) ? ERROR_UNKNOWN : SUCCESS);
   }
*/
import "C"
import (
	"fmt"
	"math"
	"unsafe"

	"huawei.com/npu-exporter/devmanager/common"
	"huawei.com/npu-exporter/hwlog"
)

// DriverInterface interface for dcmi
type DriverInterface interface {
	Init()
	ShutDown()

	GetDeviceCount() (int32, error)
	GetDeviceHealth(int32, int32) (int32, error)
	GetDeviceNetWorkHealth(int32, int32) (uint32, error)
	GetDeviceUtilizationRate(int32, int32, deviceType) (int32, error)
	GetDeviceTemperature(int32, int32) (int32, error)
	GetDeviceVoltage(int32, int32) (float32, error)
	GetDevicePowerInfo(int32, int32) (float32, error)
	GetDeviceFrequency(int32, int32, deviceType) (int32, error)
	GetMemoryInfoV3(int32, int32) (*MemoryInfo, error)
	GetHbmInfo(int32, int32) (*HbmInfo, error)
	GetDeviceErrorCodeV2(int32, int32) (int32, int64, error)
	GetChipInfo(int32, int32) (*ChipInfo, error)
	GetPhysicIDFromLogicID(uint32) (int32, error)
	GetLogicIDFromPhysicID(uint32) (int32, error)
	GetDeviceLogicID(int32, int32) (int32, error)
	GetDeviceIPAddress(int32, int32) (string, error)

	GetCardList() (int32, []int32, error)
	GetDeviceNumInCard(int32) (int32, error)
	SetDestroyVirtualDevice(int32, int32, uint32) error
	CreateVirtualDevice(int32, int32, int32, uint32) (CgoDcmiCreateVDevOut, error)
	GetDeviceVDevResource(int32, int32, uint32) (CgoVDevQueryStru, error)
	GetDeviceTotalResource(int32, int32) (CgoDcmiSocTotalResource, error)
	GetDeviceFreeResource(int32, int32) (CgoDcmiSocFreeResource, error)
	GetDeviceInfo(int32, int32) (CgoVDevInfo, error)
	GetCardIDDeviceID(uint32) (int32, int32, error)
	CreateVDevice(uint32, uint32) (uint32, error)
	GetVDeviceInfo(uint32) (CgoVDevInfo, error)
	DestroyVDevice(uint32, uint32) error
}

// DcManager for manager dcmi interface
type DcManager struct{}

// NewDcManager new dcmi manager instance
func NewDcManager() *DcManager {
	return &DcManager{}
}

// Init load symbol and initialize dcmi
func Init() {
	if err := C.dcmiInit_dl(); err != C.SUCCESS {
		fmt.Printf("dcmi lib load failed, error code: %d\n", int32(err))
		return
	}
	if err := C.dcmi_init(); err != C.SUCCESS {
		fmt.Printf("dcmi init failed, error code: %d\n", int32(err))
	}
}

// ShutDown clean the dynamically loaded resource
func (d *DcManager) ShutDown() {
	if err := C.dcmiShutDown(); err != C.SUCCESS {
		hwlog.RunLog.Errorf("dcmi shut down failed, error code: %d\n", int32(err))
	}
}

// GetCardList get card list
func (d *DcManager) GetCardList() (int32, []int32, error) {
	var ids [hiAIMaxCardNum]C.int
	var cNum C.int
	// follow: dcmi_get_card_num_list will be replaced in feature
	if err := C.dcmi_get_card_num_list(&cNum, &ids[0], hiAIMaxCardNum); err != 0 {
		return retError, nil, fmt.Errorf("get card list failed, error code: %d", int32(err))
	}
	// checking card's quantity
	if cNum <= 0 || cNum > hiAIMaxCardNum {
		return retError, nil, fmt.Errorf("get error card quantity: %d", int32(cNum))
	}
	var cardNum = int32(cNum)
	var i int32
	var cardIDList []int32
	for i = 0; i < cardNum; i++ {
		cardID := int32(ids[i])
		if cardID < 0 {
			hwlog.RunLog.Errorf("get invalid card ID: %d", cardID)
			continue
		}
		cardIDList = append(cardIDList, cardID)
	}
	return cardNum, cardIDList, nil
}

// GetDeviceNumInCard get device number in the npu card
func (d *DcManager) GetDeviceNumInCard(cardID int32) (int32, error) {
	var deviceNum C.int
	if err := C.dcmi_get_device_num_in_card(C.int(cardID), &deviceNum); err != 0 {
		return retError, fmt.Errorf("get device count on the card failed, error code: %d", int32(err))
	}
	if deviceNum <= 0 {
		return retError, fmt.Errorf("get error device quantity: %d", int32(deviceNum))
	}
	return int32(deviceNum), nil
}

// GetDeviceLogicID get device logicID
func (d *DcManager) GetDeviceLogicID(cardID, deviceID int32) (uint32, error) {
	var logicID C.int
	if err := C.dcmi_get_device_logic_id(&logicID, C.int(cardID), C.int(deviceID)); err != 0 {
		return unretError, fmt.Errorf("get logicID failed, error code: %d", int32(err))
	}

	// check whether logicID is invalid
	if logicID < 0 || uint32(logicID) > uint32(math.MaxInt8) {
		return unretError, fmt.Errorf("get invalid logicID: %d", int32(logicID))
	}
	return uint32(logicID), nil
}

// SetDestroyVirtualDevice destroy virtual device
func (d *DcManager) SetDestroyVirtualDevice(cardID, deviceID int32, vDevID uint32) error {
	if err := C.dcmi_set_destroy_vdevice(C.int(cardID), C.int(deviceID), C.uint(vDevID)); err != 0 {
		return fmt.Errorf("destroy virtual device failed, error code: %d", int32(err))
	}
	return nil
}

func convertCreateVDevOut(cCreateVDevOut C.struct_dcmi_create_vdev_out) CgoDcmiCreateVDevOut {
	cgoCreateVDevOut := CgoDcmiCreateVDevOut{
		VDevID:     uint32(cCreateVDevOut.vdev_id),
		PcieBus:    uint32(cCreateVDevOut.pcie_bus),
		PcieDevice: uint32(cCreateVDevOut.pcie_device),
		PcieFunc:   uint32(cCreateVDevOut.pcie_func),
		VfgID:      uint32(cCreateVDevOut.vfg_id),
	}
	return cgoCreateVDevOut
}

// CreateVirtualDevice create virtual device
func (d *DcManager) CreateVirtualDevice(cardID, deviceID, vDevID int32, aiCore uint32) (CgoDcmiCreateVDevOut,
	error) {
	switch aiCore {
	case aiCoreNum1, aiCoreNum2, aiCoreNum4, aiCoreNum8, aiCoreNum16:
	default:
		return CgoDcmiCreateVDevOut{}, fmt.Errorf("input invalid aiCore: %d", aiCore)
	}
	// templateName like vir01,vir02,vir04,vir08,vir16
	templateName := fmt.Sprintf("%s%02d", vDeviceCreateTemplateNamePrefix, aiCore)
	cTemplateName := C.CString(templateName)
	defer C.free(unsafe.Pointer(cTemplateName))

	var createVDevOut C.struct_dcmi_create_vdev_out
	if err := C.dcmi_create_vdevice(C.int(cardID), C.int(deviceID), C.int(vDevID), (*C.char)(cTemplateName),
		&createVDevOut); err != 0 {
		return CgoDcmiCreateVDevOut{}, fmt.Errorf("create vdevice failed, error is: %d", int32(err))
	}

	return convertCreateVDevOut(createVDevOut), nil
}

func convertToCharArr(charArr []rune, cgoArr [dcmiVDevResNameLen]C.char) []rune {
	for _, v := range cgoArr {
		if v != 0 {
			charArr = append(charArr, rune(v))
		}
	}
	return charArr
}

func convertBaseResource(cBaseResource C.struct_dcmi_base_resource) CgoDcmiBaseResource {
	baseResource := CgoDcmiBaseResource{
		token:       uint64(cBaseResource.token),
		tokenMax:    uint64(cBaseResource.token_max),
		taskTimeout: uint64(cBaseResource.task_timeout),
		vfgID:       uint32(cBaseResource.vfg_id),
		vipMode:     uint8(cBaseResource.vip_mode),
	}
	return baseResource
}

func convertComputingResource(cComputingResource C.struct_dcmi_computing_resource) CgoDcmiComputingResource {
	computingResource := CgoDcmiComputingResource{
		aic:                float32(cComputingResource.aic),
		aiv:                float32(cComputingResource.aiv),
		dsa:                uint16(cComputingResource.dsa),
		rtsq:               uint16(cComputingResource.rtsq),
		acsq:               uint16(cComputingResource.acsq),
		cdqm:               uint16(cComputingResource.cdqm),
		cCore:              uint16(cComputingResource.c_core),
		ffts:               uint16(cComputingResource.ffts),
		sdma:               uint16(cComputingResource.sdma),
		pcieDma:            uint16(cComputingResource.pcie_dma),
		memorySize:         uint64(cComputingResource.memory_size),
		eventID:            uint32(cComputingResource.event_id),
		notifyID:           uint32(cComputingResource.notify_id),
		streamID:           uint32(cComputingResource.stream_id),
		modelID:            uint32(cComputingResource.model_id),
		topicScheduleAicpu: uint16(cComputingResource.topic_schedule_aicpu),
		hostCtrlCPU:        uint16(cComputingResource.host_ctrl_cpu),
		hostAicpu:          uint16(cComputingResource.host_aicpu),
		deviceAicpu:        uint16(cComputingResource.device_aicpu),
		topicCtrlCPUSlot:   uint16(cComputingResource.topic_ctrl_cpu_slot),
	}
	return computingResource
}

func convertMediaResource(cMediaResource C.struct_dcmi_media_resource) CgoDcmiMediaResource {
	mediaResource := CgoDcmiMediaResource{
		jpegd: float32(cMediaResource.jpegd),
		jpege: float32(cMediaResource.jpege),
		vpc:   float32(cMediaResource.vpc),
		vdec:  float32(cMediaResource.vdec),
		pngd:  float32(cMediaResource.pngd),
		venc:  float32(cMediaResource.venc),
	}
	return mediaResource
}

func convertVDevQuertyInfo(cVDevQueryInfo C.struct_dcmi_vdev_query_info) CgoVDevQueryInfo {
	var name []rune
	name = convertToCharArr(name, cVDevQueryInfo.name)

	vDevQueryInfo := CgoVDevQueryInfo{
		name:            string(name),
		status:          uint32(cVDevQueryInfo.status),
		isContainerUsed: uint32(cVDevQueryInfo.is_container_used),
		vfid:            uint32(cVDevQueryInfo.vfid),
		vfgID:           uint32(cVDevQueryInfo.vfg_id),
		containerID:     uint64(cVDevQueryInfo.container_id),
		base:            convertBaseResource(cVDevQueryInfo.base),
		computing:       convertComputingResource(cVDevQueryInfo.computing),
		media:           convertMediaResource(cVDevQueryInfo.media),
	}
	return vDevQueryInfo
}

func convertVDevQueryStru(cVDevQueryStru C.struct_dcmi_vdev_query_stru) CgoVDevQueryStru {
	vDevQueryStru := CgoVDevQueryStru{
		vDevID:    uint32(cVDevQueryStru.vdev_id),
		queryInfo: convertVDevQuertyInfo(cVDevQueryStru.query_info),
	}
	return vDevQueryStru
}

// GetDeviceVDevResource get virtual device resource info
func (d *DcManager) GetDeviceVDevResource(cardID, deviceID int32, vDevID uint32) (CgoVDevQueryStru, error) {
	var cMainCmd = C.enum_dcmi_main_cmd(MainCmdVDevMng)
	subCmd := VmngSubCmdGetVDevResource
	var vDevResource C.struct_dcmi_vdev_query_stru
	size := C.uint(unsafe.Sizeof(vDevResource))
	vDevResource.vdev_id = C.uint(vDevID)
	if err := C.dcmi_get_device_info(C.int(cardID), C.int(deviceID), cMainCmd, C.uint(subCmd),
		unsafe.Pointer(&vDevResource), &size); err != 0 {
		return CgoVDevQueryStru{}, fmt.Errorf("get device info failed, error is: %d", int32(err))
	}
	return convertVDevQueryStru(vDevResource), nil
}

func convertSocTotalResource(cSocTotalResource C.struct_dcmi_soc_total_resource) CgoDcmiSocTotalResource {
	socTotalResource := CgoDcmiSocTotalResource{
		vDevNum:   uint32(cSocTotalResource.vdev_num),
		vfgNum:    uint32(cSocTotalResource.vfg_num),
		vfgBitmap: uint32(cSocTotalResource.vfg_bitmap),
		base:      convertBaseResource(cSocTotalResource.base),
		computing: convertComputingResource(cSocTotalResource.computing),
		media:     convertMediaResource(cSocTotalResource.media),
	}
	for i := uint32(0); i < uint32(cSocTotalResource.vdev_num) && i < dcmiMaxVdevNum; i++ {
		socTotalResource.vDevID = append(socTotalResource.vDevID, uint32(cSocTotalResource.vdev_id[i]))
	}
	return socTotalResource
}

// GetDeviceTotalResource get device total resource info
func (d *DcManager) GetDeviceTotalResource(cardID, deviceID int32) (CgoDcmiSocTotalResource, error) {
	var cMainCmd = C.enum_dcmi_main_cmd(MainCmdVDevMng)
	subCmd := VmngSubCmdGetTotalResource
	var totalResource C.struct_dcmi_soc_total_resource
	size := C.uint(unsafe.Sizeof(totalResource))
	if err := C.dcmi_get_device_info(C.int(cardID), C.int(deviceID), cMainCmd, C.uint(subCmd),
		unsafe.Pointer(&totalResource), &size); err != 0 {
		return CgoDcmiSocTotalResource{}, fmt.Errorf("get device info failed, error is: %d", int32(err))
	}
	if uint32(totalResource.vdev_num) > dcmiMaxVdevNum {
		return CgoDcmiSocTotalResource{}, fmt.Errorf("get error virtual quantity: %d", uint32(totalResource.vdev_num))
	}

	return convertSocTotalResource(totalResource), nil
}

func convertSocFreeResource(cSocFreeResource C.struct_dcmi_soc_free_resource) CgoDcmiSocFreeResource {
	socFreeResource := CgoDcmiSocFreeResource{
		vfgNum:    uint32(cSocFreeResource.vfg_num),
		vfgBitmap: uint32(cSocFreeResource.vfg_bitmap),
		base:      convertBaseResource(cSocFreeResource.base),
		computing: convertComputingResource(cSocFreeResource.computing),
		media:     convertMediaResource(cSocFreeResource.media),
	}
	return socFreeResource
}

// GetDeviceFreeResource get device free resource info
func (d *DcManager) GetDeviceFreeResource(cardID, deviceID int32) (CgoDcmiSocFreeResource, error) {
	var cMainCmd = C.enum_dcmi_main_cmd(MainCmdVDevMng)
	subCmd := VmngSubCmdGetFreeResource
	var freeResource C.struct_dcmi_soc_free_resource
	size := C.uint(unsafe.Sizeof(freeResource))
	if err := C.dcmi_get_device_info(C.int(cardID), C.int(deviceID), cMainCmd, C.uint(subCmd),
		unsafe.Pointer(&freeResource), &size); err != 0 {
		return CgoDcmiSocFreeResource{}, fmt.Errorf("get device info failed, error is: %d", int32(err))
	}
	return convertSocFreeResource(freeResource), nil
}

// GetDeviceInfo get device resource info
func (d *DcManager) GetDeviceInfo(cardID, deviceID int32) (CgoVDevInfo, error) {
	var unitType C.enum_dcmi_unit_type
	if err := C.dcmi_get_device_type(C.int(cardID), C.int(deviceID), &unitType); err != 0 {
		return CgoVDevInfo{}, fmt.Errorf("get device type failed, error is: %d", int32(err))
	}
	if int32(unitType) != npuType {
		return CgoVDevInfo{}, fmt.Errorf("not support unit type: %d", int32(unitType))
	}

	cgoDcmiSocTotalResource, err := d.GetDeviceTotalResource(cardID, deviceID)
	if err != nil {
		return CgoVDevInfo{}, fmt.Errorf("get device total resource failed, error is: %v", err)
	}

	cgoDcmiSocFreeResource, err := d.GetDeviceFreeResource(cardID, deviceID)
	if err != nil {
		return CgoVDevInfo{}, fmt.Errorf("get device free resource failed, error is: %v", err)
	}

	dcmiVDevInfo := CgoVDevInfo{
		VDevNum:       cgoDcmiSocTotalResource.vDevNum,
		CoreNumUnused: cgoDcmiSocFreeResource.computing.aic,
	}
	for i := 0; i < len(cgoDcmiSocTotalResource.vDevID); i++ {
		dcmiVDevInfo.VDevID = append(dcmiVDevInfo.VDevID, cgoDcmiSocTotalResource.vDevID[i])
	}
	for _, vDevID := range cgoDcmiSocTotalResource.vDevID {
		cgoVDevQueryStru, err := d.GetDeviceVDevResource(cardID, deviceID, vDevID)
		if err != nil {
			return CgoVDevInfo{}, fmt.Errorf("get device virtual resource failed, error is: %v", err)
		}
		dcmiVDevInfo.Status = append(dcmiVDevInfo.Status, cgoVDevQueryStru.queryInfo.status)
		dcmiVDevInfo.VfID = append(dcmiVDevInfo.VfID, cgoVDevQueryStru.queryInfo.vfid)
		dcmiVDevInfo.CID = append(dcmiVDevInfo.CID, cgoVDevQueryStru.queryInfo.containerID)
		dcmiVDevInfo.CoreNum = append(dcmiVDevInfo.CoreNum, cgoVDevQueryStru.queryInfo.computing.aic)
	}
	return dcmiVDevInfo, nil
}

// GetCardIDDeviceID get card id and device id from logic id
func (d *DcManager) GetCardIDDeviceID(logicID uint32) (int32, int32, error) {
	if logicID > uint32(math.MaxInt8) {
		return retError, retError, fmt.Errorf("input invalid logicID: %d", logicID)
	}

	_, cards, err := d.GetCardList()
	if err != nil {
		return retError, retError, fmt.Errorf("get card list failed, error is: %v", err)
	}

	for _, cardID := range cards {
		deviceNum, err := d.GetDeviceNumInCard(cardID)
		if err != nil {
			hwlog.RunLog.Errorf("get device num in card failed, error is: %v", err)
			continue
		}
		for deviceID := int32(0); deviceID < deviceNum; deviceID++ {
			logicIDGet, err := d.GetDeviceLogicID(cardID, deviceID)
			if err != nil {
				hwlog.RunLog.Errorf("get device logic id failed, error is: %v", err)
				continue
			}
			if logicID == logicIDGet {
				return cardID, deviceID, nil
			}
		}
	}
	errInfo := fmt.Errorf("the card id and device id corresponding to the logic id are not found")
	return retError, retError, errInfo
}

// CreateVDevice create virtual device by logic id
func (d *DcManager) CreateVDevice(logicID, aiCore uint32) (uint32, error) {
	cardID, deviceID, err := d.GetCardIDDeviceID(logicID)
	if err != nil {
		return unretError, fmt.Errorf("get card id and device id failed, error is: %v", err)
	}

	cgoDcmiSocFreeResource, err := d.GetDeviceFreeResource(cardID, deviceID)
	if err != nil {
		return unretError, fmt.Errorf("get virtual device info failed, error is: %v", err)
	}

	if cgoDcmiSocFreeResource.computing.aic < float32(aiCore) {
		return unretError, fmt.Errorf("the remaining core resource is insufficient, free core: %f",
			cgoDcmiSocFreeResource.computing.aic)
	}

	var vDevID int32
	createVDevOut, err := d.CreateVirtualDevice(cardID, deviceID, vDevID, aiCore)
	if err != nil {
		return unretError, fmt.Errorf("create virtual device failed, error is: %v", err)
	}
	return createVDevOut.VDevID, nil
}

// GetVDeviceInfo get virtual device info by logic id
func (d *DcManager) GetVDeviceInfo(logicID uint32) (CgoVDevInfo, error) {
	cardID, deviceID, err := d.GetCardIDDeviceID(logicID)
	if err != nil {
		return CgoVDevInfo{}, fmt.Errorf("get card id and device id failed, error is: %v", err)
	}

	dcmiVDevInfo, err := d.GetDeviceInfo(cardID, deviceID)
	if err != nil {
		return CgoVDevInfo{}, fmt.Errorf("get virtual device info failed, error is: %v", err)
	}
	return dcmiVDevInfo, nil
}

// DestroyVDevice destroy spec virtual device by logic id
func (d *DcManager) DestroyVDevice(logicID, vDevID uint32) error {
	cardID, deviceID, err := d.GetCardIDDeviceID(logicID)
	if err != nil {
		return fmt.Errorf("get card id and device id failed, error is: %v", err)
	}

	if err = d.SetDestroyVirtualDevice(cardID, deviceID, vDevID); err != nil {
		return fmt.Errorf("destroy virtual device failed, error is: %v", err)
	}
	return nil
}

// GetDeviceCount get device count
func (d *DcManager) GetDeviceCount() (int32, error) {
	var cardNum C.int
	var cardList [common.HiAIMaxDeviceNum]C.int
	if err := C.dcmi_get_card_list(&cardNum, &cardList[0], common.HiAIMaxDeviceNum); err != 0 {
		errInfo := fmt.Errorf("get device count failed, error code: %d", int32(err))
		hwlog.RunLog.Error(errInfo)
		return retError, errInfo
	}
	// Invalid number of devices.
	if cardNum < 0 || cardNum > common.MaxChipNum {
		errInfo := fmt.Errorf("get device count failed, the number of devices is: %d", int32(cardNum))
		hwlog.RunLog.Error(errInfo)
		return retError, errInfo
	}
	return int32(cardNum), nil
}

// GetDeviceHealth get device health
func (d *DcManager) GetDeviceHealth(cardID, deviceID int32) (int32, error) {
	var health C.uint
	if err := C.dcmi_get_device_health(C.int(cardID), C.int(deviceID), &health); err != 0 {
		errInfo := fmt.Errorf("get device (cardID: %d, deviceID: %d) health state failed, error code: %d",
			cardID, deviceID, int32(err))
		hwlog.RunLog.Error(errInfo)
		return retError, errInfo
	}
	if common.IsGreaterThanOrEqualInt32(int64(health)) {
		errInfo := fmt.Errorf("get wrong health state , device (cardID: %d, deviceID: %d) health: %d",
			cardID, deviceID, int64(health))
		hwlog.RunLog.Error(errInfo)
		return retError, errInfo
	}
	return int32(health), nil
}

// GetDeviceUtilizationRate get device utils rate by id
func (d *DcManager) GetDeviceUtilizationRate(cardID int32, deviceID int32, devType deviceType) (int32, error) {
	var rate C.uint
	if err := C.dcmi_get_device_utilization_rate(C.int(cardID), C.int(deviceID), C.int(devType), &rate); err != 0 {
		hwlog.RunLog.Errorf("get device (cardID: %d, deviceID: %d) utilize rate failed, error code: %d, "+
			"try again ... ", cardID, deviceID, int32(err))
		for i := 0; i < common.RetryTime; i++ {
			err = C.dcmi_get_device_utilization_rate(C.int(cardID), C.int(deviceID), C.int(devType), &rate)
			if err == 0 && common.IsValidUtilizationRate(uint32(rate)) {
				return int32(rate), nil
			}
		}
		return retError, fmt.Errorf("get device (cardID: %d, deviceID: %d) utilization rate failed, error "+
			"code: %d", cardID, deviceID, int32(err))
	}
	if !common.IsValidUtilizationRate(uint32(rate)) {
		return retError, fmt.Errorf("get wrong device utilize rate, device (cardID: %d, deviceID: %d) "+
			"utilize rate: %d", cardID, deviceID, uint32(rate))
	}
	return int32(rate), nil
}

// GetDeviceTemperature get the device temperature
func (d *DcManager) GetDeviceTemperature(cardID int32, deviceID int32) (int32, error) {
	var temp C.int
	if err := C.dcmi_get_device_temperature(C.int(cardID), C.int(deviceID), &temp); err != 0 {
		errInfo := fmt.Errorf("get device (cardID: %d, deviceID: %d) temperature failed ,error code is : %d",
			cardID, deviceID, int32(err))
		return retError, errInfo
	}
	parsedTemp := int32(temp)
	if parsedTemp < int32(common.DefaultTemperatureWhenQueryFailed) {
		errInfo := fmt.Errorf("get wrong device temperature, devcie (cardID: %d, deviceID: %d), temperature: %d",
			cardID, deviceID, parsedTemp)
		return retError, errInfo
	}
	return parsedTemp, nil
}
