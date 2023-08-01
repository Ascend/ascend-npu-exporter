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

// Package collector for Prometheus
package collector

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"huawei.com/npu-exporter/v5/collector/container"
	"huawei.com/npu-exporter/v5/common-utils/cache"
	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	"huawei.com/npu-exporter/v5/common-utils/utils"
	"huawei.com/npu-exporter/v5/devmanager"
	"huawei.com/npu-exporter/v5/devmanager/common"
	"huawei.com/npu-exporter/v5/devmanager/dcmi"
	"huawei.com/npu-exporter/v5/versions"
)

var (
	versionInfoDesc = prometheus.NewDesc("npu_exporter_version_info",
		"exporter version with value '1'", []string{"exporterVersion"}, nil)
	machineInfoNPUDesc = prometheus.NewDesc("machine_npu_nums",
		"Amount of npu installed on the machine.", nil, nil)
	npuChipInfoDescNpuName = prometheus.NewDesc("npu_chip_info_name",
		"the Ascend npu name with value '1'", []string{"id", "name", "vdie_id"}, nil)
	npuChipInfoDescUtil = prometheus.NewDesc("npu_chip_info_utilization",
		"the ai core utilization", []string{"id", "vdie_id"}, nil)
	npuChipInfoDescTemp = prometheus.NewDesc("npu_chip_info_temperature",
		"the npu temperature", []string{"id", "vdie_id"}, nil)
	npuChipInfoDescPower = prometheus.NewDesc("npu_chip_info_power",
		"the npu power", []string{"id", "vdie_id"}, nil)
	npuChipInfoDescVoltage = prometheus.NewDesc("npu_chip_info_voltage",
		"the npu voltage", []string{"id", "vdie_id"}, nil)
	npuChipInfoDescUsedMemory = prometheus.NewDesc("npu_chip_info_used_memory",
		"the npu used memory", []string{"id", "vdie_id"}, nil)
	npuChipInfoDescTotalMemory = prometheus.NewDesc("npu_chip_info_total_memory",
		"the npu total memory", []string{"id", "vdie_id"}, nil)
	npuChipInfoDescHealthStatus = prometheus.NewDesc("npu_chip_info_health_status",
		"the npu health status", []string{"id", "vdie_id"}, nil)
	npuChipInfoDescHbmUsedMemory = prometheus.NewDesc("npu_chip_info_hbm_used_memory",
		"the npu hbm used memory", []string{"id", "vdie_id"}, nil)
	npuChipInfoDescHbmTotalMemory = prometheus.NewDesc("npu_chip_info_hbm_total_memory",
		"the npu hbm total memory", []string{"id", "vdie_id"}, nil)
	npuChipInfoDescErrorCode = prometheus.NewDesc("npu_chip_info_error_code",
		"the npu error code", []string{"id", "vdie_id"}, nil)
	npuChipInfoDescLinkStatus = prometheus.NewDesc("npu_chip_info_link_status",
		"the npu link status", []string{"id", "vdie_id"}, nil)
	npuChipInfoDescNetworkStatus = prometheus.NewDesc("npu_chip_info_network_status",
		"the npu network health status", []string{"id", "vdie_id"}, nil)
	npuChipInfoDescBandwidthTx = prometheus.NewDesc("npu_chip_info_bandwidth_tx",
		"the npu interface transport speed, unit is 'MB/s'", []string{"id", "vdie_id"}, nil)
	npuChipInfoDescBandwidthRx = prometheus.NewDesc("npu_chip_info_bandwidth_rx",
		"the npu interface receive speed, unit is 'MB/s'", []string{"id", "vdie_id"}, nil)
	npuChipInfoDescDevProcessInfo = prometheus.NewDesc("npu_chip_info_process_info",
		"the npu process info, unit is 'MB'", []string{"id", "vdie_id", "process_id"}, nil)
	npuContainerInfo = prometheus.NewDesc("npu_container_info",
		"the container name and deviceID relationship", []string{"containerID", "containerName", "npuID", "vdie_id"},
		nil)
	npuContainerTotalMemory = prometheus.NewDesc("container_npu_total_memory",
		"the npu total memory in container, unit is 'MB'", []string{"id", "namespace", "pod_name", "container_name",
			"vdie_id"}, nil)
	npuContainerUsedMemory = prometheus.NewDesc("container_npu_used_memory",
		"the npu used memory in container, unit is 'MB'", []string{"id", "namespace", "pod_name", "container_name",
			"vdie_id"}, nil)
	npuContainerUtilization = prometheus.NewDesc("container_npu_utilization",
		"the npu ai core utilization in container, unit is '%'", []string{"id", "namespace", "pod_name",
			"container_name", "vdie_id"}, nil)
	npuContainerInfoInit sync.Once
	npuChipInfoInit      sync.Once
)

