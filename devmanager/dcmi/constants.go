//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package dcmi this for constants
package dcmi

// MainCmd main command enum
type MainCmd uint32

// VDevMngSubCmd virtual device manager sub command
type VDevMngSubCmd uint32

const (
	// dcmiMaxVdevNum is max number of vdevice, value is from driver specification
	dcmiMaxVdevNum = 32
	// dcmiVDevResNameLen length of vnpu resource name
	dcmiVDevResNameLen = 16

	maxChipNameLen = 32
	productTypeLen = 64

	// vDeviceCreateTemplateNamePrefix prefix of vnpu template name
	vDeviceCreateTemplateNamePrefix = "vir"

	// MainCmdVDevMng virtual device manager
	MainCmdVDevMng MainCmd = 52

	// VmngSubCmdGetVDevResource get virtual device resource info
	VmngSubCmdGetVDevResource VDevMngSubCmd = 0
	// VmngSubCmdGetTotalResource get total resource info
	VmngSubCmdGetTotalResource VDevMngSubCmd = 1
	// VmngSubCmdGetFreeResource get free resource info
	VmngSubCmdGetFreeResource VDevMngSubCmd = 2
)
