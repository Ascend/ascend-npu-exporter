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

// Package common this for util method
package common

import (
	"math"
	"regexp"
	"strings"
)

// IsGreaterThanOrEqualInt32 check num range
func IsGreaterThanOrEqualInt32(num int64) bool {
	if num >= int64(math.MaxInt32) {
		return true
	}

	return false
}

// IsValidUtilizationRate valid utilization rate is 0-100
func IsValidUtilizationRate(num uint32) bool {
	if num > uint32(Percent) || num < 0 {
		return false
	}

	return true
}

// IsValidChipInfo valid chip info is or not empty
func IsValidChipInfo(chip *ChipInfo) bool {
	return chip.Name != "" || chip.Type != "" || chip.Version != ""
}

// IsValidCardID valid card id
func IsValidCardID(cardID int32) bool {
	// for cardID, please watch the maximum value of the driver is changed in the future version
	return cardID >= 0 && cardID < HiAIMaxCardID
}

// IsValidDeviceID valid device id
func IsValidDeviceID(deviceID int32) bool {
	return deviceID >= 0 && deviceID < HiAIMaxDeviceNum
}

// IsValidLogicIDOrPhyID valid logic id
func IsValidLogicIDOrPhyID(id int32) bool {
	return id >= 0 && id < HiAIMaxCardNum*HiAIMaxDeviceNum
}

// IsValidCardIDAndDeviceID check two params both needs meet the requirement
func IsValidCardIDAndDeviceID(cardID, deviceID int32) bool {
	if !IsValidCardID(cardID) {
		return false
	}

	return IsValidDeviceID(deviceID)
}

// IsValidDevNumInCard valid devNum in card
func IsValidDevNumInCard(num int32) bool {
	return num > 0 && num <= HiAIMaxDeviceNum
}

// GetDeviceTypeByChipName get device type by chipName
func GetDeviceTypeByChipName(chipName string) string {
	if strings.Contains(chipName, "310P") {
		return Ascend310P
	}
	if strings.Contains(chipName, "310") {
		return Ascend310
	}
	reg910B := regexp.MustCompile(Pattern1980B)
	if reg910B.MatchString(chipName) {
		return Ascend910B
	}
	reg910A := regexp.MustCompile(Pattern1980)
	if reg910A.MatchString(chipName) {
		return Ascend910
	}
	return ""
}

func get910TemplateNameList() map[string]struct{} {
	return map[string]struct{}{"vir16": {}, "vir08": {}, "vir04": {}, "vir02": {}, "vir01": {}}
}

func get310PTemplateNameList() map[string]struct{} {
	return map[string]struct{}{"vir04": {}, "vir02": {}, "vir01": {}, "vir04_3c": {}, "vir02_1c": {},
		"vir04_4c_dvpp": {}, "vir04_3c_ndvpp": {}}
}

// IsValidTemplateName check template name meet the requirement
func IsValidTemplateName(devType, templateName string) bool {
	isTemplateNameValid := false
	switch devType {
	case Ascend310P:
		_, isTemplateNameValid = get310PTemplateNameList()[templateName]
	case Ascend910:
		_, isTemplateNameValid = get910TemplateNameList()[templateName]
	default:
	}
	return isTemplateNameValid
}

// RemoveDuplicate remove duplicate device
func RemoveDuplicate(list *[]string) []string {
	listValueMap := make(map[string]string, len(*list))
	var rmDupValueList []string
	for _, value := range *list {
		listValueMap[value] = value
	}
	for _, value := range listValueMap {
		rmDupValueList = append(rmDupValueList, value)
	}
	return rmDupValueList
}
