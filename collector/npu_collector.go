// Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.

// Package collector for Prometheus
package collector

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"huawei.com/npu-exporter/collector/container"
	"huawei.com/npu-exporter/dsmi"
	"huawei.com/npu-exporter/hwlog"
	"math"
	"os"
	"reflect"
	"strconv"
	"sync"
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
	npuContainerInfo              = prometheus.NewDesc("npu_container_info", "the container name and deviceID relationship", []string{"containerID", "containerName", "npuID"}, nil)
	npuContainerInfoInit          sync.Once
	npuChipInfoInit               sync.Once
)

type npuCollector struct {
	cache         *cache.Cache
	devicesParser *container.DevicesParser
	updateTime    time.Duration
	cacheTime     time.Duration
}

// NewNpuCollector new a instance of prometheus Collector
func NewNpuCollector(cacheTime time.Duration, updateTime time.Duration, stop chan os.Signal,
	deviceParser *container.DevicesParser) prometheus.Collector {
	npuCollect := &npuCollector{
		cache:         cache.New(cacheTime, five*time.Minute),
		cacheTime:     cacheTime,
		updateTime:    updateTime,
		devicesParser: deviceParser,
	}
	go start(npuCollect, stop, dsmi.GetDeviceManager())
	return npuCollect
}

var getNPUInfo = func(dmgr dsmi.DeviceMgrInterface) []HuaWeiNPUCard {
	var npuList []HuaWeiNPUCard
	cardNum, cards, err := dmgr.GetCardList()
	if cardNum == 0 || err != nil {
		hwlog.RunLog.Warn("Downgrade to user DSMI only,maybe need check ENV of LD_LIBRARY_PATH")
		return assembleNPUInfoV1(dmgr)
	}
	var logicID int32 = 0
	for _, cardID := range cards {
		deviceNum, err := dmgr.GetDeviceNumOnCard(cardID)
		if err != nil {
			continue
		}
		var deviceList []*HuaWeiAIChip
		for i := int32(0); i < deviceNum; i++ {
			var chipInfo *HuaWeiAIChip
			logID, err := dmgr.GetDeviceLogicID(cardID, i)
			if err == nil {
				chipInfo = assembleNPUInfoV2(cardID, logID, dmgr)
			} else {
				chipInfo = assembleNPUInfoV2(cardID, logicID, dmgr)
			}
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
	if phyID > int32(math.MaxInt8) || err != nil {
		return nil
	}
	chipInfo := packChipInfo(logicID, dmgr)
	chipInfo.DeviceID = int(phyID)
	if dsmi.GetChipTypeNow() == dsmi.Ascend710 {
		cardPower, err := dmgr.GetCardPower(cardID)
		if err != nil {
			cardPower = float32(dsmi.DefaultErrorValue)
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
		if phyID > int32(math.MaxInt8) || err != nil {
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

var start = func(n *npuCollector, stop <-chan os.Signal, dmgr dsmi.DeviceMgrInterface) {
	defer func() {
		if err := recover(); err != nil {
			hwlog.RunLog.Errorf("go routine failed with %v", err)
		}
	}()

	if n == nil || stop == nil {
		hwlog.RunLog.Error("Invalid param in function start")
		return
	}

	if err := n.devicesParser.Init(); err != nil {
		hwlog.RunLog.Errorf("failed to init devices parser: %v", err)
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
			n.cache.Set(key, npuInfo, n.cacheTime)
			hwlog.RunLog.Infof("update cache,key is %s", key)
		case result := <-n.devicesParser.RecvResult():
			n.cache.Set(containersDevicesInfoKey, result, n.cacheTime)
			hwlog.RunLog.Infof("update cache,key is %s", containersDevicesInfoKey)
		case err := <-n.devicesParser.RecvErr():
			hwlog.RunLog.Errorf("received error from device parser: %v", err)
		case _, ok := <-stop:
			if !ok {
				hwlog.RunLog.Error("closed")
				return
			}
			ticker.Stop()
			hwlog.RunLog.Warn("received the stop signal,STOPPED")
			dsmi.ShutDown()
			os.Exit(0)
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
}

// Collect implements prometheus.Collector
func (n *npuCollector) Collect(ch chan<- prometheus.Metric) {
	if !validate(ch) {
		hwlog.RunLog.Error("Invalid param in function Collect")
		return
	}
	obj, found := n.cache.Get(key)
	npuChipInfoInit.Do(func() {
		if !found {
			hwlog.RunLog.Warn("no cache, start to get npulist and rebuild cache")
			npuInfo := getNPUInfo(dsmi.GetDeviceManager())
			n.cache.Set(key, npuInfo, n.cacheTime)
			hwlog.RunLog.Warn("rebuild cache successfully")
			obj = npuInfo
		}
	})
	npuList, ok := obj.([]HuaWeiNPUCard)
	if !ok {
		hwlog.RunLog.Error("Error cache and convert failed")
		n.cache.Delete(key)
	}
	ch <- prometheus.MustNewConstMetric(versionInfoDesc, prometheus.GaugeValue, 1, []string{hwlog.BuildVersion}...)
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

	updateContainerNPUInfo(ch, n)
}

func updateContainerNPUInfo(ch chan<- prometheus.Metric, n *npuCollector) {
	if ch == nil {
		hwlog.RunLog.Error("metric channel is nil")
		return
	}
	obj, found := n.cache.Get(containersDevicesInfoKey)
	// only run once to prevent wait when container info get failed
	npuContainerInfoInit.Do(func() {
		if !found {
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
		return
	}
	for _, v := range cntNpuInfos {
		for _, deviceID := range v.Devices {
			ch <- prometheus.MustNewConstMetric(
				npuContainerInfo,
				prometheus.GaugeValue, 1,
				[]string{v.ID, v.Name, strconv.Itoa(deviceID)}...)
		}
	}
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
		hwlog.RunLog.Error("Invalid param in function updateNPUMemoryInfo")
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
			float64(chip.Meminf.MemorySize*uint64(chip.Meminf.Utilization)/uint64(dsmi.Percent)),
			[]string{strconv.FormatInt(int64(chip.DeviceID), base)}...))
	ch <- prometheus.NewMetricWithTimestamp(
		npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescTotalMemory, prometheus.GaugeValue, float64(chip.Meminf.MemorySize),
			[]string{strconv.FormatInt(int64(chip.DeviceID), base)}...))
}

func updateNPUCommonInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip) {
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
		temp = dsmi.DefaultTemperatureWhenQueryFailed
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
	// return the first errorCode and data value type is int64
	_, errCode, err := dmgr.GetDeviceErrCode(logicID)
	if err != nil {
		errCode = dsmi.DefaultErrorValue // valid data range 0-128
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
