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
	"context"
	"math"
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
	"huawei.com/npu-exporter/v5/devmanager/dcmi"
	"huawei.com/npu-exporter/v5/devmanager/hccn"
	"huawei.com/npu-exporter/v5/versions"
)

// metric label name
const (
	npuID       = "id"
	modelName   = "model_name"
	npuUUID     = "vdie_id"
	vNpuUUID    = "v_dev_id"
	npuPCIEInfo = "pcie_bus_info"
	namespace   = "namespace"
	podName     = "pod_name"
	isVirtual   = "is_virtual"
)

const (
	txPower0 = "Tx_Power0"
	txPower1 = "Tx_Power1"
	txPower2 = "Tx_Power2"
	txPower3 = "Tx_Power3"

	rxPower0 = "Rx_Power0"
	rxPower1 = "Rx_Power1"
	rxPower2 = "Rx_Power2"
	rxPower3 = "Rx_Power3"

	present     = "present"
	temperature = "temperature"
	voltage     = "Vcc"
)

const (
	macRxMacPauseNum       = "mac_rx_mac_pause_num"
	macTxMacPauseNum       = "mac_tx_mac_pause_num"
	macRxPfcPktNum         = "mac_rx_pfc_pkt_num"
	macTxPfcPktNum         = "mac_tx_pfc_pkt_num"
	macRxBadPktNum         = "mac_rx_bad_pkt_num"
	macTxBadPktNum         = "mac_tx_bad_pkt_num"
	roCERxAllPktNum        = "roce_rx_all_pkt_num"
	roCETxAllPktNum        = "roce_tx_all_pkt_num"
	roCERxErrPktNum        = "roce_rx_err_pkt_num"
	roCETxErrPktNum        = "roce_tx_err_pkt_num"
	roCERxCnpPktNum        = "roce_rx_cnp_pkt_num"
	roCETxCnpPktNum        = "roce_tx_cnp_pkt_num"
	macRxBadOctNum         = "mac_rx_bad_oct_num"
	macTxBadOctNum         = "mac_tx_bad_oct_num"
	roCEUnexpectedAckNum   = "roce_unexpected_ack_num"
	roCEOutOfOrderNum      = "roce_out_of_order_num"
	roCEVerificationErrNum = "roce_verification_err_num"
	roCEQpStatusErrNum     = "roce_qp_status_err_num"
	roCENewPktRtyNum       = "roce_new_pkt_rty_num"
)

