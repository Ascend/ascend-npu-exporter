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

// Package collector for Prometheus
package collector

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"huawei.com/npu-exporter/dsmi"
	"k8s.io/klog"
	"math"
	"os"
	"reflect"
	"strconv"
	"time"
)

var (
	versionInfoDesc               = prometheus.NewDesc("npu_exporter_version_info", "exporter version with value '1'", []string{"exporterVersion"}, nil)
	machineInfoNPUDesc            = prometheus.NewDesc("machine_npu_nums", "Amount of npu installed on the machine.", nil, nil)
	npuChipInfoDescNpuName        = prometheus.NewDesc("npu_chip_info_name", "the Ascend npu name with value '1'", []string{"id", "name"}, nil)
	npuChipInfoDescUtil           = prometheus.NewDesc("npu_chip_info_utilization", "the ai core utilization", []string{"id"}, nil)
	npuChipInfoDescTemp           = prometheus.NewDesc("npu_chip_info_temperature", "the npu temperature", []string{"id"}, nil)
	npuChipInfoDescPower          = prometheus.NewDesc("npu_chip_info_power", "the npu power", []string{"id"}, nil)
	npuChipInfoDescVoltage        = prometheus.NewDesc("npu_chip_info_voltage", "the npu voltage", []string{"id"}, nil)
	npuChipInfoDescUsedMemory     = prometheus.NewDesc("npu_chip_info_used_memory", "the npu used memory", []string{"id"}, nil)
	npuChipInfoDescTotalMemory    = prometheus.NewDesc("npu_chip_info_total_memory", "the npu total memory", []string{"id"}, nil)
	npuChipInfoDescHealthStatus   = prometheus.NewDesc("npu_chip_info_health_status", "the npu health status", []string{"id"}, nil)
	npuChipInfoDescHbmUsedMemory  = prometheus.NewDesc("npu_chip_info_hbm_used_memory", "the npu hbm used memory", []string{"id"}, nil)
	npuChipInfoDescHbmTotalMemory = prometheus.NewDesc("npu_chip_info_hbm_total_memory", "the npu hbm total memory", []string{"id"}, nil)
	npuChipInfoDescErrorCode      = prometheus.NewDesc("npu_chip_info_error_code", "the npu error code", []string{"id"}, nil)
)

type npuCollector struct {
	cache      *cache.Cache
	updateTime time.Duration
	cacheTime  time.Duration
}

// NewNpuCollector new a instance of prometheus Collector
func NewNpuCollector(cacheTime time.Duration, updateTime time.Duration, stop chan os.Signal) prometheus.Collector {
	npuCollect := &npuCollector{
		cache:      cache.New(cacheTime, five*time.Minute),
		cacheTime:  cacheTime,
		updateTime: updateTime,
	}
	go start(npuCollect, stop)
	return npuCollect
}