const (
	cacheSize      = 128
	nameSpaceIdx   = 0
	podNameIdx     = 1
	conNameIdx     = 2
	space          = " "
	newLine        = "\n"
	linkStatusPart = 3
	trafficPart    = 4
	noTraffic      = 0.00
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
		hwlog.RunLog.Errorf("new npu collector failed, error is %v", err)
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
			if err := n.cache.Set(npuListCacheKey, npuInfo, n.cacheTime); err != nil {
				hwlog.RunLog.Error(err)
			}
			hwlog.RunLog.Infof("update cache,key is %s", npuListCacheKey)
		case result := <-n.devicesParser.RecvResult():
			if err := n.cache.Set(containersDevicesCacheKey, result, n.cacheTime); err != nil {
				hwlog.RunLog.Error(err)
			}
			hwlog.RunLog.Infof("update cache,key is %s", containersDevicesCacheKey)
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
	ch <- npuChipInfoDescNetworkStatus
	ch <- npuChipInfoDescBandwidthTx
	ch <- npuChipInfoDescBandwidthRx
	ch <- npuChipInfoDescLinkStatus
	ch <- npuChipInfoDescDevProcessInfo
}

// Collect implements prometheus.Collector
func (n *npuCollector) Collect(ch chan<- prometheus.Metric) {
	if !validate(ch) {
		hwlog.RunLog.Error("Invalid param in function Collect")
		return
	}
	obj, err := n.cache.Get(npuListCacheKey)
	npuChipInfoInit.Do(func() {
		if err != nil {
			hwlog.RunLog.Debugf("no cache, start to get npulist and rebuild cache")
			devManager, err := devmanager.GetDeviceManager()
			if err != nil {
				hwlog.RunLog.Debugf("get device manager failed, error is: %#v ", err)
				return
			}
			npuInfo := getNPUInfo(devManager)
			if err = n.cache.Set(npuListCacheKey, npuInfo, n.cacheTime); err != nil {
				hwlog.RunLog.Errorf("no cache for prometheus, try to build cache failed, error is: %v", err)
				return
			}
			hwlog.RunLog.Debugf("rebuild cache successfully")
			obj = npuInfo
		}
	})
	npuList, ok := obj.([]HuaWeiNPUCard)
	if !ok {
		hwlog.RunLog.Error("Error npu info cache and convert failed")
		n.cache.Delete(npuListCacheKey)
	}
	containerMap := getContainerNPUInfo(ch, n)
	ch <- prometheus.MustNewConstMetric(versionInfoDesc, prometheus.GaugeValue, 1, []string{versions.BuildVersion}...)
	var totalCount = 0
	for _, card := range npuList {
		deviceCount := len(card.DeviceList)
		if deviceCount <= 0 {
			continue
		}
		totalCount += deviceCount
		for _, chip := range card.DeviceList {
			devInfo, ok := containerMap[chip.DeviceID]
			if !ok {
				devInfo = nil
			}
			updateNPUCommonInfo(ch, &card, chip)
			updateNPUMemoryInfo(ch, &card, chip)
			updateNPUOtherInfo(ch, &card, chip)
			updateProcessInfo(ch, &card, chip)
			updateContainerInfo(ch, &card, chip, devInfo)
		}
	}

	ch <- prometheus.MustNewConstMetric(machineInfoNPUDesc, prometheus.GaugeValue, float64(totalCount))
}

func getContainerNPUInfo(ch chan<- prometheus.Metric, n *npuCollector) map[int]*container.DevicesInfo {
	if ch == nil {
		hwlog.RunLog.Error("metric channel is nil")
		return nil
	}
	obj, err := n.cache.Get(containersDevicesCacheKey)
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
		hwlog.RunLog.Error("Error container npu info cache and convert failed")
		return nil
	}
	res := make(map[int]*container.DevicesInfo, initSize)
	for _, v := range cntNpuInfos {
		for _, deviceID := range v.Devices {
			res[deviceID] = &v
		}
	}
	return res
}

func updateNPUOtherInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip) {
	if !validate(ch, npu, chip, chip.ChipIfo) {
		hwlog.RunLog.Error("Invalid param in function updateNPUOtherInfo")
		return
	}
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescHealthStatus, prometheus.GaugeValue,
			float64(getHealthCode(chip.HealthStatus)), []string{strconv.FormatInt(int64(chip.DeviceID), base),
				chip.VDieID}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuChipInfoDescErrorCode,
		prometheus.GaugeValue, float64(chip.ErrorCode), []string{strconv.FormatInt(int64(chip.DeviceID), base),
			chip.VDieID}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuChipInfoDescNpuName,
		prometheus.GaugeValue, 1, []string{strconv.FormatInt(int64(chip.DeviceID), base), fmt.Sprintf("%s-%s-%s",
			chip.ChipIfo.Name, chip.ChipIfo.Type, chip.ChipIfo.Version), chip.VDieID}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuChipInfoDescLinkStatus,
		prometheus.GaugeValue, float64(getLinkStatusCode(chip.LinkStatus)),
		[]string{strconv.FormatInt(int64(chip.DeviceID), base), chip.VDieID}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescBandwidthTx, prometheus.GaugeValue, chip.TxValue,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), chip.VDieID}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescBandwidthRx, prometheus.GaugeValue, chip.RxValue,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), chip.VDieID}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescNetworkStatus, prometheus.GaugeValue,
			float64(getHealthCode(chip.NetHealthStatus)), []string{strconv.FormatInt(int64(chip.DeviceID), base),
				chip.VDieID}...))
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

func getContainerNameArray(devInfo *container.DevicesInfo) []string {
	if devInfo == nil {
		return nil
	}

	return strings.Split(devInfo.Name, "_")
}

func updateNPUMemoryInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip) {
	if !validate(ch, npu, chip, chip.HbmInfo, chip.Meminf) {
		hwlog.RunLog.Error("Invalid param in function updateNPUMemoryInfo")
		return
	}
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescHbmUsedMemory, prometheus.GaugeValue, float64(chip.HbmInfo.Usage),
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), chip.VDieID}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescHbmTotalMemory, prometheus.GaugeValue,
			float64(chip.HbmInfo.MemorySize), []string{strconv.FormatInt(int64(chip.DeviceID), base), chip.VDieID}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuChipInfoDescUsedMemory,
		prometheus.GaugeValue, float64(chip.Meminf.MemorySize-chip.Meminf.MemoryAvailable),
		[]string{strconv.FormatInt(int64(chip.DeviceID), base), chip.VDieID}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescTotalMemory, prometheus.GaugeValue,
			float64(chip.Meminf.MemorySize), []string{strconv.FormatInt(int64(chip.DeviceID), base), chip.VDieID}...))
}

func updateContainerInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip,
	devInfo *container.DevicesInfo) {
	containerName := getContainerNameArray(devInfo)
	if len(containerName) != containerNameLen {
		return
	}
	updateContainerNPUMemoryInfo(ch, npu, chip, containerName)
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuContainerUtilization,
		prometheus.GaugeValue, float64(chip.Utilization), []string{strconv.FormatInt(int64(chip.DeviceID), base),
			containerName[nameSpaceIdx], containerName[podNameIdx], containerName[conNameIdx], chip.VDieID}...))
	ch <- prometheus.MustNewConstMetric(npuContainerInfo, prometheus.GaugeValue, 1,
		[]string{devInfo.ID, strings.Join(containerName, "_"), strconv.Itoa(chip.DeviceID), chip.VDieID}...)
}

func updateContainerNPUMemoryInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip,
	containerName []string) {
	if strings.Contains(chip.ChipIfo.Name, "910") {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuContainerTotalMemory, prometheus.GaugeValue,
				float64(chip.HbmInfo.MemorySize), []string{strconv.FormatInt(int64(chip.DeviceID), base),
					containerName[nameSpaceIdx], containerName[podNameIdx], containerName[conNameIdx],
					chip.VDieID}...))
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuContainerUsedMemory, prometheus.GaugeValue, float64(chip.HbmInfo.Usage),
				[]string{strconv.FormatInt(int64(chip.DeviceID), base), containerName[nameSpaceIdx],
					containerName[podNameIdx], containerName[conNameIdx],
					chip.VDieID}...))
		return
	}
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuContainerTotalMemory,
		prometheus.GaugeValue, float64(chip.Meminf.MemorySize), []string{strconv.FormatInt(int64(chip.DeviceID), base),
			containerName[nameSpaceIdx], containerName[podNameIdx], containerName[conNameIdx], chip.VDieID}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuContainerUsedMemory,
		prometheus.GaugeValue, float64(chip.Meminf.MemorySize-chip.Meminf.MemoryAvailable),
		[]string{strconv.FormatInt(int64(chip.DeviceID), base), containerName[nameSpaceIdx],
			containerName[podNameIdx], containerName[conNameIdx], chip.VDieID}...))
}

func updateNPUCommonInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip) {
	if !validate(ch, npu, chip) {
		hwlog.RunLog.Error("Invalid param in function updateNpuCommonInfo")
		return
	}
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuChipInfoDescUtil,
		prometheus.GaugeValue, float64(chip.Utilization), []string{strconv.FormatInt(int64(chip.DeviceID), base),
			chip.VDieID}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuChipInfoDescTemp,
		prometheus.GaugeValue, float64(chip.Temperature), []string{strconv.FormatInt(int64(chip.DeviceID), base),
			chip.VDieID}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuChipInfoDescPower,
		prometheus.GaugeValue, float64(chip.Power), []string{strconv.FormatInt(int64(chip.DeviceID), base),
			chip.VDieID}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuChipInfoDescVoltage,
		prometheus.GaugeValue, float64(chip.Voltage), []string{strconv.FormatInt(int64(chip.DeviceID), base),
			chip.VDieID}...))
}

func updateProcessInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip) {
	if chip.DevProcessInfo.ProcNum == 0 {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuChipInfoDescDevProcessInfo, prometheus.GaugeValue, 0,
				[]string{strconv.FormatInt(int64(chip.DeviceID), base), chip.VDieID, ""}...))
		return
	}

	for i := int32(0); i < chip.DevProcessInfo.ProcNum; i++ {
		procInfo := chip.DevProcessInfo.DevProcArray[i]
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuChipInfoDescDevProcessInfo, prometheus.GaugeValue, procInfo.MemUsage,
				[]string{strconv.FormatInt(int64(chip.DeviceID), base), chip.VDieID,
					strconv.FormatInt(int64(procInfo.Pid), base)}...))
	}
}

var packChipInfo = func(logicID int32, dmgr devmanager.DeviceInterface) *HuaWeiAIChip {
	chip := &HuaWeiAIChip{}
	packChipInfoPart1(logicID, dmgr, chip)
	packChipInfoPart2(logicID, dmgr, chip)
	networkPackInfo(logicID, dmgr, chip)
	return chip
}

func packChipInfoPart1(logicID int32, dmgr devmanager.DeviceInterface, hwChip *HuaWeiAIChip) {
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

	hwChip.Frequency = int(freq)
	hwChip.Power = power
	hwChip.HealthStatus = getHealth(logicID, dmgr)
	hwChip.Temperature = int(temp)
	hwChip.Voltage = vol
	hwChip.Meminf = mem
	hwChip.ChipIfo = chip
	hwChip.HbmInfo = hbmInfo
}

func packChipInfoPart2(logicID int32, dmgr devmanager.DeviceInterface, hwChip *HuaWeiAIChip) {
	util, err := dmgr.GetDeviceUtilizationRate(logicID, common.AICore)
	if err != nil {
		util = common.InvaidVal // valid data range 0-100
	}
	_, errCode, err := dmgr.GetDeviceErrorCode(logicID)
	if err != nil {
		errCode = common.RetError // valid data range 0-128
	}
	vdieID, err := dmgr.GetDieID(logicID, dcmi.VDIE)
	if err != nil {
		hwlog.RunLog.Debug(err)
	}
	info, err := dmgr.GetDevProcessInfo(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		info = new(common.DevProcessInfo)
	}

	hwChip.ErrorCode = errCode
	hwChip.Utilization = int(util)
	hwChip.VDieID = vdieID
	hwChip.DevProcessInfo = info
}

