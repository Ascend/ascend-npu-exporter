//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package devmanager this for constants
package devmanager

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
