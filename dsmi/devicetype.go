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

// Package dsmi interface
package dsmi

// DeviceType devive type enum
type DeviceType int32

const (
	// Memory  Ascend310 & Ascend910
	Memory DeviceType = 1
	// AICore Ascend310 & Ascend910
	AI_Core DeviceType = 2
	// AICPU  Ascend310 & Ascend910
	AI_CPU DeviceType = 3
	// CTRLCPU  Ascend310 & Ascend910
	CTRL_CPU DeviceType = 4
	// MEMBandWidth memory brandwidth  Ascend310 & Ascend910
	MEM_BandWidth DeviceType = 5
	// HBM             Ascend910 only
	HBM                 DeviceType = 6
	AI_Core_Current_Fre DeviceType = 7
	// AICoreNormalFre AI core current frequency  Ascend910 only
	// AICoreNormalFre AI core normal frequency  Ascend910 only
	AI_Core_Normal_Fre DeviceType = 9
	// HBMBandWidth Ascend910 only
	HBM_BandWidth DeviceType = 10 // Ascend910 only

)
