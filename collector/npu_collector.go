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

// Package collector for Prometheus
package collector

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"huawei.com/npu-exporter/v5/collector/container"
	"huawei.com/npu-exporter/v5/common-utils/cache"
	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	"huawei.com/npu-exporter/v5/devmanager"
	"huawei.com/npu-exporter/v5/devmanager/common"
	"huawei.com/npu-exporter/v5/versions"
)

var (
	versionInfoDesc = prometheus.NewDesc("npu_exporter_version_info",
		"exporter version with value '1'", []string{"exporterVersion"}, nil)
	machineInfoNPUDesc = prometheus.NewDesc("machine_npu_nums",
		"Amount of npu installed on the machine.", nil, nil)
	npuChipInfoDescNpuName = prometheus.NewDesc("npu_chip_info_name",
		"the Ascend npu name with value '1'", []string{"id", "name"}, nil)
	npuChipInfoDescUtil = prometheus.NewDesc("npu_chip_info_utilization",
		"the ai core utilization", []string{"id"}, nil)
	npuChipInfoDescTemp = prometheus.NewDesc("npu_chip_info_temperature",
		"the npu temperature", []string{"id"}, nil)
	npuChipInfoDescPower = prometheus.NewDesc("npu_chip_info_power",
		"the npu power", []string{"id"}, nil)
	npuChipInfoDescVoltage = prometheus.NewDesc("npu_chip_info_voltage",
		"the npu voltage", []string{"id"}, nil)
	npuChipInfoDescUsedMemory = prometheus.NewDesc("npu_chip_info_used_memory",
		"the npu used memory", []string{"id"}, nil)
	npuChipInfoDescTotalMemory = prometheus.NewDesc("npu_chip_info_total_memory",
		"the npu total memory", []string{"id"}, nil)
	npuChipInfoDescHealthStatus = prometheus.NewDesc("npu_chip_info_health_status",
		"the npu health status", []string{"id"}, nil)
	npuChipInfoDescHbmUsedMemory = prometheus.NewDesc("npu_chip_info_hbm_used_memory",
		"the npu hbm used memory", []string{"id"}, nil)
	npuChipInfoDescHbmTotalMemory = prometheus.NewDesc("npu_chip_info_hbm_total_memory",
		"the npu hbm total memory", []string{"id"}, nil)
	npuChipInfoDescErrorCode = prometheus.NewDesc("npu_chip_info_error_code",
		"the npu error code", []string{"id"}, nil)
	npuContainerInfo = prometheus.NewDesc("npu_container_info",
		"the container name and deviceID relationship", []string{"containerID", "containerName", "npuID"}, nil)
	npuContainerTotalMemory = prometheus.NewDesc("container_npu_total_memory",
		"the npu total memory in container, unit is 'MB'", []string{"id",
			"namespace", "pod_name", "container_name"}, nil)
	npuContainerUsedMemory = prometheus.NewDesc("container_npu_used_memory",
		"the npu used memory in container, unit is 'MB'", []string{"id",
			"namespace", "pod_name", "container_name"}, nil)
	npuContainerUtilization = prometheus.NewDesc("container_npu_utilization",
		"the npu ai core utilization in container, unit is '%'", []string{"id",
			"namespace", "pod_name", "container_name"}, nil)
	npuContainerInfoInit sync.Once
	npuChipInfoInit      sync.Once
)

const (
	cacheSize    = 128
	nameSpaceIdx = 0
	podNameIdx   = 1
	conNameIdx   = 2
)

type npuCollector struct {
	cache         *cache.ConcurrencyLRUCache
	devicesParser *container.DevicesParser
	updateTime    time.Duration
	cacheTime     time.Duration
}

