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
	"strconv"
	"time"
)

var (
	versionInfoDesc               = prometheus.NewDesc("npu_exporter_version_info", "exporter version with value '1'", []string{"exporterVersion"}, nil)
	machineInfoNPUDesc            = prometheus.NewDesc("machine_npu_nums", "Amount of npu installed on the machine.", nil, nil)
	npuChipInfDescNpuName         = prometheus.NewDesc("npu_chip_info_name", "the Ascent npu name with value '1'", []string{"id", "name"}, nil)
	npuChipInfoDescUtil           = prometheus.NewDesc("npu_chip_info_utilization", "the ai core utilization", []string{"id"}, nil)
	npuChipInfoDescTemp           = prometheus.NewDesc("npu_chip_info_temperature", "the npu temperature", []string{"id"}, nil)
	npuChipInfoDescPower          = prometheus.NewDesc("npu_chip_info_power", "the npu power", []string{"id"}, nil)
	npuChipInfoDescVoltage        = prometheus.NewDesc("npu_chip_info_voltage", "the npu voltage", []string{"id"}, nil)
	npuChipInfoDescUsedMemory     = prometheus.NewDesc("npu_chip_info_used_memory", "the npu used memory", []string{"id"}, nil)
	npuChipInfoDescTotalMemory    = prometheus.NewDesc("npu_chip_info_total_memory", "the npu total memory", []string{"id"}, nil)
	npuChipInfoDescHealthStatus   = prometheus.NewDesc("npu_chip_info_health_status", "the npu health status", []string{"id"}, nil)
	npuChipInfoDescHbmUsedMemory  = prometheus.NewDesc("npu_chip_info_hbm_used_memory", "the npu hbm used memory", []string{"id"}, nil)
	npuChipInfoDescHbmTotalMemory = prometheus.NewDesc("npu_chip_info_hbm_total_memory", "the npu hbm total memory", []string{"id"}, nil)
	npuChipInfDescErrorCode       = prometheus.NewDesc("npu_chip_info_error_code", "the npu error code", []string{"id"}, nil)
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

var getNPUInfo = func(dmgr dsmi.DeviceMgrInterface) []HuaWeiNPUDevice {
	var npuList []HuaWeiNPUDevice
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
		npuCard := HuaWeiNPUDevice{
			CardID:     int(phyID), // user phyID
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
			klog.Infof("update cache, key is %s", key)
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
	ch <- npuChipInfoDescTotalMemory
	ch <- npuChipInfoDescUsedMemory
	ch <- npuChipInfDescErrorCode
	ch <- npuChipInfDescNpuName
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
	}
	npuList, ok := obj.([]HuaWeiNPUDevice)
	if !ok {
		klog.Error("Error cache and convert failed")
		n.cache.Delete(key)
	}
	ch <- prometheus.MustNewConstMetric(versionInfoDesc, prometheus.GaugeValue, 1, []string{BuildVersion}...)
	ch <- prometheus.MustNewConstMetric(machineInfoNPUDesc, prometheus.GaugeValue, float64(len(npuList)))
	for _, npu := range npuList {
		if len(npu.DeviceList) < 0 {
			continue
		}
		for _, chip := range npu.DeviceList {
			updateNpuCommonInfo(ch, &npu, chip)
			updateNPUMemoryInfo(ch, &npu, chip)
			updateNPUOtherInfo(ch, &npu, chip)
		}
	}

}

func updateNPUOtherInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUDevice, chip *HuaWeiAIChip) {
	if !validate(ch, npu, chip) {
		klog.Error("Invalid param in function updateNPUOtherInfo")
		return
	}
	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescHealthStatus, prometheus.GaugeValue,
			float64(getHealthCode(chip.HealthStatus)), []string{strconv.FormatInt(int64(npu.CardID), base)}...))
	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfDescErrorCode, prometheus.GaugeValue, float64(chip.ErrorCode),
			[]string{strconv.FormatInt(int64(npu.CardID), base)}...))
	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfDescNpuName, prometheus.GaugeValue, 1,
			[]string{strconv.FormatInt(int64(npu.CardID), base),
				fmt.Sprintf("%s-%s-%s", chip.ChipIfo.ChipName, chip.ChipIfo.ChipType, chip.ChipIfo.ChipVer)}...))
}

func validate(ch chan<- prometheus.Metric, objs ...interface{}) bool {
	if ch == nil {
		return false
	}
	for _, v := range objs {
		if v == nil {
			return false
		}
	}
	return true
}

func updateNPUMemoryInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUDevice, chip *HuaWeiAIChip) {
	if !validate(ch, npu, chip) {
		klog.Error("Invalid param in function updateNPUMemoryInfo")
		return
	}
	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescHbmUsedMemory, prometheus.GaugeValue,
			float64(chip.HbmInfo.MemoryUsage),
			[]string{strconv.FormatInt(int64(npu.CardID), base)}...))
	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescHbmTotalMemory, prometheus.GaugeValue,
			float64(chip.HbmInfo.MemorySize),
			[]string{strconv.FormatInt(int64(npu.CardID), base)}...))

	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescUsedMemory, prometheus.GaugeValue,
			float64(chip.Meminf.MemorySize*chip.Meminf.Utilization/dsmi.Percent),
			[]string{strconv.FormatInt(int64(npu.CardID), base)}...))
	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescTotalMemory, prometheus.GaugeValue, float64(chip.Meminf.MemorySize),
			[]string{strconv.FormatInt(int64(npu.CardID), base)}...))
}

func updateNpuCommonInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUDevice, chip *HuaWeiAIChip) {
	if !validate(ch, npu, chip) {
		klog.Error("Invalid param in function updateNpuCommonInfo")
		return
	}
	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescUtil, prometheus.GaugeValue, float64(chip.Utilization),
			[]string{strconv.FormatInt(int64(npu.CardID), base)}...))
	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescTemp, prometheus.GaugeValue, float64(chip.Temperature),
			[]string{strconv.FormatInt(int64(npu.CardID), base)}...))

	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescPower, prometheus.GaugeValue, float64(chip.Power),
			[]string{strconv.FormatInt(int64(npu.CardID), base)}...))

	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescVoltage, prometheus.GaugeValue, float64(chip.Voltage),
			[]string{strconv.FormatInt(int64(npu.CardID), base)}...))
}

var packChipInfo = func(logicID int32, dmgr dsmi.DeviceMgrInterface) *HuaWeiAIChip {
	freq, err := dmgr.GetDeviceFrequency(logicID, dsmi.AICore)
	if err != nil {
		freq = defaultValueWhenQueryFailed
	}
	power, err := dmgr.GetDevicePower(logicID)
	if err != nil {
		power = defaultValueWhenQueryFailed
	}
	temp, err := dmgr.GetDeviceTemperature(logicID)
	if err != nil {
		temp = defaultTemperatureWhenQueryFailed
	}
	vol, err := dmgr.GetDeviceVoltage(logicID)
	if err != nil {
		vol = defaultValueWhenQueryFailed
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
		util = defaultValueWhenQueryFailed
	}
	_, errCode, err := dmgr.GetDeviceErrCode(logicID)
	if err != nil {
		errCode = defaultValueWhenQueryFailed
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