var getNPUInfo = func(dmgr dsmi.DeviceMgrInterface) []HuaWeiNPUCard {
	var npuList []HuaWeiNPUCard
	cardNum, cards, err := dmgr.GetCardList()
	if cardNum == 0 || err != nil {
		klog.Warning("Downgrade to user DSMI only,maybe need check ENV of LD_LIBRARY_PATH")
		return assembleNPUInfoV1(dmgr)
	}
	var logicID int32 = 0
	for _, cardID := range cards {
		deviceNum, err := dmgr.GetDeviceNumOnCard(cardID)
		if err != nil {
			klog.Errorf("Can't get the device count on the card %d", cardID)
			continue
		}
		var deviceList []*HuaWeiAIChip
		for i := int32(0); i < deviceNum; i++ {
			chipInfo := assembleNPUInfoV2(cardID, logicID, dmgr)
			if chipInfo != nil {
				deviceList = append(deviceList, chipInfo)
			}
			logicID++
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

func assembleNPUInfoV2(cardID int32, logicID int32, dmgr dsmi.DeviceMgrInterface) *HuaWeiAIChip {
	phyID, err := dmgr.GetPhyIDFromLogicID(uint32(logicID))
	// check cardId, convert it to int type later
	if phyID > math.MaxInt8 || err != nil {
		return nil
	}
	chipInfo := packChipInfo(logicID, dmgr)
	chipInfo.DeviceID = int(phyID)
	if dsmi.GetChipTypeNow() == dsmi.Ascend710 {
		cardPower, err := dmgr.GetCardPower(cardID)
		if err != nil {
			cardPower = dsmi.DefaultErrorValue
		}
		// Ascend710 use cardPower to replace chipPower
		chipInfo.Power = cardPower
	}
	return chipInfo
}

var assembleNPUInfoV1 = func(dmgr dsmi.DeviceMgrInterface) []HuaWeiNPUCard {
	var npuList []HuaWeiNPUCard
	num, devices, err := dmgr.GetDeviceList()
	if num == 0 || err != nil {
		return npuList
	}
	for _, logicID := range devices {
		phyID, err := dmgr.GetPhyIDFromLogicID(uint32(logicID))
		// check cardId, convert it to int type later
		if phyID > math.MaxInt8 || err != nil {
			continue
		}
		chipInfo := packChipInfo(logicID, dmgr)
		chipInfo.DeviceID = int(phyID)
		npuCard := HuaWeiNPUCard{
			CardID:     dsmi.DefaultErrorValue,
			DeviceList: []*HuaWeiAIChip{chipInfo},
			Timestamp:  time.Now(),
		}
		npuList = append(npuList, npuCard)
	}
	return npuList
}

var start = func(n *npuCollector, stop <-chan os.Signal) {
	defer func() {
		if err := recover(); err != nil {
			klog.Errorf("go routine failed with %v", err)
		}
	}()
	if n == nil || stop == nil {
		klog.Error("Invalid param in function start")
		return
	}
	ticker := time.NewTicker(n.updateTime)
	klog.Infof("Starting update cache every %d seconds", n.updateTime/time.Second)
	for {
		select {
		case _, ok := <-ticker.C:
			if !ok {
				return
			}
			npuInfo := getNPUInfo(dsmi.GetDeviceManager())
			n.cache.Set(key, npuInfo, n.cacheTime)
			klog.Infof("update cache,key is %s", key)
		case _, ok := <-stop:
			if !ok {
				klog.Error("closed")
				return
			}
			ticker.Stop()
			klog.Warning("received the stop signal,STOPPED")
			dsmi.ShutDown()
			os.Exit(0)
		}
	}
}

// Describe implements prometheus.Collector
func (n *npuCollector) Describe(ch chan<- *prometheus.Desc) {
	if ch == nil {
		klog.Error("Invalid param in function Describe")
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
}

// Collect implements prometheus.Collector
func (n *npuCollector) Collect(ch chan<- prometheus.Metric) {
	if !validate(ch) {
		klog.Error("Invalid param in function Collect")
		return
	}
	obj, found := n.cache.Get(key)
	if !found {
		klog.Warning("no cache, start to get npulist and rebuild cache")
		npuInfo := getNPUInfo(dsmi.GetDeviceManager())
		n.cache.Set(key, npuInfo, n.cacheTime)
		klog.Warning("rebuild cache successfully")
		obj = npuInfo
	}
	npuList, ok := obj.([]HuaWeiNPUCard)
	if !ok {
		klog.Error("Error cache and convert failed")
		n.cache.Delete(key)
	}
	ch <- prometheus.MustNewConstMetric(versionInfoDesc, prometheus.GaugeValue, 1, []string{BuildVersion}...)
	var totalCount = 0
	for _, card := range npuList {
		deviceCount := len(card.DeviceList)
		if deviceCount <= 0 {
			continue
		}
		totalCount += deviceCount
		for _, chip := range card.DeviceList {
			updateNPUCommonInfo(ch, &card, chip)
			updateNPUMemoryInfo(ch, &card, chip)
			updateNPUOtherInfo(ch, &card, chip)
		}
	}
	ch <- prometheus.MustNewConstMetric(machineInfoNPUDesc, prometheus.GaugeValue, float64(totalCount))
}

func updateNPUOtherInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip) {
	if !validate(ch, npu, chip, chip.ChipIfo) {
		klog.Error("Invalid param in function updateNPUOtherInfo")
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
				fmt.Sprintf("%s-%s-%s", chip.ChipIfo.ChipName, chip.ChipIfo.ChipType, chip.ChipIfo.ChipVer)}...))
}

func validate(ch chan<- prometheus.Metric, objs ...interface{}) bool {
	if ch == nil {
		return false
	}
	for _, v := range objs {
		if reflect.ValueOf(v).IsNil() {
			return false
		}
	}
	return true
}

func updateNPUMemoryInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip) {
	if !validate(ch, npu, chip, chip.HbmInfo, chip.Meminf) {
		klog.Error("Invalid param in function updateNPUMemoryInfo")
		return
	}
	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescHbmUsedMemory, prometheus.GaugeValue,
			float64(chip.HbmInfo.MemoryUsage),
			[]string{strconv.FormatInt(int64(chip.DeviceID), base)}...))
	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescHbmTotalMemory, prometheus.GaugeValue,
			float64(chip.HbmInfo.MemorySize),
			[]string{strconv.FormatInt(int64(chip.DeviceID), base)}...))

	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescUsedMemory, prometheus.GaugeValue,
			float64(chip.Meminf.MemorySize*chip.Meminf.Utilization/dsmi.Percent),
			[]string{strconv.FormatInt(int64(chip.DeviceID), base)}...))
	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescTotalMemory, prometheus.GaugeValue, float64(chip.Meminf.MemorySize),
			[]string{strconv.FormatInt(int64(chip.DeviceID), base)}...))
}

func updateNPUCommonInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip) {
	if !validate(ch, npu, chip) {
		klog.Error("Invalid param in function updateNpuCommonInfo")
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
}

var packChipInfo = func(logicID int32, dmgr dsmi.DeviceMgrInterface) *HuaWeiAIChip {
	freq, err := dmgr.GetDeviceFrequency(logicID, dsmi.AICore)
	if err != nil {
		freq = dsmi.DefaultErrorValue
	}
	power, err := dmgr.GetDevicePower(logicID)
	if err != nil {
		power = dsmi.DefaultErrorValue
	}
	temp, err := dmgr.GetDeviceTemperature(logicID)
	if err != nil {
		temp = defaultTemperatureWhenQueryFailed
	}
	vol, err := dmgr.GetDeviceVoltage(logicID)
	if err != nil {
		vol = dsmi.DefaultErrorValue
	}
	mem, err := dmgr.GetDeviceMemoryInfo(logicID)
	if err != nil {
		mem = &dsmi.MemoryInfo{}
	}
	chip, err := dmgr.GetChipInfo(logicID)
	if err != nil {
		chip = &dsmi.ChipInfo{}
	}
	hbmInfo, err := dmgr.GetDeviceHbmInfo(logicID)
	if err != nil {
		hbmInfo = &dsmi.HbmInfo{}
	}
	util, err := dmgr.GetDeviceUtilizationRate(logicID, dsmi.AICore)
	if err != nil {
		util = dsmi.DefaultErrorValue // valid data range 0-100
	}
	_, errCode, err := dmgr.GetDeviceErrCode(logicID)
	if err != nil {
		errCode = dsmi.DefaultErrorValue // valid data range 0-128
	}
	return &HuaWeiAIChip{
		ErrorCode:    int(errCode),
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

func getHealth(cardID int32, dmgr dsmi.DeviceMgrInterface) HealthEnum {
	health, err := dmgr.GetDeviceHealth(cardID)
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