// NewNpuCollector create an instance of prometheus Collector
func NewNpuCollector(ctx context.Context, cacheTime time.Duration, updateTime time.Duration,
	deviceParser *container.DevicesParser) (prometheus.Collector, error) {
	npuCollect := &npuCollector{
		cache:         cache.New(cacheSize),
		cacheTime:     cacheTime,
		updateTime:    updateTime,
		devicesParser: deviceParser,
	}
	devManager, err := devmanager.AutoInit("")
	if err != nil {
		hwlog.RunLog.Errorf("new npu collector failed, error is %#v", err)
		return nil, err
	}
	go start(ctx, npuCollect, devManager)
	return npuCollect, nil
}

func getNPUInfo(dmgr devmanager.DeviceInterface) []HuaWeiNPUCard {
	var npuList []HuaWeiNPUCard
	cardNum, cards, err := dmgr.GetCardList()
	if cardNum == 0 || err != nil {
		hwlog.RunLog.Errorf("failed to get npu info, error is: %#v", err)
		return npuList
	}

	for _, cardID := range cards {
		deviceNum, err := dmgr.GetDeviceNumInCard(cardID)
		if err != nil {
			continue
		}
		var deviceList []*HuaWeiAIChip
		for i := int32(0); i < deviceNum; i++ {
			var chipInfo *HuaWeiAIChip
			logID, err := dmgr.GetDeviceLogicID(cardID, i)
			if err != nil {
				continue
			}
			chipInfo = assembleNPUInfo(cardID, logID, dmgr)
			if chipInfo != nil {
				deviceList = append(deviceList, chipInfo)
			}
		}
		npuCard := HuaWeiNPUCard{
			CardID:     int(cardID),
			DeviceList: deviceList,
			Timestamp:  time.Now(),
		}
		npuList = append(npuList, npuCard)
	}
	return npuList
}

func assembleNPUInfo(cardID int32, logicID int32, dmgr devmanager.DeviceInterface) *HuaWeiAIChip {
	phyID, err := dmgr.GetPhysicIDFromLogicID(logicID)
	// check cardId, convert it to int type later
	if err != nil {
		return nil
	}
	chipInfo := packChipInfo(logicID, dmgr)
	chipInfo.DeviceID = int(phyID)
	if dmgr.GetDevType() == common.Ascend310P {
		cardPower, err := dmgr.GetMcuPowerInfo(cardID)
		if err != nil {
			hwlog.RunLog.Error(err)
			cardPower = float32(common.InvaidVal)
		}
		// Ascend310P use cardPower to replace chipPower
		chipInfo.Power = cardPower
	}
	return chipInfo
}

func start(ctx context.Context, n *npuCollector, dmgr devmanager.DeviceInterface) {
	defer func() {
		if err := dmgr.ShutDown(); err != nil {
			hwlog.RunLog.Error(err)
		}
		if err := recover(); err != nil {
			hwlog.RunLog.Errorf("go routine failed with %#v", err)
		}
	}()
	if n == nil {
		hwlog.RunLog.Error("Invalid param in function start")
		return
	}
	if err := n.devicesParser.Init(); err != nil {
		hwlog.RunLog.Errorf("failed to init devices parser: %#v", err)
	}
	defer n.devicesParser.Close()
	n.devicesParser.Timeout = n.updateTime
	ticker := time.NewTicker(n.updateTime)
	hwlog.RunLog.Infof("Starting update cache every %d seconds", n.updateTime/time.Second)
	for {
		select {
		case _, ok := <-ticker.C:
			if !ok {
				return
			}
			n.devicesParser.FetchAndParse(nil)
			npuInfo := getNPUInfo(dmgr)
			if err := n.cache.Set(key, npuInfo, n.cacheTime); err != nil {
				hwlog.RunLog.Error(err)
			}
			hwlog.RunLog.Infof("update cache,key is %s", key)
		case result := <-n.devicesParser.RecvResult():
			if err := n.cache.Set(containersDevicesInfoKey, result, n.cacheTime); err != nil {
				hwlog.RunLog.Error(err)
			}
			hwlog.RunLog.Infof("update cache,key is %s", containersDevicesInfoKey)
		case err := <-n.devicesParser.RecvErr():
			hwlog.RunLog.Errorf("received error from device parser: %#v", err)
		case _, ok := <-ctx.Done():
			if !ok {
				hwlog.RunLog.Error("closed")
			}
			ticker.Stop()
			hwlog.RunLog.Warn("received the stop signal,STOPPED")
			return
		}
	}
}