var (
	versionInfoDesc = prometheus.NewDesc("npu_exporter_version_info",
		"exporter version with value '1'", []string{"exporterVersion"}, nil)
	machineInfoNPUDesc = prometheus.NewDesc("machine_npu_nums",
		"Amount of npu installed on the machine.", nil, nil)
	npuChipInfoDescNpuName = prometheus.NewDesc("npu_chip_info_name",
		"the Ascend npu name with value '1'", []string{npuID, "name", npuUUID, npuPCIEInfo}, nil)
	npuChipInfoDescUtil = prometheus.NewDesc("npu_chip_info_utilization",
		"the ai core utilization", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipInfoDescTemp = prometheus.NewDesc("npu_chip_info_temperature",
		"the npu temperature", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipInfoDescPower = prometheus.NewDesc("npu_chip_info_power",
		"the npu power", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipInfoDescVoltage = prometheus.NewDesc("npu_chip_info_voltage",
		"the npu voltage", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipInfoDescUsedMemory = prometheus.NewDesc("npu_chip_info_used_memory",
		"the npu used memory", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipInfoDescTotalMemory = prometheus.NewDesc("npu_chip_info_total_memory",
		"the npu total memory", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipInfoDescHealthStatus = prometheus.NewDesc("npu_chip_info_health_status",
		"the npu health status", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipInfoDescHbmUsedMemory = prometheus.NewDesc("npu_chip_info_hbm_used_memory",
		"the npu hbm used memory", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipInfoDescHbmTotalMemory = prometheus.NewDesc("npu_chip_info_hbm_total_memory",
		"the npu hbm total memory", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipInfoDescErrorCode = prometheus.NewDesc("npu_chip_info_error_code",
		"the npu error code", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipInfoDescLinkStatus = prometheus.NewDesc("npu_chip_info_link_status",
		"the npu link status", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipInfoDescNetworkStatus = prometheus.NewDesc("npu_chip_info_network_status",
		"the npu network health status", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipInfoDescBandwidthTx = prometheus.NewDesc("npu_chip_info_bandwidth_tx",
		"the npu interface transport speed, unit is 'MB/s'", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipInfoDescBandwidthRx = prometheus.NewDesc("npu_chip_info_bandwidth_rx",
		"the npu interface receive speed, unit is 'MB/s'", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipLinkSpeed = prometheus.NewDesc("npu_chip_link_speed",
		"the npu interface receive link speed, unit is 'Mb/s'", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipLinkUpNum = prometheus.NewDesc("npu_chip_link_up_num",
		"the npu interface receive link-up num", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipMacRxPauseNum = prometheus.NewDesc("npu_chip_mac_rx_pause_num",
		"the npu interface receive mac-rx-pause-num", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipMacTxPauseNum = prometheus.NewDesc("npu_chip_mac_tx_pause_num",
		"the npu interface receive mac-tx-pause-num", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipMacRxPfcPktNum = prometheus.NewDesc("npu_chip_mac_rx_pfc_pkt_num",
		"the npu interface receive mac-rx-pfc-pkt-num", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipMacTxPfcPktNum = prometheus.NewDesc("npu_chip_mac_tx_pfc_pkt_num",
		"the npu interface receive mac-tx-pfc-pkt-num", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipMacRxBadPktNum = prometheus.NewDesc("npu_chip_mac_rx_bad_pkt_num",
		"the npu interface receive mac-rx-bad-pkt-num", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipMacTxBadPktNum = prometheus.NewDesc("npu_chip_mac_tx_bad_pkt_num",
		"the npu interface receive mac-tx-bad-pkt-num", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipRoceRxAllPktNum = prometheus.NewDesc("npu_chip_roce_rx_all_pkt_num",
		"the npu interface receive roce-rx-all-pkt-num", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipRoceTxAllPktNum = prometheus.NewDesc("npu_chip_roce_tx_all_pkt_num",
		"the npu interface receive roce-tx-all-pkt-num", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipRoceRxErrPktNum = prometheus.NewDesc("npu_chip_roce_rx_err_pkt_num",
		"the npu interface receive roce-rx-err-pkt-num", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipRoceTxErrPktNum = prometheus.NewDesc("npu_chip_roce_tx_err_pkt_num",
		"the npu interface receive roce-tx-err-pkt-num", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipRoceRxCnpPktNum = prometheus.NewDesc("npu_chip_roce_rx_cnp_pkt_num",
		"the npu interface receive roce-rx-cnp-pkt-num", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipRoceTxCnpPktNum = prometheus.NewDesc("npu_chip_roce_tx_cnp_pkt_num",
		"the npu interface receive roce-tx-cnp-pkt-num", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipRoceNewPktRtyNum = prometheus.NewDesc("npu_chip_roce_new_pkt_rty_num",
		"the npu interface receive roce-new-pkt-rty-num", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipMacTxBadOctNum = prometheus.NewDesc("npu_chip_mac_tx_bad_oct_num",
		"the npu interface receive mac-tx-bad-oct-num", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipMacRxBadOctNum = prometheus.NewDesc("npu_chip_mac_rx_bad_oct_num",
		"the npu interface receive mac-rx-bad-oct-num", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipRoceUnexpectedAcktNum = prometheus.NewDesc("npu_chip_roce_unexpected_ack_num",
		"the npu interface receive roce-unexpected-ack-num", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipRoceOutOfOrderNum = prometheus.NewDesc("npu_chip_roce_out_of_order_num",
		"the npu interface receive roce-out-of-order-num", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipRoceVerificationErrNum = prometheus.NewDesc("npu_chip_roce_verification_err_num",
		"the npu interface receive roce-verification-err-num", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipRoceQpStatusErrNum = prometheus.NewDesc("npu_chip_roce_qp_status_err_num",
		"the npu interface receive roce-qp-status-err-num", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipOpticalState = prometheus.NewDesc("npu_chip_optical_state",
		"the npu interface receive optical-state", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipOpticalTxPower0 = prometheus.NewDesc("npu_chip_optical_tx_power_0",
		"the npu interface receive optical-tx-power-0", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipOpticalTxPower1 = prometheus.NewDesc("npu_chip_optical_tx_power_1",
		"the npu interface receive optical-tx-power-1", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipOpticalTxPower2 = prometheus.NewDesc("npu_chip_optical_tx_power_2",
		"the npu interface receive optical-tx-power-2", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipOpticalTxPower3 = prometheus.NewDesc("npu_chip_optical_tx_power_3",
		"the npu interface receive optical-tx-power-3", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipOpticalRxPower0 = prometheus.NewDesc("npu_chip_optical_rx_power_0",
		"the npu interface receive optical-rx-power-0", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipOpticalRxPower1 = prometheus.NewDesc("npu_chip_optical_rx_power_1",
		"the npu interface receive optical-rx-power-1", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipOpticalRxPower2 = prometheus.NewDesc("npu_chip_optical_rx_power_2",
		"the npu interface receive optical-rx-power-2", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipOpticalRxPower3 = prometheus.NewDesc("npu_chip_optical_rx_power_3",
		"the npu interface receive optical-rx-power-3", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipOpticalVcc = prometheus.NewDesc("npu_chip_optical_vcc",
		"the npu interface receive optical-vcc", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipOpticalTemp = prometheus.NewDesc("npu_chip_optical_temp",
		"the npu interface receive optical-temperature", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuChipInfoDescDevProcessInfo = prometheus.NewDesc("npu_chip_info_process_info",
		"the npu process info, unit is 'MB'. if process run on host, container_id and container_name will be empty",
		[]string{npuID, modelName, npuUUID, "process_id", "container_id", "container_name", npuPCIEInfo}, nil)
	npuChipInfoDescAICoreFreqInfo = prometheus.NewDesc("npu_chip_info_aicore_current_freq",
		"the npu ai core current frequency, unit is 'MHz'", []string{npuID, modelName, npuUUID, npuPCIEInfo}, nil)
	npuContainerInfo = prometheus.NewDesc("npu_container_info",
		"the container name and deviceID relationship", []string{"containerID", "containerName", "npuID", modelName, npuUUID,
			npuPCIEInfo}, nil)
	npuContainerTotalMemory = prometheus.NewDesc("container_npu_total_memory",
		"the npu total memory in container, unit is 'MB'", []string{npuID, namespace, podName, "container_name",
			modelName, npuUUID, npuPCIEInfo}, nil)
	npuContainerUsedMemory = prometheus.NewDesc("container_npu_used_memory",
		"the npu used memory in container, unit is 'MB'", []string{npuID, namespace, podName, "container_name",
			modelName, npuUUID, npuPCIEInfo}, nil)
	npuContainerUtilization = prometheus.NewDesc("container_npu_utilization",
		"the npu ai core utilization in container, unit is '%'", []string{npuID, namespace, podName,
			"container_name", modelName, npuUUID, npuPCIEInfo}, nil)
	podAiCoreUtilizationRate = prometheus.NewDesc("vnpu_pod_aicore_utilization",
		"the vnpu aicore utilization rate, unit is '%'",
		[]string{npuID, modelName, vNpuUUID, "aicore_count", namespace, podName, "container_name", isVirtual}, nil)
	podTotalMemory = prometheus.NewDesc("vnpu_pod_total_memory", "the vnpu total memory on pod, unit is 'KB'",
		[]string{npuID, modelName, vNpuUUID, "aicore_count", namespace, podName, "container_name", isVirtual}, nil)
	podUsedMemory = prometheus.NewDesc("vnpu_pod_used_memory", "the vnpu used memory on pod, unit is 'KB'",
		[]string{npuID, modelName, vNpuUUID, "aicore_count", namespace, podName, "container_name", isVirtual}, nil)
	npuContainerInfoInit sync.Once
	npuChipInfoInit      sync.Once
)

var netInfoMap sync.Map

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
	decimalPlaces  = 2
	bitSize        = 64
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

func setNetInfoWithMap(phyID int32, netInfo NpuNetInfo) {
	netInfoMap.Store(phyID, netInfo)
}

func getNetInfoFromMap(oldNetInfo map[int32]NpuNetInfo) map[int32]NpuNetInfo {
	newNetInfo := oldNetInfo
	netInfoMap.Range(func(key, value interface{}) bool {
		phyID, ok := key.(int32)
		if !ok {
			hwlog.RunLog.Warnf("failed to get phyID of netInfo from map, which is: %v", key)
			return true
		}
		netInfo, ok := value.(NpuNetInfo)
		if !ok {
			hwlog.RunLog.Warnf("failed to get value of netInfo from map, which is: %v", value)
			return true
		}
		newNetInfo[phyID] = netInfo
		return true
	})

	return newNetInfo
}

func startToGetNetInfo(dmgr devmanager.DeviceInterface, updateTime time.Duration) {
	cardNum, cards, err := dmgr.GetCardList()
	if err != nil || cardNum == 0 {
		hwlog.RunLog.Errorf("failed to get npu info, error is: %v", err)
		return
	}

	for _, cardID := range cards {
		deviceNum, err := dmgr.GetDeviceNumInCard(cardID)
		if err != nil {
			hwlog.RunLog.Errorf("get device num of card: %v failed: %v", cardID, err)
			continue
		}
		for i := int32(0); i < deviceNum; i++ {
			logicID, err := dmgr.GetDeviceLogicID(cardID, i)
			if err != nil {
				hwlog.RunLog.Errorf("get logic ID of card: %v device:%v failed: %v", cardID, i, err)
				continue
			}

			phyID, err := dmgr.GetPhysicIDFromLogicID(logicID)
			if err != nil {
				hwlog.RunLog.Errorf("failed to get phy id when assemble net info: %v", err)
				continue
			}
			go assembleNPUNetInfo(phyID, dmgr, updateTime)
		}
	}
}

func getNPUInfo(dmgr devmanager.DeviceInterface) []HuaWeiNPUCard {
	var npuList []HuaWeiNPUCard
	cardNum, cards, err := dmgr.GetCardList()
	if err != nil || cardNum == 0 {
		hwlog.RunLog.Errorf("failed to get npu info, error is: %v", err)
		return npuList
	}

	for _, cardID := range cards {
		deviceNum, err := dmgr.GetDeviceNumInCard(cardID)
		if err != nil {
			hwlog.RunLog.Errorf("get device num of card %v failed: %v", cardID, err)
			continue
		}
		var deviceList []*HuaWeiAIChip
		for i := int32(0); i < deviceNum; i++ {
			var chipInfo *HuaWeiAIChip
			logicID, err := dmgr.GetDeviceLogicID(cardID, i)
			if err != nil {
				hwlog.RunLog.Errorf("get logic ID of card %v device %v failed: %v", cardID, i, err)
				continue
			}
			chipInfo = assembleNPUInfo(cardID, logicID, dmgr)
			if chipInfo == nil {
				continue
			}
			if !strings.Contains(chipInfo.ChipIfo.Name, "310P") || chipInfo.VDevInfos.TotalResource.VDevNum == 0 {
				deviceList = append(deviceList, chipInfo)
				continue
			}
			deviceList = append(deviceList, getVNPUInfo(*chipInfo)...)
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

func assembleNPUNetInfo(phyID int32, dmgr devmanager.DeviceInterface, updateTime time.Duration) {
	if !dmgr.IsTrainingCard() {
		return
	}
	for {
		setNetInfoWithMap(phyID, networkPackInfo(phyID))
		time.Sleep(updateTime)
	}
}

func assembleNPUInfo(cardID int32, logicID int32, dmgr devmanager.DeviceInterface) *HuaWeiAIChip {
	phyID, err := dmgr.GetPhysicIDFromLogicID(logicID)
	// check cardId, convert it to int type later
	if err != nil {
		hwlog.RunLog.Errorf("failed to get phy id when assemble npu info: %v", err)
		return nil
	}
	chipInfo := packChipInfo(logicID, dmgr)
	chipInfo.DeviceID = int(phyID)

	if dmgr.GetDevType() == common.Ascend310P {
		cardPower, err := dmgr.GetMcuPowerInfo(cardID)
		if err != nil {
			hwlog.RunLog.Error(err)
			cardPower = float32(common.InvalidVal)
		}
		// Ascend310P use cardPower to replace chipPower
		chipInfo.Power = cardPower
		vDevInfos, err := dmgr.GetVirtualDeviceInfo(logicID)
		if err != nil || vDevInfos.TotalResource.VDevNum == 0 {
			return chipInfo
		}
		chipInfo.VDevInfos = vDevInfos
	}
	return chipInfo
}

func getVNPUInfo(chipInfo HuaWeiAIChip) []*HuaWeiAIChip {
	var aiChips []*HuaWeiAIChip
	for _, activityVDev := range chipInfo.VDevInfos.VDevActivityInfo {
		vDevInfo := chipInfo
		vDevInfo.VDevActivityInfo = activityVDev
		aiChips = append(aiChips, &vDevInfo)
	}
	return aiChips
}

func start(ctx context.Context, n *npuCollector, dmgr devmanager.DeviceInterface) {
	defer func() {
		if err := dmgr.ShutDown(); err != nil {
			hwlog.RunLog.Error(err)
		}
		if err := recover(); err != nil {
			hwlog.RunLog.Errorf("go routine failed with %v", err)
		}
	}()
	if n == nil {
		hwlog.RunLog.Error("Invalid param in function start")
		return
	}
	if err := n.devicesParser.Init(); err != nil {
		hwlog.RunLog.Errorf("failed to init devices parser: %v", err)
	}
	defer n.devicesParser.Close()
	n.devicesParser.Timeout = n.updateTime
	hwlog.RunLog.Infof("Starting update cache every %d seconds", n.updateTime/time.Second)

	group := &sync.WaitGroup{}

	npuBaseInfoCollect(group, n, dmgr)
	npuNetworkInfoCollect(group, n, dmgr)
	containerInfoCollect(group, n)

	group.Wait()
	hwlog.RunLog.Info("received the stop signal,STOPPED")
	return
}

func npuBaseInfoCollect(group *sync.WaitGroup, n *npuCollector, dmgr devmanager.DeviceInterface) {
	group.Add(1)
	go func() {
		defer group.Done()
		ticker := time.NewTicker(n.updateTime)
		defer ticker.Stop()
		for {
			npuInfo := getNPUInfo(dmgr)
			if err := n.cache.Set(npuListCacheKey, npuInfo, n.cacheTime); err != nil {
				hwlog.RunLog.Error(err)
			} else {
				hwlog.RunLog.Infof("update cache,key is %s", npuListCacheKey)
			}
			if _, ok := <-ticker.C; !ok {
				hwlog.RunLog.Errorf("%s ticker failed, task shutdown", npuListCacheKey)
				return
			}
		}
	}()
}

func npuNetworkInfoCollect(group *sync.WaitGroup, n *npuCollector, dmgr devmanager.DeviceInterface) {
	group.Add(1)
	netInfo := make(map[int32]NpuNetInfo, initSize)
	startToGetNetInfo(dmgr, n.updateTime)
	go func() {
		defer group.Done()
		ticker := time.NewTicker(n.updateTime)
		defer ticker.Stop()
		for {
			obj, err := n.cache.Get(npuNetworkCacheKey)
			if err != nil {
				hwlog.RunLog.Warnf("get info of %s failed: %v, so use initial net info", npuNetworkCacheKey, err)
			} else {
				if oldNetWorkInfo, ok := obj.(map[int32]NpuNetInfo); ok {
					netInfo = oldNetWorkInfo
				} else {
					hwlog.RunLog.Warn("format of net info in cache is not right")
				}
			}
			// get current net info from map to update cache
			newNetInfo := getNetInfoFromMap(netInfo)
			if err := n.cache.Set(npuNetworkCacheKey, newNetInfo, n.cacheTime); err != nil {
				hwlog.RunLog.Error(err)
			} else {
				hwlog.RunLog.Infof("update cache,key is %s", npuNetworkCacheKey)
			}
			if _, ok := <-ticker.C; !ok {
				hwlog.RunLog.Errorf("%s ticker failed, task shutdown", npuNetworkCacheKey)
				return
			}
		}
	}()
}

func containerInfoCollect(group *sync.WaitGroup, n *npuCollector) {
	group.Add(1)
	go func() {
		defer group.Done()
		ticker := time.NewTicker(n.updateTime)
		defer ticker.Stop()
		for {
			n.devicesParser.FetchAndParse(nil)
			select {
			case result := <-n.devicesParser.RecvResult():
				if err := n.cache.Set(containersDevicesCacheKey, result, n.cacheTime); err != nil {
					hwlog.RunLog.Error(err)
				}
				hwlog.RunLog.Infof("update cache,key is %s", containersDevicesCacheKey)
			case err := <-n.devicesParser.RecvErr():
				hwlog.RunLog.Errorf("received error from device parser: %v", err)
			}
			if _, ok := <-ticker.C; !ok {
				hwlog.RunLog.Errorf("%s ticker failed, task shutdown", containersDevicesCacheKey)
				return
			}
		}
	}()
}

func describeBaseChipInfo(ch chan<- *prometheus.Desc) {
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

func describeOpticalInfo(ch chan<- *prometheus.Desc) {
	ch <- npuChipOpticalState
	ch <- npuChipOpticalTxPower0
	ch <- npuChipOpticalTxPower1
	ch <- npuChipOpticalTxPower2
	ch <- npuChipOpticalTxPower3
	ch <- npuChipOpticalRxPower0
	ch <- npuChipOpticalRxPower1
	ch <- npuChipOpticalRxPower2
	ch <- npuChipOpticalRxPower3
	ch <- npuChipOpticalVcc
	ch <- npuChipOpticalTemp
}

func describeRoCEInfo(ch chan<- *prometheus.Desc) {
	ch <- npuChipInfoDescNetworkStatus
	ch <- npuChipInfoDescBandwidthTx
	ch <- npuChipInfoDescBandwidthRx
	ch <- npuChipInfoDescLinkStatus
	ch <- npuChipLinkSpeed
	ch <- npuChipLinkUpNum
	ch <- npuChipMacRxPauseNum
	ch <- npuChipMacTxPauseNum
	ch <- npuChipMacRxPfcPktNum
	ch <- npuChipMacTxPfcPktNum
	ch <- npuChipMacRxBadPktNum
	ch <- npuChipMacTxBadPktNum
	ch <- npuChipRoceRxAllPktNum
	ch <- npuChipRoceTxAllPktNum
	ch <- npuChipRoceRxErrPktNum
	ch <- npuChipRoceTxErrPktNum
	ch <- npuChipRoceRxCnpPktNum
	ch <- npuChipRoceTxCnpPktNum
	ch <- npuChipRoceNewPktRtyNum
	ch <- npuChipMacTxBadOctNum
	ch <- npuChipMacRxBadOctNum
	ch <- npuChipRoceUnexpectedAcktNum
	ch <- npuChipRoceOutOfOrderNum
	ch <- npuChipRoceVerificationErrNum
	ch <- npuChipRoceQpStatusErrNum
}

// Describe implements prometheus.Collector
func (n *npuCollector) Describe(ch chan<- *prometheus.Desc) {
	if ch == nil {
		hwlog.RunLog.Error("Invalid param in function Describe")
		return
	}
	describeBaseChipInfo(ch)
	describeOpticalInfo(ch)
	describeRoCEInfo(ch)
	ch <- npuContainerInfo
	ch <- npuContainerTotalMemory
	ch <- npuContainerUsedMemory
	ch <- npuContainerUtilization
	ch <- npuChipInfoDescDevProcessInfo
	ch <- npuChipInfoDescAICoreFreqInfo
	ch <- podAiCoreUtilizationRate
	ch <- podTotalMemory
	ch <- podUsedMemory
}

// Collect implements prometheus.Collector
func (n *npuCollector) Collect(ch chan<- prometheus.Metric) {
	if !validate(ch) {
		hwlog.RunLog.Error("Invalid param in function Collect")
		return
	}
	npuList := getNPUInfoInCache(ch, n)
	networkInfoMap := getNetworkInfoInCache(ch, n)
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
			deviceID := chip.DeviceID
			if devNetWorkInfo, ok := networkInfoMap[int32(deviceID)]; ok {
				chip.NetInfo = &devNetWorkInfo
			} else {
				hwlog.RunLog.Warn("no network information at the moment, so use initial info")
				chip.NetInfo = &NpuNetInfo{}
			}

			if chip.VDevActivityInfo.IsVirtualDev {
				deviceID = int(chip.VDevActivityInfo.VDevID)
			}
			devInfo, ok := containerMap[deviceID]
			if !ok {
				devInfo = container.DevicesInfo{}
			}
			updateNPUCommonInfo(ch, &card, chip)
			updateNPUMemoryInfo(ch, &card, chip)
			updateNPUNetworkInfo(ch, &card, chip)
			updateProcessInfo(ch, &card, chip, devInfo)
			updateContainerInfo(ch, &card, chip, devInfo)
			updatePodVNPUInfo(ch, &card, chip, devInfo)
		}
	}

	ch <- prometheus.MustNewConstMetric(machineInfoNPUDesc, prometheus.GaugeValue, float64(totalCount))
}

func getNPUInfoInCache(ch chan<- prometheus.Metric, n *npuCollector) []HuaWeiNPUCard {
	if ch == nil {
		hwlog.RunLog.Error("metric channel is nil")
		return nil
	}
	obj, err := n.cache.Get(npuListCacheKey)
	npuChipInfoInit.Do(func() {
		if err != nil {
			hwlog.RunLog.Debugf("no cache, start to get npulist and rebuild cache")
			devManager, err := devmanager.GetDeviceManager()
			if err != nil {
				hwlog.RunLog.Debugf("get device manager failed, error is: %v ", err)
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
		return nil
	}

	return npuList
}

func getNetworkInfoInCache(ch chan<- prometheus.Metric, n *npuCollector) map[int32]NpuNetInfo {
	res := make(map[int32]NpuNetInfo, initSize)
	if ch == nil {
		hwlog.RunLog.Error("metric channel is nil")
		return res
	}
	obj, err := n.cache.Get(npuNetworkCacheKey)
	if err != nil {
		hwlog.RunLog.Warn("npu network info not found in cache, please wait for the cache to be rebuilt")
		return res
	}
	networkInfoList, ok := obj.(map[int32]NpuNetInfo)
	if !ok {
		hwlog.RunLog.Error("Error npu network info cache and convert failed")
		n.cache.Delete(npuNetworkCacheKey)
		return res
	}

	return networkInfoList
}

func getContainerNPUInfo(ch chan<- prometheus.Metric, n *npuCollector) map[int]container.DevicesInfo {
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
				hwlog.RunLog.Warn("rebuild container info cache timeout")
				return
			}
			hwlog.RunLog.Warn("rebuild cache successfully")
		}
	})
	cntNpuInfos, ok := obj.(container.DevicesInfos)
	if !ok {
		hwlog.RunLog.Error("Error container npu info cache and convert failed")
		n.cache.Delete(containersDevicesCacheKey)
		return nil
	}
	res := make(map[int]container.DevicesInfo, initSize)
	for _, v := range cntNpuInfos {
		for _, deviceID := range v.Devices {
			res[deviceID] = v
		}
	}
	return res
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

func getContainerNameArray(devInfo container.DevicesInfo) []string {
	if devInfo.Name == "" {
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
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescHbmTotalMemory, prometheus.GaugeValue,
			float64(chip.HbmInfo.MemorySize), []string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID,
				chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuChipInfoDescUsedMemory,
		prometheus.GaugeValue, float64(chip.Meminf.MemorySize-chip.Meminf.MemoryAvailable),
		[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescTotalMemory, prometheus.GaugeValue,
			float64(chip.Meminf.MemorySize), []string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID,
				chip.PCIeBusInfo}...))
}

func updateStatInfoOfMac(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip) {
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipMacRxPauseNum, prometheus.GaugeValue, chip.NetInfo.StatInfo.MacRxPauseNum,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipMacTxPauseNum, prometheus.GaugeValue, chip.NetInfo.StatInfo.MacTxPauseNum,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipMacRxPfcPktNum, prometheus.GaugeValue, chip.NetInfo.StatInfo.MacRxPfcPktNum,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipMacTxPfcPktNum, prometheus.GaugeValue, chip.NetInfo.StatInfo.MacTxPfcPktNum,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipMacRxBadPktNum, prometheus.GaugeValue, chip.NetInfo.StatInfo.MacRxBadPktNum,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipMacTxBadPktNum, prometheus.GaugeValue, chip.NetInfo.StatInfo.MacTxBadPktNum,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipMacTxBadOctNum, prometheus.GaugeValue, chip.NetInfo.StatInfo.MacTxBadOctNum,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipMacRxBadOctNum, prometheus.GaugeValue, chip.NetInfo.StatInfo.MacRxBadOctNum,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))

}

func updateStatInfoOfRoCE(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip) {
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipRoceRxAllPktNum, prometheus.GaugeValue, chip.NetInfo.StatInfo.RoceRxAllPktNum,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipRoceTxAllPktNum, prometheus.GaugeValue, chip.NetInfo.StatInfo.RoceTxAllPktNum,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipRoceRxErrPktNum, prometheus.GaugeValue, chip.NetInfo.StatInfo.RoceRxErrPktNum,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipRoceTxErrPktNum, prometheus.GaugeValue, chip.NetInfo.StatInfo.RoceTxErrPktNum,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipRoceRxCnpPktNum, prometheus.GaugeValue, chip.NetInfo.StatInfo.RoceRxCnpPktNum,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipRoceTxCnpPktNum, prometheus.GaugeValue, chip.NetInfo.StatInfo.RoceTxCnpPktNum,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipRoceNewPktRtyNum, prometheus.GaugeValue, chip.NetInfo.StatInfo.RoceNewPktRtyNum,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipRoceUnexpectedAcktNum, prometheus.GaugeValue, chip.NetInfo.StatInfo.RoceUnexpectedAckNum,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipRoceOutOfOrderNum, prometheus.GaugeValue, chip.NetInfo.StatInfo.RoceOutOfOrderNum,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipRoceVerificationErrNum, prometheus.GaugeValue, chip.NetInfo.StatInfo.RoceVerificationErrNum,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipRoceQpStatusErrNum, prometheus.GaugeValue, chip.NetInfo.StatInfo.RoceQpStatusErrNum,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
}

func updateOpticalInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip) {
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipOpticalState, prometheus.GaugeValue, chip.NetInfo.OpticalInfo.OpticalState,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipOpticalTxPower0, prometheus.GaugeValue, chip.NetInfo.OpticalInfo.OpticalTxPower0,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipOpticalTxPower1, prometheus.GaugeValue, chip.NetInfo.OpticalInfo.OpticalTxPower1,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipOpticalTxPower2, prometheus.GaugeValue, chip.NetInfo.OpticalInfo.OpticalTxPower2,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipOpticalTxPower3, prometheus.GaugeValue, chip.NetInfo.OpticalInfo.OpticalTxPower3,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipOpticalRxPower0, prometheus.GaugeValue, chip.NetInfo.OpticalInfo.OpticalRxPower0,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipOpticalRxPower1, prometheus.GaugeValue, chip.NetInfo.OpticalInfo.OpticalRxPower1,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipOpticalRxPower2, prometheus.GaugeValue, chip.NetInfo.OpticalInfo.OpticalRxPower2,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipOpticalRxPower3, prometheus.GaugeValue, chip.NetInfo.OpticalInfo.OpticalRxPower3,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipOpticalVcc, prometheus.GaugeValue, chip.NetInfo.OpticalInfo.OpticalVcc,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipOpticalTemp, prometheus.GaugeValue, chip.NetInfo.OpticalInfo.OpticalTemp,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
}

func updateNPUNetworkInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip) {
	if !validate(ch, npu, chip) {
		hwlog.RunLog.Error("Invalid param in function updateNPUNetworkInfo")
		return
	}
	updateStatInfoOfMac(ch, npu, chip)
	updateStatInfoOfRoCE(ch, npu, chip)
	updateOpticalInfo(ch, npu, chip)
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescBandwidthTx, prometheus.GaugeValue, chip.NetInfo.BandwidthInfo.TxValue,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescBandwidthRx, prometheus.GaugeValue, chip.NetInfo.BandwidthInfo.RxValue,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescNetworkStatus, prometheus.GaugeValue,
			float64(getHealthCode(chip.NetHealthStatus)), []string{strconv.FormatInt(int64(chip.DeviceID), base),
				common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipLinkSpeed, prometheus.GaugeValue, chip.NetInfo.LinkSpeedInfo.Speed,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipLinkUpNum, prometheus.GaugeValue, chip.NetInfo.LinkStatInfo.LinkUPNum,
			[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
}

func updateContainerInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip,
	devInfo container.DevicesInfo) {
	containerName := getContainerNameArray(devInfo)
	if len(containerName) != containerNameLen {
		return
	}
	ch <- prometheus.MustNewConstMetric(npuContainerInfo, prometheus.GaugeValue, 1,
		[]string{devInfo.ID, strings.Join(containerName, "_"), strconv.Itoa(chip.DeviceID), common.GetNpuName(*chip.ChipIfo), chip.VDieID,
			chip.PCIeBusInfo}...)
	if common.IsValidVDevID(chip.VDevActivityInfo.VDevID) {
		return
	}
	updateContainerNPUMemoryInfo(ch, npu, chip, containerName)
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuContainerUtilization,
		prometheus.GaugeValue, float64(chip.Utilization), []string{strconv.FormatInt(int64(chip.DeviceID), base),
			containerName[nameSpaceIdx], containerName[podNameIdx], containerName[conNameIdx], common.GetNpuName(*chip.ChipIfo), chip.VDieID,
			chip.PCIeBusInfo}...))

}

func updatePodVNPUInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip,
	devInfo container.DevicesInfo) {
	if !strings.Contains(chip.ChipIfo.Name, "310P") || !common.IsValidVDevID(chip.VDevActivityInfo.VDevID) {
		return
	}
	containerName := getContainerNameArray(devInfo)
	if len(containerName) != containerNameLen {
		return
	}
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(podAiCoreUtilizationRate, prometheus.GaugeValue,
			float64(chip.VDevActivityInfo.VDevAiCoreRate), getPodDisplayInfo(chip, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(podTotalMemory, prometheus.GaugeValue,
			float64(chip.VDevActivityInfo.VDevTotalMem), getPodDisplayInfo(chip, containerName)...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(podUsedMemory, prometheus.GaugeValue,
			float64(chip.VDevActivityInfo.VDevUsedMem), getPodDisplayInfo(chip, containerName)...))
}

func updateContainerNPUMemoryInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip,
	containerName []string) {
	if strings.Contains(chip.ChipIfo.Name, common.Chip910) {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuContainerTotalMemory, prometheus.GaugeValue,
				float64(chip.HbmInfo.MemorySize), []string{strconv.FormatInt(int64(chip.DeviceID), base),
					containerName[nameSpaceIdx], containerName[podNameIdx], containerName[conNameIdx],
					common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuContainerUsedMemory, prometheus.GaugeValue, float64(chip.HbmInfo.Usage),
				[]string{strconv.FormatInt(int64(chip.DeviceID), base), containerName[nameSpaceIdx],
					containerName[podNameIdx], containerName[conNameIdx],
					common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
		return
	}
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuContainerTotalMemory,
		prometheus.GaugeValue, float64(chip.Meminf.MemorySize), []string{strconv.FormatInt(int64(chip.DeviceID), base),
			containerName[nameSpaceIdx], containerName[podNameIdx], containerName[conNameIdx], common.GetNpuName(*chip.ChipIfo), chip.VDieID,
			chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuContainerUsedMemory,
		prometheus.GaugeValue, float64(chip.Meminf.MemorySize-chip.Meminf.MemoryAvailable),
		[]string{strconv.FormatInt(int64(chip.DeviceID), base), containerName[nameSpaceIdx],
			containerName[podNameIdx], containerName[conNameIdx], common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
}

func updateNPUCommonInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip) {
	if !validate(ch, npu, chip, chip.ChipIfo) {
		hwlog.RunLog.Error("Invalid param in function updateNpuCommonInfo")
		return
	}
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuChipInfoDescLinkStatus,
		prometheus.GaugeValue, float64(hccn.GetLinkStatusCode(chip.LinkStatus)),
		[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuChipInfoDescUtil,
		prometheus.GaugeValue, float64(chip.Utilization), []string{strconv.FormatInt(int64(chip.DeviceID), base),
			common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuChipInfoDescTemp,
		prometheus.GaugeValue, float64(chip.Temperature), []string{strconv.FormatInt(int64(chip.DeviceID), base),
			common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuChipInfoDescPower,
		prometheus.GaugeValue, float64(chip.Power), []string{strconv.FormatInt(int64(chip.DeviceID), base),
			common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuChipInfoDescVoltage,
		prometheus.GaugeValue, float64(chip.Voltage), []string{strconv.FormatInt(int64(chip.DeviceID), base),
			common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
		prometheus.MustNewConstMetric(npuChipInfoDescHealthStatus, prometheus.GaugeValue,
			float64(getHealthCode(chip.HealthStatus)), []string{strconv.FormatInt(int64(chip.DeviceID), base),
				common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuChipInfoDescErrorCode,
		prometheus.GaugeValue, float64(chip.ErrorCode), []string{strconv.FormatInt(int64(chip.DeviceID), base),
			common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuChipInfoDescNpuName,
		prometheus.GaugeValue, 1, []string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID,
			chip.PCIeBusInfo}...))
	ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp, prometheus.MustNewConstMetric(npuChipInfoDescAICoreFreqInfo,
		prometheus.GaugeValue, float64(chip.AICoreCurrentFreq), []string{strconv.FormatInt(int64(chip.DeviceID), base),
			common.GetNpuName(*chip.ChipIfo), chip.VDieID, chip.PCIeBusInfo}...))
}

func updateProcessInfo(ch chan<- prometheus.Metric, npu *HuaWeiNPUCard, chip *HuaWeiAIChip,
	devInfo container.DevicesInfo) {
	containerName := ""
	containerID := ""
	cNameArray := getContainerNameArray(devInfo)
	if len(cNameArray) == containerNameLen {
		containerName = strings.Join(cNameArray, "_")
		containerID = devInfo.ID
	}
	if chip.DevProcessInfo.ProcNum == 0 {
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuChipInfoDescDevProcessInfo, prometheus.GaugeValue, 0,
				[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo),
					chip.VDieID, "", containerID, containerName, chip.PCIeBusInfo}...))
		return
	}
	for i := int32(0); i < chip.DevProcessInfo.ProcNum; i++ {
		procInfo := chip.DevProcessInfo.DevProcArray[i]
		ch <- prometheus.NewMetricWithTimestamp(npu.Timestamp,
			prometheus.MustNewConstMetric(npuChipInfoDescDevProcessInfo, prometheus.GaugeValue, procInfo.MemUsage,
				[]string{strconv.FormatInt(int64(chip.DeviceID), base), common.GetNpuName(*chip.ChipIfo), chip.VDieID,
					strconv.FormatInt(int64(procInfo.Pid), base), containerID, containerName, chip.PCIeBusInfo}...))
	}
}

var packChipInfo = func(logicID int32, dmgr devmanager.DeviceInterface) *HuaWeiAIChip {
	chip := &HuaWeiAIChip{}

	info, err := dmgr.GetChipInfo(logicID)
	if err != nil {
		hwlog.RunLog.Warnf("get chip info failed: %v", err)
		info = &common.ChipInfo{}
	}
	chip.ChipIfo = info

	packChipInfoPart2(logicID, dmgr, chip)
	packChipInfoPart1(logicID, dmgr, chip)
	return chip
}

func packChipInfoPart1(logicID int32, dmgr devmanager.DeviceInterface, hwChip *HuaWeiAIChip) {
	freq, err := dmgr.GetDeviceFrequency(logicID, common.AICoreCurrentFreq)
	if err != nil {
		freq = common.InvalidVal
	}
	power, err := dmgr.GetDevicePowerInfo(logicID)
	if err != nil {
		power = common.InvalidVal
	}
	temp, err := dmgr.GetDeviceTemperature(logicID)
	if err != nil {
		temp = common.InvalidVal
	}
	vol, err := dmgr.GetDeviceVoltage(logicID)
	if err != nil {
		vol = common.InvalidVal
	}
	mem, err := dmgr.GetDeviceMemoryInfo(logicID)
	if err != nil {
		mem = &common.MemoryInfo{}
	}
	hbmInfo, err := dmgr.GetDeviceHbmInfo(logicID)
	if err != nil {
		hbmInfo = &common.HbmInfo{}
	}

	hwChip.AICoreCurrentFreq = freq
	hwChip.Power = power
	hwChip.HealthStatus = getHealth(logicID, dmgr)
	hwChip.Temperature = int(temp)
	hwChip.Voltage = vol
	hwChip.Meminf = mem
	hwChip.HbmInfo = hbmInfo
}

func packChipInfoPart2(logicID int32, dmgr devmanager.DeviceInterface, hwChip *HuaWeiAIChip) {
	util, err := dmgr.GetDeviceUtilizationRate(logicID, common.AICore)
	if err != nil {
		util = common.InvalidVal // valid data range 0-100
	}
	_, errCode, err := dmgr.GetDeviceErrorCode(logicID)
	if err != nil {
		errCode = common.RetError // valid data range 0-128
	}
	vdieID, err := dmgr.GetDieID(logicID, dcmi.VDIE)
	if err != nil {
		hwlog.RunLog.Debug(err)
	}

	setNetHealthStatus(logicID, dmgr, hwChip)
	setProcessInfo(logicID, dmgr, hwChip)
	setPCIeBusInfo(logicID, dmgr, hwChip)
	setLinkStatus(logicID, dmgr, hwChip)
	hwChip.ErrorCode = errCode
	hwChip.Utilization = int(util)
	hwChip.VDieID = vdieID
}

func setNetHealthStatus(logicID int32, dmgr devmanager.DeviceInterface, hwChip *HuaWeiAIChip) {
	hwChip.NetHealthStatus = UnHealthy
	if !dmgr.IsTrainingCard() {
		return
	}

	netCode, err := dmgr.GetDeviceNetWorkHealth(logicID)
	hwlog.RunLog.Debugf("chip %d network healthy code is %d", logicID, netCode)
	if err != nil {
		netCode = math.MaxUint32
	}
	hwChip.NetHealthStatus = getNetworkHealthy(netCode)
}

func setProcessInfo(logicID int32, dmgr devmanager.DeviceInterface, hwChip *HuaWeiAIChip) {
	productTypes := dmgr.GetProductTypeArray()
	info, err := dmgr.GetDevProcessInfo(logicID)
	if err != nil {
		if len(productTypes) == 1 && productTypes[0] == common.Atlas200ISoc {
			hwlog.RunLog.Debugf("process info is not supported on %s", common.Atlas200ISoc)
			hwChip.DevProcessInfo = new(common.DevProcessInfo)
			return
		}
		hwlog.RunLog.Error(err)
		info = new(common.DevProcessInfo)
	}
	hwChip.DevProcessInfo = info
}

func setPCIeBusInfo(logicID int32, dmgr devmanager.DeviceInterface, hwChip *HuaWeiAIChip) {
	productTypes := dmgr.GetProductTypeArray()
	pcieInfo, err := dmgr.GetPCIeBusInfo(logicID)
	if err != nil {
		if len(productTypes) == 1 && productTypes[0] == common.Atlas200ISoc {
			hwlog.RunLog.Debugf("pcie bus info is not supported on %s", common.Atlas200ISoc)
			hwChip.PCIeBusInfo = ""
			return
		}
		hwlog.RunLog.Error(err)
		pcieInfo = ""
	}
	hwChip.PCIeBusInfo = pcieInfo
}

func setLinkStatus(logicID int32, dmgr devmanager.DeviceInterface, hwChip *HuaWeiAIChip) {
	hwChip.LinkStatus = LinkDown
	if !dmgr.IsTrainingCard() {
		return
	}

	phyID, err := dmgr.GetPhysicIDFromLogicID(logicID)
	if err != nil {
		hwlog.RunLog.Error("set link status failed")
		return
	}
	hwChip.LinkStatus = hccn.GetNPULinkStatus(phyID)
}

func getMainOptInfo(opticalInfo map[string]string) OpticalInfo {
	mainOpticalInfo := OpticalInfo{}
	mainOpticalInfo.OpticalTxPower0 = hccn.GetFloatDataFromStr(opticalInfo[txPower0])
	mainOpticalInfo.OpticalTxPower1 = hccn.GetFloatDataFromStr(opticalInfo[txPower1])
	mainOpticalInfo.OpticalTxPower2 = hccn.GetFloatDataFromStr(opticalInfo[txPower2])
	mainOpticalInfo.OpticalTxPower3 = hccn.GetFloatDataFromStr(opticalInfo[txPower3])
	mainOpticalInfo.OpticalRxPower0 = hccn.GetFloatDataFromStr(opticalInfo[rxPower0])
	mainOpticalInfo.OpticalRxPower1 = hccn.GetFloatDataFromStr(opticalInfo[rxPower1])
	mainOpticalInfo.OpticalRxPower2 = hccn.GetFloatDataFromStr(opticalInfo[rxPower2])
	mainOpticalInfo.OpticalRxPower3 = hccn.GetFloatDataFromStr(opticalInfo[rxPower3])
	mainOpticalInfo.OpticalVcc = hccn.GetFloatDataFromStr(opticalInfo[voltage])
	mainOpticalInfo.OpticalTemp = hccn.GetFloatDataFromStr(opticalInfo[temperature])

	optState := 0.0
	if opticalInfo[present] == present {
		optState = 1.0
	}
	mainOpticalInfo.OpticalState = optState

	return mainOpticalInfo
}

func getMainStatInfo(statInfo map[string]int) StatInfo {
	mainStatInfo := StatInfo{}
	mainStatInfo.MacRxPauseNum = float64(statInfo[macRxMacPauseNum])
	mainStatInfo.MacTxPauseNum = float64(statInfo[macTxMacPauseNum])
	mainStatInfo.MacRxPfcPktNum = float64(statInfo[macRxPfcPktNum])
	mainStatInfo.MacTxPfcPktNum = float64(statInfo[macTxPfcPktNum])
	mainStatInfo.MacRxBadPktNum = float64(statInfo[macRxBadPktNum])
	mainStatInfo.MacTxBadPktNum = float64(statInfo[macTxBadPktNum])
	mainStatInfo.RoceRxAllPktNum = float64(statInfo[roCERxAllPktNum])
	mainStatInfo.RoceTxAllPktNum = float64(statInfo[roCETxAllPktNum])
	mainStatInfo.RoceRxErrPktNum = float64(statInfo[roCERxErrPktNum])
	mainStatInfo.RoceTxErrPktNum = float64(statInfo[roCETxErrPktNum])
	mainStatInfo.RoceRxCnpPktNum = float64(statInfo[roCERxCnpPktNum])
	mainStatInfo.RoceTxCnpPktNum = float64(statInfo[roCETxCnpPktNum])
	mainStatInfo.MacRxBadOctNum = float64(statInfo[macRxBadOctNum])
	mainStatInfo.MacTxBadOctNum = float64(statInfo[macTxBadOctNum])
	mainStatInfo.RoceUnexpectedAckNum = float64(statInfo[roCEUnexpectedAckNum])
	mainStatInfo.RoceOutOfOrderNum = float64(statInfo[roCEOutOfOrderNum])
	mainStatInfo.RoceVerificationErrNum = float64(statInfo[roCEVerificationErrNum])
	mainStatInfo.RoceQpStatusErrNum = float64(statInfo[roCEQpStatusErrNum])
	mainStatInfo.RoceNewPktRtyNum = float64(statInfo[roCENewPktRtyNum])

	return mainStatInfo
}

func networkPackInfo(phyID int32) NpuNetInfo {
	newNetInfo := NpuNetInfo{}
	if tx, rx, err := hccn.GetNPUInterfaceTraffic(phyID); err == nil {
		newNetInfo.BandwidthInfo.RxValue = rx
		newNetInfo.BandwidthInfo.TxValue = tx
	}
	if opticalInfo, err := hccn.GetNPUOpticalInfo(phyID); err == nil {
		newNetInfo.OpticalInfo = getMainOptInfo(opticalInfo)
	}

	if statInfo, err := hccn.GetNPUStatInfo(phyID); err == nil {
		newNetInfo.StatInfo = getMainStatInfo(statInfo)
	}

	linkUpNum := hccn.GetNPULinkUpNum(phyID)
	newNetInfo.LinkStatInfo.LinkUPNum = float64(linkUpNum)

	speed := hccn.GetNPULinkSpeed(phyID)
	newNetInfo.LinkSpeedInfo.Speed = float64(speed)
	return newNetInfo
}

func getHealth(logicID int32, dmgr devmanager.DeviceInterface) string {
	health, err := dmgr.GetDeviceHealth(logicID)
	if err != nil || health != 0 {
		return UnHealthy
	}
	return Healthy
}

func getHealthCode(health string) int {
	if Healthy == health {
		return 1
	}
	return 0
}

func getNetworkHealthy(netCode uint32) string {
	if netCode == common.NetworkInit || netCode == common.NetworkSuccess {
		return Healthy
	}

	return UnHealthy
}

func getPodDisplayInfo(chip *HuaWeiAIChip, containerName []string) []string {
	return []string{
		strconv.Itoa(chip.DeviceID),
		common.GetNpuName(*chip.ChipIfo),
		strconv.Itoa(int(chip.VDevActivityInfo.VDevID)),
		strconv.FormatFloat(chip.VDevActivityInfo.VDevAiCore, 'f', decimalPlaces, bitSize),
		containerName[nameSpaceIdx],
		containerName[podNameIdx],
		containerName[conNameIdx],
		strconv.FormatBool(chip.VDevActivityInfo.IsVirtualDev),
	}
}