func networkPackInfo(logicID int32, dmgr devmanager.DeviceInterface, hwChip *HuaWeiAIChip) {
	hwChip.LinkStatus = LinkDown
	hwChip.NetHealthStatus = UnHealthy

	if !strings.Contains(hwChip.ChipIfo.Name, "910") {
		return
	}

	phyID, err := dmgr.GetPhysicIDFromLogicID(logicID)
	if err != nil {
		return
	}
	if tx, rx, err := getNPUInterfaceTraffic(phyID); err == nil {
		hwChip.TxValue = tx
		hwChip.RxValue = rx
	}
	netCode, err := dmgr.GetDeviceNetWorkHealth(logicID)
	hwlog.RunLog.Debugf("chip %d network healthy code is %d", logicID, netCode)
	if err != nil {
		netCode = math.MaxUint32
	}

	hwChip.NetHealthStatus = getNetworkHealthy(netCode)
	hwChip.LinkStatus = getNPULinkStatus(phyID)
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

func getNPULinkStatus(phyID int32) LinkEnum {
	args := []string{"-i", strconv.Itoa(int(phyID)), "-link", "-g"}
	// command example: hccn_tool -i 0 -link -g
	// success result example is: link status: DOWN
	outStr, err := hccnToolGetLink(args...)
	hwlog.RunLog.Debugf("hccn_tool command exec result: %#v", outStr)
	if err != nil {
		hwlog.RunLog.Errorf("get npu link status failed, %#v", err)
		return LinkDown
	}
	replacedStr := strings.ReplaceAll(outStr, "\n", "")
	outArr := strings.Split(replacedStr, space)
	if len(outArr) != linkStatusPart {
		return LinkDown
	}
	var lastIndex = 2
	status := outArr[lastIndex]
	hwlog.RunLog.Debugf("hccn_tool get npu link status: %s", status)
	return LinkEnum(status)
}

func getNPUInterfaceTraffic(phyID int32) (float64, float64, error) {
	args := []string{"-i", strconv.Itoa(int(phyID)), "-bandwidth", "-g"}
	// command example: hccn_tool -i 0 -bandwidth -g
	// success result has two lines:
	// Bandwidth TX: 0.00 MB/sec
	// Bandwidth RX: 0.00 MB/sec
	outStr, err := hccnToolGetLink(args...)
	hwlog.RunLog.Debugf("hccn_tool command exec result: %#v", outStr)
	if err != nil {
		hwlog.RunLog.Errorf("get npu interface status failed, %#v", err)
		return noTraffic, noTraffic, err
	}

	var (
		speedIndex = 2
		tx         = 0.00
		rx         = 0.00
		txStr      = "TX:"
		rxStr      = "RX:"
		base64     = 64
	)
	lines := strings.Split(outStr, newLine)
	for _, line := range lines {
		if line == "" {
			continue
		}

		trafficArr := strings.Split(line, space)
		hwlog.RunLog.Debugf("npu bandwidth split as: %#v", trafficArr)
		if len(trafficArr) != trafficPart {
			continue
		}
		if strings.Contains(line, txStr) {
			if tmpTx, err := strconv.ParseFloat(trafficArr[speedIndex], base64); err == nil {
				tx = tmpTx
			}
			continue
		}
		if strings.Contains(line, rxStr) {
			if tmpRx, err := strconv.ParseFloat(trafficArr[speedIndex], base64); err == nil {
				rx = tmpRx
			}
		}
	}
	return tx, rx, nil
}

func hccnToolGetLink(args ...string) (string, error) {
	const hccn_tool = "/usr/local/Ascend/driver/tools/hccn_tool"
	if _, err := utils.CheckPath(hccn_tool); err != nil {
		return "", err
	}
	hwlog.RunLog.Debugf("command is: %s %s", hccn_tool, args)
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(hccn_tool, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", errors.New(string(stderr.Bytes()))
	}

	return string(stdout.Bytes()), nil
}

func getLinkStatusCode(status LinkEnum) int {
	if LinkUp == status {
		return 1
	}
	return 0
}

func getNetworkHealthy(netCode uint32) HealthEnum {
	if netCode == common.NetworkInit || netCode == common.NetworkSuccess {
		return Healthy
	}

	return UnHealthy
}