// Describe implements prometheus.Collector
func (n *npuCollector) Describe(ch chan<- *prometheus.Desc) {
	if ch == nil {
		hwlog.RunLog.Error("Invalid param in function Describe")
		return
	}
	ch <- versionInfoDesc
	ch <- machineInfoNPUDesc
	ch <- npuChipInfoDescUtil
	ch <- npuChipInfoDescTemp
	ch <- npuChipInfoDescPower
	ch <- npuChipInfoDescVoltage
	ch <- npuChipInfoDescHealthStatus
	ch <- npuChipInfoDescHbmUsedMemory
	ch <- npuChipInfoDescHbmTotalMemory
	ch <- npuChipInfoDescUsedMemory
	ch <- npuChipInfoDescTotalMemory
	ch <- npuChipInfoDescErrorCode
	ch <- npuChipInfoDescNpuName
	ch <- npuContainerInfo
	ch <- npuContainerTotalMemory
	ch <- npuContainerUsedMemory
	ch <- npuContainerUtilization
}

// Collect implements prometheus.Collector
func (n *npuCollector) Collect(ch chan<- prometheus.Metric) {
	if !validate(ch) {
		hwlog.RunLog.Error("Invalid param in function Collect")
		return
	}
	obj, err := n.cache.Get(key)
	npuChipInfoInit.Do(func() {
		if err != nil {
			hwlog.RunLog.Warn("no cache, start to get npulist and rebuild cache")
			devManager, err := devmanager.AutoInit("")
			if err != nil {
				hwlog.RunLog.Warnf("get device manager failed, error is: %#v ", err)
				return
			}
			npuInfo := getNPUInfo(devManager)
			if err = n.cache.Set(key, npuInfo, n.cacheTime); err != nil {
				hwlog.RunLog.Error(err)
				return
			}
			hwlog.RunLog.Warn("rebuild cache successfully")
			obj = npuInfo
		}
	})
	npuList, ok := obj.([]HuaWeiNPUCard)
	if !ok {
		hwlog.RunLog.Error("Error cache and convert failed")
		n.cache.Delete(key)
	}
	containerMap := updateContainerNPUInfo(ch, n)
	ch <- prometheus.MustNewConstMetric(versionInfoDesc, prometheus.GaugeValue, 1, []string{versions.BuildVersion}...)
	var totalCount = 0
	for _, card := range npuList {
		deviceCount := len(card.DeviceList)
		if deviceCount <= 0 {
			continue
		}
		totalCount += deviceCount
		for _, chip := range card.DeviceList {
			containerName, ok := containerMap[chip.DeviceID]
			if !ok {
				containerName = nil
			}
			updateNPUCommonInfo(ch, &card, chip, containerName)
			updateNPUMemoryInfo(ch, &card, chip, containerName)
			updateNPUOtherInfo(ch, &card, chip)
		}
	}

	ch <- prometheus.MustNewConstMetric(machineInfoNPUDesc, prometheus.GaugeValue, float64(totalCount))
}

func updateContainerNPUInfo(ch chan<- prometheus.Metric, n *npuCollector) map[int][]string {
	if ch == nil {
		hwlog.RunLog.Error("metric channel is nil")
		return nil
	}
	obj, err := n.cache.Get(containersDevicesInfoKey)
	// only run once to prevent wait when container info get failed
	npuContainerInfoInit.Do(func() {
		if err != nil {
			hwlog.RunLog.Warn("containers' devices info not found in cache, rebuilding")
			resultChan := make(chan container.DevicesInfos, 1)
			n.devicesParser.FetchAndParse(resultChan)
			select {
			case obj = <-resultChan:
			case <-time.After(time.Second):
				hwlog.RunLog.Warn("rebuild cache timeout")
				return
			}
			hwlog.RunLog.Warn("rebuild cache successfully")
		}
	})
	cntNpuInfos, ok := obj.(container.DevicesInfos)
	if !ok {
		hwlog.RunLog.Error("Error cache and convert failed")
		return nil
	}
	res := make(map[int][]string, initSize)
	for _, v := range cntNpuInfos {
		for _, deviceID := range v.Devices {
			res[deviceID] = strings.Split(v.Name, "_")
			ch <- prometheus.MustNewConstMetric(
				npuContainerInfo,
				prometheus.GaugeValue, 1,
				[]string{v.ID, v.Name, strconv.Itoa(deviceID)}...)
		}
	}
	return res
}

func updateNPUOtherInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip) {
	if !validate(ch, npu, chip, chip.ChipIfo) {
		hwlog.RunLog.Error("Invalid param in function updateNPUOtherInfo")
		return
	}
	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescHealthStatus, prometheus.GaugeValue,
			float64(getHealthCode(chip.HealthStatus)), []string{strconv.FormatInt(int64(chip.DeviceID), base)}...))
	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescErrorCode, prometheus.GaugeValue, float64(chip.ErrorCode),
			[]string{strconv.FormatInt(int64(chip.DeviceID), base)}...))
	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescNpuName, prometheus.GaugeValue, 1,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base),
				fmt.Sprintf("%s-%s-%s", chip.ChipIfo.Name, chip.ChipIfo.Type, chip.ChipIfo.Version)}...))
}

func validate(ch chan<- prometheus.Metric, objs ...interface{}) bool {
	if ch == nil {
		return false
	}
	for _, v := range objs {
		val := reflect.ValueOf(v)
		if val.Kind() != reflect.Ptr {
			return false
		}
		if val.IsNil() {
			return false
		}
	}
	return true
}

func updateNPUMemoryInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip,
	containerName []string) {
	if !validate(ch, npu, chip, chip.HbmInfo, chip.Meminf) {
		hwlog.RunLog.Error("Invalid param in function updateNPUMemoryInfo")
		return
	}
	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescHbmUsedMemory, prometheus.GaugeValue,
			float64(chip.HbmInfo.Usage),
			[]string{strconv.FormatInt(int64(chip.DeviceID), base)}...))
	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescHbmTotalMemory, prometheus.GaugeValue,
			float64(chip.HbmInfo.MemorySize),
			[]string{strconv.FormatInt(int64(chip.DeviceID), base)}...))

	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescUsedMemory, prometheus.GaugeValue,
			float64(chip.Meminf.MemorySize-chip.Meminf.MemoryAvailable),
			[]string{strconv.FormatInt(int64(chip.DeviceID), base)}...))
	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescTotalMemory, prometheus.GaugeValue, float64(chip.Meminf.MemorySize),
			[]string{strconv.FormatInt(int64(chip.DeviceID), base)}...))
	updateContainerInfo(ch, npu, chip, containerName)
}

func updateContainerInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip, containerName []string) {
	if len(containerName) != containerNameLen {
		return
	}
	if strings.Contains(chip.ChipIfo.Name, "910") {
		ch <- prometheus.NewMetricWithTimestamp(
			npu.Timestamp,
			prometheus.MustNewConstMetric(npuContainerTotalMemory, prometheus.GaugeValue,
				float64(chip.HbmInfo.MemorySize), []string{strconv.FormatInt(int64(chip.DeviceID), base),
					containerName[nameSpaceIdx], containerName[podNameIdx], containerName[conNameIdx]}...))
		ch <- prometheus.NewMetricWithTimestamp(
			npu.Timestamp,
			prometheus.MustNewConstMetric(npuContainerUsedMemory, prometheus.GaugeValue,
				float64(chip.HbmInfo.Usage),
				[]string{strconv.FormatInt(int64(chip.DeviceID), base),
					containerName[nameSpaceIdx], containerName[podNameIdx], containerName[conNameIdx]}...))
		return
	}
	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuContainerTotalMemory, prometheus.GaugeValue,
			float64(chip.Meminf.MemorySize), []string{strconv.FormatInt(int64(chip.DeviceID), base),
				containerName[nameSpaceIdx], containerName[podNameIdx], containerName[conNameIdx]}...))
	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuContainerUsedMemory, prometheus.GaugeValue,
			float64(chip.Meminf.MemorySize-chip.Meminf.MemoryAvailable),
			[]string{strconv.FormatInt(int64(chip.DeviceID), base),
				containerName[nameSpaceIdx], containerName[podNameIdx], containerName[conNameIdx]}...))
}

func updateNPUCommonInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip,
	containerName []string) {
	if !validate(ch, npu, chip) {
		hwlog.RunLog.Error("Invalid param in function updateNpuCommonInfo")
		return
	}
	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescUtil, prometheus.GaugeValue, float64(chip.Utilization),
			[]string{strconv.FormatInt(int64(chip.DeviceID), base)}...))
	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescTemp, prometheus.GaugeValue, float64(chip.Temperature),
			[]string{strconv.FormatInt(int64(chip.DeviceID), base)}...))

	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescPower, prometheus.GaugeValue, float64(chip.Power),
			[]string{strconv.FormatInt(int64(chip.DeviceID), base)}...))

	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescVoltage, prometheus.GaugeValue, float64(chip.Voltage),
			[]string{strconv.FormatInt(int64(chip.DeviceID), base)}...))
	if len(containerName) == containerNameLen {
		ch <- prometheus.NewMetricWithTimestamp(
			npu.Timestamp,
			prometheus.MustNewConstMetric(npuContainerUtilization, prometheus.GaugeValue,
				float64(chip.Utilization), []string{strconv.FormatInt(int64(chip.DeviceID),
					base), containerName[nameSpaceIdx], containerName[podNameIdx], containerName[conNameIdx]}...))
	}
}

var packChipInfo = func(logicID int32, dmgr devmanager.DeviceInterface) *HuaWeiAIChip {
	freq, err := dmgr.GetDeviceFrequency(logicID, common.AICore)
	if err != nil {
		freq = common.InvaidVal
	}
	power, err := dmgr.GetDevicePowerInfo(logicID)
	if err != nil {
		power = common.InvaidVal
	}
	temp, err := dmgr.GetDeviceTemperature(logicID)
	if err != nil {
		temp = common.InvaidVal
	}
	vol, err := dmgr.GetDeviceVoltage(logicID)
	if err != nil {
		vol = common.InvaidVal
	}
	mem, err := dmgr.GetDeviceMemoryInfo(logicID)
	if err != nil {
		mem = &common.MemoryInfo{}
	}
	chip, err := dmgr.GetChipInfo(logicID)
	if err != nil {
		chip = &common.ChipInfo{}
	}
	hbmInfo, err := dmgr.GetDeviceHbmInfo(logicID)
	if err != nil {
		hbmInfo = &common.HbmInfo{}
	}
	util, err := dmgr.GetDeviceUtilizationRate(logicID, common.AICore)
	if err != nil {
		util = common.InvaidVal // valid data range 0-100
	}

	_, errCode, err := dmgr.GetDeviceErrorCode(logicID)
	if err != nil {
		errCode = common.RetError // valid data range 0-128
	}
	return &HuaWeiAIChip{
		ErrorCode:    errCode,
		Utilization:  int(util),
		Frequency:    int(freq),
		Power:        power,
		HealthStatus: getHealth(logicID, dmgr),
		Temperature:  int(temp),
		Voltage:      vol,
		Meminf:       mem,
		ChipIfo:      chip,
		HbmInfo:      hbmInfo,
	}
}

func getHealth(logicID int32, dmgr devmanager.DeviceInterface) HealthEnum {
	health, err := dmgr.GetDeviceHealth(logicID)
	if err != nil || health != 0 {
		return UnHealthy
	}
	return Healthy
}

func getHealthCode(health HealthEnum) int {
	if Healthy == health {
		return 1
	}
	return 0
}
