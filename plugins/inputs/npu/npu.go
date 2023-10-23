package npu

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	"huawei.com/npu-exporter/v5/devmanager"
	"huawei.com/npu-exporter/v5/devmanager/common"
	"huawei.com/npu-exporter/v5/devmanager/hccn"
)

const (
	defaultLogPath = "/var/log/mindx-dl/npu-exporter/npu-plugin.log"

	aiCore = common.DeviceType(2)
	hbm    = common.DeviceType(6)

	mega                = 1024 * 1024
	maxLogBackups       = 2
	defaultLogCacheSize = 2 * 1024
	defaultLogFileSize  = 2
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

//go:embed sample.conf
var sampleConfig string

type NpuWatch struct {
	NpuLogPath  string `toml:"npu_log_path"`
	NpuLogLevel int    `toml:"npu_log_level"`
	devManager  devmanager.DeviceInterface
}

func (*NpuWatch) SampleConfig() string {
	return sampleConfig
}

// Init is for setup, and validating config.
func (npu *NpuWatch) Init() error {
	if npu.NpuLogPath == "" {
		npu.NpuLogPath = defaultLogPath
	}
	var hwLogConfig = &hwlog.LogConfig{
		LogFileName: npu.NpuLogPath,
		ExpiredTime: hwlog.DefaultExpiredTime,
		CacheSize:   defaultLogCacheSize,
		FileMaxSize: defaultLogFileSize,
		LogLevel:    npu.NpuLogLevel,
		MaxAge:      hwlog.DefaultMinSaveAge,
		MaxBackups:  maxLogBackups}

	if err := hwlog.InitRunLogger(hwLogConfig, context.Background()); err != nil {
		fmt.Printf("hwlog init failed, error is %#v\n", err)
		return err
	}
	dmgr, err := devmanager.AutoInit("")
	if err != nil {
		return fmt.Errorf("init dev manager failed: %v", err)
	}
	npu.devManager = dmgr
	return nil
}

// parseOptInfoForCTYun parse optical info of NPU for CT Yun
func parseOptInfoForCTYun(opticalInfo map[string]string) map[string]interface{} {
	ctYunOpticalInfo := make(map[string]interface{})
	var ctYunFloatDataKeys = []string{
		txPower0,
		txPower1,
		txPower2,
		txPower3,
		rxPower0,
		rxPower1,
		rxPower2,
		rxPower3,
		voltage,
		temperature,
	}
	var ctYunTelegrafKeys = []string{
		"npu_chip_optical_tx_power_0",
		"npu_chip_optical_tx_power_1",
		"npu_chip_optical_tx_power_2",
		"npu_chip_optical_tx_power_3",
		"npu_chip_optical_rx_power_0",
		"npu_chip_optical_rx_power_1",
		"npu_chip_optical_rx_power_2",
		"npu_chip_optical_rx_power_3",
		"npu_chip_optical_vcc",
		"npu_chip_optical_temp",
	}

	for i, ctYunOpticalKey := range ctYunFloatDataKeys {
		floatData := hccn.GetFloatDataFromStr(opticalInfo[ctYunOpticalKey])
		ctYunOpticalInfo[ctYunTelegrafKeys[i]] = floatData
	}

	optState := 0
	if opticalInfo[present] == present {
		optState = 1
	}
	ctYunOpticalInfo["npu_chip_optical_state"] = optState

	return ctYunOpticalInfo
}

func (npu *NpuWatch) packDcmiInfo(devID int32, fields map[string]interface{}, acc telegraf.Accumulator) {
	health, err := npu.devManager.GetDeviceHealth(devID)
	if err != nil {
		acc.AddError(fmt.Errorf("get health of npu failed: %v", err))
	}
	fields["npu_chip_info_health_status"] = hccn.GetHealthCode(health)

	netCode, err := npu.devManager.GetDeviceNetWorkHealth(devID)
	if err != nil {
		acc.AddError(fmt.Errorf("get npu Net health failed: %v", err))
	}
	fields["npu_chip_info_network_status"] = hccn.GetNetworkHealthy(netCode)

	info, err := npu.devManager.GetDevProcessInfo(devID)
	if err != nil {
		acc.AddError(fmt.Errorf("get npu process info failed: %v", err))
	} else {
		fields["npu_chip_info_process_info_num"] = info.ProcNum
	}

	temp, err := npu.devManager.GetDeviceTemperature(devID)
	if err != nil {
		acc.AddError(fmt.Errorf("get npu temperature failed: %v", err))
	}
	fields["npu_chip_info_temperature"] = float64(temp)

	aiCoreUtil, err := npu.devManager.GetDeviceUtilizationRate(devID, aiCore)
	if err != nil {
		acc.AddError(fmt.Errorf("get ai core rate of npu failed: %v", err))
	}
	fields["npu_chip_info_utilization"] = float64(aiCoreUtil)

	hbmUtil, err := npu.devManager.GetDeviceUtilizationRate(devID, hbm)
	if err != nil {
		acc.AddError(fmt.Errorf("get hbm rate of npu failed: %v", err))
	}
	fields["npu_chip_info_hbm_utilization"] = float64(hbmUtil)

	hbmInfo, err := npu.devManager.GetDeviceHbmInfo(devID)
	if err != nil {
		acc.AddError(fmt.Errorf("get hbm info of npu failed: %v", err))
	} else {
		fields["npu_chip_info_hbm_used_memory"] = hbmInfo.Usage * mega
	}

	power, err := npu.devManager.GetDevicePowerInfo(devID)
	if err != nil {
		acc.AddError(fmt.Errorf("get hbm rate of npu failed: %v", err))
	}
	fields["npu_chip_info_power"] = power

	codeNum, errCodes, err := npu.devManager.GetDeviceAllErrorCode(devID)
	if err != nil {
		acc.AddError(fmt.Errorf("get err code failed: %v", err))
	}
	// conversion of "codeNum" here is safe because codeNum <= 128
	for i := 0; i < int(codeNum); i++ {
		errCodeKey := "npu_chip_info_error_code_" + strconv.Itoa(i)
		fields[errCodeKey] = errCodes[i]
	}
}

func (npu *NpuWatch) packHccnInfo(devID int32, fields map[string]interface{}, acc telegraf.Accumulator) error {
	phyID, err := npu.devManager.GetPhysicIDFromLogicID(devID)
	if err != nil {
		acc.AddError(fmt.Errorf("get devID of npu failed: %v", err))
		return err
	}

	linkStatus := hccn.GetNPULinkStatus(phyID)
	fields["npu_chip_info_link_status"] = hccn.GetLinkStatusCode(linkStatus)

	tx, rx, err := hccn.GetNPUInterfaceTraffic(phyID)
	if err != nil {
		acc.AddError(fmt.Errorf("get bandwidth of npu failed: %v", err))
	} else {
		fields["npu_chip_info_bandwidth_rx"] = rx * mega
		fields["npu_chip_info_bandwidth_tx"] = tx * mega
	}

	speed := hccn.GetNPULinkSpeed(phyID)
	fields["npu_chip_link_speed"] = speed * mega

	linkUpCnt := hccn.GetNPULinkUpNum(phyID)
	fields["npu_chip_link_up_num"] = linkUpCnt

	statInfo, err := hccn.GetNPUStatInfo(phyID)
	if err != nil {
		acc.AddError(fmt.Errorf("get stat info of npu failed: %v", err))
	} else {
		fields["npu_chip_mac_rx_pause_num"] = statInfo["mac_rx_pause_num"]
		fields["npu_chip_mac_tx_pause_num"] = statInfo["mac_tx_pause_num"]
		fields["npu_chip_mac_rx_pfc_pkt_num"] = statInfo["mac_rx_pfc_pkt_num"]
		fields["npu_chip_mac_tx_pfc_pkt_num"] = statInfo["mac_tx_pfc_pkt_num"]
		fields["npu_chip_mac_rx_bad_pkt_num"] = statInfo["mac_rx_bad_pkt_num"]
		fields["npu_chip_mac_tx_bad_pkt_num"] = statInfo["mac_tx_bad_pkt_num"]
		fields["npu_chip_roce_rx_all_pkt_num"] = statInfo["roce_rx_all_pkt_num"]
		fields["npu_chip_roce_tx_all_pkt_num"] = statInfo["roce_tx_all_pkt_num"]

		fields["npu_chip_roce_rx_err_pkt_num"] = statInfo["roce_rx_err_pkt_num"]
		fields["npu_chip_roce_tx_err_pkt_num"] = statInfo["roce_tx_err_pkt_num"]

		fields["npu_chip_roce_rx_cnp_pkt_num"] = statInfo["roce_rx_cnp_pkt_num"]
		fields["npu_chip_roce_tx_cnp_pkt_num"] = statInfo["roce_tx_cnp_pkt_num"]

		fields["npu_chip_mac_tx_bad_oct_num"] = statInfo["mac_tx_bad_oct_num"]
		fields["npu_chip_mac_rx_bad_oct_num"] = statInfo["mac_rx_bad_oct_num"]

		fields["npu_chip_roce_unexpected_ack_num"] = statInfo["roce_unexpected_ack_num"]
		fields["npu_chip_roce_out_of_order_num"] = statInfo["roce_out_of_order_num"]
		fields["npu_chip_roce_verification_err_num"] = statInfo["roce_verification_err_num"]
		fields["npu_chip_roce_qp_status_err_num"] = statInfo["roce_qp_status_err_num"]
		fields["npu_chip_roce_new_pkt_rty_num"] = statInfo["roce_new_pkt_rty_num"]
	}

	opticalInfo, err := hccn.GetNPUOpticalInfo(phyID)
	if err != nil {
		acc.AddError(fmt.Errorf("get optical info of npu failed: %v", err))
	}
	ctYunOpticalInfo := parseOptInfoForCTYun(opticalInfo)
	for k, v := range ctYunOpticalInfo {
		fields[k] = v
	}
	return nil
}

func (npu *NpuWatch) Gather(acc telegraf.Accumulator) error {
	if npu.devManager == nil {
		return errors.New("empty dev object")
	}
	devNum, devList, err := npu.devManager.GetDeviceList()
	if err != nil {
		acc.AddError(fmt.Errorf("get npu list failed: %s", err))
		return err
	}

	const devName = "ascend"
	devTag := make(map[string]string)
	devTagValue := "unsupported"
	if cardType := npu.devManager.GetDevType(); cardType == common.Ascend910B || cardType == common.Ascend910 {
		devTagValue = common.Chip910
	}

	for i := int32(0); i < devNum; i++ {
		fields := make(map[string]interface{})

		npu.packDcmiInfo(devList[i], fields, acc)
		if err := npu.packHccnInfo(devList[i], fields, acc); err != nil {
			return err
		}

		devTag["device"] = devTagValue + "-" + strconv.Itoa(int(devList[i]))
		acc.AddFields(devName, fields, devTag)
	}

	return nil
}

func init() {
	inputs.Add("npu", func() telegraf.Input { return &NpuWatch{} })
}
