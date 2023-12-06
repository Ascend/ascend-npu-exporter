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

// Package hccn this for npu hccn info
package hccn

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"

	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	"huawei.com/npu-exporter/v5/common-utils/utils"
	"huawei.com/npu-exporter/v5/devmanager/common"
)

const (
	space   = " "
	newLine = "\n"
	colon   = ":"

	// LinkUp npu interface up
	LinkUp string = "UP"
	// LinkDown npu interface down
	LinkDown string = "DOWN"

	opticalPartLen = 2
	secondIndex    = 2
	linkStatusPart = 3
	base64         = 64

	cardHealthy = 0

	normalCode   = 1
	abnormalCode = 0
)

func hccnToolGetInfo(args ...string) (string, error) {
	const hccn_tool = "/usr/local/Ascend/driver/tools/hccn_tool"
	if _, err := utils.CheckPath(hccn_tool); err != nil {
		return "", err
	}
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(hccn_tool, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return string(stdout.Bytes()), nil
}

// GetNPULinkStatus exec "hccn_tool -i * -link -g" to get link status
func GetNPULinkStatus(phyID int32) string {
	args := []string{"-i", strconv.Itoa(int(phyID)), "-link", "-g"}
	// command example: hccn_tool -i 0 -link -g
	// success result example is: link status: DOWN
	outStr, err := hccnToolGetInfo(args...)
	hwlog.RunLog.Debugf("hccn_tool command exec result: %v", outStr)
	if err != nil {
		hwlog.RunLog.Errorf("get npu link status failed, %s", err)
		return LinkDown
	}
	replacedStr := strings.ReplaceAll(outStr, newLine, "")
	outArr := strings.Split(replacedStr, space)
	if len(outArr) != linkStatusPart {
		return LinkDown
	}

	status := outArr[secondIndex]
	hwlog.RunLog.Debugf("hccn_tool get npu link status: %s", status)
	return status
}

// GetNPULinkSpeed exec "hccn_tool -i * -speed -g" to get link speed
func GetNPULinkSpeed(phyID int32) int {
	args := []string{"-i", strconv.Itoa(int(phyID)), "-speed", "-g"}
	// command example: hccn_tool -i 0 -speed -g
	// success result example is: Speed: 100000 Mb/s
	outStr, err := hccnToolGetInfo(args...)
	if err != nil {
		hwlog.RunLog.Errorf("get npu link speed failed, %s", err)
		return abnormalCode
	}
	replacedStr := strings.ReplaceAll(outStr, newLine, "")
	outArr := strings.Split(replacedStr, space)
	if len(outArr) != linkStatusPart {
		return abnormalCode
	}
	const midIndex = 1
	speed, err := strconv.Atoi(outArr[midIndex])
	if err != nil {
		hwlog.RunLog.Errorf("covert speed from string failed: %s", err)
		return abnormalCode
	}

	return speed
}

// GetNPULinkUpNum exec "hccn_tool -i * -link_stat -g" to get link up count
func GetNPULinkUpNum(phyID int32) int {
	args := []string{"-i", strconv.Itoa(int(phyID)), "-link_stat", "-g"}
	// command example: hccn_tool -i 0 -link_stat -g
	// success result include: [device x]link up count : y
	outStr, err := hccnToolGetInfo(args...)
	if err != nil {
		hwlog.RunLog.Errorf("get npu link stat failed, %s", err)
		return 0
	}

	const (
		linkUpArrLen = 6
		linkUpStr    = "link up count"
	)
	linkUPCount := 0
	lines := strings.Split(outStr, newLine)
	for _, line := range lines {
		if line == "" || !strings.Contains(line, linkUpStr) {
			continue
		}

		linkUpArr := strings.Fields(line)
		if len(linkUpArr) != linkUpArrLen {
			return abnormalCode
		}
		if linkUPCount, err = strconv.Atoi(linkUpArr[linkUpArrLen-1]); err != nil {
			hwlog.RunLog.Errorf("covert link up num from string failed: %s", err)
			return abnormalCode
		}
		return linkUPCount
	}

	return abnormalCode
}

// GetNPUStatInfo exec "hccn_tool -i * -stat -g" to get stat info
func GetNPUStatInfo(phyID int32) (map[string]int, error) {
	args := []string{"-i", strconv.Itoa(int(phyID)), "-stat", "-g"}
	// command example: hccn_tool -i 0 -stat -g
	// success result include: [device x]link up count : y
	outStr, err := hccnToolGetInfo(args...)
	if err != nil {
		hwlog.RunLog.Errorf("get npu stat inf failed, %s", err)
		return nil, err
	}
	lines := strings.Split(outStr, newLine)
	statInfoMap := make(map[string]int)
	const statPartLen = 2
	for _, line := range lines {
		statParts := strings.Split(line, colon)
		if len(statParts) != statPartLen || statParts[1] == "" {
			continue
		}
		statNum, err := strconv.Atoi(statParts[1])
		if err != nil {
			hwlog.RunLog.Errorf("covert stat num of [%s] from string failed: %s", statParts[1], err)
			continue
		}
		statInfoMap[statParts[0]] = statNum
	}

	return statInfoMap, nil
}

// GetNPUOpticalInfo exec "hccn_tool -i * -optical -g" to get optical info
func GetNPUOpticalInfo(phyID int32) (map[string]string, error) {
	args := []string{"-i", strconv.Itoa(int(phyID)), "-optical", "-g"}
	// command example: hccn_tool -i 0 -optical -g
	// success result include: [device x]link up count : y
	outStr, err := hccnToolGetInfo(args...)
	if err != nil {
		hwlog.RunLog.Errorf("get npu stat inf failed, %s", err)
		return nil, err
	}
	lines := strings.Split(outStr, newLine)
	opticalInfoMap := make(map[string]string)
	for _, line := range lines {
		opticalParts := strings.Split(line, colon)
		if len(opticalParts) != opticalPartLen {
			continue
		}
		opticalKey := strings.ReplaceAll(strings.TrimSpace(opticalParts[0]), space, "_")
		opticalValue := strings.TrimSpace(opticalParts[1])
		opticalInfoMap[opticalKey] = opticalValue
	}

	return opticalInfoMap, nil
}

// GetNPUInterfaceTraffic exec "hccn_tool -i * -bandwidth -g" to get bandwidth info
func GetNPUInterfaceTraffic(phyID int32) (float64, float64, error) {
	const (
		noTraffic      = 0.00
		trafficPartLen = 4
		txStr          = "TX:"
		rxStr          = "RX:"
	)

	args := []string{"-i", strconv.Itoa(int(phyID)), "-bandwidth", "-g"}
	// command example: hccn_tool -i 0 -bandwidth -g
	// success result has two lines:
	// Bandwidth TX: 0.00 MB/sec
	// Bandwidth RX: 0.00 MB/sec
	outStr, err := hccnToolGetInfo(args...)
	hwlog.RunLog.Debugf("hccn_tool command exec result: %v", outStr)
	if err != nil {
		hwlog.RunLog.Errorf("get npu interface traffic failed, %s", err)
		return noTraffic, noTraffic, err
	}

	var (
		tx = 0.00
		rx = 0.00
	)

	lines := strings.Split(outStr, newLine)
	for _, line := range lines {
		if line == "" {
			continue
		}

		trafficArr := strings.Split(line, space)
		hwlog.RunLog.Debugf("npu bandwidth split as: %v", trafficArr)
		if len(trafficArr) != trafficPartLen {
			continue
		}
		if strings.Contains(line, txStr) {
			if tmpTx, err := strconv.ParseFloat(trafficArr[secondIndex], base64); err == nil {
				tx = tmpTx
			}
			continue
		}
		if strings.Contains(line, rxStr) {
			if tmpRx, err := strconv.ParseFloat(trafficArr[secondIndex], base64); err == nil {
				rx = tmpRx
			}
		}
	}
	return tx, rx, nil
}

// GetFloatDataFromStr get float data from string with space
func GetFloatDataFromStr(str string) float64 {
	dataParts := strings.Split(str, space)
	if len(dataParts) != opticalPartLen {
		return abnormalCode
	}

	floatData, err := strconv.ParseFloat(dataParts[0], base64)
	if err != nil {
		hwlog.RunLog.Errorf("get float data from string err: %s", err)
		return abnormalCode
	}
	return floatData
}

// GetHealthCode return union healthy code
func GetHealthCode(healthCode uint32) int {
	if healthCode == cardHealthy {
		return normalCode
	}
	return abnormalCode
}

// GetLinkStatusCode return union link status code
func GetLinkStatusCode(status string) int {
	if status == LinkUp {
		return normalCode
	}
	return abnormalCode
}

// GetNetworkHealthy return union network healthy code
func GetNetworkHealthy(netCode uint32) int {
	if netCode == common.NetworkInit || netCode == common.NetworkSuccess {
		return normalCode
	}
	return abnormalCode
}
