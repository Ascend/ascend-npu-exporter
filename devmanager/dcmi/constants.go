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

// Package dcmi this for constants
package dcmi

// MainCmd main command enum
type MainCmd uint32

// VDevMngSubCmd virtual device manager sub command
type VDevMngSubCmd uint32

// DcmiDieType present chip die type
type DcmiDieType int32

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

	// NDIE NDie ID, only Ascend910 has
	NDIE DcmiDieType = 0
	// VDIE VDie ID, it can be the uuid of chip
	VDIE DcmiDieType = 1
	// DieIDCount die id array max length
	DieIDCount = 5

	// ipAddrTypeV6 ip address type of IPv6
	ipAddrTypeV6 = 1
)
