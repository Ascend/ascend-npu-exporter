//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package devmanager this for constants
package dcmi

// deviceType device type enum
type deviceType int32

const (
	// Memory  Ascend310 & Ascend910
	Memory deviceType = 1
	// AICore Ascend310 & Ascend910
	AICore deviceType = 2
	// AICPU  Ascend310 & Ascend910
	AICPU deviceType = 3
	// CTRLCPU  Ascend310 & Ascend910
	CTRLCPU deviceType = 4
	// MEMBandWidth memory bandwidth Ascend310 & Ascend910
	MEMBandWidth deviceType = 5
	// HBM Ascend910 only
	HBM deviceType = 6
	// AICoreCurrentFreq AI core current frequency Ascend910 only
	AICoreCurrentFreq deviceType = 7
	// DDR now is not supported
	DDR deviceType = 8
	// AICoreNormalFreq AI core normal frequency Ascend910 only
	AICoreNormalFreq deviceType = 9
	// HBMBandWidth Ascend910 only
	HBMBandWidth deviceType = 10
	// VectorCore now is not supported
	VectorCore deviceType = 12
)

// MainCmd main command enum
type MainCmd uint32

// VDevMngSubCmd virtual device manager sub command
type VDevMngSubCmd uint32

const (
	// retError return error when the function failed
	retError = -1

	// unRetError return unsigned int error
	unretError = 100

	// dcmiMaxVdevNum is max number of vdevice, value is from driver specification
	dcmiMaxVdevNum = 32

	hiAIMaxCardNum = 16

	dcmiVDevResNameLen = 16

	vDeviceCreateTemplateNamePrefix = "vir"

	// MainCmdVDevMng virtual device manager
	MainCmdVDevMng MainCmd = 52

	// VmngSubCmdGetVDevResource get virtual device resource info
	VmngSubCmdGetVDevResource VDevMngSubCmd = 0
	// VmngSubCmdGetTotalResource get total resource info
	VmngSubCmdGetTotalResource VDevMngSubCmd = 1
	// VmngSubCmdGetFreeResource get free resource info
	VmngSubCmdGetFreeResource VDevMngSubCmd = 2

	npuType = 0

	aiCoreNum1  = 1
	aiCoreNum2  = 2
	aiCoreNum4  = 4
	aiCoreNum8  = 8
	aiCoreNum16 = 16
)
